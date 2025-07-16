package game

import (
	"fmt"
	"time"
)

type Phase string

const (
	SpringMovement Phase = "spring_movement"
	FallMovement   Phase = "fall_movement"
	SpringRetreat  Phase = "spring_retreat"
	FallRetreat    Phase = "fall_retreat"
	WinterBuild    Phase = "winter_build"
)

type OrderType string

const (
	Move    OrderType = "move"
	Hold    OrderType = "hold"
	Support OrderType = "support"
	Convoy  OrderType = "convoy"
)

type OrderResult string

const (
	Success         OrderResult = "success"
	Failure         OrderResult = "failure"
	Bounced         OrderResult = "bounced"
	Dislodged       OrderResult = "dislodged"
	ConvoyDisrupted OrderResult = "convoy_disrupted"
)

type Order struct {
	ID            string
	UnitType      UnitType
	From          string
	FromCoast     string
	Type          OrderType
	To            string
	ToCoast       string
	Owner         Nation
	Result        OrderResult
	FailureReason string

	SupportTarget      string
	SupportDestination string

	ConvoyTarget string
}

type GameState struct {
	Board          *Board
	Phase          Phase
	Year           int
	RawOrders      map[Nation][]string
	SupplyCenters  map[Nation][]string
	DislodgedUnits []*Unit // Units that have been dislodged and need to retreat
	CreatedAt      time.Time
}

type Game struct {
	ID           string
	Name         string
	Players      map[Nation]string
	CurrentState *GameState
	History      []*GameState
	Status       GameStatus
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type GameStatus string

const (
	WaitingForPlayers GameStatus = "waiting_for_players"
	InProgress        GameStatus = "in_progress"
	Completed         GameStatus = "completed"
	Abandoned         GameStatus = "abandoned"
)

func NewGameState(board *Board, phase Phase, year int) *GameState {
	return &GameState{
		Board:         board,
		Phase:         phase,
		Year:          year,
		RawOrders:     make(map[Nation][]string),
		SupplyCenters: make(map[Nation][]string),
		CreatedAt:     time.Now(),
	}
}

func NewGame(id, name string) *Game {
	return &Game{
		ID:        id,
		Name:      name,
		Players:   make(map[Nation]string),
		History:   make([]*GameState, 0),
		Status:    WaitingForPlayers,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

func (gs *GameState) AddRawOrder(nation Nation, orderText string) error {
	if gs.RawOrders[nation] == nil {
		gs.RawOrders[nation] = make([]string, 0)
	}
	gs.RawOrders[nation] = append(gs.RawOrders[nation], orderText)
	return nil
}

func (gs *GameState) GetRawOrdersForNation(nation Nation) []string {
	return gs.RawOrders[nation]
}

func (g *Game) AddPlayer(nation Nation, playerID string) error {
	if g.Status != WaitingForPlayers {
		return fmt.Errorf("cannot add players to game in status %s", g.Status)
	}

	if _, exists := g.Players[nation]; exists {
		return fmt.Errorf("nation %s is already taken", nation)
	}

	g.Players[nation] = playerID
	g.UpdatedAt = time.Now()
	return nil
}

func (g *Game) IsPlayerInGame(playerID string) bool {
	for _, id := range g.Players {
		if id == playerID {
			return true
		}
	}
	return false
}

func (g *Game) GetPlayerNation(playerID string) Nation {
	for nation, id := range g.Players {
		if id == playerID {
			return nation
		}
	}
	return ""
}

func (gs *GameState) AdvanceToNextPhase(dislodgedUnits []*Unit) {
	switch gs.Phase {
	case SpringMovement:
		if len(dislodgedUnits) > 0 {
			gs.Phase = SpringRetreat
		} else {
			gs.Phase = FallMovement
		}
	case FallMovement:
		if len(dislodgedUnits) > 0 {
			gs.Phase = FallRetreat
		} else {
			gs.Phase = WinterBuild
			gs.Year++
		}
	case SpringRetreat:
		gs.Phase = FallMovement
	case FallRetreat:
		gs.Phase = WinterBuild
		gs.Year++
	case WinterBuild:
		gs.Phase = SpringMovement
	}
	gs.RawOrders = make(map[Nation][]string)
}

func (gs *GameState) Clone() *GameState {
	newBoard := NewBoard()

	for name, province := range gs.Board.Provinces {
		newProvince := &Province{
			Name:           province.Name,
			ShortCode:      province.ShortCode,
			DisplayName:    province.DisplayName,
			Type:           province.Type,
			SupplyCenter:   province.SupplyCenter,
			CoastNeighbors: make(map[string][]string),
			ArmyNeighbors:  make([]string, len(province.ArmyNeighbors)),
			FleetNeighbors: make([]string, len(province.FleetNeighbors)),
		}

		for coast, neighbors := range province.CoastNeighbors {
			newProvince.CoastNeighbors[coast] = make([]string, len(neighbors))
			copy(newProvince.CoastNeighbors[coast], neighbors)
		}

		copy(newProvince.ArmyNeighbors, province.ArmyNeighbors)
		copy(newProvince.FleetNeighbors, province.FleetNeighbors)
		newBoard.Provinces[name] = newProvince
	}

	for name, unit := range gs.Board.Units {
		if unit != nil {
			newBoard.Units[name] = &Unit{
				Type:     unit.Type,
				Owner:    unit.Owner,
				Province: unit.Province,
				Coast:    unit.Coast,
			}
		}
	}

	newRawOrders := make(map[Nation][]string)
	for nation, orders := range gs.RawOrders {
		newRawOrders[nation] = make([]string, len(orders))
		copy(newRawOrders[nation], orders)
	}

	newSupplyCenters := make(map[Nation][]string)
	for nation, centers := range gs.SupplyCenters {
		newSupplyCenters[nation] = make([]string, len(centers))
		copy(newSupplyCenters[nation], centers)
	}

	newDislodgedUnits := make([]*Unit, len(gs.DislodgedUnits))
	for i, unit := range gs.DislodgedUnits {
		if unit != nil {
			newDislodgedUnits[i] = &Unit{
				Type:     unit.Type,
				Owner:    unit.Owner,
				Province: unit.Province,
				Coast:    unit.Coast,
			}
		}
	}

	return &GameState{
		Board:          newBoard,
		Phase:          gs.Phase,
		Year:           gs.Year,
		RawOrders:      newRawOrders,
		SupplyCenters:  newSupplyCenters,
		DislodgedUnits: newDislodgedUnits,
		CreatedAt:      gs.CreatedAt,
	}
}
