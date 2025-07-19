package websocket

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/websocket"

	"diplomacy-cli/backend/internal/storage"
)

// Upgrader configures the websocket upgrader
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// In production, implement proper origin checking
		return true
	},
}

// Handler manages WebSocket connections
type Handler struct {
	hub         *Hub
	authService *storage.AuthService
}

// NewHandler creates a new WebSocket handler
func NewHandler(hub *Hub, authService *storage.AuthService) *Handler {
	return &Handler{
		hub:         hub,
		authService: authService,
	}
}

// HandleGameWebSocket handles WebSocket connections for a specific game
func (h *Handler) HandleGameWebSocket(w http.ResponseWriter, r *http.Request) {
	// Extract game ID from URL path
	gameID, err := h.extractGameID(r)
	if err != nil {
		http.Error(w, "Invalid game ID", http.StatusBadRequest)
		return
	}

	// Authenticate user
	user, err := h.authenticateWebSocket(r)
	if err != nil {
		http.Error(w, "Authentication failed", http.StatusUnauthorized)
		return
	}

	// Upgrade HTTP connection to WebSocket
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade failed: %v", err)
		return
	}

	// Create client
	client := &Client{
		ID:     generateClientID(),
		UserID: user.ID,
		GameID: gameID,
		Conn:   conn,
		Send:   make(chan []byte, 256),
		Hub:    h.hub,
	}

	// Register client with hub
	h.hub.register <- client

	// Start goroutines for reading and writing
	go client.writePump()
	go client.readPump()
}

// extractGameID extracts the game ID from the request URL
func (h *Handler) extractGameID(r *http.Request) (int64, error) {
	// Extract from URL path like /ws/game/123
	path := r.URL.Path
	parts := strings.Split(path, "/")

	if len(parts) < 4 || parts[1] != "ws" || parts[2] != "game" {
		return 0, fmt.Errorf("invalid WebSocket path format")
	}

	gameIDStr := parts[3]
	gameID, err := strconv.ParseInt(gameIDStr, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid game ID: %w", err)
	}

	return gameID, nil
}

// authenticateWebSocket authenticates a WebSocket connection
func (h *Handler) authenticateWebSocket(r *http.Request) (*storage.User, error) {
	// Try to get token from Authorization header first
	authHeader := r.Header.Get("Authorization")
	if authHeader != "" {
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) == 2 && parts[0] == "Bearer" {
			token := parts[1]
			user, err := h.authService.GetUserFromToken(token)
			if err == nil {
				return user, nil
			}
		}
	}

	// Try to get token from query parameter as fallback
	token := r.URL.Query().Get("token")
	if token == "" {
		return nil, fmt.Errorf("no authentication token provided")
	}

	user, err := h.authService.GetUserFromToken(token)
	if err != nil {
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	return user, nil
}

// BroadcastGameUpdate broadcasts a game state update to all clients in a game
func (h *Handler) BroadcastGameUpdate(gameID int64, updateType string, data interface{}) {
	h.hub.BroadcastToGame(gameID, updateType, data)
}

// BroadcastUserNotification sends a notification to a specific user
func (h *Handler) BroadcastUserNotification(gameID int64, userID int64, notificationType string, data interface{}) {
	h.hub.BroadcastToUser(gameID, userID, notificationType, data)
}

// GetGameConnectionCount returns the number of active connections for a game
func (h *Handler) GetGameConnectionCount(gameID int64) int {
	return h.hub.GetGameClientCount(gameID)
}

// GetConnectedGames returns a list of games with active connections
func (h *Handler) GetConnectedGames() []int64 {
	return h.hub.GetConnectedGames()
}

// generateClientID generates a random client ID
func generateClientID() string {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		// Fallback to timestamp-based ID if random fails
		return fmt.Sprintf("client_%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(bytes)
}
