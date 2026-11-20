package main

import (
	"net/http"
	"os"
	"strings"

	"github.com/philyuchkoff/slozy-oss/cmd/slozy-server/internal/handlers"
	"github.com/philyuchkoff/slozy-oss/cmd/slozy-server/internal/models"
)

// compose applies middleware chain to a handler
func compose(handler http.HandlerFunc, middlewares ...func(http.HandlerFunc) http.HandlerFunc) http.HandlerFunc {
	for _, mw := range middlewares {
		handler = mw(handler)
	}
	return handler
}

// setupRoutes configures all HTTP routes and returns a configured ServeMux
func setupRoutes(
	cfg *models.Config,
	sloHandler *handlers.SLOHandler,
	templateHandler *handlers.TemplateHandler,
	securityHeaders func(http.HandlerFunc) http.HandlerFunc,
	corsMiddleware func(http.HandlerFunc) http.HandlerFunc,
	rateLimitMiddleware func(http.HandlerFunc) http.HandlerFunc,
	requestSizeLimit func(http.HandlerFunc) http.HandlerFunc,
	contentTypeValidation func(http.HandlerFunc) http.HandlerFunc,
	loggingMiddleware func(http.HandlerFunc) http.HandlerFunc,
) *http.ServeMux {
	mux := http.NewServeMux()

	// Root handler for serving index.html
	mux.Handle("/", compose(
		func(w http.ResponseWriter, r *http.Request) {
			if !strings.HasPrefix(r.URL.Path, "/api/") && r.URL.Path != "/data/" {
				http.ServeFile(w, r, cfg.StaticDirectory+"/index.html")
				return
			}
			http.NotFound(w, r)
		},
		securityHeaders,
		loggingMiddleware,
	))

	// Setup static file server
	staticFS := http.FileServer(http.Dir(cfg.StaticDirectory))
	mux.Handle("/static/", http.StripPrefix("/static/", compose(
		func(w http.ResponseWriter, r *http.Request) {
			staticFS.ServeHTTP(w, r)
		},
		securityHeaders,
	)))

	// Setup secure data file server - only allows .yaml files
	mux.HandleFunc("/data/", compose(
		func(w http.ResponseWriter, r *http.Request) {
			// Only allow GET method
			if r.Method != http.MethodGet {
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
				return
			}

			// Get the file path from the URL
			filePath := strings.TrimPrefix(r.URL.Path, "/data/")

			// Validate the path
			if strings.Contains(filePath, "..") || strings.Contains(filePath, "\\") {
				http.Error(w, "Invalid file path", http.StatusBadRequest)
				return
			}

			// Ensure it has .yaml extension
			if !strings.HasSuffix(filePath, ".yaml") {
				http.Error(w, "Only .yaml files are allowed", http.StatusBadRequest)
				return
			}

			// Construct full file path avoiding directory traversal
			fullPath := cfg.DataDirectory + "/" + filePath

			// Check if file exists
			if _, err := os.Stat(fullPath); os.IsNotExist(err) {
				http.Error(w, "File not found", http.StatusNotFound)
				return
			}

			// Serve the file
			http.ServeFile(w, r, fullPath)
		},
		securityHeaders,
		loggingMiddleware,
	))

	// API routes with comprehensive middleware
	mux.Handle("/api/slo/create", compose(
		sloHandler.CreateSLO,
		securityHeaders,
		corsMiddleware,
		rateLimitMiddleware,
		requestSizeLimit,
		contentTypeValidation,
		loggingMiddleware,
	))

	mux.Handle("/api/slos", compose(
		sloHandler.ListSLOs,
		securityHeaders,
		corsMiddleware,
		rateLimitMiddleware,
		loggingMiddleware,
	))

	mux.Handle("/api/templates", compose(
		templateHandler.GetTemplates,
		securityHeaders,
		corsMiddleware,
		rateLimitMiddleware,
		loggingMiddleware,
	))

	mux.Handle("/api/templates/create-defaults", compose(
		templateHandler.CreateDefaultTemplates,
		securityHeaders,
		corsMiddleware,
		rateLimitMiddleware,
		loggingMiddleware,
	))

	mux.Handle("/health", compose(
		healthHandler,
		securityHeaders,
		loggingMiddleware,
	))

	return mux
}
