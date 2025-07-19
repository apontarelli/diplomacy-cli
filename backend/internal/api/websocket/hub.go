package websocket

import (
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// Client represents a WebSocket client connection
type Client struct {
	ID     string
	UserID int64
	GameID int64
	Conn   *websocket.Conn
	Send   chan []byte
	Hub    *Hub
}

// Hub maintains the set of active clients and broadcasts messages to them
type Hub struct {
	// Registered clients by game ID
	games map[int64]map[*Client]bool

	// Register requests from clients
	register chan *Client

	// Unregister requests from clients
	unregister chan *Client

	// Broadcast messages to all clients in a game
	broadcast chan *BroadcastMessage

	// Mutex for thread-safe operations
	mutex sync.RWMutex
}

// BroadcastMessage represents a message to broadcast to clients
type BroadcastMessage struct {
	GameID int64       `json:"game_id"`
	Type   string      `json:"type"`
	Data   interface{} `json:"data"`
	UserID *int64      `json:"user_id,omitempty"` // If set, only send to this user
}

// Message represents a WebSocket message
type Message struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}

// NewHub creates a new WebSocket hub
func NewHub() *Hub {
	return &Hub{
		games:      make(map[int64]map[*Client]bool),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		broadcast:  make(chan *BroadcastMessage),
	}
}

// Run starts the hub's main loop
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.registerClient(client)

		case client := <-h.unregister:
			h.unregisterClient(client)

		case message := <-h.broadcast:
			h.broadcastMessage(message)
		}
	}
}

// registerClient adds a client to the hub
func (h *Hub) registerClient(client *Client) {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	if h.games[client.GameID] == nil {
		h.games[client.GameID] = make(map[*Client]bool)
	}
	h.games[client.GameID][client] = true

	log.Printf("Client %s (User %d) connected to game %d", client.ID, client.UserID, client.GameID)

	// Send welcome message
	welcomeMsg := Message{
		Type: "connected",
		Data: map[string]interface{}{
			"client_id": client.ID,
			"game_id":   client.GameID,
		},
	}
	h.sendToClient(client, welcomeMsg)
}

// unregisterClient removes a client from the hub
func (h *Hub) unregisterClient(client *Client) {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	if clients, ok := h.games[client.GameID]; ok {
		if _, ok := clients[client]; ok {
			delete(clients, client)
			close(client.Send)

			// Clean up empty game rooms
			if len(clients) == 0 {
				delete(h.games, client.GameID)
			}

			log.Printf("Client %s (User %d) disconnected from game %d", client.ID, client.UserID, client.GameID)
		}
	}
}

// broadcastMessage sends a message to all clients in a game
func (h *Hub) broadcastMessage(message *BroadcastMessage) {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	clients, ok := h.games[message.GameID]
	if !ok {
		return
	}

	msg := Message{
		Type: message.Type,
		Data: message.Data,
	}

	data, err := json.Marshal(msg)
	if err != nil {
		log.Printf("Error marshaling message: %v", err)
		return
	}

	for client := range clients {
		// If UserID is specified, only send to that user
		if message.UserID != nil && client.UserID != *message.UserID {
			continue
		}

		select {
		case client.Send <- data:
		default:
			// Client's send channel is full, close it
			close(client.Send)
			delete(clients, client)
		}
	}
}

// sendToClient sends a message to a specific client
func (h *Hub) sendToClient(client *Client, message Message) {
	data, err := json.Marshal(message)
	if err != nil {
		log.Printf("Error marshaling message: %v", err)
		return
	}

	select {
	case client.Send <- data:
	default:
		close(client.Send)
		h.unregisterClient(client)
	}
}

// BroadcastToGame sends a message to all clients in a specific game
func (h *Hub) BroadcastToGame(gameID int64, messageType string, data interface{}) {
	message := &BroadcastMessage{
		GameID: gameID,
		Type:   messageType,
		Data:   data,
	}

	select {
	case h.broadcast <- message:
	default:
		log.Printf("Broadcast channel full, dropping message for game %d", gameID)
	}
}

// BroadcastToUser sends a message to a specific user in a game
func (h *Hub) BroadcastToUser(gameID int64, userID int64, messageType string, data interface{}) {
	message := &BroadcastMessage{
		GameID: gameID,
		Type:   messageType,
		Data:   data,
		UserID: &userID,
	}

	select {
	case h.broadcast <- message:
	default:
		log.Printf("Broadcast channel full, dropping message for user %d in game %d", userID, gameID)
	}
}

// GetGameClientCount returns the number of connected clients for a game
func (h *Hub) GetGameClientCount(gameID int64) int {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	if clients, ok := h.games[gameID]; ok {
		return len(clients)
	}
	return 0
}

// GetConnectedGames returns a list of game IDs with active connections
func (h *Hub) GetConnectedGames() []int64 {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	games := make([]int64, 0, len(h.games))
	for gameID := range h.games {
		games = append(games, gameID)
	}
	return games
}

// writePump pumps messages from the hub to the websocket connection
func (c *Client) writePump() {
	ticker := time.NewTicker(54 * time.Second)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.Conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Add queued messages to the current websocket message
			n := len(c.Send)
			for i := 0; i < n; i++ {
				w.Write([]byte{'\n'})
				w.Write(<-c.Send)
			}

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// readPump pumps messages from the websocket connection to the hub
func (c *Client) readPump() {
	defer func() {
		c.Hub.unregister <- c
		c.Conn.Close()
	}()

	c.Conn.SetReadLimit(512)
	c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}

		// Handle incoming messages from client
		var msg Message
		if err := json.Unmarshal(message, &msg); err != nil {
			log.Printf("Error unmarshaling client message: %v", err)
			continue
		}

		// Process client message (e.g., ping, subscribe to events, etc.)
		c.handleClientMessage(msg)
	}
}

// handleClientMessage processes messages received from the client
func (c *Client) handleClientMessage(msg Message) {
	switch msg.Type {
	case "ping":
		// Respond with pong
		pongMsg := Message{
			Type: "pong",
			Data: msg.Data,
		}
		c.Hub.sendToClient(c, pongMsg)

	case "subscribe":
		// Handle subscription requests (future enhancement)
		log.Printf("Client %s requested subscription: %v", c.ID, msg.Data)

	default:
		log.Printf("Unknown message type from client %s: %s", c.ID, msg.Type)
	}
}
