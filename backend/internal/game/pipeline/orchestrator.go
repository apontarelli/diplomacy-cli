package pipeline

import (
	"diplomacy-cli/backend/internal/game"
	"diplomacy-cli/backend/internal/game/resolution"
	"diplomacy-cli/backend/internal/game/validation"
	"fmt"
)

// TurnProcessor orchestrates the complete pipeline from raw orders to new game state
type TurnProcessor struct {
	resolver *validation.ProvinceResolver
	registry *validation.OrderParserRegistry
}

// NewTurnProcessor creates a new turn processor
func NewTurnProcessor() *TurnProcessor {
	return &TurnProcessor{
		registry: validation.NewOrderParserRegistry(),
	}
}

// ProcessTurn executes the complete pipeline: raw orders → parsed orders → resolution → new state
func (tp *TurnProcessor) ProcessTurn(gameState *game.GameState) (*game.GameState, error) {
	// Initialize resolver with current board state
	tp.resolver = validation.NewProvinceResolver(gameState.Board)

	// Step 1: Syntax Validation - Parse all raw orders
	parsedOrders, syntaxErrors := tp.parseAllOrders(gameState)
	if len(syntaxErrors) > 0 {
		return nil, fmt.Errorf("syntax errors: %v", syntaxErrors)
	}

	// Step 2: Semantic Validation - Validate orders against game rules
	validatedOrders, semanticErrors := tp.validateOrders(parsedOrders, gameState)
	if len(semanticErrors) > 0 {
		return nil, fmt.Errorf("semantic errors: %v", semanticErrors)
	}

	// Step 3: Resolution - Execute multi-pass resolution algorithm
	resolutionResults, err := tp.resolveOrders(validatedOrders, gameState.Board)
	if err != nil {
		return nil, fmt.Errorf("resolution failed: %w", err)
	}

	// Step 4: State Transition - Apply results to create new game state
	newState, err := tp.applyResults(gameState, validatedOrders, resolutionResults)
	if err != nil {
		return nil, fmt.Errorf("state transition failed: %w", err)
	}

	return newState, nil
}

// parseAllOrders converts raw order strings to parsed Order objects
func (tp *TurnProcessor) parseAllOrders(gameState *game.GameState) ([]*game.Order, []error) {
	var allOrders []*game.Order
	var errors []error

	for nation, rawOrders := range gameState.RawOrders {
		for _, rawOrder := range rawOrders {
			// Tokenize the raw order
			tokens := validation.Tokenize(rawOrder)

			// Parse tokens into structured order
			order, err := tp.registry.ParseOrder(tokens, gameState.Phase, tp.resolver)
			if err != nil {
				errors = append(errors, fmt.Errorf("nation %s, order '%s': %w", nation, rawOrder, err))
				continue
			}

			// Set the owner nation
			order.Owner = nation
			allOrders = append(allOrders, order)
		}
	}

	return allOrders, errors
}

// validateOrders performs semantic validation on parsed orders
func (tp *TurnProcessor) validateOrders(orders []*game.Order, gameState *game.GameState) ([]*game.Order, []error) {
	var validOrders []*game.Order
	var errors []error

	for _, order := range orders {
		// Semantic validation checks:
		// 1. Unit exists at the specified location
		// 2. Unit belongs to the nation giving the order
		// 3. Order is valid for the current game phase
		// 4. Move destinations are adjacent (for moves)
		// 5. Support/convoy targets exist and are valid

		if err := tp.validateOrder(order, gameState); err != nil {
			errors = append(errors, fmt.Errorf("order %s: %w", order.ID, err))
			continue
		}

		validOrders = append(validOrders, order)
	}

	return validOrders, errors
}

// validateOrder performs semantic validation on a single order
func (tp *TurnProcessor) validateOrder(order *game.Order, gameState *game.GameState) error {
	// Check if unit exists at the specified location
	unit := gameState.Board.GetUnit(order.From)
	if unit == nil {
		return fmt.Errorf("no unit found at %s", order.From)
	}

	// Check if unit belongs to the nation giving the order
	if unit.Owner != order.Owner {
		return fmt.Errorf("unit at %s belongs to %s, not %s", order.From, unit.Owner, order.Owner)
	}

	// Check if unit type matches (if specified in order)
	if order.UnitType != "" && unit.Type != order.UnitType {
		return fmt.Errorf("unit at %s is %s, not %s", order.From, unit.Type, order.UnitType)
	}

	// Phase-specific validation
	switch gameState.Phase {
	case game.SpringMovement, game.FallMovement:
		return tp.validateMovementOrder(order, gameState)
	case game.SpringRetreat, game.FallRetreat:
		return tp.validateRetreatOrder(order, gameState)
	case game.WinterBuild:
		return tp.validateBuildOrder(order, gameState)
	default:
		return fmt.Errorf("unknown phase: %s", gameState.Phase)
	}
}

// validateMovementOrder validates orders during movement phases
func (tp *TurnProcessor) validateMovementOrder(order *game.Order, gameState *game.GameState) error {
	switch order.Type {
	case game.Move:
		// Check if unit is trying to move to its own location
		if order.From == order.To {
			return fmt.Errorf("unit cannot move to its own location")
		}
		// Convoy-aware adjacency validation
		if err := tp.validateMoveAdjacency(order, gameState); err != nil {
			return err
		}
	case game.Hold:
		// Hold orders are always valid if unit exists
		return nil
	case game.Support:
		// Validate support target exists
		if order.SupportTarget == "" {
			return fmt.Errorf("support order missing target")
		}
		// Check for self-support (unit cannot support itself)
		if order.From == order.SupportTarget {
			return fmt.Errorf("unit cannot support itself")
		}
		// For support move orders, validate that the supporting unit can reach the destination
		if order.SupportDestination != "" { // This is a support move order
			unit := gameState.Board.GetUnit(order.From)
			if !tp.areAdjacent(order.From, order.SupportDestination, unit.Type, gameState.Board) {
				return fmt.Errorf("supporting unit cannot reach destination %s", order.SupportDestination)
			}
		}
		// Additional support validation would go here
	case game.Convoy:
		// Validate convoy is by fleet
		unit := gameState.Board.GetUnit(order.From)
		if unit.Type != game.Fleet {
			return fmt.Errorf("convoy orders can only be given by fleets")
		}
		// Additional convoy validation would go here
	}

	return nil
}

// areAdjacent checks if two provinces are adjacent for the given unit type
func (tp *TurnProcessor) areAdjacent(from, to string, unitType game.UnitType, board *game.Board) bool {
	fromProvince := board.GetProvince(from)
	if fromProvince == nil {
		return false
	}

	switch unitType {
	case game.Army:
		for _, neighbor := range fromProvince.ArmyNeighbors {
			if neighbor == to {
				return true
			}
		}
	case game.Fleet:
		for _, neighbor := range fromProvince.FleetNeighbors {
			if neighbor == to {
				return true
			}
		}
		// Also check coast neighbors
		for _, neighbors := range fromProvince.CoastNeighbors {
			for _, neighbor := range neighbors {
				if neighbor == to {
					return true
				}
			}
		}
	}

	return false
}

// validateMoveAdjacency validates move adjacency with convoy awareness
func (tp *TurnProcessor) validateMoveAdjacency(order *game.Order, gameState *game.GameState) error {
	// For fleets, always check normal adjacency (fleets cannot be convoyed)
	if order.UnitType == game.Fleet {
		if !tp.areAdjacent(order.From, order.To, order.UnitType, gameState.Board) {
			return fmt.Errorf("cannot move from %s to %s: not adjacent (unit type: %v)", order.From, order.To, order.UnitType)
		}
		return nil
	}

	// For armies, check if convoy orders exist that could enable this move
	if order.UnitType == game.Army {
		// First check normal adjacency
		if tp.areAdjacent(order.From, order.To, order.UnitType, gameState.Board) {
			return nil // Adjacent move is always valid
		}

		// Not adjacent - check if convoy orders exist for this move
		if tp.hasConvoyOrders(order.From, order.To, gameState) {
			return nil // Convoy orders exist - let resolution engine validate the path
		}

		// No convoy orders and not adjacent
		return fmt.Errorf("cannot move from %s to %s: not adjacent and no convoy orders (unit type: %v)", order.From, order.To, order.UnitType)
	}

	// Unknown unit type
	return fmt.Errorf("unknown unit type: %v", order.UnitType)
}

// hasConvoyOrders checks if there are convoy orders that could enable the given army move
func (tp *TurnProcessor) hasConvoyOrders(from, to string, gameState *game.GameState) bool {
	for _, orders := range gameState.RawOrders {
		for _, rawOrder := range orders {
			// Parse the order to check if it's a convoy for this move
			tokens := validation.Tokenize(rawOrder)
			if len(tokens) < 5 {
				continue
			}

			// Check for convoy order format: "f province c from - to"
			if len(tokens) >= 6 &&
				(tokens[1].Value == "c" || tokens[1].Value == "convoy" || tokens[1].Value == "convoys") &&
				tokens[2].Value == from && tokens[4].Value == to {
				return true
			}
			if len(tokens) >= 7 &&
				(tokens[2].Value == "c" || tokens[2].Value == "convoy" || tokens[2].Value == "convoys") &&
				tokens[3].Value == from && tokens[5].Value == to {
				return true
			}
		}
	}
	return false
}

// validateRetreatOrder validates orders during retreat phases
func (tp *TurnProcessor) validateRetreatOrder(order *game.Order, gameState *game.GameState) error {
	// Retreat validation logic would go here
	return nil
}

// validateBuildOrder validates orders during build phases
func (tp *TurnProcessor) validateBuildOrder(order *game.Order, gameState *game.GameState) error {
	// Build/disband validation logic would go here
	return nil
}

// resolveOrders executes the multi-pass resolution algorithm
func (tp *TurnProcessor) resolveOrders(orders []*game.Order, board *game.Board) ([]resolution.OrderOutcome, error) {
	// Create resolution engine
	engine := resolution.NewResolutionEngine(orders, board)

	// Execute resolution
	err := engine.Resolve()
	if err != nil {
		return nil, err
	}

	// Get results
	return engine.GetResults(), nil
}

// applyResults creates a new game state with resolution results applied
func (tp *TurnProcessor) applyResults(
	oldState *game.GameState,
	orders []*game.Order,
	results []resolution.OrderOutcome,
) (*game.GameState, error) {
	// Clone the current state
	newState := oldState.Clone()

	// Track dislodged units for phase advancement
	var dislodgedUnits []*game.Unit

	// Apply each order result
	for i, order := range orders {
		if i >= len(results) {
			continue
		}

		result := results[i]

		switch order.Type {
		case game.Move:
			if err := tp.applyMoveResult(newState, order, result, &dislodgedUnits); err != nil {
				return nil, err
			}
		case game.Hold:
			if err := tp.applyHoldResult(newState, order, result, &dislodgedUnits); err != nil {
				return nil, err
			}
		case game.Support:
			// Support orders don't directly change unit positions
			// Results are already captured in the resolution
		case game.Convoy:
			// Convoy orders don't directly change unit positions
			// Results are already captured in the resolution
		}
	}

	// Advance to next phase based on dislodged units
	newState.AdvanceToNextPhase(dislodgedUnits)

	return newState, nil
}

// applyMoveResult applies the result of a move order
func (tp *TurnProcessor) applyMoveResult(
	state *game.GameState,
	order *game.Order,
	result resolution.OrderOutcome,
	dislodgedUnits *[]*game.Unit,
) error {
	unit := state.Board.GetUnit(order.From)
	if unit == nil {
		return fmt.Errorf("unit not found at %s for move order", order.From)
	}

	switch result.Result {
	case resolution.MoveSuccess:
		// Move succeeded - relocate unit
		state.Board.RemoveUnit(order.From)
		unit.Province = order.To
		unit.Coast = order.ToCoast
		state.Board.PlaceUnit(unit)

	case resolution.MoveBounced:
		// Move bounced - unit stays in place
		// No action needed

	case resolution.MoveNoConvoy:
		// Move failed due to convoy disruption - unit stays in place
		// No action needed

	case resolution.Dislodged:
		// Unit was dislodged by another unit
		*dislodgedUnits = append(*dislodgedUnits, unit)
		state.Board.RemoveUnit(order.From)
	}

	return nil
}

// applyHoldResult applies the result of a hold order
func (tp *TurnProcessor) applyHoldResult(
	state *game.GameState,
	order *game.Order,
	result resolution.OrderOutcome,
	dislodgedUnits *[]*game.Unit,
) error {
	if result.Dislodged {
		unit := state.Board.GetUnit(order.From)
		if unit != nil {
			*dislodgedUnits = append(*dislodgedUnits, unit)
			state.Board.RemoveUnit(order.From)
		}
	}
	// Otherwise, unit holds position (no action needed)

	return nil
}

// TurnResult contains the outcome of processing a turn
type TurnResult struct {
	NewState     *game.GameState
	OrderResults []OrderResult
	Errors       []error
}

// OrderResult contains the result of processing a single order
type OrderResult struct {
	Order   *game.Order
	Outcome resolution.OrderOutcome
	Success bool
	Message string
}

// ProcessTurnWithDetails returns detailed results including individual order outcomes
func (tp *TurnProcessor) ProcessTurnWithDetails(gameState *game.GameState) (*TurnResult, error) {
	result := &TurnResult{
		OrderResults: make([]OrderResult, 0),
		Errors:       make([]error, 0),
	}

	// Process the turn
	newState, err := tp.ProcessTurn(gameState)
	if err != nil {
		result.Errors = append(result.Errors, err)
		return result, err
	}

	result.NewState = newState

	// TODO: Populate detailed order results
	// This would require tracking individual order outcomes through the pipeline

	return result, nil
}
