package main

import (
	"net/http"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

// httpMetrics tracks request count and duration, labeled by method and
// status only — deliberately not by route/path, which would need a safe
// pattern label (e.g. "/posts/{slug}", not the literal requested URI)
// injected at every route registration to avoid unbounded label
// cardinality from unique slugs. Deferred until per-route breakdown is a
// genuine need.
type httpMetrics struct {
	requestsTotal   *prometheus.CounterVec
	requestDuration *prometheus.HistogramVec
}

func newHTTPMetrics(reg *prometheus.Registry) *httpMetrics {
	m := &httpMetrics{
		requestsTotal: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "blog_http_requests_total",
			Help: "Total number of HTTP requests processed, labeled by method and status.",
		}, []string{"method", "status"}),
		requestDuration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "blog_http_request_duration_seconds",
			Help:    "HTTP request duration in seconds, labeled by method and status.",
			Buckets: prometheus.DefBuckets,
		}, []string{"method", "status"}),
	}

	reg.MustRegister(m.requestsTotal, m.requestDuration)

	return m
}

// middleware reuses statusRecorder (middleware.go) rather than reinventing
// status capture — the same technique, a separate wrap, since logging and
// metrics are separate concerns even though they both need the status code.
func (m *httpMetrics) middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rec := &statusRecorder{ResponseWriter: w}

		next.ServeHTTP(rec, r)

		statusLabel := strconv.Itoa(rec.Status())

		m.requestsTotal.WithLabelValues(r.Method, statusLabel).Inc()
		m.requestDuration.WithLabelValues(r.Method, statusLabel).Observe(time.Since(start).Seconds())
	})
}
