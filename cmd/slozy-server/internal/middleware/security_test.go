package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSecurityHeaders(t *testing.T) {
	handler := SecurityHeaders(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler(w, req)

	resp := w.Result()

	expectedHeaders := map[string]string{
		"X-Content-Type-Options":  "nosniff",
		"X-Frame-Options":         "DENY",
		"X-XSS-Protection":        "1; mode=block",
		"Referrer-Policy":         "strict-origin-when-cross-origin",
		"Content-Security-Policy": "default-src 'self' 'unsafe-inline' 'unsafe-eval'; img-src 'self' data: https:; font-src 'self' data:; style-src 'self' 'unsafe-inline' https:; script-src 'self' 'unsafe-inline' https:; connect-src 'self'",
	}

	for header, expected := range expectedHeaders {
		if got := resp.Header.Get(header); got != expected {
			t.Errorf("SecurityHeaders: Expected header %s to be %q, got %q", header, expected, got)
		}
	}
}

func TestCORS(t *testing.T) {
	config := &CORSConfig{
		AllowedOrigins: []string{"http://localhost:3000"},
		AllowedMethods: []string{"GET", "POST"},
		AllowedHeaders: []string{"Content-Type"},
	}

	handler := CORS(config)(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	tests := []struct {
		name           string
		origin         string
		expectOrigin   string
		method         string
		expectMethods  string
		isPreflight    bool
		expectedStatus int
	}{
		{
			name:           "Allowed origin",
			origin:         "http://localhost:3000",
			expectOrigin:   "http://localhost:3000",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Preflight request",
			origin:         "http://localhost:3000",
			method:         "OPTIONS",
			expectOrigin:   "http://localhost:3000",
			expectMethods:  "GET, POST",
			isPreflight:    true,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Disallowed origin",
			origin:         "http://evil.com",
			expectOrigin:   "",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Development localhost",
			origin:         "http://localhost:8080",
			expectOrigin:   "http://localhost:8080",
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, "/", nil)
			req.Header.Set("Origin", tt.origin)
			if tt.method == "" {
				req.Method = "GET"
			}

			w := httptest.NewRecorder()
			handler(w, req)

			resp := w.Result()

			if resp.StatusCode != tt.expectedStatus {
				t.Errorf("CORS: Expected status %d, got %d", tt.expectedStatus, resp.StatusCode)
			}

			if got := resp.Header.Get("Access-Control-Allow-Origin"); got != tt.expectOrigin {
				t.Errorf("CORS: Expected origin %q, got %q", tt.expectOrigin, got)
			}

			if tt.expectMethods != "" {
				if got := resp.Header.Get("Access-Control-Allow-Methods"); got != tt.expectMethods {
					t.Errorf("CORS: Expected methods %q, got %q", tt.expectMethods, got)
				}
			}
		})
	}
}