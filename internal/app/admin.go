package app

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/thomasilliet/gfw/internal/store"
	"golang.org/x/crypto/bcrypt"
)

// ecosystemPayload is the create and update ecosystem request body.
type ecosystemPayload struct {
	Name        string `json:"name"`
	DisplayName string `json:"displayName"`
	IsActive    *bool  `json:"isActive"`
}

// machinePayload is the machine upsert request body.
type machinePayload struct {
	Hostname, EcosystemID, IPv4 string
	IsManaged                   bool
}

// trustedPayload is the trusted-service request body.
type trustedPayload struct {
	Name, Description string
	IsActive          bool
	Hostnames         []string
}

// dashboard serves administration activity aggregates.
func (a *App) dashboard(w http.ResponseWriter, r *http.Request) {
	window := r.URL.Query().Get("window")
	if window == "" {
		window = "24h"
	}
	if !validWindow(window) {
		problem(w, 400, "invalid_window", "window must be 24h, 7d, or 30d")
		return
	}
	v, err := a.store.Dashboard(r.Context(), window)
	if err != nil {
		problem(w, 500, "dashboard_failed", err.Error())
		return
	}
	writeJSON(w, 200, v)
}

// ecosystems lists or creates ecosystems.
func (a *App) ecosystems(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		v, err := a.store.ListEcosystems(r.Context())
		if err != nil {
			problem(w, 500, "list_failed", err.Error())
			return
		}
		writeJSON(w, 200, map[string]any{"items": v, "total": len(v)})
		return
	}
	if !sameOrigin(r) {
		problem(w, 403, "csrf_failed", "cross-origin mutation rejected")
		return
	}
	var p ecosystemPayload
	if err := decode(r, &p); err != nil {
		problem(w, 400, "invalid_payload", err.Error())
		return
	}
	v, err := a.store.CreateEcosystem(r.Context(), p.Name, p.DisplayName)
	if err != nil {
		problem(w, 409, "ecosystem_conflict", err.Error())
		return
	}
	a.store.Audit(r.Context(), actor(r), "create", "ecosystem", v.ID)
	writeJSON(w, 201, v)
}

// ecosystemUpdate changes one ecosystem.
func (a *App) ecosystemUpdate(w http.ResponseWriter, r *http.Request) {
	if !sameOrigin(r) {
		problem(w, 403, "csrf_failed", "cross-origin mutation rejected")
		return
	}
	var p ecosystemPayload
	if err := decode(r, &p); err != nil {
		problem(w, 400, "invalid_payload", err.Error())
		return
	}
	active := true
	if p.IsActive != nil {
		active = *p.IsActive
	}
	v, err := a.store.UpdateEcosystem(r.Context(), r.PathValue("id"), p.DisplayName, active)
	if err != nil {
		problem(w, 404, "ecosystem_not_found", err.Error())
		return
	}
	a.store.Audit(r.Context(), actor(r), "update", "ecosystem", v.ID)
	writeJSON(w, 200, v)
}

// ecosystemDelete removes an ecosystem and its assigned machines.
func (a *App) ecosystemDelete(w http.ResponseWriter, r *http.Request) {
	if !sameOrigin(r) {
		problem(w, 403, "csrf_failed", "cross-origin mutation rejected")
		return
	}
	id := r.PathValue("id")
	if err := a.store.DeleteEcosystem(r.Context(), id); err != nil {
		if errors.Is(err, store.ErrNotFound) {
			problem(w, 404, "ecosystem_not_found", "ecosystem does not exist")
			return
		}
		problem(w, 500, "ecosystem_delete_failed", err.Error())
		return
	}
	a.store.Audit(r.Context(), actor(r), "delete", "ecosystem", id)
	w.WriteHeader(204)
}

// machines lists or creates machines.
func (a *App) machines(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		v, err := a.store.ListMachines(r.Context(), r.URL.Query().Get("search"), r.URL.Query().Get("ecosystemId"))
		if err != nil {
			problem(w, 500, "list_failed", err.Error())
			return
		}
		writeJSON(w, 200, map[string]any{"items": v, "total": len(v)})
		return
	}
	if !sameOrigin(r) {
		problem(w, 403, "csrf_failed", "cross-origin mutation rejected")
		return
	}
	var p machinePayload
	if err := decode(r, &p); err != nil {
		problem(w, 400, "invalid_payload", err.Error())
		return
	}
	p.IsManaged = false
	v, err := a.store.UpsertMachine(r.Context(), p.Hostname, p.EcosystemID, p.IPv4, false)
	if err != nil {
		problem(w, 400, "invalid_machine", err.Error())
		return
	}
	a.store.Audit(r.Context(), actor(r), "upsert", "machine", v.ID)
	writeJSON(w, 201, v)
}

// machineDelete removes a manually managed machine.
func (a *App) machineDelete(w http.ResponseWriter, r *http.Request) {
	if !sameOrigin(r) {
		problem(w, 403, "csrf_failed", "cross-origin mutation rejected")
		return
	}
	if err := a.store.DeleteMachine(r.Context(), r.PathValue("id")); err != nil {
		problem(w, 404, "machine_not_found", "managed or missing machine")
		return
	}
	a.store.Audit(r.Context(), actor(r), "delete", "machine", r.PathValue("id"))
	w.WriteHeader(204)
}

// trustedServices lists or creates trusted services.
func (a *App) trustedServices(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		v, err := a.store.ListTrustedServices(r.Context())
		if err != nil {
			problem(w, 500, "list_failed", err.Error())
			return
		}
		writeJSON(w, 200, map[string]any{"items": v, "total": len(v)})
		return
	}
	if !sameOrigin(r) {
		problem(w, 403, "csrf_failed", "cross-origin mutation rejected")
		return
	}
	var p trustedPayload
	if err := decode(r, &p); err != nil {
		problem(w, 400, "invalid_payload", err.Error())
		return
	}
	v, err := a.store.CreateTrustedService(r.Context(), p.Name, p.Description, p.IsActive, p.Hostnames)
	if err != nil {
		problem(w, 409, "trusted_service_conflict", err.Error())
		return
	}
	a.store.Audit(r.Context(), actor(r), "create", "trusted_service", v.ID)
	writeJSON(w, 201, v)
}

// trustedServiceUpdate changes one trusted service.
func (a *App) trustedServiceUpdate(w http.ResponseWriter, r *http.Request) {
	if !sameOrigin(r) {
		problem(w, 403, "csrf_failed", "cross-origin mutation rejected")
		return
	}
	var p trustedPayload
	if err := decode(r, &p); err != nil {
		problem(w, 400, "invalid_payload", err.Error())
		return
	}
	v, err := a.store.UpdateTrustedService(r.Context(), r.PathValue("id"), p.Name, p.Description, p.IsActive, p.Hostnames)
	if err != nil {
		problem(w, 404, "trusted_service_not_found", err.Error())
		return
	}
	a.store.Audit(r.Context(), actor(r), "update", "trusted_service", v.ID)
	writeJSON(w, 200, v)
}

// trustedServiceDelete removes one trusted service.
func (a *App) trustedServiceDelete(w http.ResponseWriter, r *http.Request) {
	if !sameOrigin(r) {
		problem(w, 403, "csrf_failed", "cross-origin mutation rejected")
		return
	}
	if err := a.store.DeleteTrustedService(r.Context(), r.PathValue("id")); err != nil {
		problem(w, 404, "trusted_service_not_found", err.Error())
		return
	}
	a.store.Audit(r.Context(), actor(r), "delete", "trusted_service", r.PathValue("id"))
	w.WriteHeader(204)
}

// resolveTrusted triggers an immediate refresh of trusted DNS entries.
func (a *App) resolveTrusted(w http.ResponseWriter, r *http.Request) {
	if !sameOrigin(r) {
		problem(w, 403, "csrf_failed", "cross-origin mutation rejected")
		return
	}
	if err := a.dns.ResolveAll(r.Context()); err != nil {
		problem(w, 500, "resolution_failed", err.Error())
		return
	}
	a.store.Audit(r.Context(), actor(r), "resolve", "trusted_service", "")
	writeJSON(w, 200, map[string]string{"status": "resolved"})
}

// ingestionKeys lists or creates ingestion API keys.
func (a *App) ingestionKeys(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		v, err := a.store.ListIngestionKeys(r.Context())
		if err != nil {
			problem(w, 500, "list_failed", err.Error())
			return
		}
		writeJSON(w, 200, map[string]any{"items": v, "total": len(v)})
		return
	}
	if !sameOrigin(r) {
		problem(w, 403, "csrf_failed", "cross-origin mutation rejected")
		return
	}
	var p struct {
		Name      string     `json:"name"`
		ExpiresAt *time.Time `json:"expiresAt"`
	}
	if err := decode(r, &p); err != nil {
		problem(w, 400, "invalid_payload", err.Error())
		return
	}
	prefix := randomToken(6)
	secret := prefix + "." + randomToken(32)
	hash, _ := bcrypt.GenerateFromPassword([]byte(secret), bcrypt.DefaultCost)
	v, err := a.store.CreateIngestionKey(r.Context(), p.Name, prefix, string(hash), p.ExpiresAt)
	if err != nil {
		problem(w, 409, "key_conflict", err.Error())
		return
	}
	a.store.Audit(r.Context(), actor(r), "create", "ingestion_key", v.ID)
	writeJSON(w, 201, map[string]any{"key": v, "secret": secret})
}

// revokeIngestionKey revokes one ingestion API key.
func (a *App) revokeIngestionKey(w http.ResponseWriter, r *http.Request) {
	if !sameOrigin(r) {
		problem(w, 403, "csrf_failed", "cross-origin mutation rejected")
		return
	}
	if err := a.store.RevokeIngestionKey(r.Context(), r.PathValue("id")); err != nil {
		problem(w, 404, "key_not_found", err.Error())
		return
	}
	a.store.Audit(r.Context(), actor(r), "revoke", "ingestion_key", r.PathValue("id"))
	w.WriteHeader(204)
}

// ingestMachines authenticates an agent and upserts its machine batch.
func (a *App) ingestMachines(w http.ResponseWriter, r *http.Request) {
	token := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
	parts := strings.SplitN(token, ".", 2)
	if len(parts) != 2 {
		problem(w, 401, "invalid_api_key", "missing API key")
		return
	}
	id, hash, err := a.store.IngestionKeyHash(r.Context(), parts[0])
	if err != nil || bcrypt.CompareHashAndPassword([]byte(hash), []byte(token)) != nil {
		problem(w, 401, "invalid_api_key", "invalid API key")
		return
	}
	var p struct {
		Machines []machinePayload `json:"machines"`
	}
	if err := decode(r, &p); err != nil {
		problem(w, 400, "invalid_payload", err.Error())
		return
	}
	created := make([]store.Machine, 0, len(p.Machines))
	for i, m := range p.Machines {
		v, err := a.store.UpsertMachine(r.Context(), m.Hostname, m.EcosystemID, m.IPv4, true)
		if err != nil {
			problem(w, 400, "invalid_machine", "item "+strconv.Itoa(i)+": "+err.Error())
			return
		}
		created = append(created, v)
	}
	a.store.TouchIngestionKey(r.Context(), id)
	writeJSON(w, 200, map[string]any{"items": created, "upserted": len(created)})
}

// sameOrigin applies the origin check used by cookie-authenticated mutations.
func sameOrigin(r *http.Request) bool {
	origin := r.Header.Get("Origin")
	if origin == "" {
		return true
	}
	return strings.HasPrefix(origin, "http://"+r.Host) || strings.HasPrefix(origin, "https://"+r.Host)
}

// randomToken returns a URL-safe cryptographically random value.
func randomToken(n int) string {
	b := make([]byte, n)
	_, _ = rand.Read(b)
	return base64.RawURLEncoding.EncodeToString(b)
}
