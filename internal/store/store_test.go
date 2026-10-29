package store

import (
	"context"
	"errors"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// mockStore returns a GORM store backed by sqlmock.
func mockStore(t *testing.T) (*Store, sqlmock.Sqlmock) {
	t.Helper()
	sqlDB, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	if err != nil {
		t.Fatal(err)
	}
	db, err := gorm.Open(postgres.New(postgres.Config{Conn: sqlDB}), &gorm.Config{SkipDefaultTransaction: true, Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })
	return New(db), mock
}

// TestModelMetadata covers model table names and store construction.
func TestModelMetadata(t *testing.T) {
	s, _ := mockStore(t)
	if s == nil || (Ecosystem{}).TableName() != "ecosystems" || (Machine{}).TableName() != "machines" || (IngestionKey{}).TableName() != "ingestion_keys" {
		t.Fatal("metadata")
	}
}

// TestPingAndClose covers connection lifecycle helpers.
func TestPingAndClose(t *testing.T) {
	sqlDB, mock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
	if err != nil {
		t.Fatal(err)
	}
	db, err := gorm.Open(postgres.New(postgres.Config{Conn: sqlDB}), &gorm.Config{DisableAutomaticPing: true, Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		t.Fatal(err)
	}
	s := New(db)
	mock.ExpectPing()
	if err := s.Ping(context.Background()); err != nil {
		t.Fatal(err)
	}
	mock.ExpectClose()
	if err := s.Close(); err != nil {
		t.Fatal(err)
	}
}

// TestEcosystemAndMachineRepositories covers their query and validation paths.
func TestEcosystemAndMachineRepositories(t *testing.T) {
	s, mock := mockStore(t)
	ctx := context.Background()
	mock.ExpectQuery("SELECT e.id").WillReturnRows(sqlmock.NewRows([]string{"id", "name", "display_name", "is_active", "machine_count", "last_used_at", "created_at", "updated_at"}).AddRow("e1", "pay", "Pay", true, 1, nil, time.Now(), time.Now()))
	if rows, err := s.ListEcosystems(ctx); err != nil || len(rows) != 1 {
		t.Fatalf("list: %v %v", rows, err)
	}
	mock.ExpectQuery("INSERT INTO \"ecosystems\"").WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("e1"))
	if _, err := s.CreateEcosystem(ctx, "pay", "Pay"); err != nil {
		t.Fatal(err)
	}
	mock.ExpectExec("UPDATE \"ecosystems\"").WillReturnResult(sqlmock.NewResult(0, 0))
	if _, err := s.UpdateEcosystem(ctx, "missing", "x", true); !errors.Is(err, ErrNotFound) {
		t.Fatalf("update: %v", err)
	}
	mock.ExpectExec("UPDATE \"ecosystems\"").WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery("SELECT \\* FROM \"ecosystems\"").WillReturnRows(sqlmock.NewRows([]string{"id", "name", "display_name", "is_active", "created_at", "updated_at"}).AddRow("e1", "pay", "Payments", true, time.Now(), time.Now()))
	if v, err := s.UpdateEcosystem(ctx, "e1", "Payments", true); err != nil || v.ID != "e1" {
		t.Fatalf("successful update: %+v %v", v, err)
	}
	mock.ExpectExec("DELETE FROM \"ecosystems\"").WillReturnResult(sqlmock.NewResult(0, 0))
	if err := s.DeleteEcosystem(ctx, "missing"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("delete missing: %v", err)
	}
	mock.ExpectExec("DELETE FROM \"ecosystems\"").WillReturnResult(sqlmock.NewResult(0, 1))
	if err := s.DeleteEcosystem(ctx, "e1"); err != nil {
		t.Fatalf("delete: %v", err)
	}
	mock.ExpectQuery("SELECT m.id").WillReturnRows(sqlmock.NewRows([]string{"id", "hostname", "ecosystem_id", "ecosystem_name", "ipv4", "last_sync_at", "is_managed", "created_at", "updated_at"}).AddRow("m1", "one", "e1", "pay", "10.0.0.1", time.Now(), false, time.Now(), time.Now()))
	if rows, err := s.ListMachines(ctx, "one", "e1"); err != nil || len(rows) != 1 {
		t.Fatalf("machines: %v %v", rows, err)
	}
	if _, err := s.UpsertMachine(ctx, "one", "e1", "bad", false); err == nil {
		t.Fatal("invalid IP accepted")
	}
	mock.ExpectQuery("INSERT INTO machines").WillReturnRows(sqlmock.NewRows([]string{"id", "hostname", "ecosystem_id", "ecosystem_name", "ipv4", "last_sync_at", "is_managed", "created_at", "updated_at"}).AddRow("m1", "one", "e1", "pay", "10.0.0.1", time.Now(), false, time.Now(), time.Now()))
	if _, err := s.UpsertMachine(ctx, "one", "e1", "10.0.0.1", false); err != nil {
		t.Fatal(err)
	}
	mock.ExpectExec("DELETE FROM \"machines\"").WillReturnResult(sqlmock.NewResult(0, 0))
	if err := s.DeleteMachine(ctx, "missing"); !errors.Is(err, ErrNotFound) {
		t.Fatal(err)
	}
}

// TestFirewallRepositories covers effective lists, IP details, and missing resources.
func TestFirewallRepositories(t *testing.T) {
	s, mock := mockStore(t)
	ctx := context.Background()
	mock.ExpectQuery("SELECT id,name FROM \"ecosystems\"").WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).AddRow("e1", "pay"))
	mock.ExpectQuery("SELECT host\\(m.ipv4\\)").WillReturnRows(sqlmock.NewRows([]string{"ipv4", "source_type", "source_id", "hostname", "service", "last_used_at", "last_sync_at", "stale"}).AddRow("10.0.0.1", "machine", "m1", "one", "", nil, nil, false))
	if _, rows, err := s.EffectiveIPs(ctx, "pay"); err != nil || len(rows) != 1 {
		t.Fatalf("effective: %v", err)
	}
	mock.ExpectQuery("SELECT id,name FROM \"ecosystems\"").WillReturnRows(sqlmock.NewRows([]string{"id", "name"}))
	if _, _, err := s.EffectiveIPs(ctx, "missing"); !errors.Is(err, ErrNotFound) {
		t.Fatal(err)
	}
	if _, err := s.IPDetails(ctx, "bad"); err == nil {
		t.Fatal("invalid IP")
	}
	mock.ExpectQuery("SELECT host\\(m.ipv4\\)").WillReturnRows(sqlmock.NewRows([]string{"ipv4", "source_type", "source_id", "hostname", "service", "last_used_at", "last_sync_at", "stale"}))
	if _, err := s.IPDetails(ctx, "10.0.0.9"); !errors.Is(err, ErrNotFound) {
		t.Fatal(err)
	}
}

// TestIngestionKeys covers listing, creation, lookup, revocation, and touch.
func TestIngestionKeys(t *testing.T) {
	s, mock := mockStore(t)
	ctx := context.Background()
	now := time.Now()
	columns := []string{"id", "name", "prefix", "expires_at", "revoked_at", "last_used_at", "created_at"}
	mock.ExpectQuery("SELECT \\* FROM \"ingestion_keys\"").WillReturnRows(sqlmock.NewRows(columns).AddRow("k1", "agent", "pre", nil, nil, nil, now))
	if rows, err := s.ListIngestionKeys(ctx); err != nil || len(rows) != 1 {
		t.Fatal(err)
	}
	mock.ExpectExec("INSERT INTO \"ingestion_keys\"").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectQuery("SELECT \\* FROM \"ingestion_keys\"").WillReturnRows(sqlmock.NewRows(columns).AddRow("k1", "agent", "pre", nil, nil, nil, now))
	if _, err := s.CreateIngestionKey(ctx, "agent", "pre", "hash", nil); err != nil {
		t.Fatal(err)
	}
	mock.ExpectExec("UPDATE \"ingestion_keys\"").WillReturnResult(sqlmock.NewResult(0, 0))
	if err := s.RevokeIngestionKey(ctx, "none"); !errors.Is(err, ErrNotFound) {
		t.Fatal(err)
	}
	mock.ExpectQuery("SELECT id,secret_hash").WillReturnRows(sqlmock.NewRows([]string{"id", "secret_hash"}).AddRow("k1", "hash"))
	if id, _, err := s.IngestionKeyHash(ctx, "pre"); err != nil || id != "k1" {
		t.Fatal(err)
	}
	mock.ExpectQuery("SELECT id,secret_hash").WillReturnRows(sqlmock.NewRows([]string{"id", "secret_hash"}))
	if _, _, err := s.IngestionKeyHash(ctx, "missing"); !errors.Is(err, ErrNotFound) {
		t.Fatal(err)
	}
	mock.ExpectExec("UPDATE \"ingestion_keys\"").WillReturnResult(sqlmock.NewResult(0, 1))
	s.TouchIngestionKey(ctx, "k1")
}

// TestTrustedReadAndDelete covers trusted-service read helpers and deletion.
func TestTrustedReadAndDelete(t *testing.T) {
	s, mock := mockStore(t)
	ctx := context.Background()
	now := time.Now()
	mock.ExpectQuery("SELECT \\* FROM \"trusted_services\"").WillReturnRows(sqlmock.NewRows([]string{"id", "name", "description", "is_active", "created_at", "updated_at"}).AddRow("t1", "dns", "DNS", true, now, now))
	mock.ExpectQuery("SELECT id,hostname").WillReturnRows(sqlmock.NewRows([]string{"id", "hostname", "last_attempt_at", "last_success_at", "last_error", "status"}).AddRow("h1", "example.test", nil, nil, nil, "pending"))
	mock.ExpectQuery("SELECT id,host\\(ipv4\\)").WillReturnRows(sqlmock.NewRows([]string{"id", "ipv4", "first_seen_at", "last_seen_at", "last_used_at", "is_current"}).AddRow("r1", "10.0.0.1", now, now, nil, true))
	if rows, err := s.ListTrustedServices(ctx); err != nil || len(rows) != 1 || len(rows[0].Hostnames) != 1 {
		t.Fatalf("trusted: %v %v", rows, err)
	}
	mock.ExpectQuery("SELECT h.id").WillReturnRows(sqlmock.NewRows([]string{"id", "hostname", "last_attempt_at", "last_success_at", "last_error"}).AddRow("h1", "example.test", nil, nil, nil))
	if rows, err := s.ListActiveHostnames(ctx); err != nil || len(rows) != 1 {
		t.Fatal(err)
	}
	mock.ExpectExec("DELETE FROM \"trusted_services\"").WillReturnResult(sqlmock.NewResult(0, 0))
	if err := s.DeleteTrustedService(ctx, "none"); !errors.Is(err, ErrNotFound) {
		t.Fatal(err)
	}
	mock.ExpectExec("INSERT INTO audit_events").WillReturnResult(sqlmock.NewResult(1, 1))
	s.Audit(ctx, "a", "b", "c", "d")
}

// TestSchemaHelpers covers partition pruning.
func TestSchemaHelpers(t *testing.T) {
	s, mock := mockStore(t)
	ctx := context.Background()
	mock.ExpectQuery("SELECT inhrelid").WillReturnRows(sqlmock.NewRows([]string{"inhrelid"}).AddRow("firewall_requests_2020_01").AddRow("odd"))
	mock.ExpectExec(regexp.QuoteMeta("DROP TABLE IF EXISTS firewall_requests_2020_01")).WillReturnResult(sqlmock.NewResult(0, 0))
	if err := s.DropExpiredPartitions(ctx, time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC), 30); err != nil {
		t.Fatal(err)
	}
}

// TestPartitionsAndSchemaFailures covers partition DDL error paths.
func TestPartitionsAndSchemaFailures(t *testing.T) {
	s, mock := mockStore(t)
	ctx := context.Background()
	for i := 0; i < 4; i++ {
		mock.ExpectExec("CREATE (TABLE|INDEX)").WillReturnResult(sqlmock.NewResult(0, 0))
	}
	if err := s.EnsureHistoryPartitions(ctx, time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC), -12); err != nil {
		t.Fatal(err)
	}
}

// TestTrustedWrites covers create, update, resolution success, and resolution failure.
func TestTrustedWrites(t *testing.T) {
	s, mock := mockStore(t)
	ctx := context.Background()
	now := time.Now()
	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO \"trusted_services\"").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectQuery("SELECT \\* FROM \"trusted_services\"").WillReturnRows(sqlmock.NewRows([]string{"id", "name", "description", "is_active", "created_at", "updated_at"}).AddRow("t1", "dns", "DNS", true, now, now))
	mock.ExpectExec("INSERT INTO \"trusted_hostnames\"").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()
	mock.ExpectQuery("SELECT id,hostname").WillReturnRows(sqlmock.NewRows([]string{"id", "hostname", "last_attempt_at", "last_success_at", "last_error", "status"}))
	if _, err := s.CreateTrustedService(ctx, "dns", "DNS", true, []string{"EXAMPLE.TEST"}); err != nil {
		t.Fatal(err)
	}
	mock.ExpectBegin()
	mock.ExpectExec("UPDATE \"trusted_services\"").WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("INSERT INTO trusted_hostnames").WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("DELETE FROM \"trusted_hostnames\"").WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()
	mock.ExpectQuery("SELECT \\* FROM \"trusted_services\"").WillReturnRows(sqlmock.NewRows([]string{"id", "name", "description", "is_active", "created_at", "updated_at"}).AddRow("t1", "dns", "DNS", true, now, now))
	mock.ExpectQuery("SELECT id,hostname").WillReturnRows(sqlmock.NewRows([]string{"id", "hostname", "last_attempt_at", "last_success_at", "last_error", "status"}))
	if _, err := s.UpdateTrustedService(ctx, "t1", "dns", "DNS", true, []string{"example.test"}); err != nil {
		t.Fatal(err)
	}
	mock.ExpectBegin()
	mock.ExpectExec("UPDATE \"trusted_hostnames\"").WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()
	if err := s.SaveResolution(ctx, "h1", nil, errors.New("dns")); err != nil {
		t.Fatal(err)
	}
	mock.ExpectBegin()
	mock.ExpectExec("UPDATE \"trusted_hostnames\"").WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("UPDATE \"trusted_resolved_ips\"").WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("INSERT INTO trusted_resolved_ips").WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()
	if err := s.SaveResolution(ctx, "h1", []string{"10.0.0.1"}, nil); err != nil {
		t.Fatal(err)
	}
}

// TestDashboard covers all aggregate scans and rate calculations.
func TestDashboard(t *testing.T) {
	s, mock := mockStore(t)
	ctx := context.Background()
	mock.ExpectQuery("SELECT count\\(\\*\\) FILTER").WillReturnRows(sqlmock.NewRows([]string{"total", "previous", "success", "errors", "not_modified", "average", "p95"}).AddRow(10, 5, 8, 2, 1, 12.5, 20.0))
	mock.ExpectQuery("WITH current_ips").WillReturnRows(sqlmock.NewRows([]string{"total", "active", "passive", "machine_active", "machine_passive", "trusted_active", "trusted_passive"}).AddRow(4, 3, 1, 2, 1, 1, 0))
	mock.ExpectQuery("WITH buckets").WillReturnRows(sqlmock.NewRows([]string{"at", "requests", "errors"}).AddRow(time.Now(), 2, 1))
	mock.ExpectQuery("SELECT coalesce").WillReturnRows(sqlmock.NewRows([]string{"name", "requests", "unique_ips"}).AddRow("pay", 2, 1))
	mock.ExpectQuery("SELECT format").WillReturnRows(sqlmock.NewRows([]string{"name", "count"}).AddRow("json", 2))
	mock.ExpectQuery("SELECT source_type").WillReturnRows(sqlmock.NewRows([]string{"name", "count"}).AddRow("machine", 1))
	d, err := s.Dashboard(ctx, "7d")
	if err != nil || d.Requests.SuccessRate != 80 || d.Activity.ActiveRate != 75 || len(d.Series) != 1 {
		t.Fatalf("dashboard: %+v %v", d, err)
	}
}

// TestRecordRequest covers transactional history insertion and source updates.
func TestRecordRequest(t *testing.T) {
	s, mock := mockStore(t)
	ctx := context.Background()
	now := time.Now()
	mock.ExpectBegin()
	mock.ExpectQuery("INSERT INTO firewall_requests").WillReturnRows(sqlmock.NewRows([]string{"id", "occurred_at"}).AddRow("r1", now))
	mock.ExpectExec("INSERT INTO firewall_request_ips").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("UPDATE \"machines\"").WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("INSERT INTO firewall_request_ips").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("UPDATE \"trusted_resolved_ips\"").WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()
	input := HistoryInput{RouteKind: "ecosystem", EcosystemName: "pay", Format: "json", StatusCode: 200, Entries: []EffectiveIP{{IPv4: "10.0.0.1", SourceType: "machine", SourceID: "m1"}, {IPv4: "10.0.0.2", SourceType: "trusted", SourceID: "t1"}}}
	if err := s.RecordRequest(ctx, input); err != nil {
		t.Fatal(err)
	}
}
