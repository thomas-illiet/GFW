package auth

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/thomasilliet/gfw/internal/config"
	"golang.org/x/oauth2"
)

// User is the authenticated identity stored in a signed session.
type User struct {
	Subject   string    `json:"subject"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Role      string    `json:"role"`
	ExpiresAt time.Time `json:"expiresAt"`
}

// ctxKey isolates authenticated-user context values.
type ctxKey struct{}

// Manager implements OIDC login, signed sessions, and role authorization.
type Manager struct {
	cfg      config.Config
	provider *oidc.Provider
	verifier *oidc.IDTokenVerifier
	oauth    oauth2.Config
	secret   []byte
}

// New constructs an authentication manager and discovers the OIDC provider when enabled.
func New(ctx context.Context, c config.Config) (*Manager, error) {
	m := &Manager{cfg: c, secret: []byte(c.SessionSecret)}
	if c.DevAuth {
		m.secret = []byte("development-secret-not-for-production")
		return m, nil
	}
	p, err := oidc.NewProvider(ctx, c.OIDCIssuer)
	if err != nil {
		return nil, err
	}
	m.provider = p
	m.verifier = p.Verifier(&oidc.Config{ClientID: c.OIDCClientID})
	m.oauth = oauth2.Config{ClientID: c.OIDCClientID, ClientSecret: c.OIDCSecret, Endpoint: p.Endpoint(), RedirectURL: c.OIDCRedirect, Scopes: []string{oidc.ScopeOpenID, "profile", "email", "groups"}}
	return m, nil
}

// UserFrom reads the authenticated user from a request context.
func UserFrom(ctx context.Context) (User, bool) { u, ok := ctx.Value(ctxKey{}).(User); return u, ok }

// Middleware attaches a valid session identity to each request.
func (m *Manager) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if m.cfg.DevAuth {
			u := User{Subject: "developer", Name: "Developer", Email: "dev@localhost", Role: "admin", ExpiresAt: time.Now().Add(time.Hour)}
			next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), ctxKey{}, u)))
			return
		}
		cookie, err := r.Cookie("gfw_session")
		if err != nil {
			next.ServeHTTP(w, r)
			return
		}
		u, err := m.parse(cookie.Value)
		if err == nil && u.ExpiresAt.After(time.Now()) {
			next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), ctxKey{}, u)))
			return
		}
		next.ServeHTTP(w, r)
	})
}

// Require enforces authentication and a minimum allowed role.
func (m *Manager) Require(roles ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			u, ok := UserFrom(r.Context())
			if !ok {
				writeAuthError(w, http.StatusUnauthorized, "authentication_required")
				return
			}
			for _, role := range roles {
				if roleRank(u.Role) >= roleRank(role) {
					next.ServeHTTP(w, r)
					return
				}
			}
			writeAuthError(w, http.StatusForbidden, "insufficient_role")
		})
	}
}

// Login starts the OIDC authorization-code flow.
func (m *Manager) Login(w http.ResponseWriter, r *http.Request) {
	if m.cfg.DevAuth {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}
	state := random()
	http.SetCookie(w, &http.Cookie{Name: "gfw_oidc_state", Value: state, Path: "/auth", HttpOnly: true, Secure: strings.HasPrefix(m.cfg.PublicURL, "https://"), SameSite: http.SameSiteLaxMode, MaxAge: 600})
	http.Redirect(w, r, m.oauth.AuthCodeURL(state), http.StatusFound)
}

// Callback validates the OIDC response and creates a signed session.
func (m *Manager) Callback(w http.ResponseWriter, r *http.Request) {
	state, err := r.Cookie("gfw_oidc_state")
	if err != nil || r.URL.Query().Get("state") != state.Value {
		writeAuthError(w, 400, "invalid_state")
		return
	}
	token, err := m.oauth.Exchange(r.Context(), r.URL.Query().Get("code"))
	if err != nil {
		writeAuthError(w, 401, "oidc_exchange_failed")
		return
	}
	raw, _ := token.Extra("id_token").(string)
	idToken, err := m.verifier.Verify(r.Context(), raw)
	if err != nil {
		writeAuthError(w, 401, "invalid_id_token")
		return
	}
	var claims struct {
		Sub, Name, Email string
		Groups           []string
	}
	if err = idToken.Claims(&claims); err != nil {
		writeAuthError(w, 401, "invalid_claims")
		return
	}
	role := roleFromGroups(claims.Groups)
	u := User{Subject: claims.Sub, Name: claims.Name, Email: claims.Email, Role: role, ExpiresAt: time.Now().Add(8 * time.Hour)}
	value, _ := m.sign(u)
	http.SetCookie(w, &http.Cookie{Name: "gfw_session", Value: value, Path: "/", HttpOnly: true, Secure: strings.HasPrefix(m.cfg.PublicURL, "https://"), SameSite: http.SameSiteLaxMode, MaxAge: 8 * 3600})
	http.Redirect(w, r, "/", http.StatusFound)
}

// roleFromGroups maps OIDC read, write, and admin groups to internal roles.
func roleFromGroups(groups []string) string {
	role := "viewer"
	for _, group := range groups {
		switch strings.ToLower(strings.TrimSpace(group)) {
		case "gfw-admin", "admin":
			return "admin"
		case "gfw-write", "write", "gfw-editor", "editor":
			role = "editor"
		case "gfw-read", "read", "gfw-viewer", "viewer":
			// Viewer is the baseline role for every authenticated user.
		}
	}
	return role
}

// Logout clears the browser session.
func (m *Manager) Logout(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{Name: "gfw_session", Value: "", Path: "/", HttpOnly: true, MaxAge: -1})
	http.Redirect(w, r, "/login", http.StatusFound)
}

// sign serializes and authenticates a user session.
func (m *Manager) sign(u User) (string, error) {
	b, err := json.Marshal(u)
	if err != nil {
		return "", err
	}
	payload := base64.RawURLEncoding.EncodeToString(b)
	mac := hmac.New(sha256.New, m.secret)
	mac.Write([]byte(payload))
	return payload + "." + base64.RawURLEncoding.EncodeToString(mac.Sum(nil)), nil
}

// parse verifies and decodes a signed user session.
func (m *Manager) parse(value string) (User, error) {
	parts := strings.Split(value, ".")
	if len(parts) != 2 {
		return User{}, errors.New("invalid cookie")
	}
	mac := hmac.New(sha256.New, m.secret)
	mac.Write([]byte(parts[0]))
	sig, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil || !hmac.Equal(sig, mac.Sum(nil)) {
		return User{}, errors.New("invalid signature")
	}
	body, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return User{}, err
	}
	var u User
	err = json.Unmarshal(body, &u)
	return u, err
}

// random returns a cryptographically random OIDC state value.
func random() string {
	b := make([]byte, 24)
	_, _ = rand.Read(b)
	return base64.RawURLEncoding.EncodeToString(b)
}

// roleRank converts a role to its authorization ordering.
func roleRank(role string) int {
	switch role {
	case "admin":
		return 3
	case "editor":
		return 2
	default:
		return 1
	}
}

// writeAuthError writes the authentication error envelope.
func writeAuthError(w http.ResponseWriter, status int, code string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": code})
}
