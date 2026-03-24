package middleware

import (
	"net/http"
	"time"

	"github.com/markbfromdc/cloudcode/internal/logging"
)

// responseWriter wraps http.ResponseWriter to capture the status code.
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// RequestLogger returns middleware that logs each HTTP request with timing information.
func RequestLogger(log *logging.Logger) func(http.Handler) http.Handler {
	reqLog := log.WithField("component", "http")

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			rw := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}
			next.ServeHTTP(rw, r)

			duration := time.Since(start)
			reqLog.Info("%s %s %d %s", r.Method, r.URL.Path, rw.statusCode, duration)
		})
	}
}
