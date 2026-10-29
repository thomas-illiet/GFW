package app

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/thomasilliet/gfw/internal/store"
)

// firewallResponse is the stable JSON firewall list representation.
type firewallResponse struct {
	Ecosystem   string              `json:"ecosystem,omitempty"`
	Version     string              `json:"version"`
	GeneratedAt time.Time           `json:"generatedAt"`
	Items       []store.EffectiveIP `json:"items"`
}

// firewallEcosystem publishes the effective list for one ecosystem.
func (a *App) firewallEcosystem(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	format, ok := normalizeFormat(r.URL.Query().Get("format"))
	if !ok {
		problem(w, 400, "invalid_format", "format must be json or txt")
		return
	}
	name, items, err := a.store.EffectiveIPs(r.Context(), r.PathValue("ecosystem"))
	if err == store.ErrNotFound {
		problem(w, 404, "ecosystem_not_found", "ecosystem not found")
		return
	}
	if err != nil {
		problem(w, 500, "list_failed", err.Error())
		return
	}
	body, etag, err := encodeList(name, format, items)
	if err != nil {
		problem(w, 500, "encode_failed", err.Error())
		return
	}
	notModified := r.Header.Get("If-None-Match") == etag
	status := http.StatusOK
	if notModified {
		status = http.StatusNotModified
	}
	history := store.HistoryInput{RouteKind: "ecosystem", EcosystemName: name, Format: format, StatusCode: status, DurationMS: int(time.Since(start).Milliseconds()), ETag: etag, NotModified: notModified, Entries: items}
	if err := a.store.RecordRequest(r.Context(), history); err != nil {
		problem(w, 500, "history_failed", "request could not be recorded")
		return
	}
	w.Header().Set("ETag", etag)
	w.Header().Set("Cache-Control", "no-cache")
	if notModified {
		w.WriteHeader(304)
		return
	}
	if format == "txt" {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	} else {
		w.Header().Set("Content-Type", "application/json")
	}
	w.WriteHeader(200)
	_, _ = w.Write(body)
}

// firewallIP publishes provenance details for one address.
func (a *App) firewallIP(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	items, err := a.store.IPDetails(r.Context(), r.PathValue("ipv4"))
	if err == store.ErrNotFound {
		problem(w, 404, "ip_not_found", "IP not found")
		return
	}
	if err != nil {
		problem(w, 400, "invalid_ipv4", err.Error())
		return
	}
	body, _ := json.Marshal(map[string]any{"ipv4": r.PathValue("ipv4"), "items": items})
	hash := sha256.Sum256(body)
	etag := `"` + hex.EncodeToString(hash[:]) + `"`
	notModified := r.Header.Get("If-None-Match") == etag
	status := 200
	if notModified {
		status = 304
	}
	if err := a.store.RecordRequest(r.Context(), store.HistoryInput{RouteKind: "ip", Format: "json", StatusCode: status, DurationMS: int(time.Since(start).Milliseconds()), ETag: etag, NotModified: notModified, Entries: items}); err != nil {
		problem(w, 500, "history_failed", "request could not be recorded")
		return
	}
	w.Header().Set("ETag", etag)
	if notModified {
		w.WriteHeader(304)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write(body)
}

// encodeList creates a deterministic representation and semantic ETag.
func encodeList(name, format string, items []store.EffectiveIP) ([]byte, string, error) {
	stableVersion := version(items)
	var body []byte
	var err error
	if format == "txt" {
		seen := map[string]bool{}
		var ips []string
		for _, item := range items {
			if !seen[item.IPv4] {
				seen[item.IPv4] = true
				ips = append(ips, item.IPv4)
			}
		}
		sort.Strings(ips)
		body = []byte(strings.Join(ips, "\n"))
		if len(ips) > 0 {
			body = append(body, '\n')
		}
	} else {
		body, err = json.Marshal(firewallResponse{Ecosystem: name, Version: stableVersion, GeneratedAt: time.Now().UTC(), Items: items})
	}
	if err != nil {
		return nil, "", err
	}
	hash := sha256.Sum256([]byte(format + "|" + name + "|" + stableVersion))
	return body, `"` + hex.EncodeToString(hash[:]) + `"`, nil
}

// version hashes address provenance into a stable list version.
func version(items []store.EffectiveIP) string {
	h := sha256.New()
	for _, v := range items {
		fmt.Fprintf(h, "%s|%s|%s\n", v.IPv4, v.SourceType, v.SourceID)
	}
	return hex.EncodeToString(h.Sum(nil))
}
