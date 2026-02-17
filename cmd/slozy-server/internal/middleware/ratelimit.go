package middleware

import (
	"net/http"
	"strings"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

// RateLimiter represents a rate limiter
type RateLimiter struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

// RateLimitConfig represents rate limit configuration
type RateLimitConfig struct {
	// Requests per second
	BurstSizeBPS int
	// Maximum burst size
	BurstSizeMAX int
	// Cleanup interval
	CleanupInterval time.Duration
}

// DefaultRateLimitConfig returns default rate limit configuration
func DefaultRateLimitConfig() *RateLimitConfig {
	return &RateLimitConfig{
		BurstSizeBPS:    5,
		BurstSizeMAX:    10,
		CleanupInterval: time.Minute,
	}
}

// RateLimitMiddleware creates a rate limiting middleware
func RateLimitMiddleware(config *RateLimitConfig) func(http.HandlerFunc) http.HandlerFunc {
	if config == nil {
		config = DefaultRateLimitConfig()
	}

	// Map of IP to rate limiter
	clients := make(map[string]*RateLimiter)
	var mu sync.Mutex

	// Cleanup goroutine
	go func() {
		ticker := time.NewTicker(config.CleanupInterval)
		defer ticker.Stop()

		for range ticker.C {
			mu.Lock()
			for ip, client := range clients {
				if time.Since(client.lastSeen) > 5*time.Minute {
					delete(clients, ip)
				}
			}
			mu.Unlock()
		}
	}()

	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			// Get client IP
			ip := getClientIP(r)

			// Create or get rate limiter for this IP
			mu.Lock()
			if _, exists := clients[ip]; !exists {
				clients[ip] = &RateLimiter{
					limiter: rate.NewLimiter(rate.Limit(config.BurstSizeBPS), config.BurstSizeMAX),
				}
			}
			clients[ip].lastSeen = time.Now()

			if !clients[ip].limiter.Allow() {
				mu.Unlock()
				http.Error(w, "Rate limit exceeded. Please try again later.", http.StatusTooManyRequests)
				return
			}

			mu.Unlock()
			next(w, r)
		}
	}
}

// getClientIP extracts the real client IP from request
func getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header (behind proxy)
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		// X-Forwarded-For can contain multiple IPs, take the first one
		ips := strings.Split(xff, ",")
		if len(ips) > 0 {
			ip := strings.TrimSpace(ips[0])
			if ip != "" {
				return ip
			}
		}
	}

	// Check X-Real-IP header
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		ip := strings.TrimSpace(xri)
		if ip != "" {
			return ip
		}
	}

	// Fall back to RemoteAddr
	ip := r.RemoteAddr
	if idx := strings.LastIndex(ip, ":"); idx != -1 {
		ip = ip[:idx]
	}
	return ip
}

// EndpointRateLimitConfig represents endpoint-specific rate limits
type EndpointRateLimitConfig struct {
	Prefix string        // URL prefix
	Config  *RateLimitConfig
}

// EndpointRateLimitMiddleware creates endpoint-specific rate limiting
func EndpointRateLimitMiddleware(endpointConfigs []EndpointRateLimitConfig, defaultConfig *RateLimitConfig) func(http.HandlerFunc) http.HandlerFunc {
	if defaultConfig == nil {
		defaultConfig = DefaultRateLimitConfig()
	}

	// Create endpoint-specific limiters
	endpointLimiters := make(map[string]func(http.HandlerFunc) http.HandlerFunc)
	for _, conf := range endpointConfigs {
		endpointLimiters[conf.Prefix] = RateLimitMiddleware(conf.Config)
	}
	defaultLimiter := RateLimitMiddleware(defaultConfig)

	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			// Find matching endpoint config
			limiter := defaultLimiter
			for prefix, l := range endpointLimiters {
				if strings.HasPrefix(r.URL.Path, prefix) {
					limiter = l
					break
				}
			}

			limited := limiter(next)
			limited(w, r)
		}
	}
}