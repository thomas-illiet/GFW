// Package migration owns the embedded PostgreSQL schema and its Goose runner.
package migration

import (
	"context"
	"database/sql"
	"embed"
	"errors"
	"fmt"
	"io/fs"
	"strings"
	"time"

	"github.com/pressly/goose/v3"
	"github.com/pressly/goose/v3/lock"
)

const versionTable = "goose_db_version"

// files is the single production source of schema migrations.
//
//go:embed sql/*.sql
var files embed.FS

// Runner applies and inspects the embedded, forward-only schema history.
type Runner struct {
	db       *sql.DB
	provider provider
	sources  []*goose.Source
}

type provider interface {
	Up(context.Context) ([]*goose.MigrationResult, error)
	Status(context.Context) ([]*goose.MigrationStatus, error)
	GetVersions(context.Context) (int64, int64, error)
	ListSources() []*goose.Source
}

// Status describes one known migration.
type Status struct {
	Version   int64
	Path      string
	State     string
	AppliedAt time.Time
}

// New creates a PostgreSQL migration runner with a cross-process session lock.
func New(db *sql.DB) (*Runner, error) { return newRunner(db, true) }

// newRunner allows unit tests to avoid PostgreSQL-specific advisory locking.
func newRunner(db *sql.DB, withLock bool) (*Runner, error) {
	if db == nil {
		return nil, errors.New("migration database is required")
	}
	migrationFS, err := fs.Sub(files, "sql")
	if err != nil {
		return nil, fmt.Errorf("open embedded migrations: %w", err)
	}
	options := []goose.ProviderOption{goose.WithDisableGlobalRegistry(true)}
	if withLock {
		locker, err := lock.NewPostgresSessionLocker(lock.WithLockTimeout(1, 300), lock.WithUnlockTimeout(1, 60))
		if err != nil {
			return nil, fmt.Errorf("configure migration lock: %w", err)
		}
		options = append(options, goose.WithSessionLocker(locker))
	}
	provider, err := goose.NewProvider(goose.DialectPostgres, db, migrationFS, options...)
	if err != nil {
		return nil, fmt.Errorf("configure migrations: %w", err)
	}
	return &Runner{db: db, provider: provider, sources: provider.ListSources()}, nil
}

// Up applies every pending migration under a PostgreSQL advisory lock.
func (r *Runner) Up(ctx context.Context) (int, error) {
	if err := r.rejectLegacy(ctx); err != nil {
		return 0, err
	}
	results, err := r.provider.Up(ctx)
	if err != nil {
		return 0, fmt.Errorf("apply migrations: %w", err)
	}
	return len(results), nil
}

// CheckCurrent verifies the exact migration history without changing the database.
func (r *Runner) CheckCurrent(ctx context.Context) error {
	if err := r.rejectLegacy(ctx); err != nil {
		return err
	}
	exists, err := r.tableExists(ctx, versionTable)
	if err != nil {
		return err
	}
	if !exists {
		return errors.New("database schema is not initialized; run `gfw migrate up`")
	}
	rows, err := r.db.QueryContext(ctx, `SELECT version_id FROM goose_db_version WHERE is_applied AND version_id > 0 ORDER BY version_id`)
	if err != nil {
		return fmt.Errorf("read migration history: %w", err)
	}
	defer rows.Close()
	var applied []int64
	for rows.Next() {
		var version int64
		if err := rows.Scan(&version); err != nil {
			return fmt.Errorf("read migration version: %w", err)
		}
		applied = append(applied, version)
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("read migration history: %w", err)
	}
	if len(applied) != len(r.sources) {
		return errors.New("database schema is not current; run `gfw migrate up`")
	}
	for i, source := range r.sources {
		if applied[i] != source.Version {
			return errors.New("database migration history is inconsistent; recreate pre-v0.1.0 databases or restore the missing migration")
		}
	}
	return nil
}

// Status returns the applied or pending state of every embedded migration.
func (r *Runner) Status(ctx context.Context) ([]Status, error) {
	if err := r.requireVersionTable(ctx); err != nil {
		return nil, err
	}
	items, err := r.provider.Status(ctx)
	if err != nil {
		return nil, fmt.Errorf("migration status: %w", err)
	}
	result := make([]Status, 0, len(items))
	for _, item := range items {
		result = append(result, Status{Version: item.Source.Version, Path: item.Source.Path, State: string(item.State), AppliedAt: item.AppliedAt})
	}
	return result, nil
}

// Versions returns the database and embedded target versions.
func (r *Runner) Versions(ctx context.Context) (int64, int64, error) {
	if err := r.requireVersionTable(ctx); err != nil {
		return 0, 0, err
	}
	current, target, err := r.provider.GetVersions(ctx)
	if err != nil {
		return 0, 0, fmt.Errorf("migration versions: %w", err)
	}
	return current, target, nil
}

// requireVersionTable keeps read-only inspection commands from initializing state.
func (r *Runner) requireVersionTable(ctx context.Context) error {
	if err := r.rejectLegacy(ctx); err != nil {
		return err
	}
	exists, err := r.tableExists(ctx, versionTable)
	if err != nil {
		return err
	}
	if !exists {
		return errors.New("database schema is not initialized; run `gfw migrate up`")
	}
	return nil
}

// rejectLegacy prevents Goose from being run over an unreleased, incompatible schema.
func (r *Runner) rejectLegacy(ctx context.Context) error {
	legacy, err := r.tableExists(ctx, "schema_migrations")
	if err != nil {
		return err
	}
	if legacy {
		return errors.New("legacy pre-v0.1.0 schema detected; recreate this database before running `gfw migrate up`")
	}
	return nil
}

// tableExists checks a simple, internal table name without mutating the schema.
func (r *Runner) tableExists(ctx context.Context, name string) (bool, error) {
	if strings.ContainsAny(name, `"'.;`) {
		return false, errors.New("invalid migration table name")
	}
	var exists bool
	if err := r.db.QueryRowContext(ctx, `SELECT to_regclass($1) IS NOT NULL`, "public."+name).Scan(&exists); err != nil {
		return false, fmt.Errorf("inspect database schema: %w", err)
	}
	return exists, nil
}
