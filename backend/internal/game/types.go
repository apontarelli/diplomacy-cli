package game

import "time"

type GameStatus string

const (
	GameStatusForming   GameStatus = "forming"
	GameStatusActive    GameStatus = "active"
	GameStatusCompleted GameStatus = "completed"
)

type PlayerStatus string

const (
	PlayerStatusActive     PlayerStatus = "active"
	PlayerStatusEliminated PlayerStatus = "eliminated"
	PlayerStatusCivil      PlayerStatus = "civil_disorder"
)

type UnitType string

const (
	UnitTypeArmy  UnitType = "army"
	UnitTypeFleet UnitType = "fleet"
)

type Game struct {
	ID        string
	Name      string
	Status    GameStatus
	CreatedAt time.Time
}

type Turn struct {
	ID        int
	GameID    string
	TurnCode  string
	StateJSON string
	CreatedAt time.Time
}

type Player struct {
	ID     string
	GameID string
	Nation string
	Status PlayerStatus
}

type Unit struct {
	ID          string
	GameID      string
	OwnerID     string
	UnitType    UnitType
	TerritoryID string
}

type TerritoryState struct {
	ID          int
	GameID      string
	TurnID      int
	TerritoryID string
	OwnerID     string
}

type GameState struct {
	Game        Game
	Players     []Player
	Units       []Unit
	Territories []TerritoryState
}