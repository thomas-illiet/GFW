// Package metrics exposes dependency-free Prometheus instrumentation.
package metrics

import (
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Registry stores bounded HTTP metrics grouped by method, route pattern, and status.
type Registry struct {
	mu       sync.RWMutex
	requests map[labels]*measurement
}

type labels struct {
	method, route, status string
}

type measurement struct {
	count       uint64
	durationSum float64
}

type responseRecorder struct {
	http.ResponseWriter
	status int
}

func (r *responseRecorder) WriteHeader(status int) {
	if r.status != 0 {
		return
	}
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}

func (r *responseRecorder) Write(body []byte) (int, error) {
	if r.status == 0 {
		r.status = http.StatusOK
	}
	return r.ResponseWriter.Write(body)
}

// New returns an empty metrics registry.
func New() *Registry { return &Registry{requests: make(map[labels]*measurement)} }

// Middleware records one observation after every application request.
func (r *Registry) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		started := time.Now()
		recorder := &responseRecorder{ResponseWriter: w}
		next.ServeHTTP(recorder, request)
		status := recorder.status
		if status == 0 {
			status = http.StatusOK
		}
		route := request.Pattern
		if route == "" {
			route = "unmatched"
		} else {
			route = strings.TrimPrefix(route, request.Method+" ")
		}
		key := labels{method: request.Method, route: route, status: strconv.Itoa(status)}
		r.mu.Lock()
		value := r.requests[key]
		if value == nil {
			value = &measurement{}
			r.requests[key] = value
		}
		value.count++
		value.durationSum += time.Since(started).Seconds()
		r.mu.Unlock()
	})
}

// Handler serves the Prometheus text exposition format on /metrics.
func (r *Registry) Handler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodGet || request.URL.Path != "/metrics" {
			http.NotFound(w, request)
			return
		}
		w.Header().Set("Content-Type", "text/plain; version=0.0.4; charset=utf-8")
		_, _ = fmt.Fprint(w, r.render())
	})
}

func (r *Registry) render() string {
	r.mu.RLock()
	keys := make([]labels, 0, len(r.requests))
	values := make(map[labels]measurement, len(r.requests))
	for key, value := range r.requests {
		keys = append(keys, key)
		values[key] = *value
	}
	r.mu.RUnlock()
	sort.Slice(keys, func(i, j int) bool {
		return keys[i].method+keys[i].route+keys[i].status < keys[j].method+keys[j].route+keys[j].status
	})
	var output strings.Builder
	output.WriteString("# HELP gfw_up Whether the Great Firewall process is running.\n# TYPE gfw_up gauge\ngfw_up 1\n")
	output.WriteString("# HELP gfw_http_requests_total Total application HTTP requests.\n# TYPE gfw_http_requests_total counter\n")
	for _, key := range keys {
		value := values[key]
		fmt.Fprintf(&output, "gfw_http_requests_total{method=%s,route=%s,status=%s} %d\n", strconv.Quote(key.method), strconv.Quote(key.route), strconv.Quote(key.status), value.count)
	}
	output.WriteString("# HELP gfw_http_request_duration_seconds Application request duration.\n# TYPE gfw_http_request_duration_seconds summary\n")
	for _, key := range keys {
		value := values[key]
		labelSet := fmt.Sprintf("method=%s,route=%s,status=%s", strconv.Quote(key.method), strconv.Quote(key.route), strconv.Quote(key.status))
		fmt.Fprintf(&output, "gfw_http_request_duration_seconds_sum{%s} %g\n", labelSet, value.durationSum)
		fmt.Fprintf(&output, "gfw_http_request_duration_seconds_count{%s} %d\n", labelSet, value.count)
	}
	return output.String()
}
