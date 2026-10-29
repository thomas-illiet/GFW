package auth

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/thomasilliet/gfw/internal/config"
	"golang.org/x/oauth2"
)

// TestDevAuthentication covers development construction, middleware, login, and logout.
func TestDevAuthentication(t *testing.T) {
	m, err := New(context.Background(), config.Config{DevAuth: true, PublicURL: "http://localhost"})
	if err != nil {
		t.Fatal(err)
	}
	var got User
	h := m.Middleware(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) { got, _ = UserFrom(r.Context()) }))
	h.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil))
	if got.Role != "admin" {
		t.Fatalf("unexpected user: %+v", got)
	}
	w := httptest.NewRecorder()
	m.Login(w, httptest.NewRequest(http.MethodGet, "/auth/login", nil))
	if w.Code != http.StatusFound {
		t.Fatalf("login status %d", w.Code)
	}
	w = httptest.NewRecorder()
	m.Logout(w, httptest.NewRequest(http.MethodPost, "/auth/logout", nil))
	if w.Code != http.StatusFound || len(w.Result().Cookies()) == 0 {
		t.Fatal("logout cookie missing")
	}
}

// TestProductionLoginAndCallbackFailures covers state generation and callback guards.
func TestProductionLoginAndCallbackFailures(t *testing.T) {
	m := &Manager{cfg: config.Config{PublicURL: "https://gfw.test"}, oauth: oauth2.Config{ClientID: "id", RedirectURL: "https://gfw.test/auth/callback", Endpoint: oauth2.Endpoint{AuthURL: "https://issuer.test/auth", TokenURL: "http://127.0.0.1:1/token"}}}
	w := httptest.NewRecorder()
	m.Login(w, httptest.NewRequest(http.MethodGet, "/auth/login", nil))
	if w.Code != http.StatusFound || !strings.Contains(w.Header().Get("Location"), "state=") {
		t.Fatal("login redirect")
	}
	w = httptest.NewRecorder()
	m.Callback(w, httptest.NewRequest(http.MethodGet, "/auth/callback", nil))
	if w.Code != 400 {
		t.Fatalf("callback state: %d", w.Code)
	}
	req := httptest.NewRequest(http.MethodGet, "/auth/callback?state=ok&code=bad", nil)
	req.AddCookie(&http.Cookie{Name: "gfw_oidc_state", Value: "ok"})
	w = httptest.NewRecorder()
	m.Callback(w, req)
	if w.Code != 401 {
		t.Fatalf("callback exchange: %d", w.Code)
	}
	if len(random()) == 0 {
		t.Fatal("random state empty")
	}
}

// TestAnonymousAndExpiredMiddleware covers requests without a valid session.
func TestAnonymousAndExpiredMiddleware(t *testing.T) {
	m, _ := New(context.Background(), config.Config{DevAuth: true})
	m.cfg.DevAuth = false
	for _, cookie := range []*http.Cookie{nil, &http.Cookie{Name: "gfw_session", Value: "invalid"}} {
		called := false
		req := httptest.NewRequest("GET", "/", nil)
		if cookie != nil {
			req.AddCookie(cookie)
		}
		m.Middleware(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) { _, ok := UserFrom(r.Context()); called = !ok })).ServeHTTP(httptest.NewRecorder(), req)
		if !called {
			t.Fatal("anonymous request was authenticated")
		}
	}
}

// TestAdminTokenMiddleware covers accepted and rejected administrator API credentials.
func TestAdminTokenMiddleware(t *testing.T) {
	m := &Manager{cfg: config.Config{AdminToken: "01234567890123456789012345678901"}}
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		u, ok := UserFrom(r.Context())
		if !ok || u.Role != "admin" || u.Subject != "admin-token" {
			t.Fatalf("unexpected token identity: %+v", u)
		}
		w.WriteHeader(http.StatusNoContent)
	})
	for _, test := range []struct {
		header string
		want   int
	}{
		{header: "Bearer 01234567890123456789012345678901", want: http.StatusNoContent},
		{header: "bearer 01234567890123456789012345678901", want: http.StatusNoContent},
		{header: "Bearer invalid", want: http.StatusUnauthorized},
		{header: "Basic 01234567890123456789012345678901", want: http.StatusUnauthorized},
		{want: http.StatusUnauthorized},
	} {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Authorization", test.header)
		w := httptest.NewRecorder()
		m.Middleware(m.Require("admin")(next)).ServeHTTP(w, req)
		if w.Code != test.want {
			t.Fatalf("header %q got %d want %d", test.header, w.Code, test.want)
		}
	}
}

// TestSessionSigning covers valid, malformed, tampered, and expired cookies.
func TestSessionSigning(t *testing.T) {
	m, _ := New(context.Background(), config.Config{DevAuth: true})
	u := User{Subject: "sub", Role: "editor", ExpiresAt: time.Now().Add(time.Hour)}
	value, err := m.sign(u)
	if err != nil {
		t.Fatal(err)
	}
	parsed, err := m.parse(value)
	if err != nil || parsed.Subject != "sub" {
		t.Fatalf("parse failed: %+v %v", parsed, err)
	}
	for _, bad := range []string{"bad", value + "x", strings.Split(value, ".")[0] + ".%%%"} {
		if _, err := m.parse(bad); err == nil {
			t.Fatalf("accepted %q", bad)
		}
	}
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{Name: "gfw_session", Value: value})
	called := false
	m.cfg.DevAuth = false
	m.Middleware(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) { called = true })).ServeHTTP(httptest.NewRecorder(), req)
	if !called {
		t.Fatal("middleware did not continue")
	}
}

// TestRequireAndRoles covers authentication and authorization decisions.
func TestRequireAndRoles(t *testing.T) {
	m := &Manager{}
	next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusNoContent) })
	for _, test := range []struct {
		role string
		want int
	}{{"", 401}, {"viewer", 403}, {"editor", 204}, {"admin", 204}} {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		if test.role != "" {
			req = req.WithContext(context.WithValue(req.Context(), ctxKey{}, User{Role: test.role}))
		}
		w := httptest.NewRecorder()
		m.Require("editor")(next).ServeHTTP(w, req)
		if w.Code != test.want {
			t.Fatalf("role %q got %d want %d", test.role, w.Code, test.want)
		}
	}
	if roleRank("admin") != 3 || roleRank("editor") != 2 || roleRank("other") != 1 {
		t.Fatal("role ranking")
	}
}

// TestRoleFromGroups covers the production OIDC group-to-role contract.
func TestRoleFromGroups(t *testing.T) {
	tests := []struct {
		name   string
		groups []string
		want   string
	}{
		{name: "authenticated without group", want: "viewer"},
		{name: "read", groups: []string{"read"}, want: "viewer"},
		{name: "prefixed read", groups: []string{"gfw-read"}, want: "viewer"},
		{name: "write", groups: []string{"write"}, want: "editor"},
		{name: "prefixed write", groups: []string{"gfw-write"}, want: "editor"},
		{name: "admin", groups: []string{"admin"}, want: "admin"},
		{name: "prefixed admin", groups: []string{"gfw-admin"}, want: "admin"},
		{name: "admin wins", groups: []string{"write", "ADMIN", "read"}, want: "admin"},
		{name: "legacy aliases", groups: []string{"viewer", "editor"}, want: "editor"},
		{name: "surrounding spaces", groups: []string{" Write "}, want: "editor"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := roleFromGroups(test.groups); got != test.want {
				t.Fatalf("roleFromGroups(%v) = %q, want %q", test.groups, got, test.want)
			}
		})
	}
}
