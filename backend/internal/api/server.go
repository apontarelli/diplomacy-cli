package api

import (
	"net/http"

	"diplomacy-cli/backend/internal/api/handlers"
	"diplomacy-cli/backend/internal/api/middleware"
	"diplomacy-cli/backend/internal/storage"
)

// Server represents the HTTP server
type Server struct {
	db          *storage.Database
	authService *storage.AuthService
	rateLimiter *middleware.RateLimiter
}

// NewServer creates a new API server
func NewServer(db *storage.Database, authService *storage.AuthService) *Server {
	return &Server{
		db:          db,
		authService: authService,
		rateLimiter: middleware.NewRateLimiter(),
	}
}

// Handler returns the HTTP handler for the server
func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()

	// Initialize handlers
	gameHandler := handlers.NewGameHandler(s.db)
	orderHandler := handlers.NewOrderHandler(s.db)
	authHandler := handlers.NewAuthHandler(s.authService)

	// Initialize middleware
	authMiddleware := middleware.NewAuthMiddleware(s.authService)

	// Health check endpoint (no auth required)
	mux.HandleFunc("GET /health", s.handleHealth)

	// Authentication endpoints (no auth required for login/register)
	mux.Handle("POST /api/auth/login",
		s.rateLimiter.RateLimit("POST /api/auth/login")(
			http.HandlerFunc(authHandler.Login)))

	mux.Handle("POST /api/auth/register",
		s.rateLimiter.RateLimit("POST /api/auth/register")(
			http.HandlerFunc(authHandler.Register)))

	mux.Handle("POST /api/auth/refresh",
		s.rateLimiter.RateLimit("POST /api/auth/refresh")(
			http.HandlerFunc(authHandler.RefreshToken)))

	mux.HandleFunc("GET /api/auth/token-info", authHandler.GetTokenInfo)

	// Protected authentication endpoints (require auth)
	mux.Handle("GET /api/auth/profile",
		authMiddleware.RequireAuth(
			http.HandlerFunc(authHandler.GetProfile)))

	mux.Handle("POST /api/auth/logout",
		authMiddleware.RequireAuth(
			http.HandlerFunc(authHandler.Logout)))

	// Game endpoints (auth required)
	mux.Handle("POST /api/games",
		authMiddleware.RequireAuth(
			s.rateLimiter.RateLimit("POST /api/games")(
				http.HandlerFunc(gameHandler.CreateGame))))

	mux.Handle("POST /api/games/{id}/join",
		authMiddleware.RequireAuth(
			s.rateLimiter.RateLimit("POST /api/games/{id}/join")(
				http.HandlerFunc(gameHandler.JoinGame))))

	mux.Handle("POST /api/games/{id}/orders",
		authMiddleware.RequireAuth(
			s.rateLimiter.RateLimit("POST /api/games/{id}/orders")(
				http.HandlerFunc(orderHandler.SubmitOrders))))

	return mux
}

// handleHealth provides a simple health check endpoint
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"ok"}`))
}
