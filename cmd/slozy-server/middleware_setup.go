package main

import (
	"net/http"
	"time"

	"github.com/philyuchkoff/slozy-oss/cmd/slozy-server/internal/middleware"
)

// setupMiddleware creates all middleware configurations for the server
func setupMiddleware() (
	securityHeaders func(http.HandlerFunc) http.HandlerFunc,
	corsMiddleware func(http.HandlerFunc) http.HandlerFunc,
	rateLimitMiddleware func(http.HandlerFunc) http.HandlerFunc,
	requestSizeLimit func(http.HandlerFunc) http.HandlerFunc,
	contentTypeValidation func(http.HandlerFunc) http.HandlerFunc,
	loggingMiddleware func(http.HandlerFunc) http.HandlerFunc,
) {
	securityHeaders = middleware.SecurityHeaders

	corsConfig := middleware.DefaultCORSConfig()
	corsMiddleware = middleware.CORS(corsConfig)

	// Rate limiting configurations
	apiRateLimit := &middleware.RateLimitConfig{
		BurstSizeBPS:    10, // 10 requests per second
		BurstSizeMAX:    20, // Burst of 20 requests
		CleanupInterval: time.Minute,
	}

	createRateLimit := &middleware.RateLimitConfig{
		BurstSizeBPS:    5,  // 5 SLO creates per second
		BurstSizeMAX:    10, // Burst of 10 requests
		CleanupInterval: time.Minute,
	}

	endpointConfigs := []middleware.EndpointRateLimitConfig{
		{Prefix: "/api/slo/create", Config: createRateLimit},
	}
	rateLimitMiddleware = middleware.EndpointRateLimitMiddleware(endpointConfigs, apiRateLimit)

	// Request size validation (1MB max)
	requestSizeLimit = middleware.MaxRequestSize(1024 * 1024)

	// Content type validation for API endpoints
	contentTypeValidation = middleware.ValidateContentType("application/json", "multipart/form-data")

	loggingMiddleware = middleware.LoggingMiddleware

	return
}
