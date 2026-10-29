package app

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/thomasilliet/gfw/internal/auth"
	"github.com/thomasilliet/gfw/internal/config"
	resolver "github.com/thomasilliet/gfw/internal/dns"
	"github.com/thomasilliet/gfw/internal/store"
)

// maintenanceInterval controls periodic partition maintenance.
var maintenanceInterval = 24 * time.Hour

// App owns HTTP routing and background workers.
type App struct {
	cfg             config.Config
	store           Repository
	auth            *auth.Manager
	dns             *resolver.Worker
	mux             *http.ServeMux
	cancel          context.CancelFunc
	maintenanceDone chan struct{}
}

// Repository defines the persistence operations used by HTTP handlers and maintenance.
type Repository interface {
	Ping(context.Context) error
	Dashboard(context.Context, string) (store.Dashboard, error)
	ListEcosystems(context.Context) ([]store.Ecosystem, error)
	CreateEcosystem(context.Context, string, string) (store.Ecosystem, error)
	UpdateEcosystem(context.Context, string, string, bool) (store.Ecosystem, error)
	ListMachines(context.Context, string, string) ([]store.Machine, error)
	UpsertMachine(context.Context, string, string, string, bool) (store.Machine, error)
	DeleteMachine(context.Context, string) error
	ListTrustedServices(context.Context) ([]store.TrustedService, error)
	CreateTrustedService(context.Context, string, string, bool, []string) (store.TrustedService, error)
	UpdateTrustedService(context.Context, string, string, string, bool, []string) (store.TrustedService, error)
	DeleteTrustedService(context.Context, string) error
	ListIngestionKeys(context.Context) ([]store.IngestionKey, error)
	CreateIngestionKey(context.Context, string, string, string, *time.Time) (store.IngestionKey, error)
	RevokeIngestionKey(context.Context, string) error
	IngestionKeyHash(context.Context, string) (string, string, error)
	TouchIngestionKey(context.Context, string)
	EffectiveIPs(context.Context, string) (string, []store.EffectiveIP, error)
	IPDetails(context.Context, string) ([]store.EffectiveIP, error)
	RecordRequest(context.Context, store.HistoryInput) error
	ListActiveHostnames(context.Context) ([]store.TrustedHostname, error)
	SaveResolution(context.Context, string, []string, error) error
	Audit(context.Context, string, string, string, string)
	EnsureHistoryPartitions(context.Context, time.Time, int) error
	DropExpiredPartitions(context.Context, time.Time, int) error
}

// New constructs an application and starts its background workers.
func New(ctx context.Context, cfg config.Config, s Repository) (*App, error) {
	am, err := auth.New(ctx, cfg)
	if err != nil {
		return nil, err
	}
	workerCtx, cancel := context.WithCancel(ctx)
	a := &App{cfg: cfg, store: s, auth: am, mux: http.NewServeMux(), cancel: cancel, maintenanceDone: make(chan struct{})}
	a.dns = resolver.New(s, cfg.DNSInterval, cfg.DNSTimeout)
	a.dns.Start(workerCtx)
	a.routes()
	go a.maintenance(workerCtx)
	return a, nil
}

// Close stops application workers.
func (a *App) Close() {
	if a.cancel != nil {
		a.cancel()
	}
	a.dns.Close()
	if a.maintenanceDone != nil {
		<-a.maintenanceDone
	}
}

// Handler returns the fully wrapped HTTP handler.
func (a *App) Handler() http.Handler { return requestLog(securityHeaders(a.auth.Middleware(a.mux))) }

// routes registers all public and administration routes.
func (a *App) routes() {
	a.mux.HandleFunc("GET /health/live", func(w http.ResponseWriter, _ *http.Request) { writeJSON(w, 200, map[string]string{"status": "ok"}) })
	a.mux.HandleFunc("GET /health/ready", a.ready)
	a.mux.HandleFunc("GET /auth/login", a.auth.Login)
	a.mux.HandleFunc("GET /auth/callback", a.auth.Callback)
	a.mux.HandleFunc("POST /auth/logout", a.auth.Logout)
	a.mux.Handle("GET /auth/session", a.auth.Require("viewer")(http.HandlerFunc(a.session)))
	a.mux.HandleFunc("GET /api/v1/firewall/ecosystems/{ecosystem}/ips", a.firewallEcosystem)
	a.mux.HandleFunc("GET /api/v1/firewall/ips/{ipv4}", a.firewallIP)
	a.mux.Handle("GET /api/v1/admin/dashboard", a.auth.Require("viewer")(http.HandlerFunc(a.dashboard)))
	a.mux.Handle("GET /api/v1/admin/ecosystems", a.auth.Require("viewer")(http.HandlerFunc(a.ecosystems)))
	a.mux.Handle("POST /api/v1/admin/ecosystems", a.auth.Require("admin")(http.HandlerFunc(a.ecosystems)))
	a.mux.Handle("PATCH /api/v1/admin/ecosystems/{id}", a.auth.Require("admin")(http.HandlerFunc(a.ecosystemUpdate)))
	a.mux.Handle("GET /api/v1/admin/machines", a.auth.Require("viewer")(http.HandlerFunc(a.machines)))
	a.mux.Handle("POST /api/v1/admin/machines", a.auth.Require("editor")(http.HandlerFunc(a.machines)))
	a.mux.Handle("DELETE /api/v1/admin/machines/{id}", a.auth.Require("editor")(http.HandlerFunc(a.machineDelete)))
	a.mux.Handle("GET /api/v1/admin/trusted-services", a.auth.Require("viewer")(http.HandlerFunc(a.trustedServices)))
	a.mux.Handle("POST /api/v1/admin/trusted-services", a.auth.Require("editor")(http.HandlerFunc(a.trustedServices)))
	a.mux.Handle("PATCH /api/v1/admin/trusted-services/{id}", a.auth.Require("editor")(http.HandlerFunc(a.trustedServiceUpdate)))
	a.mux.Handle("DELETE /api/v1/admin/trusted-services/{id}", a.auth.Require("editor")(http.HandlerFunc(a.trustedServiceDelete)))
	a.mux.Handle("POST /api/v1/admin/trusted-services/resolve", a.auth.Require("admin")(http.HandlerFunc(a.resolveTrusted)))
	a.mux.Handle("GET /api/v1/admin/ingestion-keys", a.auth.Require("admin")(http.HandlerFunc(a.ingestionKeys)))
	a.mux.Handle("POST /api/v1/admin/ingestion-keys", a.auth.Require("admin")(http.HandlerFunc(a.ingestionKeys)))
	a.mux.Handle("DELETE /api/v1/admin/ingestion-keys/{id}", a.auth.Require("admin")(http.HandlerFunc(a.revokeIngestionKey)))
	a.mux.HandleFunc("POST /api/v1/ingestion/machines:upsert", a.ingestMachines)
	a.mux.HandleFunc("GET /", a.frontend)
}

// maintenance periodically creates and prunes history partitions.
func (a *App) maintenance(ctx context.Context) {
	defer close(a.maintenanceDone)
	ticker := time.NewTicker(maintenanceInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case now := <-ticker.C:
			if err := a.store.EnsureHistoryPartitions(ctx, now, 2); err != nil {
				slog.Error("ensure partitions", "error", err)
			}
			if err := a.store.DropExpiredPartitions(ctx, now, a.cfg.RetentionDays); err != nil {
				slog.Error("drop partitions", "error", err)
			}
		}
	}
}

// ready reports database readiness.
func (a *App) ready(w http.ResponseWriter, r *http.Request) {
	if err := a.store.Ping(r.Context()); err != nil {
		problem(w, 503, "database_unavailable", err.Error())
		return
	}
	writeJSON(w, 200, map[string]string{"status": "ready"})
}

// session returns the authenticated user.
func (a *App) session(w http.ResponseWriter, r *http.Request) {
	u, _ := auth.UserFrom(r.Context())
	writeJSON(w, 200, u)
}

// frontend serves generated assets and falls back to the SPA shell.
func (a *App) frontend(w http.ResponseWriter, r *http.Request) {
	clean := filepath.Clean(r.URL.Path)
	path := filepath.Join(a.cfg.FrontendDir, clean)
	if info, err := os.Stat(path); err == nil && !info.IsDir() {
		http.ServeFile(w, r, path)
		return
	}
	index := filepath.Join(a.cfg.FrontendDir, "index.html")
	if _, err := os.Stat(index); err != nil {
		problem(w, 404, "frontend_not_built", "Run npm run generate in frontend")
		return
	}
	http.ServeFile(w, r, index)
}

// writeJSON writes one JSON response.
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// problem writes the API error envelope.
func problem(w http.ResponseWriter, status int, code, detail string) {
	writeJSON(w, status, map[string]string{"error": code, "detail": detail})
}

// decode reads one size-limited strict JSON body.
func decode(r *http.Request, dst any) error {
	dec := json.NewDecoder(io.LimitReader(r.Body, 2<<20))
	dec.DisallowUnknownFields()
	return dec.Decode(dst)
}

// securityHeaders adds browser hardening headers.
func securityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("Referrer-Policy", "no-referrer")
		// Nuxt's generated SPA shell contains a small inline import map and bootstrap payload.
		w.Header().Set("Content-Security-Policy", "default-src 'self'; img-src 'self' data:; style-src 'self' 'unsafe-inline'; script-src 'self' 'unsafe-inline'; connect-src 'self'")
		next.ServeHTTP(w, r)
	})
}

// requestLog records request method, path, and latency.
func requestLog(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		slog.Info("request", "method", r.Method, "path", r.URL.Path, "duration", time.Since(start))
	})
}

// actor returns the audit identity associated with a request.
func actor(r *http.Request) string {
	if u, ok := auth.UserFrom(r.Context()); ok {
		return u.Email
	}
	return "anonymous"
}

// validWindow reports whether a dashboard window is supported.
func validWindow(v string) bool { return v == "24h" || v == "7d" || v == "30d" }

// normalizeFormat normalizes and validates a firewall output format.
func normalizeFormat(v string) (string, bool) {
	v = strings.ToLower(v)
	if v == "" {
		return "json", true
	}
	return v, v == "json" || v == "txt"
}
