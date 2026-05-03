package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// HttpRequestsTotal counts all HTTP requests by method and normalized path.
	HttpRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "app_http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "path"},
	)

	// HttpErrorsTotal counts HTTP responses with status codes >= 400.
	HttpErrorsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "app_http_errors_total",
			Help: "Total number of HTTP error responses",
		},
		[]string{"method", "path", "status_code"},
	)

	// HttpRequestDuration stores request duration distribution.
	HttpRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "app_http_request_duration_seconds",
			Help:    "HTTP request duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path"},
	)

	// ActiveRequests shows how many requests are being processed right now.
	ActiveRequests = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "app_active_requests",
			Help: "Number of active HTTP requests",
		},
	)

	// StudentRequestsTotal is an extra business metric for /students/{id} requests.
	StudentRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "app_student_requests_total",
			Help: "Total number of requests for students by student_id",
		},
		[]string{"student_id"},
	)

	// StudentRequestDuration measures duration only for the business route /students/{id}.
	StudentRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "app_student_request_duration_seconds",
			Help:    "Duration of student endpoint requests in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"student_id"},
	)
)
