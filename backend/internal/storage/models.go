package storage

import (
	"time"
)

// Game represents a Diplomacy game instance
type Game struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	Phase     string    `json:"phase"` // spring_movement, fall_movement, spring_retreat, fall_retreat, winter_build
	Year      int       `json:"year"`
	Status    string    `json:"status"` // waiting_for_players, in_progress, completed, abandoned
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Player represents a player in a game
type Player struct {
	ID        int64     `json:"id"`
	GameID    int64     `json:"game_id"`
	UserID    string    `json:"user_id"` // External user identifier
	Nation    string    `json:"nation"`  // austria, england, france, germany, italy, russia, turkey
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Unit represents a military unit on the board
type Unit struct {
	ID        int64     `json:"id"`
	GameID    int64     `json:"game_id"`
	PlayerID  *int64    `json:"player_id"` // Nullable for neutral/dislodged units
	Type      string    `json:"type"`      // army, fleet
	Province  string    `json:"province"`
	Coast     string    `json:"coast"` // Optional coast specification
	Owner     string    `json:"owner"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Order represents a player's order for a unit
type Order struct {
	ID                 int64     `json:"id"`
	GameID             int64     `json:"game_id"`
	PlayerID           int64     `json:"player_id"`
	UnitID             *int64    `json:"unit_id"` // Nullable for build orders
	Type               string    `json:"type"`    // move, hold, support, convoy
	Target             string    `json:"target"`  // Target province for moves/supports
	Coast              string    `json:"coast"`   // Target coast specification
	Resolved           bool      `json:"resolved"`
	Result             string    `json:"result"`              // success, failure, bounced, dislodged, convoy_disrupted
	SupportTarget      string    `json:"support_target"`      // Unit being supported
	SupportDestination string    `json:"support_destination"` // Destination of supported move
	ConvoyTarget       string    `json:"convoy_target"`       // Unit being convoyed
	FailureReason      string    `json:"failure_reason"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}

// SupplyCenter represents ownership of supply centers
type SupplyCenter struct {
	ID        int64     `json:"id"`
	GameID    int64     `json:"game_id"`
	Province  string    `json:"province"`
	Owner     string    `json:"owner"` // Nullable for neutral supply centers
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// GameHistory represents historical game states for replay/analysis
type GameHistory struct {
	ID        int64     `json:"id"`
	GameID    int64     `json:"game_id"`
	Phase     string    `json:"phase"`
	Year      int       `json:"year"`
	StateData string    `json:"state_data"` // JSON serialized game state
	Timestamp time.Time `json:"timestamp"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// User represents a user account (for future authentication)
type User struct {
	ID           int64     `json:"id"`
	Username     string    `json:"username"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"password_hash"`
	DisplayName  string    `json:"display_name"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}
