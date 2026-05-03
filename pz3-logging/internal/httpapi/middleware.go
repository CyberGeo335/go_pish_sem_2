package httpapi

import (
	"net/http"
	"strconv"
	"time"

	"go.uber.org/zap"
)

const RequestIDHeader = "X-Request-ID"

// LoggingMiddleware logs incoming HTTP requests and completed responses.
func LoggingMiddleware(log *zap.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		lrw := NewLoggingResponseWriter(w)

		requestID := time.Now().UnixNano()
		requestIDString := strconv.FormatInt(requestID, 10)
		lrw.Header().Set(RequestIDHeader, requestIDString)

		log.Info("incoming request",
			zap.Int64("request_id", requestID),
			zap.String("method", r.Method),
			zap.String("path", r.URL.Path),
			zap.String("remote_addr", r.RemoteAddr),
		)

		next.ServeHTTP(lrw, r)

		duration := time.Since(start)

		log.Info("request completed",
			zap.Int64("request_id", requestID),
			zap.String("method", r.Method),
			zap.String("path", r.URL.Path),
			zap.Int("status_code", lrw.StatusCode()),
			zap.Duration("duration", duration),
		)
	})
}
