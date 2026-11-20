package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/philyuchkoff/slozy-oss/cmd/slozy-server/internal/config"
	"github.com/philyuchkoff/slozy-oss/cmd/slozy-server/internal/handlers"
	"github.com/philyuchkoff/slozy-oss/cmd/slozy-server/internal/logger"
)

func main() {
	// Load configuration
	cfg := config.LoadConfig()

	// Create directories
	if err := os.MkdirAll(cfg.DataDirectory, 0755); err != nil {
		logger.Fatal("Failed to create data directory", "error", err)
	}
	if err := os.MkdirAll(cfg.TemplatesDirectory, 0755); err != nil {
		logger.Fatal("Failed to create templates directory", "error", err)
	}
	if err := os.MkdirAll(cfg.StaticDirectory, 0755); err != nil {
		logger.Fatal("Failed to create static directory", "error", err)
	}

	// Initialize handlers
	sloHandler := handlers.NewSLOHandler(cfg.DataDirectory)
	templateHandler := handlers.NewTemplateHandler(cfg.TemplatesDirectory)

	// Setup middleware
	securityHeaders, corsMiddleware, rateLimitMiddleware, requestSizeLimit, contentTypeValidation, loggingMiddleware := setupMiddleware()

	// Setup routes
	mux := setupRoutes(cfg, sloHandler, templateHandler,
		securityHeaders, corsMiddleware, rateLimitMiddleware,
		requestSizeLimit, contentTypeValidation, loggingMiddleware,
	)

	// Create server with timeouts
	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Port),
		Handler:      mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Print startup information
	printServerInfo(cfg)

	// Start server in goroutine
	go func() {
		logger.Info("Server starting", "addr", server.Addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Server failed", "error", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Info("Shutting down server...")

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Error("Server forced to shutdown", "error", err)
	} else {
		logger.Info("Server gracefully shutdown")
	}
}
