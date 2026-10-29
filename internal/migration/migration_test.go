package migration

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/pressly/goose/v3"
)

type fakeProvider struct {
	up              []*goose.MigrationResult
	upErr           error
	status          []*goose.MigrationStatus
	statusErr       error
	current, target int64
	versionsErr     error
}

func (f *fakeProvider) Up(context.Context) ([]*goose.MigrationResult, error) { return f.up, f.upErr }
func (f *fakeProvider) Status(context.Context) ([]*goose.MigrationStatus, error) {
	return f.status, f.statusErr
}
func (f *fakeProvider) GetVersions(context.Context) (int64, int64, error) {
	return f.current, f.target, f.versionsErr
}
func (f *fakeProvider) ListSources() []*goose.Source {
	return []*goose.Source{{Version: 1, Path: "00001_initial.sql"}}
}

func testRunner(t *testing.T) (*Runner, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	runner, err := newRunner(db, false)
	if err != nil {
		t.Fatal(err)
	}
	return runner, mock
}

func TestUpStatusAndVersions(t *testing.T) {
	ctx := context.Background()
	t.Run("up", func(t *testing.T) {
		runner, mock := testRunner(t)
		fake := &fakeProvider{up: []*goose.MigrationResult{{}}}
		runner.provider = fake
		mock.ExpectQuery("SELECT to_regclass").WithArgs("public.schema_migrations").WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))
		if count, err := runner.Up(ctx); err != nil || count != 1 {
			t.Fatalf("up: %d %v", count, err)
		}
		fake.upErr = errors.New("ddl")
		mock.ExpectQuery("SELECT to_regclass").WithArgs("public.schema_migrations").WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))
		if _, err := runner.Up(ctx); err == nil {
			t.Fatal("expected up failure")
		}
	})
	t.Run("status-version", func(t *testing.T) {
		runner, mock := testRunner(t)
		fake := &fakeProvider{status: []*goose.MigrationStatus{{Source: &goose.Source{Version: 1, Path: "00001_initial.sql"}, State: goose.StateApplied, AppliedAt: time.Now()}}, current: 1, target: 1}
		runner.provider = fake
		expectTable(mock, true)
		items, err := runner.Status(ctx)
		if err != nil || len(items) != 1 || items[0].State != "applied" {
			t.Fatalf("status: %+v %v", items, err)
		}
		expectTable(mock, true)
		current, target, err := runner.Versions(ctx)
		if err != nil || current != 1 || target != 1 {
			t.Fatalf("versions: %d %d %v", current, target, err)
		}
	})
}

func expectTable(mock sqlmock.Sqlmock, exists bool) {
	mock.ExpectQuery("SELECT to_regclass").WithArgs("public.schema_migrations").WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))
	mock.ExpectQuery("SELECT to_regclass").WithArgs("public.goose_db_version").WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(exists))
}

func TestCheckCurrent(t *testing.T) {
	runner, mock := testRunner(t)
	expectTable(mock, true)
	mock.ExpectQuery("SELECT version_id").WillReturnRows(sqlmock.NewRows([]string{"version_id"}).AddRow(1))
	if err := runner.CheckCurrent(context.Background()); err != nil {
		t.Fatal(err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestCheckCurrentFailures(t *testing.T) {
	ctx := context.Background()
	t.Run("legacy", func(t *testing.T) {
		runner, mock := testRunner(t)
		mock.ExpectQuery("SELECT to_regclass").WithArgs("public.schema_migrations").WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))
		if err := runner.CheckCurrent(ctx); err == nil {
			t.Fatal("expected legacy schema error")
		}
	})
	t.Run("missing", func(t *testing.T) {
		runner, mock := testRunner(t)
		expectTable(mock, false)
		if err := runner.CheckCurrent(ctx); err == nil {
			t.Fatal("expected missing schema error")
		}
	})
	t.Run("pending", func(t *testing.T) {
		runner, mock := testRunner(t)
		expectTable(mock, true)
		mock.ExpectQuery("SELECT version_id").WillReturnRows(sqlmock.NewRows([]string{"version_id"}))
		if err := runner.CheckCurrent(ctx); err == nil {
			t.Fatal("expected pending schema error")
		}
	})
	t.Run("query", func(t *testing.T) {
		runner, mock := testRunner(t)
		mock.ExpectQuery("SELECT to_regclass").WithArgs("public.schema_migrations").WillReturnError(errors.New("database unavailable"))
		if err := runner.CheckCurrent(ctx); err == nil {
			t.Fatal("expected database error")
		}
	})
}
