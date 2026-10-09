package server

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/octoplorer/octopulse/internal/telemetry"
)

func TestMetricsUseRouteTemplatesAndStayOffPublicRouter(t *testing.T) {
	s, _, _ := testServer(t)
	s.Metrics = telemetry.New()
	s.Metrics.ObserveStore(s.Store)
	req := httptest.NewRequest(
		http.MethodPost,
		"http://localhost/api/heartbeat/private-monitor/private-token?secret=private-query",
		strings.NewReader(`{"status":"up"}`),
	)
	req.Header.Set("Content-Type", "application/json")
	res := httptest.NewRecorder()
	s.Handler().ServeHTTP(res, req)
	if res.Header().Get("X-Request-ID") == "" {
		t.Fatal("missing request identifier")
	}
	metrics := httptest.NewRecorder()
	s.Metrics.Handler().ServeHTTP(metrics, httptest.NewRequest("GET", "http://localhost/metrics", nil))
	text := metrics.Body.String()
	if metrics.Code != 200 {
		t.Fatalf("missing bounded route or backlog metric: %d %s", metrics.Code, text)
	}
	hasRouteMetric := strings.Contains(text, `/api/heartbeat/{id}/{token}`)
	hasBacklogMetric := strings.Contains(text, "octopulse_delivery_pending 0")
	if !hasRouteMetric || !hasBacklogMetric {
		t.Fatalf("missing bounded route or backlog metric: %d %s", metrics.Code, text)
	}
	for _, secret := range []string{"private-monitor", "private-token", "private-query"} {
		if strings.Contains(text, secret) {
			t.Fatalf("metrics leaked %s", secret)
		}
	}
	public := httptest.NewRecorder()
	s.Handler().ServeHTTP(public, httptest.NewRequest("GET", "http://public.example/metrics", nil))
	if strings.Contains(public.Body.String(), "octopulse_") {
		t.Fatal("operational metrics exposed on public host")
	}
}

func TestDatabaseFailureIsNotReportedAsInvalidCredentials(t *testing.T) {
	s, _, _ := testServer(t)
	s.Store.Close()
	req := httptest.NewRequest("GET", "http://localhost/api/v1/monitors", nil)
	req.AddCookie(&http.Cookie{Name: "octopulse_session", Value: "session-token"})
	res := httptest.NewRecorder()
	s.Handler().ServeHTTP(res, req)
	if res.Code != http.StatusServiceUnavailable {
		t.Fatalf("database failure returned %d: %s", res.Code, res.Body.String())
	}
}
