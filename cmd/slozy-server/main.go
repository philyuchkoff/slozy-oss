package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
<<<<<<< HEAD
	"path/filepath"
=======
	"os/signal"
>>>>>>> e090066 (feat: graceful shutdown implemented)
	"strings"
	"syscall"
	"time"

	"github.com/philyuchkoff/slozy/cmd/slozy-server/internal/config"
	"github.com/philyuchkoff/slozy/cmd/slozy-server/internal/handlers"
	"github.com/philyuchkoff/slozy/cmd/slozy-server/internal/models"
	"github.com/philyuchkoff/slozy/cmd/slozy-server/internal/middleware"
)

func main() {
	// Load configuration
	cfg := config.LoadConfig()

	// Create directories
	if err := os.MkdirAll(cfg.DataDirectory, 0755); err != nil {
		log.Fatal("Failed to create data directory:", err)
	}
	if err := os.MkdirAll(cfg.TemplatesDirectory, 0755); err != nil {
		log.Fatal("Failed to create templates directory:", err)
	}
	if err := os.MkdirAll(cfg.StaticDirectory, 0755); err != nil {
		log.Fatal("Failed to create static directory:", err)
	}

	// Initialize handlers
	sloHandler := handlers.NewSLOHandler(cfg.DataDirectory)
	templateHandler := handlers.NewTemplateHandler(cfg.TemplatesDirectory)

<<<<<<< HEAD
	// Setup static file server
	staticFS := http.FileServer(http.Dir(cfg.StaticDirectory))
	http.Handle("/static/", http.StripPrefix("/static/", staticFS))

	// Setup secure data file server - only allows .yaml files
	http.HandleFunc("/data/", middleware.LoggingMiddleware(func(w http.ResponseWriter, r *http.Request) {
		// Only allow GET method
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Get the file path from the URL
		filePath := strings.TrimPrefix(r.URL.Path, "/data/")
		filePath = filepath.Clean(filePath)

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

		// Construct full file path
		fullPath := filepath.Join(cfg.DataDirectory, filePath)

		// Ensure the file is within the data directory
		if !strings.HasPrefix(fullPath, cfg.DataDirectory) {
			http.Error(w, "Access denied", http.StatusForbidden)
			return
		}

		// Check if file exists and is a regular file
	 fileInfo, err := os.Stat(fullPath)
		if err != nil {
			if os.IsNotExist(err) {
				http.Error(w, "File not found", http.StatusNotFound)
			} else {
				http.Error(w, "Internal server error", http.StatusInternalServerError)
			}
			return
		}

		if fileInfo.IsDir() {
			http.Error(w, "Directory access not allowed", http.StatusBadRequest)
			return
		}

		// Serve the file
		http.ServeFile(w, r, fullPath)
	}))

	// Setup security middleware
	securityHeaders := middleware.SecurityHeaders
	corsConfig := middleware.DefaultCORSConfig()
	corsMiddleware := middleware.CORS(corsConfig)

	// Rate limiting configurations
	apiRateLimit := &middleware.RateLimitConfig{
		BurstSizeBPS:    10,  // 10 requests per second
		BurstSizeMAX:    20,  // Burst of 20 requests
		CleanupInterval: time.Minute,
	}

	createRateLimit := &middleware.RateLimitConfig{
		BurstSizeBPS:    5,   // 5 SLO creates per second
		BurstSizeMAX:    10,  // Burst of 10 requests
		CleanupInterval: time.Minute,
	}

	endpointConfigs := []middleware.EndpointRateLimitConfig{
		{Prefix: "/api/slo/create", Config: createRateLimit},
	}
	rateLimitMiddleware := middleware.EndpointRateLimitMiddleware(endpointConfigs, apiRateLimit)

	// Request size validation (1MB max)
	requestSizeLimit := middleware.MaxRequestSize(1024 * 1024)

	// Content type validation for API endpoints
	contentTypeValidation := middleware.ValidateContentType("application/json", "multipart/form-data")

	// Apply middleware chain to routes
	loggingMiddleware := middleware.LoggingMiddleware

	// Helper function to compose middleware
	compose := func(handler http.HandlerFunc, middlewares ...func(http.HandlerFunc) http.HandlerFunc) http.HandlerFunc {
		for _, mw := range middlewares {
			handler = mw(handler)
		}
		return handler
	}

	// Root handler for serving index.html
	http.Handle("/", compose(
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

	// API routes
	http.Handle("/api/slo/create", compose(
		sloHandler.CreateSLO,
		securityHeaders,
		corsMiddleware,
		rateLimitMiddleware,
		requestSizeLimit,
		contentTypeValidation,
		loggingMiddleware,
	))

	http.Handle("/api/slos", compose(
		sloHandler.ListSLOs,
		securityHeaders,
		corsMiddleware,
		rateLimitMiddleware,
		loggingMiddleware,
	))

	http.Handle("/api/templates", compose(
		templateHandler.GetTemplates,
		securityHeaders,
		corsMiddleware,
		rateLimitMiddleware,
		loggingMiddleware,
	))

	http.Handle("/api/templates/create-defaults", compose(
		templateHandler.CreateDefaultTemplates,
		securityHeaders,
		corsMiddleware,
		rateLimitMiddleware,
		loggingMiddleware,
	))

	http.Handle("/health", compose(
		healthHandler,
		securityHeaders,
		loggingMiddleware,
	))
=======
	// Create router with logging middleware
	mux := http.NewServeMux()
	loggingMiddleware := middleware.LoggingMiddleware

	// Root handler for serving index.html
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasPrefix(r.URL.Path, "/api/") && r.URL.Path != "/data/" {
			http.ServeFile(w, r, cfg.StaticDirectory+"/index.html")
			return
		}
		http.NotFound(w, r)
	})

	// API routes with logging middleware applied
	mux.HandleFunc("/api/slo/create", loggingMiddleware(sloHandler.CreateSLO))
	mux.HandleFunc("/api/slos", loggingMiddleware(sloHandler.ListSLOs))
	mux.HandleFunc("/api/templates", loggingMiddleware(templateHandler.GetTemplates))
	mux.HandleFunc("/api/templates/create-defaults", loggingMiddleware(templateHandler.CreateDefaultTemplates))
	mux.HandleFunc("/health", loggingMiddleware(healthHandler))

	// Setup static file server
	staticFS := http.FileServer(http.Dir(cfg.StaticDirectory))
	mux.Handle("/static/", http.StripPrefix("/static/", staticFS))

	// Setup data file server
	dataFS := http.FileServer(http.Dir(cfg.DataDirectory))
	mux.Handle("/data/", http.StripPrefix("/data/", dataFS))

	// Create server with timeouts
	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Port),
		Handler:      mux,
		ReadTimeout:  15 * time.Second,  // Maximum time to read request
		WriteTimeout: 15 * time.Second,  // Maximum time to write response
		IdleTimeout:  60 * time.Second,  // Maximum time to wait for next request
	}
>>>>>>> e090066 (feat: graceful shutdown implemented)

	// Print startup information
	printServerInfo(cfg)

	// Start server in goroutine
	go func() {
		log.Printf("Server starting on %s", server.Addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("Server failed:", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	} else {
		log.Println("Server gracefully shutdown")
	}
}

// printServerInfo prints server startup information
func printServerInfo(cfg *models.Config) {
	fmt.Printf("🚀 Starting SLOzy with full alerting support on http://localhost:%d\n", cfg.Port)
	fmt.Printf("📁 Data directory: %s/\n", cfg.DataDirectory)
	fmt.Printf("📁 Templates directory: %s/\n", cfg.TemplatesDirectory)
	fmt.Printf("📁 Static directory: %s/\n", cfg.StaticDirectory)

	// Telegram status
	if cfg.TelegramBotToken != "" && cfg.TelegramChatID != "" {
		tokenPreview := cfg.TelegramBotToken
		if len(tokenPreview) > 10 {
			tokenPreview = tokenPreview[:10] + "..."
		}
		fmt.Printf("🤖 Telegram bot configured\n")
		fmt.Printf("   • Bot token: %s\n", tokenPreview)
		fmt.Printf("   • Chat ID: %s\n", cfg.TelegramChatID)
	} else {
		fmt.Printf("⚠️  Telegram bot not configured (set TELEGRAM_BOT_TOKEN and TELEGRAM_CHAT_ID env vars)\n")
		fmt.Printf("   To enable Telegram alerts, run:\n")
		fmt.Printf("   export TELEGRAM_BOT_TOKEN='your_bot_token'\n")
		fmt.Printf("   export TELEGRAM_CHAT_ID='your_chat_id'\n")
	}

	fmt.Printf("🔔 Alerting features:\n")
	fmt.Printf("   • OpenSLO Alert Policies\n")
	fmt.Printf("   • Prometheus Alert Rules\n")
	fmt.Printf("   • AlertManager Config with Telegram support\n")
	fmt.Printf("   • Grafana Dashboards\n")
	fmt.Printf("   • README with instructions\n")
	fmt.Printf("📦 All files are now downloaded as ZIP archive\n")
	fmt.Printf("📋 Templates loaded from %s/\n", cfg.TemplatesDirectory)
	fmt.Println("\nPress Ctrl+C to stop")
}

// healthHandler returns server health status
func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(models.HealthResponse{
		Status: "healthy",
		Time:   time.Now(),
	})
}
