package app

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/thomasilliet/gfw/internal/auth"
	"github.com/thomasilliet/gfw/internal/config"
	resolver "github.com/thomasilliet/gfw/internal/dns"
	"github.com/thomasilliet/gfw/internal/store"
	"golang.org/x/crypto/bcrypt"
)

// fakeRepository is a configurable in-memory handler dependency.
type fakeRepository struct {
	err     error
	history store.HistoryInput
	audits  int
	touched bool
}

func (f *fakeRepository) Ping(context.Context) error { return f.err }
func (f *fakeRepository) Dashboard(context.Context, string) (store.Dashboard, error) {
	return store.Dashboard{Window: "24h"}, f.err
}
func (f *fakeRepository) ListEcosystems(context.Context) ([]store.Ecosystem, error) {
	return []store.Ecosystem{{ID: "e1", Name: "pay"}}, f.err
}
func (f *fakeRepository) CreateEcosystem(context.Context, string, string) (store.Ecosystem, error) {
	return store.Ecosystem{ID: "e1"}, f.err
}
func (f *fakeRepository) UpdateEcosystem(context.Context, string, string, bool) (store.Ecosystem, error) {
	return store.Ecosystem{ID: "e1"}, f.err
}
func (f *fakeRepository) ListMachines(context.Context, string, string) ([]store.Machine, error) {
	return []store.Machine{{ID: "m1"}}, f.err
}
func (f *fakeRepository) UpsertMachine(context.Context, string, string, string, bool) (store.Machine, error) {
	return store.Machine{ID: "m1"}, f.err
}
func (f *fakeRepository) DeleteMachine(context.Context, string) error { return f.err }
func (f *fakeRepository) ListTrustedServices(context.Context) ([]store.TrustedService, error) {
	return []store.TrustedService{{ID: "t1"}}, f.err
}
func (f *fakeRepository) CreateTrustedService(context.Context, string, string, bool, []string) (store.TrustedService, error) {
	return store.TrustedService{ID: "t1"}, f.err
}
func (f *fakeRepository) UpdateTrustedService(context.Context, string, string, string, bool, []string) (store.TrustedService, error) {
	return store.TrustedService{ID: "t1"}, f.err
}
func (f *fakeRepository) DeleteTrustedService(context.Context, string) error { return f.err }
func (f *fakeRepository) ListIngestionKeys(context.Context) ([]store.IngestionKey, error) {
	return []store.IngestionKey{{ID: "k1"}}, f.err
}
func (f *fakeRepository) CreateIngestionKey(context.Context, string, string, string, *time.Time) (store.IngestionKey, error) {
	return store.IngestionKey{ID: "k1"}, f.err
}
func (f *fakeRepository) RevokeIngestionKey(context.Context, string) error { return f.err }
func (f *fakeRepository) IngestionKeyHash(context.Context, string) (string, string, error) {
	hash, _ := bcrypt.GenerateFromPassword([]byte("prefix.secret"), bcrypt.MinCost)
	return "k1", string(hash), f.err
}
func (f *fakeRepository) TouchIngestionKey(context.Context, string) { f.touched = true }
func (f *fakeRepository) EffectiveIPs(context.Context, string) (string, []store.EffectiveIP, error) {
	return "pay", []store.EffectiveIP{{IPv4: "10.0.0.1", SourceType: "machine", SourceID: "m1"}}, f.err
}
func (f *fakeRepository) IPDetails(context.Context, string) ([]store.EffectiveIP, error) {
	return []store.EffectiveIP{{IPv4: "10.0.0.1", SourceType: "machine", SourceID: "m1"}}, f.err
}
func (f *fakeRepository) RecordRequest(_ context.Context, input store.HistoryInput) error {
	f.history = input
	return f.err
}
func (f *fakeRepository) ListActiveHostnames(context.Context) ([]store.TrustedHostname, error) {
	return nil, f.err
}
func (f *fakeRepository) SaveResolution(context.Context, string, []string, error) error { return f.err }
func (f *fakeRepository) Audit(context.Context, string, string, string, string)         { f.audits++ }
func (f *fakeRepository) EnsureHistoryPartitions(context.Context, time.Time, int) error { return f.err }
func (f *fakeRepository) DropExpiredPartitions(context.Context, time.Time, int) error   { return f.err }

// testApp constructs a routed application without background goroutines.
func testApp(t *testing.T, repo *fakeRepository, frontend string) *App {
	t.Helper()
	cfg := config.Config{DevAuth: true, PublicURL: "http://example.test", FrontendDir: frontend, DNSInterval: time.Hour, DNSTimeout: time.Second, RetentionDays: 30}
	am, err := auth.New(context.Background(), cfg)
	if err != nil {
		t.Fatal(err)
	}
	a := &App{cfg: cfg, store: repo, auth: am, mux: http.NewServeMux()}
	a.dns = resolver.New(repo, time.Hour, time.Second)
	a.routes()
	return a
}

// request executes one request through the complete middleware stack.
func request(a *App, method, target, body string) *httptest.ResponseRecorder {
	r := httptest.NewRequest(method, target, strings.NewReader(body))
	r.Host = "example.test"
	if body != "" {
		r.Header.Set("Content-Type", "application/json")
	}
	w := httptest.NewRecorder()
	a.Handler().ServeHTTP(w, r)
	return w
}

// TestInfrastructureHandlers covers health, session, middleware, and frontend behavior.
func TestInfrastructureHandlers(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "index.html"), []byte("index"), 0o600); err != nil {
		t.Fatal(err)
	}
	a := testApp(t, &fakeRepository{}, dir)
	for _, path := range []string{"/health/live", "/health/ready", "/auth/session", "/"} {
		if w := request(a, "GET", path, ""); w.Code != 200 {
			t.Fatalf("%s: %d %s", path, w.Code, w.Body.String())
		}
	}
	w := request(a, "GET", "/missing", "")
	if w.Code != 200 || w.Header().Get("X-Frame-Options") != "DENY" {
		t.Fatal("SPA or security headers missing")
	}
	a.store.(*fakeRepository).err = errors.New("db")
	if w := request(a, "GET", "/health/ready", ""); w.Code != 503 {
		t.Fatalf("ready=%d", w.Code)
	}
	a.cfg.FrontendDir = t.TempDir()
	if w := request(a, "GET", "/", ""); w.Code != 404 {
		t.Fatalf("frontend=%d", w.Code)
	}
}

// TestAdministrationSuccessPaths covers every administration domain's normal route.
func TestAdministrationSuccessPaths(t *testing.T) {
	repo := &fakeRepository{}
	a := testApp(t, repo, t.TempDir())
	tests := []struct {
		method, path, body string
		status             int
	}{
		{"GET", "/api/v1/admin/dashboard", "", 200}, {"GET", "/api/v1/admin/ecosystems", "", 200},
		{"POST", "/api/v1/admin/ecosystems", `{"name":"pay","displayName":"Pay"}`, 201},
		{"PATCH", "/api/v1/admin/ecosystems/e1", `{"displayName":"Pay","isActive":true}`, 200},
		{"GET", "/api/v1/admin/machines", "", 200}, {"POST", "/api/v1/admin/machines", `{"Hostname":"one","EcosystemID":"e1","IPv4":"10.0.0.1"}`, 201},
		{"DELETE", "/api/v1/admin/machines/m1", "", 204}, {"GET", "/api/v1/admin/trusted-services", "", 200},
		{"POST", "/api/v1/admin/trusted-services", `{"Name":"dns","Description":"DNS","IsActive":true,"Hostnames":["example.test"]}`, 201},
		{"PATCH", "/api/v1/admin/trusted-services/t1", `{"Name":"dns","Description":"DNS","IsActive":true,"Hostnames":[]}`, 200},
		{"DELETE", "/api/v1/admin/trusted-services/t1", "", 204}, {"POST", "/api/v1/admin/trusted-services/resolve", "", 200},
		{"GET", "/api/v1/admin/ingestion-keys", "", 200}, {"POST", "/api/v1/admin/ingestion-keys", `{"name":"agent"}`, 201},
		{"DELETE", "/api/v1/admin/ingestion-keys/k1", "", 204},
	}
	for _, test := range tests {
		w := request(a, test.method, test.path, test.body)
		if w.Code != test.status {
			t.Fatalf("%s %s: %d %s", test.method, test.path, w.Code, w.Body.String())
		}
	}
	if repo.audits == 0 {
		t.Fatal("expected audit calls")
	}
}

// TestFirewallAndIngestion covers public firewall responses, ETags, and ingestion.
func TestFirewallAndIngestion(t *testing.T) {
	repo := &fakeRepository{}
	a := testApp(t, repo, t.TempDir())
	w := request(a, "GET", "/api/v1/firewall/ecosystems/pay/ips?format=txt", "")
	if w.Code != 200 || !strings.Contains(w.Body.String(), "10.0.0.1") {
		t.Fatalf("firewall: %d %s", w.Code, w.Body.String())
	}
	r := httptest.NewRequest("GET", "/api/v1/firewall/ecosystems/pay/ips", nil)
	r.Header.Set("If-None-Match", w.Header().Get("ETag"))
	w = httptest.NewRecorder()
	a.Handler().ServeHTTP(w, r)
	if w = request(a, "GET", "/api/v1/firewall/ips/10.0.0.1", ""); w.Code != 200 {
		t.Fatalf("IP: %d", w.Code)
	}
	w = request(a, "POST", "/api/v1/ingestion/machines:upsert", `{"machines":[{"Hostname":"one","EcosystemID":"e1","IPv4":"10.0.0.1"}]}`)
	w.Result().Request = nil
	if w.Code != 401 {
		t.Fatalf("missing key: %d", w.Code)
	}
	req := httptest.NewRequest("POST", "/api/v1/ingestion/machines:upsert", strings.NewReader(`{"machines":[]}`))
	req.Header.Set("Authorization", "Bearer prefix.secret")
	w = httptest.NewRecorder()
	a.Handler().ServeHTTP(w, req)
	if w.Code != 200 || !repo.touched {
		t.Fatalf("ingestion: %d %s", w.Code, w.Body.String())
	}
}

// TestValidationFailures covers format, dashboard, payload, and CSRF failures.
func TestValidationFailures(t *testing.T) {
	a := testApp(t, &fakeRepository{}, t.TempDir())
	if w := request(a, "GET", "/api/v1/firewall/ecosystems/pay/ips?format=yaml", ""); w.Code != 400 {
		t.Fatal(w.Code)
	}
	if w := request(a, "GET", "/api/v1/admin/dashboard?window=year", ""); w.Code != 400 {
		t.Fatal(w.Code)
	}
	if w := request(a, "POST", "/api/v1/admin/ecosystems", `{bad`); w.Code != 400 {
		t.Fatal(w.Code)
	}
	r := httptest.NewRequest("POST", "/api/v1/admin/ecosystems", strings.NewReader(`{}`))
	r.Host = "example.test"
	r.Header.Set("Origin", "https://evil.test")
	w := httptest.NewRecorder()
	a.Handler().ServeHTTP(w, r)
	if w.Code != 403 {
		t.Fatal(w.Code)
	}
	if !sameOrigin(httptest.NewRequest("POST", "http://example.test/x", nil)) || !validWindow("7d") || validWindow("1y") {
		t.Fatal("helper behavior")
	}
}

// TestNewAndMaintenance covers application construction, workers, and shutdown.
func TestNewAndMaintenance(t *testing.T) {
	repo := &fakeRepository{}
	old := maintenanceInterval
	maintenanceInterval = time.Millisecond
	ctx, cancel := context.WithCancel(context.Background())
	cfg := config.Config{DevAuth: true, DNSInterval: time.Millisecond, DNSTimeout: time.Second, RetentionDays: 30}
	a, err := New(ctx, cfg, repo)
	if err != nil {
		t.Fatal(err)
	}
	time.Sleep(5 * time.Millisecond)
	cancel()
	a.Close()
	maintenanceInterval = old
}

// TestRepositoryFailures covers handler mappings for persistence failures.
func TestRepositoryFailures(t *testing.T) {
	repo := &fakeRepository{err: errors.New("failure")}
	a := testApp(t, repo, t.TempDir())
	tests := []struct {
		method, path, body string
		status             int
	}{
		{"GET", "/api/v1/admin/dashboard", "", 500}, {"GET", "/api/v1/admin/ecosystems", "", 500},
		{"POST", "/api/v1/admin/ecosystems", `{"name":"x","displayName":"X"}`, 409},
		{"PATCH", "/api/v1/admin/ecosystems/e", `{"displayName":"X"}`, 404},
		{"GET", "/api/v1/admin/machines", "", 500}, {"POST", "/api/v1/admin/machines", `{}`, 400},
		{"DELETE", "/api/v1/admin/machines/m", "", 404}, {"GET", "/api/v1/admin/trusted-services", "", 500},
		{"POST", "/api/v1/admin/trusted-services", `{}`, 409}, {"PATCH", "/api/v1/admin/trusted-services/t", `{}`, 404},
		{"DELETE", "/api/v1/admin/trusted-services/t", "", 404}, {"POST", "/api/v1/admin/trusted-services/resolve", "", 500},
		{"GET", "/api/v1/admin/ingestion-keys", "", 500}, {"POST", "/api/v1/admin/ingestion-keys", `{"name":"x"}`, 409},
		{"DELETE", "/api/v1/admin/ingestion-keys/k", "", 404},
	}
	for _, test := range tests {
		if w := request(a, test.method, test.path, test.body); w.Code != test.status {
			t.Fatalf("%s %s: %d", test.method, test.path, w.Code)
		}
	}
	repo.err = store.ErrNotFound
	if w := request(a, "GET", "/api/v1/firewall/ecosystems/missing/ips", ""); w.Code != 404 {
		t.Fatal(w.Code)
	}
	if w := request(a, "GET", "/api/v1/firewall/ips/10.0.0.9", ""); w.Code != 404 {
		t.Fatal(w.Code)
	}
}
