package api

import (
	"net/http"

	"diplomacy-cli/backend/internal/api/handlers"
	"diplomacy-cli/backend/internal/api/middleware"
	"diplomacy-cli/backend/internal/api/websocket"
	"diplomacy-cli/backend/internal/storage"
)

// Server represents the HTTP server
type Server struct {
	db          *storage.Database
	authService *storage.AuthService
	rateLimiter *middleware.RateLimiter
	wsHub       *websocket.Hub
	wsHandler   *websocket.Handler
}

// NewServer creates a new API server
func NewServer(db *storage.Database, authService *storage.AuthService) *Server {
	// Create WebSocket hub and handler
	wsHub := websocket.NewHub()
	wsHandler := websocket.NewHandler(wsHub, authService)

	// Start the hub in a goroutine
	go wsHub.Run()

	return &Server{
		db:          db,
		authService: authService,
		rateLimiter: middleware.NewRateLimiter(),
		wsHub:       wsHub,
		wsHandler:   wsHandler,
	}
}

// Handler returns the HTTP handler for the server
func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()

	// Initialize handlers
	gameHandler := handlers.NewGameHandler(s.db, s.wsHandler)
	orderHandler := handlers.NewOrderHandler(s.db, s.wsHandler)
	authHandler := handlers.NewAuthHandler(s.authService)
	templateHandler := handlers.NewTemplateHandler(s.db)

	// Initialize middleware
	authMiddleware := middleware.NewAuthMiddleware(s.authService)

	// Serve static files from frontend dist
	fileServer := http.FileServer(http.Dir("../frontend/dist/"))
	mux.Handle("/assets/", http.StripPrefix("/", fileServer))
	mux.Handle("/js/", http.StripPrefix("/", fileServer))
	mux.Handle("/vite.svg", http.StripPrefix("/", fileServer))

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

	// WebSocket endpoints
	mux.HandleFunc("GET /ws/game/{id}", s.wsHandler.HandleGameWebSocket)

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

	// HTML/HTMX endpoints
	mux.HandleFunc("GET /games", templateHandler.GameListPage)
	mux.HandleFunc("GET /games/{id}", templateHandler.GamePage)
	mux.HandleFunc("GET /games/{id}/board", templateHandler.GameBoardPartial)
	mux.HandleFunc("GET /games/{id}/orders", templateHandler.OrderFormPartial)
	mux.HandleFunc("GET /province-details", templateHandler.ProvinceDetailsPartial)

	// Root redirect to games
	mux.HandleFunc("GET /{$}", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/games", http.StatusFound)
	})

	return mux
}

// GetWebSocketHandler returns the WebSocket handler for broadcasting
func (s *Server) GetWebSocketHandler() *websocket.Handler {
	return s.wsHandler
}

// handleHealth provides a simple health check endpoint
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"ok"}`))
}
