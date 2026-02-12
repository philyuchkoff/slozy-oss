package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
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

	// Setup static file server
	staticFS := http.FileServer(http.Dir(cfg.StaticDirectory))
	http.Handle("/static/", http.StripPrefix("/static/", staticFS))

	// Setup data file server
	dataFS := http.FileServer(http.Dir(cfg.DataDirectory))
	http.Handle("/data/", http.StripPrefix("/data/", dataFS))

	// Setup routes with logging middleware
	loggingMiddleware := middleware.LoggingMiddleware

	// Root handler for serving index.html
	http.HandleFunc("/", loggingMiddleware(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasPrefix(r.URL.Path, "/api/") && r.URL.Path != "/data/" {
			http.ServeFile(w, r, cfg.StaticDirectory+"/index.html")
			return
		}
		http.NotFound(w, r)
	}))

	// API routes
	http.HandleFunc("/api/slo/create", loggingMiddleware(sloHandler.CreateSLO))
	http.HandleFunc("/api/slos", loggingMiddleware(sloHandler.ListSLOs))
	http.HandleFunc("/api/templates", loggingMiddleware(templateHandler.GetTemplates))
	http.HandleFunc("/api/templates/create-defaults", loggingMiddleware(templateHandler.CreateDefaultTemplates))
	http.HandleFunc("/health", loggingMiddleware(healthHandler))

	// Print startup information
	printServerInfo(cfg)

	// Start server
	if err := http.ListenAndServe(fmt.Sprintf(":%d", cfg.Port), nil); err != nil {
		log.Fatal("Server failed:", err)
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