// Package command defines the testable Cobra command tree.
package command

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/thomasilliet/gfw/internal/app"
	"github.com/thomasilliet/gfw/internal/config"
	"github.com/thomasilliet/gfw/internal/migration"
	"github.com/thomasilliet/gfw/internal/store"
)

// BuildInfo describes the compiled artifact.
type BuildInfo struct{ Version, Commit, Date string }

// dependencies contains replaceable command side effects.
type dependencies struct {
	open     func(context.Context, string) (*store.Store, error)
	serve    func(context.Context, config.Config, *store.Store) error
	migrator func(*store.Store) (migrationService, error)
}

// migrationService is the command-facing, testable migration contract.
type migrationService interface {
	Up(context.Context) (int, error)
	CheckCurrent(context.Context) error
	Status(context.Context) ([]migration.Status, error)
	Versions(context.Context) (int64, int64, error)
}

// New creates the complete Cobra command tree.
func New(build BuildInfo) *cobra.Command {
	return newRoot(build, dependencies{open: store.Open, serve: runServer, migrator: newMigrator})
}

// newRoot assembles commands with injectable side effects.
func newRoot(build BuildInfo, deps dependencies) *cobra.Command {
	loader := config.NewLoader()
	var configPath string
	root := &cobra.Command{Use: "gfw", Short: "Great Firewall control plane", SilenceUsage: true}
	addFlags(root.PersistentFlags(), &configPath)
	load := func(cmd *cobra.Command) (config.Config, error) {
		if err := loader.BindFlags(cmd.Flags()); err != nil {
			return config.Config{}, err
		}
		return loader.Load(configPath)
	}
	serveCmd := &cobra.Command{Use: "serve", Short: "start the HTTP server"}
	serveCmd.RunE = func(cmd *cobra.Command, _ []string) error {
		cfg, err := load(cmd)
		if err != nil {
			return err
		}
		db, err := deps.open(cmd.Context(), cfg.DatabaseURL)
		if err != nil {
			return err
		}
		defer db.Close()
		migrations, err := deps.migrator(db)
		if err != nil {
			return err
		}
		if err := migrations.CheckCurrent(cmd.Context()); err != nil {
			return err
		}
		return deps.serve(cmd.Context(), cfg, db)
	}
	withMigrations := func(cmd *cobra.Command, run func(migrationService, *store.Store) error) error {
		cfg, err := load(cmd)
		if err != nil {
			return err
		}
		db, err := deps.open(cmd.Context(), cfg.DatabaseURL)
		if err != nil {
			return err
		}
		defer db.Close()
		migrations, err := deps.migrator(db)
		if err != nil {
			return err
		}
		return run(migrations, db)
	}
	migrateCmd := &cobra.Command{Use: "migrate", Short: "manage forward-only database migrations"}
	up := func(cmd *cobra.Command, _ []string) error {
		return withMigrations(cmd, func(migrations migrationService, db *store.Store) error {
			count, err := migrations.Up(cmd.Context())
			if err != nil {
				return err
			}
			if err := db.EnsureHistoryPartitions(cmd.Context(), time.Now().UTC(), 14); err != nil {
				return fmt.Errorf("create history partitions: %w", err)
			}
			_, err = fmt.Fprintf(cmd.OutOrStdout(), "applied %d migration(s)\n", count)
			return err
		})
	}
	migrateCmd.RunE = up
	upCmd := &cobra.Command{Use: "up", Short: "apply all pending migrations", RunE: up}
	statusCmd := &cobra.Command{Use: "status", Short: "print migration status"}
	statusCmd.RunE = func(cmd *cobra.Command, _ []string) error {
		return withMigrations(cmd, func(migrations migrationService, _ *store.Store) error {
			items, err := migrations.Status(cmd.Context())
			if err != nil {
				return err
			}
			for _, item := range items {
				if _, err := fmt.Fprintf(cmd.OutOrStdout(), "%05d\t%s\t%s\n", item.Version, item.State, item.Path); err != nil {
					return err
				}
			}
			return nil
		})
	}
	migrationVersionCmd := &cobra.Command{Use: "version", Short: "print current and target schema versions"}
	migrationVersionCmd.RunE = func(cmd *cobra.Command, _ []string) error {
		return withMigrations(cmd, func(migrations migrationService, _ *store.Store) error {
			current, target, err := migrations.Versions(cmd.Context())
			if err != nil {
				return err
			}
			_, err = fmt.Fprintf(cmd.OutOrStdout(), "current %d, target %d\n", current, target)
			return err
		})
	}
	migrateCmd.AddCommand(upCmd, statusCmd, migrationVersionCmd)
	versionCmd := &cobra.Command{Use: "version", Short: "print build metadata"}
	versionCmd.RunE = func(cmd *cobra.Command, _ []string) error {
		_, err := fmt.Fprintf(cmd.OutOrStdout(), "%s (commit %s, built %s)\n", build.Version, build.Commit, build.Date)
		return err
	}
	root.AddCommand(serveCmd, migrateCmd, versionCmd)
	root.RunE = serveCmd.RunE
	return root
}

// newMigrator connects the shared SQL pool to the embedded Goose migration set.
func newMigrator(db *store.Store) (migrationService, error) {
	sqlDB, err := db.SQLDB()
	if err != nil {
		return nil, err
	}
	return migration.New(sqlDB)
}

// addFlags declares the stable CLI configuration interface.
func addFlags(flags *pflag.FlagSet, configPath *string) {
	flags.StringVar(configPath, "config", "", "YAML configuration file")
	flags.String("environment", "", "deployment environment (development, staging, or production)")
	flags.String("listen-address", "", "HTTP listen address")
	flags.String("metrics-listen-address", "", "Prometheus metrics listen address")
	flags.String("database-url", "", "PostgreSQL connection URL")
	flags.String("public-url", "", "browser-visible base URL")
	flags.String("frontend-dir", "", "generated frontend directory")
	flags.String("oidc-issuer", "", "OIDC issuer URL")
	flags.String("oidc-client-id", "", "OIDC client ID")
	flags.String("oidc-client-secret", "", "OIDC client secret")
	flags.String("oidc-redirect-url", "", "OIDC callback URL")
	flags.String("session-secret", "", "session signing secret")
	flags.String("admin-token", "", "administrator API bearer token")
	flags.Bool("dev-auth", false, "enable development authentication")
	flags.Duration("dns-interval", 0, "DNS refresh interval")
	flags.Duration("dns-timeout", 0, "DNS lookup timeout")
	flags.Int("retention-days", 0, "history retention in days")
}

// runServer starts the application and performs graceful shutdown.
func runServer(parent context.Context, cfg config.Config, db *store.Store) error {
	ctx, stop := signal.NotifyContext(parent, syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	application, err := app.New(ctx, cfg, db)
	if err != nil {
		return err
	}
	defer application.Close()
	server := &http.Server{Addr: cfg.ListenAddress, Handler: application.Handler(), ReadHeaderTimeout: 10 * time.Second, ReadTimeout: 30 * time.Second, WriteTimeout: 60 * time.Second, IdleTimeout: 90 * time.Second}
	metricsServer := &http.Server{Addr: cfg.MetricsAddress, Handler: application.MetricsHandler(), ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 10 * time.Second, WriteTimeout: 10 * time.Second, IdleTimeout: 30 * time.Second}
	errCh := make(chan error, 2)
	go func() { errCh <- server.ListenAndServe() }()
	go func() { errCh <- metricsServer.ListenAndServe() }()
	shutdown := func() error {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()
		apiErr := server.Shutdown(shutdownCtx)
		metricsErr := metricsServer.Shutdown(shutdownCtx)
		return errors.Join(apiErr, metricsErr)
	}
	select {
	case <-ctx.Done():
		return shutdown()
	case err := <-errCh:
		shutdownErr := shutdown()
		if errors.Is(err, http.ErrServerClosed) {
			return shutdownErr
		}
		return errors.Join(err, shutdownErr)
	}
}
