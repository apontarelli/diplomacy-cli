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

type Season string

const (
	SeasonSpring Season = "spring"
	SeasonFall   Season = "fall"
)

type Phase string

const (
	PhaseMovement Phase = "movement"
	PhaseRetreat  Phase = "retreat"
	PhaseBuild    Phase = "build"
)

type TurnStatus string

const (
	TurnStatusActive    TurnStatus = "active"
	TurnStatusCompleted TurnStatus = "completed"
)

type OrderType string

const (
	OrderTypeMove    OrderType = "move"
	OrderTypeHold    OrderType = "hold"
	OrderTypeSupport OrderType = "support"
	OrderTypeConvoy  OrderType = "convoy"
)

type OrderStatus string

const (
	OrderStatusPending   OrderStatus = "pending"
	OrderStatusSubmitted OrderStatus = "submitted"
	OrderStatusCancelled OrderStatus = "cancelled"
)

type ResolutionResult string

const (
	ResolutionSuccess ResolutionResult = "success"
	ResolutionFailed  ResolutionResult = "failed"
	ResolutionBounced ResolutionResult = "bounced"
	ResolutionCut     ResolutionResult = "cut"
)

type OrderResult struct {
	OrderID    string
	Result     ResolutionResult
	Reason     string
	Dislodged  bool
	NewPosition string
}

type Game struct {
	ID        string
	Name      string
	Status    GameStatus
	CreatedAt time.Time
}

type Turn struct {
	ID        int
	GameID    string
	Year      int
	Season    Season
	Phase     Phase
	Status    TurnStatus
	Deadline  *time.Time
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

type Order struct {
	ID           string
	GameID       string
	TurnID       int
	PlayerID     string
	UnitID       string
	OrderType    OrderType
	FromTerritory string
	ToTerritory   string
	SupportUnit   string
	Status       OrderStatus
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type GameState struct {
	Game        Game
	Players     []Player
	Units       []Unit
	Territories []TerritoryState
	Orders      []Order
}