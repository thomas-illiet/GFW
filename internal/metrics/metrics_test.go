package metrics

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestMetricsMiddlewareAndHandler(t *testing.T) {
	registry := New()
	mux := http.NewServeMux()
	mux.HandleFunc("GET /items/{id}", func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusCreated) })
	registry.Middleware(mux).ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/items/one", nil))

	w := httptest.NewRecorder()
	registry.Handler().ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/metrics", nil))
	body := w.Body.String()
	for _, expected := range []string{
		"gfw_up 1",
		`gfw_http_requests_total{method="GET",route="/items/{id}",status="201"} 1`,
		`gfw_http_request_duration_seconds_count{method="GET",route="/items/{id}",status="201"} 1`,
	} {
		if !strings.Contains(body, expected) {
			t.Fatalf("metrics missing %q:\n%s", expected, body)
		}
	}
	if got := w.Header().Get("Content-Type"); !strings.Contains(got, "version=0.0.4") {
		t.Fatalf("unexpected content type %q", got)
	}
}

func TestMetricsEndpointRejectsOtherRequests(t *testing.T) {
	registry := New()
	for _, request := range []*http.Request{
		httptest.NewRequest(http.MethodPost, "/metrics", nil),
		httptest.NewRequest(http.MethodGet, "/other", nil),
	} {
		w := httptest.NewRecorder()
		registry.Handler().ServeHTTP(w, request)
		if w.Code != http.StatusNotFound {
			t.Fatalf("got %d", w.Code)
		}
	}
}
