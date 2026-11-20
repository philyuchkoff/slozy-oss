package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/philyuchkoff/slozy-oss/cmd/slozy-server/internal/models"
)

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
	fmt.Printf("🛡️  Security features enabled:\n")
	fmt.Printf("   • Rate limiting (5-10 req/s for API)\n")
	fmt.Printf("   • Request size limit (1MB)\n")
	fmt.Printf("   • CORS headers\n")
	fmt.Printf("   • Content type validation\n")
	fmt.Printf("   • Directory traversal protection\n")
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
