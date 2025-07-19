package websocket

import (
	"encoding/json"
	"errors"
	"net/http/httptest"
	"testing"
	"time"

	"diplomacy-cli/backend/internal/storage"
)

var ErrInvalidCredentials = errors.New("invalid credentials")

// MockAuthService for testing
type MockAuthService struct{}

func (m *MockAuthService) GetUserFromToken(token string) (*storage.User, error) {
	if token == "valid_token" {
		return &storage.User{
			ID:          1,
			Username:    "testuser",
			Email:       "test@example.com",
			DisplayName: "Test User",
		}, nil
	}
	return nil, ErrInvalidCredentials
}

func (m *MockAuthService) ValidateToken(token string) (*storage.Claims, error) {
	if token == "valid_token" {
		return &storage.Claims{
			UserID:   1,
			Username: "testuser",
			Email:    "test@example.com",
		}, nil
	}
	return nil, ErrInvalidCredentials
}

func TestHub_BroadcastToGame(t *testing.T) {
	hub := NewHub()
	go hub.Run()

	// Create a mock client
	client := &Client{
		ID:     "test_client",
		UserID: 1,
		GameID: 123,
		Send:   make(chan []byte, 256),
		Hub:    hub,
	}

	// Register the client
	hub.register <- client

	// Give the hub time to process the registration
	time.Sleep(10 * time.Millisecond)

	// Broadcast a message
	testData := map[string]string{"test": "data"}
	hub.BroadcastToGame(123, "test_event", testData)

	// First message should be the welcome message, skip it
	select {
	case <-client.Send:
		// Skip welcome message
	case <-time.After(50 * time.Millisecond):
		t.Error("Timeout waiting for welcome message")
	}

	// Check if the broadcast message was received
	select {
	case message := <-client.Send:
		var msg Message
		err := json.Unmarshal(message, &msg)
		if err != nil {
			t.Fatalf("Failed to unmarshal message: %v", err)
		}

		if msg.Type != "test_event" {
			t.Errorf("Expected message type 'test_event', got '%s'", msg.Type)
		}

		// Verify the data
		data, ok := msg.Data.(map[string]interface{})
		if !ok {
			t.Errorf("Expected data to be a map, got %T", msg.Data)
		}

		if data["test"] != "data" {
			t.Errorf("Expected data['test'] to be 'data', got '%v'", data["test"])
		}

	case <-time.After(100 * time.Millisecond):
		t.Error("Timeout waiting for broadcast message")
	}

	// Clean up
	hub.unregister <- client
}

func TestHub_BroadcastToUser(t *testing.T) {
	hub := NewHub()
	go hub.Run()

	// Create two mock clients for the same game but different users
	client1 := &Client{
		ID:     "test_client_1",
		UserID: 1,
		GameID: 123,
		Send:   make(chan []byte, 256),
		Hub:    hub,
	}

	client2 := &Client{
		ID:     "test_client_2",
		UserID: 2,
		GameID: 123,
		Send:   make(chan []byte, 256),
		Hub:    hub,
	}

	// Register both clients
	hub.register <- client1
	hub.register <- client2

	// Give the hub time to process the registrations
	time.Sleep(10 * time.Millisecond)

	// Broadcast a message to user 1 only
	testData := map[string]string{"user": "specific"}
	hub.BroadcastToUser(123, 1, "user_event", testData)

	// Skip welcome messages for both clients
	select {
	case <-client1.Send:
		// Skip welcome message for client1
	case <-time.After(50 * time.Millisecond):
		t.Error("Timeout waiting for welcome message on client1")
	}

	select {
	case <-client2.Send:
		// Skip welcome message for client2
	case <-time.After(50 * time.Millisecond):
		t.Error("Timeout waiting for welcome message on client2")
	}

	// Check if client1 received the user-specific message
	select {
	case message := <-client1.Send:
		var msg Message
		err := json.Unmarshal(message, &msg)
		if err != nil {
			t.Fatalf("Failed to unmarshal message: %v", err)
		}

		if msg.Type != "user_event" {
			t.Errorf("Expected message type 'user_event', got '%s'", msg.Type)
		}

	case <-time.After(100 * time.Millisecond):
		t.Error("Timeout waiting for user-specific message on client1")
	}

	// Check that client2 did NOT receive the user-specific message
	select {
	case <-client2.Send:
		t.Error("Client2 should not have received the user-specific message")
	case <-time.After(50 * time.Millisecond):
		// This is expected - client2 should not receive the message
	}

	// Clean up
	hub.unregister <- client1
	hub.unregister <- client2
}

func TestHandler_extractGameID(t *testing.T) {
	handler := &Handler{}

	tests := []struct {
		path        string
		expectedID  int64
		expectError bool
	}{
		{"/ws/game/123", 123, false},
		{"/ws/game/456", 456, false},
		{"/ws/game/invalid", 0, true},
		{"/invalid/path", 0, true},
		{"/ws/game/", 0, true},
	}

	for _, test := range tests {
		req := httptest.NewRequest("GET", test.path, nil)
		gameID, err := handler.extractGameID(req)

		if test.expectError {
			if err == nil {
				t.Errorf("Expected error for path '%s', but got none", test.path)
			}
		} else {
			if err != nil {
				t.Errorf("Unexpected error for path '%s': %v", test.path, err)
			}
			if gameID != test.expectedID {
				t.Errorf("Expected game ID %d for path '%s', got %d", test.expectedID, test.path, gameID)
			}
		}
	}
}

// Note: Authentication tests would require a proper mock setup
// For now, we'll focus on testing the core WebSocket functionality

func TestEventHelpers(t *testing.T) {
	// Test NewGameStateUpdate
	gameUpdate := NewGameStateUpdate(123, "spring", 1901, "active", map[string]string{"test": "data"})
	if gameUpdate.GameID != 123 {
		t.Errorf("Expected GameID 123, got %d", gameUpdate.GameID)
	}
	if gameUpdate.Phase != "spring" {
		t.Errorf("Expected Phase 'spring', got '%s'", gameUpdate.Phase)
	}

	// Test NewPlayerJoinedEvent
	playerEvent := NewPlayerJoinedEvent(123, 456, 789, "testuser", "england")
	if playerEvent.GameID != 123 {
		t.Errorf("Expected GameID 123, got %d", playerEvent.GameID)
	}
	if playerEvent.Nation != "england" {
		t.Errorf("Expected Nation 'england', got '%s'", playerEvent.Nation)
	}

	// Test NewResolutionCompletedEvent
	resolutionEvent := NewResolutionCompletedEvent(123, "spring", 1901, map[string]string{"result": "success"})
	if resolutionEvent.GameID != 123 {
		t.Errorf("Expected GameID 123, got %d", resolutionEvent.GameID)
	}
	if resolutionEvent.Results == nil {
		t.Error("Expected Results to be set, got nil")
	}
}
