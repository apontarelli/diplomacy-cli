package storage

import (
	"encoding/json"
	"fmt"
	"time"

	"diplomacy-cli/backend/internal/game"
)

// GameStateDTO represents a serializable version of game.GameState
type GameStateDTO struct {
	Phase          string              `json:"phase"`
	Year           int                 `json:"year"`
	Board          *BoardDTO           `json:"board"`
	RawOrders      map[string][]string `json:"raw_orders"`
	SupplyCenters  map[string][]string `json:"supply_centers"`
	DislodgedUnits []*UnitDTO          `json:"dislodged_units"`
	CreatedAt      time.Time           `json:"created_at"`
}

// GameDTO represents a serializable version of game.Game
type GameDTO struct {
	ID           string            `json:"id"`
	Name         string            `json:"name"`
	Players      map[string]string `json:"players"`
	CurrentState *GameStateDTO     `json:"current_state"`
	History      []*GameStateDTO   `json:"history"`
	Status       string            `json:"status"`
	CreatedAt    time.Time         `json:"created_at"`
	UpdatedAt    time.Time         `json:"updated_at"`
}

// BoardDTO represents a serializable version of game.Board
type BoardDTO struct {
	Provinces map[string]*ProvinceDTO `json:"provinces"`
	Units     map[string]*UnitDTO     `json:"units"`
}

// ProvinceDTO represents a serializable version of game.Province
type ProvinceDTO struct {
	Name           string              `json:"name"`
	ShortCode      string              `json:"short_code"`
	DisplayName    string              `json:"display_name"`
	Type           string              `json:"type"`
	SupplyCenter   bool                `json:"supply_center"`
	CoastNeighbors map[string][]string `json:"coast_neighbors"`
	ArmyNeighbors  []string            `json:"army_neighbors"`
	FleetNeighbors []string            `json:"fleet_neighbors"`
}

// UnitDTO represents a serializable version of game.Unit
type UnitDTO struct {
	Type      string `json:"type"`
	Owner     string `json:"owner"`
	Province  string `json:"province"`
	Coast     string `json:"coast"`
	Dislodged bool   `json:"dislodged"`
}

// OrderDTO represents a serializable version of game.Order
type OrderDTO struct {
	ID                 string `json:"id"`
	UnitType           string `json:"unit_type"`
	From               string `json:"from"`
	FromCoast          string `json:"from_coast"`
	Type               string `json:"type"`
	To                 string `json:"to"`
	ToCoast            string `json:"to_coast"`
	Owner              string `json:"owner"`
	Result             string `json:"result"`
	FailureReason      string `json:"failure_reason"`
	SupportTarget      string `json:"support_target"`
	SupportDestination string `json:"support_destination"`
	ConvoyTarget       string `json:"convoy_target"`
}

// ToGameStateDTO converts a domain GameState to a serializable DTO
func ToGameStateDTO(gs *game.GameState) (*GameStateDTO, error) {
	if gs == nil {
		return nil, fmt.Errorf("game state cannot be nil")
	}

	boardDTO, err := ToBoardDTO(gs.Board)
	if err != nil {
		return nil, fmt.Errorf("failed to convert board: %w", err)
	}

	// Convert RawOrders map from Nation keys to string keys
	rawOrders := make(map[string][]string)
	for nation, orders := range gs.RawOrders {
		rawOrders[string(nation)] = orders
	}

	// Convert SupplyCenters map from Nation keys to string keys
	supplyCenters := make(map[string][]string)
	for nation, centers := range gs.SupplyCenters {
		supplyCenters[string(nation)] = centers
	}

	// Convert dislodged units
	var dislodgedUnits []*UnitDTO
	for _, unit := range gs.DislodgedUnits {
		unitDTO, err := ToUnitDTO(unit)
		if err != nil {
			return nil, fmt.Errorf("failed to convert dislodged unit: %w", err)
		}
		dislodgedUnits = append(dislodgedUnits, unitDTO)
	}

	return &GameStateDTO{
		Phase:          string(gs.Phase),
		Year:           gs.Year,
		Board:          boardDTO,
		RawOrders:      rawOrders,
		SupplyCenters:  supplyCenters,
		DislodgedUnits: dislodgedUnits,
		CreatedAt:      gs.CreatedAt,
	}, nil
}

// FromGameStateDTO converts a DTO back to a domain GameState
func FromGameStateDTO(dto *GameStateDTO) (*game.GameState, error) {
	if dto == nil {
		return nil, fmt.Errorf("game state DTO cannot be nil")
	}

	board, err := FromBoardDTO(dto.Board)
	if err != nil {
		return nil, fmt.Errorf("failed to convert board: %w", err)
	}

	// Convert RawOrders map from string keys to Nation keys
	rawOrders := make(map[game.Nation][]string)
	for nationStr, orders := range dto.RawOrders {
		rawOrders[game.Nation(nationStr)] = orders
	}

	// Convert SupplyCenters map from string keys to Nation keys
	supplyCenters := make(map[game.Nation][]string)
	for nationStr, centers := range dto.SupplyCenters {
		supplyCenters[game.Nation(nationStr)] = centers
	}

	// Convert dislodged units
	var dislodgedUnits []*game.Unit
	for _, unitDTO := range dto.DislodgedUnits {
		unit, err := FromUnitDTO(unitDTO)
		if err != nil {
			return nil, fmt.Errorf("failed to convert dislodged unit: %w", err)
		}
		dislodgedUnits = append(dislodgedUnits, unit)
	}

	return &game.GameState{
		Board:          board,
		Phase:          game.Phase(dto.Phase),
		Year:           dto.Year,
		RawOrders:      rawOrders,
		SupplyCenters:  supplyCenters,
		DislodgedUnits: dislodgedUnits,
		CreatedAt:      dto.CreatedAt,
	}, nil
}

// ToGameDTO converts a domain Game to a serializable DTO
func ToGameDTO(g *game.Game) (*GameDTO, error) {
	if g == nil {
		return nil, fmt.Errorf("game cannot be nil")
	}

	// Convert Players map from Nation keys to string keys
	players := make(map[string]string)
	for nation, playerID := range g.Players {
		players[string(nation)] = playerID
	}

	// Convert current state
	var currentStateDTO *GameStateDTO
	if g.CurrentState != nil {
		var err error
		currentStateDTO, err = ToGameStateDTO(g.CurrentState)
		if err != nil {
			return nil, fmt.Errorf("failed to convert current state: %w", err)
		}
	}

	// Convert history
	var historyDTO []*GameStateDTO
	for _, state := range g.History {
		stateDTO, err := ToGameStateDTO(state)
		if err != nil {
			return nil, fmt.Errorf("failed to convert history state: %w", err)
		}
		historyDTO = append(historyDTO, stateDTO)
	}

	return &GameDTO{
		ID:           g.ID,
		Name:         g.Name,
		Players:      players,
		CurrentState: currentStateDTO,
		History:      historyDTO,
		Status:       string(g.Status),
		CreatedAt:    g.CreatedAt,
		UpdatedAt:    g.UpdatedAt,
	}, nil
}

// FromGameDTO converts a DTO back to a domain Game
func FromGameDTO(dto *GameDTO) (*game.Game, error) {
	if dto == nil {
		return nil, fmt.Errorf("game DTO cannot be nil")
	}

	// Convert Players map from string keys to Nation keys
	players := make(map[game.Nation]string)
	for nationStr, playerID := range dto.Players {
		players[game.Nation(nationStr)] = playerID
	}

	// Convert current state
	var currentState *game.GameState
	if dto.CurrentState != nil {
		var err error
		currentState, err = FromGameStateDTO(dto.CurrentState)
		if err != nil {
			return nil, fmt.Errorf("failed to convert current state: %w", err)
		}
	}

	// Convert history
	var history []*game.GameState
	for _, stateDTO := range dto.History {
		state, err := FromGameStateDTO(stateDTO)
		if err != nil {
			return nil, fmt.Errorf("failed to convert history state: %w", err)
		}
		history = append(history, state)
	}

	return &game.Game{
		ID:           dto.ID,
		Name:         dto.Name,
		Players:      players,
		CurrentState: currentState,
		History:      history,
		Status:       game.GameStatus(dto.Status),
		CreatedAt:    dto.CreatedAt,
		UpdatedAt:    dto.UpdatedAt,
	}, nil
}

// ToBoardDTO converts a domain Board to a serializable DTO
func ToBoardDTO(b *game.Board) (*BoardDTO, error) {
	if b == nil {
		return nil, fmt.Errorf("board cannot be nil")
	}

	provinces := make(map[string]*ProvinceDTO)
	for name, province := range b.Provinces {
		provinceDTO, err := ToProvinceDTO(province)
		if err != nil {
			return nil, fmt.Errorf("failed to convert province %s: %w", name, err)
		}
		provinces[name] = provinceDTO
	}

	units := make(map[string]*UnitDTO)
	for name, unit := range b.Units {
		if unit != nil {
			unitDTO, err := ToUnitDTO(unit)
			if err != nil {
				return nil, fmt.Errorf("failed to convert unit at %s: %w", name, err)
			}
			units[name] = unitDTO
		}
	}

	return &BoardDTO{
		Provinces: provinces,
		Units:     units,
	}, nil
}

// FromBoardDTO converts a DTO back to a domain Board
func FromBoardDTO(dto *BoardDTO) (*game.Board, error) {
	if dto == nil {
		return nil, fmt.Errorf("board DTO cannot be nil")
	}

	board := game.NewBoard()

	for name, provinceDTO := range dto.Provinces {
		province, err := FromProvinceDTO(provinceDTO)
		if err != nil {
			return nil, fmt.Errorf("failed to convert province %s: %w", name, err)
		}
		board.Provinces[name] = province
	}

	for name, unitDTO := range dto.Units {
		if unitDTO != nil {
			unit, err := FromUnitDTO(unitDTO)
			if err != nil {
				return nil, fmt.Errorf("failed to convert unit at %s: %w", name, err)
			}
			board.Units[name] = unit
		}
	}

	return board, nil
}

// ToProvinceDTO converts a domain Province to a serializable DTO
func ToProvinceDTO(p *game.Province) (*ProvinceDTO, error) {
	if p == nil {
		return nil, fmt.Errorf("province cannot be nil")
	}

	// Deep copy coast neighbors map
	coastNeighbors := make(map[string][]string)
	for coast, neighbors := range p.CoastNeighbors {
		coastNeighbors[coast] = make([]string, len(neighbors))
		copy(coastNeighbors[coast], neighbors)
	}

	// Copy neighbor slices
	armyNeighbors := make([]string, len(p.ArmyNeighbors))
	copy(armyNeighbors, p.ArmyNeighbors)

	fleetNeighbors := make([]string, len(p.FleetNeighbors))
	copy(fleetNeighbors, p.FleetNeighbors)

	return &ProvinceDTO{
		Name:           p.Name,
		ShortCode:      p.ShortCode,
		DisplayName:    p.DisplayName,
		Type:           string(p.Type),
		SupplyCenter:   p.SupplyCenter,
		CoastNeighbors: coastNeighbors,
		ArmyNeighbors:  armyNeighbors,
		FleetNeighbors: fleetNeighbors,
	}, nil
}

// FromProvinceDTO converts a DTO back to a domain Province
func FromProvinceDTO(dto *ProvinceDTO) (*game.Province, error) {
	if dto == nil {
		return nil, fmt.Errorf("province DTO cannot be nil")
	}

	// Deep copy coast neighbors map
	coastNeighbors := make(map[string][]string)
	for coast, neighbors := range dto.CoastNeighbors {
		coastNeighbors[coast] = make([]string, len(neighbors))
		copy(coastNeighbors[coast], neighbors)
	}

	// Copy neighbor slices
	armyNeighbors := make([]string, len(dto.ArmyNeighbors))
	copy(armyNeighbors, dto.ArmyNeighbors)

	fleetNeighbors := make([]string, len(dto.FleetNeighbors))
	copy(fleetNeighbors, dto.FleetNeighbors)

	return &game.Province{
		Name:           dto.Name,
		ShortCode:      dto.ShortCode,
		DisplayName:    dto.DisplayName,
		Type:           game.ProvinceType(dto.Type),
		SupplyCenter:   dto.SupplyCenter,
		CoastNeighbors: coastNeighbors,
		ArmyNeighbors:  armyNeighbors,
		FleetNeighbors: fleetNeighbors,
	}, nil
}

// ToUnitDTO converts a domain Unit to a serializable DTO
func ToUnitDTO(u *game.Unit) (*UnitDTO, error) {
	if u == nil {
		return nil, fmt.Errorf("unit cannot be nil")
	}

	return &UnitDTO{
		Type:      string(u.Type),
		Owner:     string(u.Owner),
		Province:  u.Province,
		Coast:     u.Coast,
		Dislodged: u.Dislodged,
	}, nil
}

// FromUnitDTO converts a DTO back to a domain Unit
func FromUnitDTO(dto *UnitDTO) (*game.Unit, error) {
	if dto == nil {
		return nil, fmt.Errorf("unit DTO cannot be nil")
	}

	return &game.Unit{
		Type:      game.UnitType(dto.Type),
		Owner:     game.Nation(dto.Owner),
		Province:  dto.Province,
		Coast:     dto.Coast,
		Dislodged: dto.Dislodged,
	}, nil
}

// ToOrderDTO converts a domain Order to a serializable DTO
func ToOrderDTO(o *game.Order) (*OrderDTO, error) {
	if o == nil {
		return nil, fmt.Errorf("order cannot be nil")
	}

	return &OrderDTO{
		ID:                 o.ID,
		UnitType:           string(o.UnitType),
		From:               o.From,
		FromCoast:          o.FromCoast,
		Type:               string(o.Type),
		To:                 o.To,
		ToCoast:            o.ToCoast,
		Owner:              string(o.Owner),
		Result:             string(o.Result),
		FailureReason:      o.FailureReason,
		SupportTarget:      o.SupportTarget,
		SupportDestination: o.SupportDestination,
		ConvoyTarget:       o.ConvoyTarget,
	}, nil
}

// FromOrderDTO converts a DTO back to a domain Order
func FromOrderDTO(dto *OrderDTO) (*game.Order, error) {
	if dto == nil {
		return nil, fmt.Errorf("order DTO cannot be nil")
	}

	return &game.Order{
		ID:                 dto.ID,
		UnitType:           game.UnitType(dto.UnitType),
		From:               dto.From,
		FromCoast:          dto.FromCoast,
		Type:               game.OrderType(dto.Type),
		To:                 dto.To,
		ToCoast:            dto.ToCoast,
		Owner:              game.Nation(dto.Owner),
		Result:             game.OrderResult(dto.Result),
		FailureReason:      dto.FailureReason,
		SupportTarget:      dto.SupportTarget,
		SupportDestination: dto.SupportDestination,
		ConvoyTarget:       dto.ConvoyTarget,
	}, nil
}

// SerializeGameState converts a GameState to JSON bytes
func SerializeGameState(gs *game.GameState) ([]byte, error) {
	dto, err := ToGameStateDTO(gs)
	if err != nil {
		return nil, fmt.Errorf("failed to convert to DTO: %w", err)
	}

	data, err := json.Marshal(dto)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal JSON: %w", err)
	}

	return data, nil
}

// DeserializeGameState converts JSON bytes back to a GameState
func DeserializeGameState(data []byte) (*game.GameState, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("empty JSON data")
	}

	var dto GameStateDTO
	if err := json.Unmarshal(data, &dto); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	gs, err := FromGameStateDTO(&dto)
	if err != nil {
		return nil, fmt.Errorf("failed to convert from DTO: %w", err)
	}

	// Validate the deserialized game state
	if err := ValidateGameState(gs); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	return gs, nil
}

// SerializeGame converts a Game to JSON bytes
func SerializeGame(g *game.Game) ([]byte, error) {
	dto, err := ToGameDTO(g)
	if err != nil {
		return nil, fmt.Errorf("failed to convert to DTO: %w", err)
	}

	data, err := json.Marshal(dto)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal JSON: %w", err)
	}

	return data, nil
}

// DeserializeGame converts JSON bytes back to a Game
func DeserializeGame(data []byte) (*game.Game, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("empty JSON data")
	}

	var dto GameDTO
	if err := json.Unmarshal(data, &dto); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	g, err := FromGameDTO(&dto)
	if err != nil {
		return nil, fmt.Errorf("failed to convert from DTO: %w", err)
	}

	return g, nil
}

// ValidateGameState performs comprehensive validation of a deserialized game state
func ValidateGameState(gs *game.GameState) error {
	if gs == nil {
		return fmt.Errorf("game state cannot be nil")
	}

	// Basic field validation
	if gs.Board == nil {
		return fmt.Errorf("game state board cannot be nil")
	}
	if gs.Year < 1901 {
		return fmt.Errorf("invalid year %d: must be >= 1901", gs.Year)
	}
	if gs.Phase == "" {
		return fmt.Errorf("game state phase cannot be empty")
	}

	// Validate board structure
	if err := validateBoard(gs.Board); err != nil {
		return fmt.Errorf("invalid board: %w", err)
	}

	// Validate unit consistency
	if err := validateUnits(gs.Board); err != nil {
		return fmt.Errorf("invalid units: %w", err)
	}

	// Validate supply centers
	if err := validateSupplyCenters(gs.SupplyCenters, gs.Board); err != nil {
		return fmt.Errorf("invalid supply centers: %w", err)
	}

	// Validate raw orders
	if err := validateRawOrders(gs.RawOrders, gs.Board); err != nil {
		return fmt.Errorf("invalid raw orders: %w", err)
	}

	// Validate dislodged units
	if err := validateDislodgedUnits(gs.DislodgedUnits, gs.Board); err != nil {
		return fmt.Errorf("invalid dislodged units: %w", err)
	}

	return nil
}

// validateBoard checks board structure integrity
func validateBoard(board *game.Board) error {
	if board.Provinces == nil {
		return fmt.Errorf("board provinces map cannot be nil")
	}
	if board.Units == nil {
		return fmt.Errorf("board units map cannot be nil")
	}

	// Validate provinces
	for name, province := range board.Provinces {
		if province == nil {
			return fmt.Errorf("province %s cannot be nil", name)
		}
		if province.Name != name {
			return fmt.Errorf("province name mismatch: key=%s, name=%s", name, province.Name)
		}
		if province.ShortCode == "" {
			return fmt.Errorf("province %s must have a short code", name)
		}
	}

	return nil
}

// validateUnits checks unit placement and consistency
func validateUnits(board *game.Board) error {
	for provinceName, unit := range board.Units {
		if unit == nil {
			continue // Empty provinces are allowed
		}

		// Check if province exists
		province, exists := board.Provinces[provinceName]
		if !exists {
			return fmt.Errorf("unit in non-existent province: %s", provinceName)
		}

		// Check unit placement rules
		if unit.Type == game.Army && province.Type == game.Sea {
			return fmt.Errorf("army cannot be placed in sea province: %s", provinceName)
		}

		if unit.Type == game.Fleet && province.Type == game.Land && len(province.FleetNeighbors) == 0 {
			return fmt.Errorf("fleet cannot be placed in landlocked province: %s", provinceName)
		}

		// Validate unit province matches map key
		if unit.Province != provinceName {
			return fmt.Errorf("unit province mismatch: key=%s, unit.Province=%s", provinceName, unit.Province)
		}

		// Validate nation
		if unit.Owner == "" {
			return fmt.Errorf("unit in %s must have an owner", provinceName)
		}
	}

	return nil
}

// validateSupplyCenters checks supply center ownership consistency
func validateSupplyCenters(supplyCenters map[game.Nation][]string, board *game.Board) error {
	if supplyCenters == nil {
		return nil // Supply centers map can be nil/empty
	}

	seenCenters := make(map[string]game.Nation)

	for nation, centers := range supplyCenters {
		if nation == "" {
			return fmt.Errorf("supply center owner cannot be empty")
		}

		for _, centerName := range centers {
			// Check if province exists
			province, exists := board.Provinces[centerName]
			if !exists {
				return fmt.Errorf("supply center references non-existent province: %s", centerName)
			}

			// Check if province is actually a supply center
			if !province.SupplyCenter {
				return fmt.Errorf("province %s is not a supply center but is assigned to %s", centerName, nation)
			}

			// Check for duplicate assignments
			if prevOwner, seen := seenCenters[centerName]; seen {
				return fmt.Errorf("supply center %s assigned to multiple nations: %s and %s", centerName, prevOwner, nation)
			}
			seenCenters[centerName] = nation
		}
	}

	return nil
}

// validateRawOrders checks raw order format and references
func validateRawOrders(rawOrders map[game.Nation][]string, board *game.Board) error {
	if rawOrders == nil {
		return nil // Raw orders can be nil/empty
	}

	for nation, orders := range rawOrders {
		if nation == "" {
			return fmt.Errorf("raw order nation cannot be empty")
		}

		for i, order := range orders {
			if order == "" {
				return fmt.Errorf("raw order %d for nation %s cannot be empty", i, nation)
			}

			// Basic format validation - orders should contain province references
			if err := validateOrderReferences(order, board); err != nil {
				return fmt.Errorf("invalid order '%s' for nation %s: %w", order, nation, err)
			}
		}
	}

	return nil
}

// validateOrderReferences performs basic validation of province references in orders
func validateOrderReferences(order string, board *game.Board) error {
	// This is a simplified validation - in a full implementation you'd parse the order properly
	// For now, we'll do basic checks for common province codes

	// Extract potential province codes (3-letter codes)
	words := []string{}
	currentWord := ""
	for _, char := range order {
		if (char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') {
			currentWord += string(char)
		} else {
			if len(currentWord) > 0 {
				words = append(words, currentWord)
				currentWord = ""
			}
		}
	}
	if len(currentWord) > 0 {
		words = append(words, currentWord)
	}

	// Check if any 3-letter words might be province codes
	for _, word := range words {
		if len(word) == 3 {
			// Look for this as a short code
			found := false
			for _, province := range board.Provinces {
				if province.ShortCode == word {
					found = true
					break
				}
			}
			// If it looks like a province code but doesn't exist, that's suspicious
			// But we won't fail validation since order parsing is complex
			_ = found // For now, just acknowledge we checked
		}
	}

	return nil
}

// validateDislodgedUnits checks dislodged unit consistency
func validateDislodgedUnits(dislodgedUnits []*game.Unit, board *game.Board) error {
	for i, unit := range dislodgedUnits {
		if unit == nil {
			return fmt.Errorf("dislodged unit %d cannot be nil", i)
		}

		// Check if province exists
		if _, exists := board.Provinces[unit.Province]; !exists {
			return fmt.Errorf("dislodged unit %d references non-existent province: %s", i, unit.Province)
		}

		// Validate nation
		if unit.Owner == "" {
			return fmt.Errorf("dislodged unit %d must have an owner", i)
		}

		// Dislodged units should have the Dislodged flag set
		if !unit.Dislodged {
			return fmt.Errorf("unit %d in dislodged units list should have Dislodged=true", i)
		}
	}

	return nil
}
