// Package metrics exposes Prometheus metrics: HTTP (RED), outbox and Kafka consumer counters.
// Go runtime and process metrics come from the default registry.
package metrics

import (
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Config is the separate port for /metrics: it must not be exposed to clients like the public API.
type Config struct {
	Port string `envconfig:"METRICS_PORT" default:"8001"`
}

const unmatchedRoute = "unmatched"

var (
	httpRequests = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "http_requests_total",
		Help: "HTTP requests by method, route pattern and status code.",
	}, []string{"method", "route", "status"})

	httpDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "http_request_duration_seconds",
		Help:    "HTTP request duration by method and route pattern.",
		Buckets: prometheus.DefBuckets,
	}, []string{"method", "route"})

	outboxSent = promauto.NewCounter(prometheus.CounterOpts{
		Name: "outbox_events_sent_total",
		Help: "Outbox events delivered to Kafka.",
	})

	outboxErrors = promauto.NewCounter(prometheus.CounterOpts{
		Name: "outbox_send_errors_total",
		Help: "Failed attempts to deliver outbox events to Kafka.",
	})

	kafkaConsumed = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "kafka_consumed_messages_total",
		Help: "Kafka messages fetched by the consumer.",
	}, []string{"topic"})
)

// Handler serves GET /metrics.
func Handler() http.Handler {
	mux := http.NewServeMux()
	mux.Handle("GET /metrics", promhttp.Handler())

	return mux
}

// Middleware records request count and duration. The route is the chi pattern
// (/api/apple/v1/get_apple/{id}), not the raw path, to keep label cardinality bounded.
func Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

		next.ServeHTTP(ww, r)

		status := ww.Status()
		if status == 0 {
			status = http.StatusOK
		}

		route := routePattern(r)

		httpRequests.WithLabelValues(r.Method, route, strconv.Itoa(status)).Inc()
		httpDuration.WithLabelValues(r.Method, route).Observe(time.Since(start).Seconds())
	})
}

func routePattern(r *http.Request) string {
	if rc := chi.RouteContext(r.Context()); rc != nil {
		if p := rc.RoutePattern(); p != "" {
			return p
		}
	}

	return unmatchedRoute
}

// OutboxSent counts events delivered to Kafka.
func OutboxSent(n int) { outboxSent.Add(float64(n)) }

// OutboxError counts a failed delivery attempt.
func OutboxError() { outboxErrors.Inc() }

// KafkaConsumed counts a message fetched from topic.
func KafkaConsumed(topic string) { kafkaConsumed.WithLabelValues(topic).Inc() }
