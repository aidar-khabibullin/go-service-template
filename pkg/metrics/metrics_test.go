package metrics_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/aidar-khabibullin/go-service-template/pkg/metrics"
	"github.com/go-chi/chi/v5"
)

func get(h http.Handler, path string) *httptest.ResponseRecorder {
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequestWithContext(context.Background(), http.MethodGet, path, nil))

	return rec
}

func scrape(t *testing.T) string {
	t.Helper()

	rec := get(metrics.Handler(), "/metrics")
	if rec.Code != http.StatusOK {
		t.Fatalf("/metrics: got %d", rec.Code)
	}

	return rec.Body.String()
}

func TestMiddlewareUsesRoutePatternNotRawPath(t *testing.T) {
	r := chi.NewRouter()
	r.Use(metrics.Middleware)
	r.Get("/items/{id}", func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusNoContent) })

	get(r, "/items/1")
	get(r, "/items/2")
	get(r, "/nope")

	out := scrape(t)

	want := `http_requests_total{method="GET",route="/items/{id}",status="204"} 2`
	if !strings.Contains(out, want) {
		t.Errorf("want %q in scrape output", want)
	}

	// raw ids must not become label values (cardinality)
	if strings.Contains(out, `route="/items/1"`) {
		t.Error("raw path leaked into the route label")
	}

	if !strings.Contains(out, `route="unmatched"`) {
		t.Error("unknown routes must be grouped under route=unmatched")
	}

	if !strings.Contains(out, "http_request_duration_seconds_bucket") {
		t.Error("duration histogram is missing")
	}
}

func TestBusinessCounters(t *testing.T) {
	metrics.OutboxSent(3)
	metrics.OutboxError()
	metrics.KafkaConsumed("apple-topic")

	out := scrape(t)

	for _, want := range []string{
		"outbox_events_sent_total 3",
		"outbox_send_errors_total 1",
		`kafka_consumed_messages_total{topic="apple-topic"} 1`,
	} {
		if !strings.Contains(out, want) {
			t.Errorf("want %q in scrape output", want)
		}
	}
}
