package middleware

import (
	"net/http"
	"regexp"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	httpRequestsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "diplomaverify_http_requests_total",
		Help: "Total HTTP requests by method, path, status",
	}, []string{"method", "path", "status"})

	httpRequestDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "diplomaverify_http_request_duration_seconds",
		Help:    "HTTP request latency in seconds",
		Buckets: prometheus.ExponentialBuckets(0.005, 2, 10), // 5ms -> 5.12s
	}, []string{"method", "path"})

	rateLimitHits = promauto.NewCounter(prometheus.CounterOpts{
		Name: "diplomaverify_rate_limit_hits_total",
		Help: "Total 429 responses due to rate limiting",
	})
)

// Path normalizer to avoid Prometheus high cardinality
var pathRegex = regexp.MustCompile(`/(api/v1/verify/qr/|api/v1/diplomas/)[^/]+`)

func MetricsHandler() http.Handler {
	return promhttp.Handler()
}

func MetricsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rw := &responseWriter{ResponseWriter: w, status: http.StatusOK}
		
		next.ServeHTTP(rw, r)
		
		duration := time.Since(start).Seconds()
		path := pathRegex.ReplaceAllString(r.URL.Path, "$1{id}")
		status := strconv.Itoa(rw.status)
		
		httpRequestsTotal.WithLabelValues(r.Method, path, status).Inc()
		httpRequestDuration.WithLabelValues(r.Method, path).Observe(duration)
		if rw.status == http.StatusTooManyRequests {
			rateLimitHits.Inc()
		}
	})
}

// responseWriter обёртка для захвата статуса ответа
type responseWriter struct {
	http.ResponseWriter
	status int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
}