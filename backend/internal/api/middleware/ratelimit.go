package middleware

import (
	"fmt"
	"net/http"
	"sync"
	"time"
)

// RateLimiter tracks request counts for rate limiting
type RateLimiter struct {
	mu       sync.RWMutex
	requests map[string]*userRequests
	cleanup  *time.Ticker
}

// userRequests tracks requests for a specific user
type userRequests struct {
	count     int
	resetTime time.Time
}

// RateLimitConfig defines rate limiting configuration
type RateLimitConfig struct {
	RequestsPerMinute int
	WindowDuration    time.Duration
}

// DefaultRateLimitConfigs provides default rate limits for different endpoints
var DefaultRateLimitConfigs = map[string]RateLimitConfig{
	"POST /api/games":             {RequestsPerMinute: 5, WindowDuration: time.Minute},
	"POST /api/games/{id}/join":   {RequestsPerMinute: 10, WindowDuration: time.Minute},
	"POST /api/games/{id}/orders": {RequestsPerMinute: 30, WindowDuration: time.Minute},
	"POST /api/auth/login":        {RequestsPerMinute: 5, WindowDuration: time.Minute},
	"POST /api/auth/register":     {RequestsPerMinute: 3, WindowDuration: time.Minute},
	"POST /api/auth/refresh":      {RequestsPerMinute: 10, WindowDuration: time.Minute},
	"default":                     {RequestsPerMinute: 60, WindowDuration: time.Minute},
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter() *RateLimiter {
	rl := &RateLimiter{
		requests: make(map[string]*userRequests),
		cleanup:  time.NewTicker(5 * time.Minute), // Cleanup every 5 minutes
	}

	// Start cleanup goroutine
	go rl.cleanupExpiredEntries()

	return rl
}

// cleanupExpiredEntries removes expired rate limit entries
func (rl *RateLimiter) cleanupExpiredEntries() {
	for range rl.cleanup.C {
		rl.mu.Lock()
		now := time.Now()
		for key, req := range rl.requests {
			if now.After(req.resetTime) {
				delete(rl.requests, key)
			}
		}
		rl.mu.Unlock()
	}
}

// Stop stops the rate limiter cleanup goroutine
func (rl *RateLimiter) Stop() {
	rl.cleanup.Stop()
}

// isAllowed checks if a request is allowed under the rate limit
func (rl *RateLimiter) isAllowed(userID int64, endpoint string, config RateLimitConfig) (bool, time.Time) {
	key := fmt.Sprintf("%d:%s", userID, endpoint)
	now := time.Now()

	rl.mu.Lock()
	defer rl.mu.Unlock()

	req, exists := rl.requests[key]
	if !exists || now.After(req.resetTime) {
		// First request or window expired, reset
		rl.requests[key] = &userRequests{
			count:     1,
			resetTime: now.Add(config.WindowDuration),
		}
		return true, now.Add(config.WindowDuration)
	}

	// Check if under limit
	if req.count < config.RequestsPerMinute {
		req.count++
		return true, req.resetTime
	}

	// Rate limit exceeded
	return false, req.resetTime
}

// RateLimit creates rate limiting middleware for a specific endpoint
func (rl *RateLimiter) RateLimit(endpoint string) func(http.Handler) http.Handler {
	config, exists := DefaultRateLimitConfigs[endpoint]
	if !exists {
		config = DefaultRateLimitConfigs["default"]
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get user from context (must be after auth middleware)
			user, ok := GetUserFromContext(r.Context())
			if !ok {
				// If no user in context, use IP-based rate limiting
				rl.rateLimitByIP(w, r, next, endpoint, config)
				return
			}

			// Check rate limit
			allowed, resetTime := rl.isAllowed(user.ID, endpoint, config)
			if !allowed {
				// Add rate limit headers
				w.Header().Set("X-RateLimit-Limit", fmt.Sprintf("%d", config.RequestsPerMinute))
				w.Header().Set("X-RateLimit-Remaining", "0")
				w.Header().Set("X-RateLimit-Reset", fmt.Sprintf("%d", resetTime.Unix()))

				// Return rate limit error
				context := map[string]interface{}{
					"limit":    config.RequestsPerMinute,
					"window":   config.WindowDuration.String(),
					"reset_at": resetTime.Format(time.RFC3339),
				}
				err := NewRateLimitError("Rate limit exceeded", "RATE_LIMIT_EXCEEDED", context)
				WriteError(w, err)
				return
			}

			// Add rate limit headers for successful requests
			remaining := config.RequestsPerMinute - rl.getCurrentCount(user.ID, endpoint)
			w.Header().Set("X-RateLimit-Limit", fmt.Sprintf("%d", config.RequestsPerMinute))
			w.Header().Set("X-RateLimit-Remaining", fmt.Sprintf("%d", remaining))
			w.Header().Set("X-RateLimit-Reset", fmt.Sprintf("%d", resetTime.Unix()))

			next.ServeHTTP(w, r)
		})
	}
}

// rateLimitByIP provides IP-based rate limiting for unauthenticated requests
func (rl *RateLimiter) rateLimitByIP(w http.ResponseWriter, r *http.Request, next http.Handler, endpoint string, config RateLimitConfig) {
	// Get client IP
	ip := r.Header.Get("X-Forwarded-For")
	if ip == "" {
		ip = r.Header.Get("X-Real-IP")
	}
	if ip == "" {
		ip = r.RemoteAddr
	}

	key := fmt.Sprintf("ip:%s:%s", ip, endpoint)
	now := time.Now()

	rl.mu.Lock()
	req, exists := rl.requests[key]
	if !exists || now.After(req.resetTime) {
		rl.requests[key] = &userRequests{
			count:     1,
			resetTime: now.Add(config.WindowDuration),
		}
		rl.mu.Unlock()
		next.ServeHTTP(w, r)
		return
	}

	if req.count < config.RequestsPerMinute {
		req.count++
		rl.mu.Unlock()
		next.ServeHTTP(w, r)
		return
	}

	rl.mu.Unlock()

	// Rate limit exceeded
	w.Header().Set("X-RateLimit-Limit", fmt.Sprintf("%d", config.RequestsPerMinute))
	w.Header().Set("X-RateLimit-Remaining", "0")
	w.Header().Set("X-RateLimit-Reset", fmt.Sprintf("%d", req.resetTime.Unix()))

	context := map[string]interface{}{
		"limit":    config.RequestsPerMinute,
		"window":   config.WindowDuration.String(),
		"reset_at": req.resetTime.Format(time.RFC3339),
	}
	err := NewRateLimitError("Rate limit exceeded", "RATE_LIMIT_EXCEEDED", context)
	WriteError(w, err)
}

// getCurrentCount gets the current request count for a user/endpoint
func (rl *RateLimiter) getCurrentCount(userID int64, endpoint string) int {
	key := fmt.Sprintf("%d:%s", userID, endpoint)

	rl.mu.RLock()
	defer rl.mu.RUnlock()

	req, exists := rl.requests[key]
	if !exists {
		return 0
	}

	if time.Now().After(req.resetTime) {
		return 0
	}

	return req.count
}

// GlobalRateLimit creates a global rate limiter middleware
func (rl *RateLimiter) GlobalRateLimit(requestsPerMinute int) func(http.Handler) http.Handler {
	config := RateLimitConfig{
		RequestsPerMinute: requestsPerMinute,
		WindowDuration:    time.Minute,
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			endpoint := "global"

			// Try to get user from context
			if user, ok := GetUserFromContext(r.Context()); ok {
				allowed, resetTime := rl.isAllowed(user.ID, endpoint, config)
				if !allowed {
					w.Header().Set("X-RateLimit-Limit", fmt.Sprintf("%d", config.RequestsPerMinute))
					w.Header().Set("X-RateLimit-Remaining", "0")
					w.Header().Set("X-RateLimit-Reset", fmt.Sprintf("%d", resetTime.Unix()))

					context := map[string]interface{}{
						"limit":    config.RequestsPerMinute,
						"window":   config.WindowDuration.String(),
						"reset_at": resetTime.Format(time.RFC3339),
					}
					err := NewRateLimitError("Global rate limit exceeded", "GLOBAL_RATE_LIMIT_EXCEEDED", context)
					WriteError(w, err)
					return
				}
			} else {
				// Fall back to IP-based limiting
				rl.rateLimitByIP(w, r, next, endpoint, config)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
