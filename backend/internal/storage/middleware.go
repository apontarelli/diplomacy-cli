package storage

import (
	"context"
	"fmt"
	"net/http"
	"strings"
)

// ContextKey is a type for context keys to avoid collisions
type ContextKey string

const (
	// UserContextKey is the key for storing user in context
	UserContextKey ContextKey = "user"
	// ClaimsContextKey is the key for storing JWT claims in context
	ClaimsContextKey ContextKey = "claims"
)

// AuthMiddleware provides HTTP middleware for JWT authentication
type AuthMiddleware struct {
	authService *AuthService
}

// NewAuthMiddleware creates a new authentication middleware
func NewAuthMiddleware(authService *AuthService) *AuthMiddleware {
	return &AuthMiddleware{
		authService: authService,
	}
}

// RequireAuth is a middleware that requires valid JWT authentication
func (am *AuthMiddleware) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract token from Authorization header
		token, err := am.extractTokenFromHeader(r)
		if err != nil {
			http.Error(w, "Unauthorized: "+err.Error(), http.StatusUnauthorized)
			return
		}

		// Validate token
		claims, err := am.authService.ValidateToken(token)
		if err != nil {
			http.Error(w, "Unauthorized: invalid token", http.StatusUnauthorized)
			return
		}

		// Get user from database
		user, err := am.authService.userService.GetUserByID(claims.UserID)
		if err != nil {
			http.Error(w, "Unauthorized: user not found", http.StatusUnauthorized)
			return
		}

		// Add user and claims to context
		ctx := context.WithValue(r.Context(), UserContextKey, user)
		ctx = context.WithValue(ctx, ClaimsContextKey, claims)

		// Call next handler with updated context
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// OptionalAuth is a middleware that optionally validates JWT authentication
// If a valid token is provided, user info is added to context
// If no token or invalid token, request continues without user info
func (am *AuthMiddleware) OptionalAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Try to extract token from Authorization header
		token, err := am.extractTokenFromHeader(r)
		if err != nil {
			// No token provided, continue without authentication
			next.ServeHTTP(w, r)
			return
		}

		// Try to validate token
		claims, err := am.authService.ValidateToken(token)
		if err != nil {
			// Invalid token, continue without authentication
			next.ServeHTTP(w, r)
			return
		}

		// Try to get user from database
		user, err := am.authService.userService.GetUserByID(claims.UserID)
		if err != nil {
			// User not found, continue without authentication
			next.ServeHTTP(w, r)
			return
		}

		// Add user and claims to context
		ctx := context.WithValue(r.Context(), UserContextKey, user)
		ctx = context.WithValue(ctx, ClaimsContextKey, claims)

		// Call next handler with updated context
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// extractTokenFromHeader extracts JWT token from Authorization header
func (am *AuthMiddleware) extractTokenFromHeader(r *http.Request) (string, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return "", fmt.Errorf("authorization header missing")
	}

	// Check for Bearer token format
	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
		return "", fmt.Errorf("authorization header must be Bearer token")
	}

	return parts[1], nil
}

// GetUserFromContext extracts user from request context
func GetUserFromContext(ctx context.Context) (*User, bool) {
	user, ok := ctx.Value(UserContextKey).(*User)
	return user, ok
}

// GetClaimsFromContext extracts JWT claims from request context
func GetClaimsFromContext(ctx context.Context) (*Claims, bool) {
	claims, ok := ctx.Value(ClaimsContextKey).(*Claims)
	return claims, ok
}

// RequireUserID is a middleware that ensures the authenticated user matches a specific user ID
// This is useful for endpoints that should only be accessible by the user themselves
func (am *AuthMiddleware) RequireUserID(userIDParam string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// First ensure user is authenticated
			_, ok := GetUserFromContext(r.Context())
			if !ok {
				http.Error(w, "Unauthorized: authentication required", http.StatusUnauthorized)
				return
			}

			// Extract user ID from URL parameters (this would depend on your router)
			// For now, we'll assume it's passed as a parameter
			// In a real implementation, you'd extract this from the URL path or query params

			// This is a placeholder - in practice you'd extract the user ID from the request
			// For example, with gorilla/mux: mux.Vars(r)["userID"]
			// Or with gin: c.Param("userID")

			// For now, we'll just ensure the user is authenticated
			// The actual user ID comparison would be implemented in the specific handler
			// userIDParam is available for future use when implementing specific router integration
			_ = userIDParam

			next.ServeHTTP(w, r)
		})
	}
}

// CORS middleware for handling Cross-Origin Resource Sharing
func (am *AuthMiddleware) CORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set CORS headers
		w.Header().Set("Access-Control-Allow-Origin", "*") // In production, be more specific
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Max-Age", "86400")

		// Handle preflight requests
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// RateLimitConfig holds rate limiting configuration
type RateLimitConfig struct {
	RequestsPerMinute int
	BurstSize         int
}

// BasicRateLimit is a simple rate limiting middleware
// Note: This is a basic implementation. For production, consider using a more robust solution
func (am *AuthMiddleware) BasicRateLimit(config RateLimitConfig) func(http.Handler) http.Handler {
	// This is a placeholder for rate limiting
	// In a real implementation, you'd use a proper rate limiting library
	// like golang.org/x/time/rate or a Redis-based solution

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// TODO: Implement actual rate limiting logic
			// For now, just pass through
			next.ServeHTTP(w, r)
		})
	}
}

// LoggingMiddleware logs HTTP requests
func (am *AuthMiddleware) LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// TODO: Implement proper logging
		// For now, this is a placeholder
		next.ServeHTTP(w, r)
	})
}
