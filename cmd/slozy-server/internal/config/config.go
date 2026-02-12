package config

import (
	"os"
	"strconv"

	"github.com/philyuchkoff/slozy/cmd/slozy-server/internal/models"
)

// LoadConfig loads configuration from environment variables and defaults
func LoadConfig() *models.Config {
	config := &models.Config{
		Port:               8080,
		DataDirectory:      "./data",
		TemplatesDirectory:  "./templates",
		StaticDirectory:     "./static",
		TelegramBotToken:    os.Getenv("TELEGRAM_BOT_TOKEN"),
		TelegramChatID:      os.Getenv("TELEGRAM_CHAT_ID"),
		SMTPHost:           getEnvDefault("SMTP_HOST", "localhost:25"),
		SMTPFrom:           getEnvDefault("SMTP_FROM", "slo-alerts@example.com"),
		GrafanaURL:         getEnvDefault("GRAFANA_URL", "https://grafana.example.com"),
	}

	// Override with environment variables if present
	if port := os.Getenv("PORT"); port != "" {
		if p, err := strconv.Atoi(port); err == nil {
			config.Port = p
		}
	}

	if dataDir := os.Getenv("DATA_DIR"); dataDir != "" {
		config.DataDirectory = dataDir
	}

	if templatesDir := os.Getenv("TEMPLATES_DIR"); templatesDir != "" {
		config.TemplatesDirectory = templatesDir
	}

	if staticDir := os.Getenv("STATIC_DIR"); staticDir != "" {
		config.StaticDirectory = staticDir
	}

	return config
}

// getEnvDefault gets environment variable with default value
func getEnvDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}