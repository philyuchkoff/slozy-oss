package middleware

import (
	"bytes"
	"io"
	"net/http"
	"strings"
)

// MaxRequestSize wraps a handler to validate request size
func MaxRequestSize(maxBytes int64) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			// Only validate POST, PUT, PATCH requests with a body
			if r.Method == "GET" || r.Method == "DELETE" || r.Method == "OPTIONS" || r.Method == "HEAD" {
				next(w, r)
				return
			}

			// If content length is available and exceeds limit
			if r.ContentLength > maxBytes {
				http.Error(w, "Request entity too large", http.StatusRequestEntityTooLarge)
				return
			}

			// If no content length or if it's within bounds, read the body
			if r.Body != nil {
				r.Body = http.MaxBytesReader(w, r.Body, maxBytes)

				// Try to read a small portion to detect if body is too large
				buf := make([]byte, 4096)
				n, err := r.Body.Read(buf)

				if err != nil && err != io.EOF {
					// Check if it's a MaxBytesError
					if err.Error() == "http: request body too large" ||
					   err.Error() == "http: request body too large" {
						http.Error(w, "Request entity too large", http.StatusRequestEntityTooLarge)
						return
					}
					// Other errors are handled by the normal flow
				}

				// Put the bytes back
				if n > 0 {
					r.Body = io.NopCloser(io.MultiReader(bytes.NewReader(buf[:n]), r.Body))
				}
			}

			next(w, r)
		}
	}
}

// ValidateContentType validates the content type
func ValidateContentType(allowedTypes ...string) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			contentType := r.Header.Get("Content-Type")
			if contentType != "" {
				// Check if the content type starts with any of the allowed types
				matched := false
				for _, allowed := range allowedTypes {
					if contentType == allowed ||
					   (strings.HasSuffix(allowed, "*") && strings.HasPrefix(contentType, allowed[:len(allowed)-1])) {
						matched = true
						break
					}
				}

				if !matched {
					http.Error(w, "Unsupported media type", http.StatusUnsupportedMediaType)
					return
				}
			}

			next(w, r)
		}
	}
}