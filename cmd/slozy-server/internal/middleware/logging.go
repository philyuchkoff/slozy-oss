package middleware

import (
	"log"
	"net/http"
	"time"
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
		log.Printf("[%s] %s %s", r.Method, r.URL.Path, r.RemoteAddr)

		rw := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}
		next(rw, r)

		log.Printf("[%s] %s %s - %d %s",
			r.Method, r.URL.Path, r.RemoteAddr,
			rw.statusCode, time.Since(start))
	}
}