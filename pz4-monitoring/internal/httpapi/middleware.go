package httpapi

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/CyberGeo335/pz4-monitoring/internal/metrics"
)

// MetricsMiddleware records request count, errors, duration, and active requests.
func MetricsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		metrics.ActiveRequests.Inc()
		defer metrics.ActiveRequests.Dec()

		lrw := NewLoggingResponseWriter(w)
		next.ServeHTTP(lrw, r)

		duration := time.Since(start).Seconds()
		path := normalizePath(r.URL.Path)

		metrics.HttpRequestsTotal.WithLabelValues(r.Method, path).Inc()
		metrics.HttpRequestDuration.WithLabelValues(r.Method, path).Observe(duration)

		if lrw.StatusCode() >= http.StatusBadRequest {
			metrics.HttpErrorsTotal.WithLabelValues(
				r.Method,
				path,
				strconv.Itoa(lrw.StatusCode()),
			).Inc()
		}
	})
}

func normalizePath(path string) string {
	switch {
	case path == "/health":
		return "/health"
	case path == "/metrics":
		return "/metrics"
	case path == "/students" || strings.HasPrefix(path, "/students/"):
		return "/students/{id}"
	case path == "":
		return "/"
	default:
		return path
	}
}
