package middleware

import (
	"net/http"
	"time"

	"github.com/philyuchkoff/slozy-oss/cmd/slozy-server/internal/logger"
)

// responseWriter is a wrapper around http.ResponseWriter that captures status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

// WriteHeader captures the status code before writing header
func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// LoggingMiddleware logs HTTP requests with timing
func LoggingMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		logger.Info("Request", "method", r.Method, "path", r.URL.Path, "remote", r.RemoteAddr)

		rw := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}
		next(rw, r)

		logger.Info("Response",
			"method", r.Method,
			"path", r.URL.Path,
			"remote", r.RemoteAddr,
			"status", rw.statusCode,
			"duration", time.Since(start).String(),
		)
	}
}