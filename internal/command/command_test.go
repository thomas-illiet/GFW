package command

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/thomasilliet/gfw/internal/config"
	"github.com/thomasilliet/gfw/internal/migration"
	"github.com/thomasilliet/gfw/internal/store"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type fakeMigrations struct {
	checkErr, upErr error
	upCount         int
}

func (f *fakeMigrations) Up(context.Context) (int, error)    { return f.upCount, f.upErr }
func (f *fakeMigrations) CheckCurrent(context.Context) error { return f.checkErr }
func (f *fakeMigrations) Status(context.Context) ([]migration.Status, error) {
	return []migration.Status{{Version: 1, Path: "00001_initial.sql", State: "applied"}}, nil
}
func (f *fakeMigrations) Versions(context.Context) (int64, int64, error) { return 1, 1, nil }

// commandStore builds a store suitable for command tests.
func commandStore(t *testing.T) (*store.Store, sqlmock.Sqlmock) {
	t.Helper()
	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	db, err := gorm.Open(postgres.New(postgres.Config{Conn: sqlDB}), &gorm.Config{SkipDefaultTransaction: true, Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })
	return store.New(db), mock
}

// TestVersionAndConfigurationFailure covers output and preflight validation.
func TestVersionAndConfigurationFailure(t *testing.T) {
	cmd := New(BuildInfo{Version: "1.2.3", Commit: "abc", Date: "today"})
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetArgs([]string{"version"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "1.2.3") {
		t.Fatal(out.String())
	}
	cmd = New(BuildInfo{})
	cmd.SetArgs([]string{"serve", "--database-url", " "})
	if err := cmd.Execute(); err == nil {
		t.Fatal("expected configuration error")
	}
}

// TestServeCommand covers schema preflight and injected server execution.
func TestServeCommand(t *testing.T) {
	s, _ := commandStore(t)
	served := false
	deps := dependencies{open: func(context.Context, string) (*store.Store, error) { return s, nil }, serve: func(_ context.Context, cfg config.Config, _ *store.Store) error { served = cfg.DevAuth; return nil }, migrator: func(*store.Store) (migrationService, error) { return &fakeMigrations{}, nil }}
	cmd := newRoot(BuildInfo{}, deps)
	cmd.SetArgs([]string{"serve", "--dev-auth"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
	if !served {
		t.Fatal("server was not called")
	}
}

// TestCommandOpenFailures covers serve and migrate connection errors.
func TestCommandOpenFailures(t *testing.T) {
	deps := dependencies{open: func(context.Context, string) (*store.Store, error) { return nil, errors.New("open") }, migrator: func(*store.Store) (migrationService, error) { return &fakeMigrations{}, nil }}
	for _, name := range []string{"serve", "migrate"} {
		cmd := newRoot(BuildInfo{}, deps)
		cmd.SetArgs([]string{name, "--dev-auth"})
		if err := cmd.Execute(); err == nil {
			t.Fatalf("%s accepted open failure", name)
		}
	}
}

// TestMigrationInspectionCommands covers the public status and version interfaces.
func TestMigrationInspectionCommands(t *testing.T) {
	for _, tc := range []struct {
		args []string
		want string
	}{
		{[]string{"migrate", "status", "--dev-auth"}, "00001\tapplied\t00001_initial.sql"},
		{[]string{"migrate", "version", "--dev-auth"}, "current 1, target 1"},
	} {
		s, _ := commandStore(t)
		deps := dependencies{
			open:     func(context.Context, string) (*store.Store, error) { return s, nil },
			migrator: func(*store.Store) (migrationService, error) { return &fakeMigrations{}, nil },
		}
		cmd := newRoot(BuildInfo{}, deps)
		var out bytes.Buffer
		cmd.SetOut(&out)
		cmd.SetArgs(tc.args)
		if err := cmd.Execute(); err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(out.String(), tc.want) {
			t.Fatalf("output %q does not contain %q", out.String(), tc.want)
		}
	}
}

// TestSchemaPreflightFailure verifies serve stops before starting HTTP.
func TestSchemaPreflightFailure(t *testing.T) {
	s, _ := commandStore(t)
	deps := dependencies{
		open:  func(context.Context, string) (*store.Store, error) { return s, nil },
		serve: func(context.Context, config.Config, *store.Store) error { t.Fatal("serve called"); return nil },
		migrator: func(*store.Store) (migrationService, error) {
			return &fakeMigrations{checkErr: errors.New("pending")}, nil
		},
	}
	cmd := newRoot(BuildInfo{}, deps)
	cmd.SetArgs([]string{"serve", "--dev-auth"})
	if err := cmd.Execute(); err == nil {
		t.Fatal("expected schema preflight failure")
	}
}
