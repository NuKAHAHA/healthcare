package middleware

import (
	"fmt"
	"healthcare-api/internal/logger"
	"net/http"
	"time"
)

// RequestLogger logs HTTP requests and responses
func RequestLogger(logger *logger.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Create a response writer wrapper to capture status code
			wrapped := &responseWriterWrapper{ResponseWriter: w}

			// Call next handler
			next.ServeHTTP(wrapped, r)

			// Log request
			duration := time.Since(start)
			logger.InfoWithContext(
				fmt.Sprintf("%s %s %d", r.Method, r.URL.Path, wrapped.statusCode),
				"http_request",
				0,
				fmt.Sprintf("method=%s path=%s status=%d duration=%dms", r.Method, r.URL.Path, wrapped.statusCode, duration.Milliseconds()),
			)
		})
	}
}

// responseWriterWrapper wraps http.ResponseWriter to capture status code
type responseWriterWrapper struct {
	http.ResponseWriter
	statusCode int
}

func (w *responseWriterWrapper) WriteHeader(code int) {
	w.statusCode = code
	w.ResponseWriter.WriteHeader(code)
}

func (w *responseWriterWrapper) Write(b []byte) (int, error) {
	if w.statusCode == 0 {
		w.statusCode = http.StatusOK
	}
	return w.ResponseWriter.Write(b)
}
