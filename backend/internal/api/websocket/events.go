package websocket

import "time"

// Event types for WebSocket messages
const (
	// Game state events
	EventGameStateUpdate = "game_state_update"
	EventGamePhaseChange = "game_phase_change"
	EventGameStarted     = "game_started"
	EventGameEnded       = "game_ended"

	// Player events
	EventPlayerJoined = "player_joined"
	EventPlayerLeft   = "player_left"

	// Order events
	EventOrdersSubmitted = "orders_submitted"
	EventOrdersResolved  = "orders_resolved"
	EventOrderDeadline   = "order_deadline"

	// Resolution events
	EventResolutionStarted   = "resolution_started"
	EventResolutionCompleted = "resolution_completed"
	EventResolutionError     = "resolution_error"

	// System events
	EventConnected    = "connected"
	EventDisconnected = "disconnected"
	EventError        = "error"
	EventPing         = "ping"
	EventPong         = "pong"
)

// GameStateUpdateData represents game state update event data
type GameStateUpdateData struct {
	GameID    int64       `json:"game_id"`
	Phase     string      `json:"phase"`
	Year      int         `json:"year"`
	Status    string      `json:"status"`
	UpdatedAt time.Time   `json:"updated_at"`
	Data      interface{} `json:"data,omitempty"`
}

// PlayerEventData represents player join/leave event data
type PlayerEventData struct {
	GameID    int64     `json:"game_id"`
	PlayerID  int64     `json:"player_id"`
	UserID    int64     `json:"user_id"`
	Username  string    `json:"username"`
	Nation    string    `json:"nation"`
	EventTime time.Time `json:"event_time"`
}

// OrderEventData represents order-related event data
type OrderEventData struct {
	GameID     int64       `json:"game_id"`
	PlayerID   int64       `json:"player_id"`
	UserID     int64       `json:"user_id"`
	OrderCount int         `json:"order_count"`
	EventTime  time.Time   `json:"event_time"`
	Orders     interface{} `json:"orders,omitempty"`
}

// ResolutionEventData represents resolution event data
type ResolutionEventData struct {
	GameID    int64       `json:"game_id"`
	Phase     string      `json:"phase"`
	Year      int         `json:"year"`
	EventTime time.Time   `json:"event_time"`
	Results   interface{} `json:"results,omitempty"`
	Error     string      `json:"error,omitempty"`
}

// DeadlineEventData represents deadline notification data
type DeadlineEventData struct {
	GameID       int64     `json:"game_id"`
	Phase        string    `json:"phase"`
	Year         int       `json:"year"`
	DeadlineTime time.Time `json:"deadline_time"`
	TimeLeft     string    `json:"time_left"`
}

// ErrorEventData represents error event data
type ErrorEventData struct {
	Code    string      `json:"code"`
	Message string      `json:"message"`
	Context interface{} `json:"context,omitempty"`
}

// Helper functions for creating event data

// NewGameStateUpdate creates a game state update event
func NewGameStateUpdate(gameID int64, phase string, year int, status string, data interface{}) GameStateUpdateData {
	return GameStateUpdateData{
		GameID:    gameID,
		Phase:     phase,
		Year:      year,
		Status:    status,
		UpdatedAt: time.Now(),
		Data:      data,
	}
}

// NewPlayerJoinedEvent creates a player joined event
func NewPlayerJoinedEvent(gameID, playerID, userID int64, username, nation string) PlayerEventData {
	return PlayerEventData{
		GameID:    gameID,
		PlayerID:  playerID,
		UserID:    userID,
		Username:  username,
		Nation:    nation,
		EventTime: time.Now(),
	}
}

// NewPlayerLeftEvent creates a player left event
func NewPlayerLeftEvent(gameID, playerID, userID int64, username, nation string) PlayerEventData {
	return PlayerEventData{
		GameID:    gameID,
		PlayerID:  playerID,
		UserID:    userID,
		Username:  username,
		Nation:    nation,
		EventTime: time.Now(),
	}
}

// NewOrdersSubmittedEvent creates an orders submitted event
func NewOrdersSubmittedEvent(gameID, playerID, userID int64, orderCount int, orders interface{}) OrderEventData {
	return OrderEventData{
		GameID:     gameID,
		PlayerID:   playerID,
		UserID:     userID,
		OrderCount: orderCount,
		EventTime:  time.Now(),
		Orders:     orders,
	}
}

// NewResolutionStartedEvent creates a resolution started event
func NewResolutionStartedEvent(gameID int64, phase string, year int) ResolutionEventData {
	return ResolutionEventData{
		GameID:    gameID,
		Phase:     phase,
		Year:      year,
		EventTime: time.Now(),
	}
}

// NewResolutionCompletedEvent creates a resolution completed event
func NewResolutionCompletedEvent(gameID int64, phase string, year int, results interface{}) ResolutionEventData {
	return ResolutionEventData{
		GameID:    gameID,
		Phase:     phase,
		Year:      year,
		EventTime: time.Now(),
		Results:   results,
	}
}

// NewResolutionErrorEvent creates a resolution error event
func NewResolutionErrorEvent(gameID int64, phase string, year int, errorMsg string) ResolutionEventData {
	return ResolutionEventData{
		GameID:    gameID,
		Phase:     phase,
		Year:      year,
		EventTime: time.Now(),
		Error:     errorMsg,
	}
}

// NewDeadlineEvent creates a deadline notification event
func NewDeadlineEvent(gameID int64, phase string, year int, deadlineTime time.Time, timeLeft string) DeadlineEventData {
	return DeadlineEventData{
		GameID:       gameID,
		Phase:        phase,
		Year:         year,
		DeadlineTime: deadlineTime,
		TimeLeft:     timeLeft,
	}
}

// NewErrorEvent creates an error event
func NewErrorEvent(code, message string, context interface{}) ErrorEventData {
	return ErrorEventData{
		Code:    code,
		Message: message,
		Context: context,
	}
}
