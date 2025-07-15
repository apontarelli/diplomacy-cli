package pipeline

import (
	"diplomacy-cli/backend/internal/game"
	"diplomacy-cli/backend/internal/game/resolution"
	"diplomacy-cli/backend/internal/game/validation"
	"fmt"
)

// TurnProcessor orchestrates the complete pipeline from raw orders to new game state
type TurnProcessor struct {
	resolver          *validation.ProvinceResolver
	registry          *validation.OrderParserRegistry
	semanticValidator *validation.SemanticValidator
}

// NewTurnProcessor creates a new turn processor
func NewTurnProcessor() *TurnProcessor {
	return &TurnProcessor{
		registry: validation.NewOrderParserRegistry(),
	}
}

// ProcessTurn executes the complete pipeline: raw orders → parsed orders → resolution → new state
func (tp *TurnProcessor) ProcessTurn(gameState *game.GameState) (*game.GameState, error) {
	// Initialize resolver and semantic validator with current board state
	tp.resolver = validation.NewProvinceResolver(gameState.Board)
	tp.semanticValidator = validation.NewSemanticValidator(gameState.Board)

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
		// Use the new semantic validator (includes unit existence, ownership, adjacency, etc.)
		if err := tp.semanticValidator.ValidateOrder(order, "", gameState); err != nil {
			errors = append(errors, fmt.Errorf("order %s: %w", order.ID, err))
			continue
		}

		// Additional ownership validation (semantic validator checks unit existence but not ownership)
		unit := gameState.Board.GetUnit(order.From)
		if unit != nil && unit.Owner != order.Owner {
			errors = append(errors, fmt.Errorf("order %s: unit at %s belongs to %s, not %s", order.ID, order.From, unit.Owner, order.Owner))
			continue
		}

		validOrders = append(validOrders, order)
	}

	return validOrders, errors
}

// Note: validateOrder and validateMovementOrder removed - now handled by semantic validator

// Note: areAdjacent method removed - now handled by semantic validator

// validateMoveAdjacency validates move adjacency with convoy awareness
// This is used by convoy logic and will be replaced when convoy system is implemented
func (tp *TurnProcessor) validateMoveAdjacency(order *game.Order, gameState *game.GameState) error {
	// Use semantic validator's adjacency logic
	unit := gameState.Board.GetUnit(order.From)
	if unit == nil {
		return fmt.Errorf("no unit found at %s", order.From)
	}

	// For fleets, always check normal adjacency (fleets cannot be convoyed)
	if unit.Type == game.Fleet {
		if err := tp.semanticValidator.ValidateAdjacency(order.From, order.To, order.FromCoast, order.ToCoast, unit.Type); err != nil {
			return fmt.Errorf("cannot move from %s to %s: %v", order.From, order.To, err)
		}
		return nil
	}

	// For armies, check if convoy orders exist that could enable this move
	if unit.Type == game.Army {
		// First check normal adjacency
		if err := tp.semanticValidator.ValidateAdjacency(order.From, order.To, order.FromCoast, order.ToCoast, unit.Type); err == nil {
			return nil // Adjacent move is always valid
		}

		// Not adjacent - check if convoy orders exist for this move
		if tp.hasConvoyOrders(order.From, order.To, gameState) {
			return nil // Convoy orders exist - let resolution engine validate the path
		}

		// No convoy orders and not adjacent
		return fmt.Errorf("cannot move from %s to %s: not adjacent and no convoy orders", order.From, order.To)
	}

	// Unknown unit type
	return fmt.Errorf("unknown unit type: %v", unit.Type)
}

// hasConvoyOrders checks if there are convoy orders that could enable the given army move
func (tp *TurnProcessor) hasConvoyOrders(from, to string, gameState *game.GameState) bool {
	fmt.Printf("🔍 hasConvoyOrders checking for convoy from %s to %s\n", from, to)
	for _, orders := range gameState.RawOrders {
		for _, rawOrder := range orders {
			// Parse the order to check if it's a convoy for this move
			tokens := validation.Tokenize(rawOrder)
			if len(tokens) < 5 {
				continue
			}

			fmt.Printf("  Checking order: %s (tokens: %d)\n", rawOrder, len(tokens))
			for i, token := range tokens {
				fmt.Printf("    [%d]: %s\n", i, token.Value)
			}

			// Check for convoy order format: "f province c from - to"
			if len(tokens) >= 6 &&
				(tokens[1].Value == "c" || tokens[1].Value == "convoy" || tokens[1].Value == "convoys") &&
				tokens[2].Value == from && tokens[4].Value == to {
				fmt.Printf("  ✅ Found convoy match (format 1)\n")
				return true
			}
			if len(tokens) >= 7 &&
				(tokens[2].Value == "c" || tokens[2].Value == "convoy" || tokens[2].Value == "convoys") &&
				tokens[3].Value == from && tokens[5].Value == to {
				fmt.Printf("  ✅ Found convoy match (format 2)\n")
				return true
			}
			// Check for format: "f province1 province2 convoys a from - to"
			if len(tokens) >= 8 &&
				(tokens[3].Value == "c" || tokens[3].Value == "convoy" || tokens[3].Value == "convoys") &&
				tokens[5].Value == from && tokens[7].Value == to {
				fmt.Printf("  ✅ Found convoy match (format 3)\n")
				return true
			}
		}
	}
	fmt.Printf("  ❌ No convoy orders found for %s to %s\n", from, to)
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

// resolveOrders executes the recursive resolution algorithm
func (tp *TurnProcessor) resolveOrders(orders []*game.Order, board *game.Board) ([]resolution.OrderOutcome, error) {
	fmt.Printf("🔍 Resolving %d orders\n", len(orders))
	for i, order := range orders {
		switch order.Type {
		case game.Move:
			fmt.Printf("  %d: %s %s %s -> %s\n", i, order.Owner, order.UnitType, order.From, order.To)
		case game.Convoy:
			fmt.Printf("  %d: %s %s %s C %s -> %s\n", i, order.Owner, order.UnitType, order.From, order.ConvoyTarget, order.To)
		case game.Support:
			if order.SupportDestination != "" {
				fmt.Printf("  %d: %s %s %s S %s -> %s\n", i, order.Owner, order.UnitType, order.From, order.SupportTarget, order.SupportDestination)
			} else {
				fmt.Printf("  %d: %s %s %s S %s\n", i, order.Owner, order.UnitType, order.From, order.SupportTarget)
			}
		case game.Hold:
			fmt.Printf("  %d: %s %s %s H\n", i, order.Owner, order.UnitType, order.From)
		default:
			fmt.Printf("  %d: %s %s %s (unknown)\n", i, order.Owner, order.UnitType, order.From)
		}
	}

	// Convert game.Order to resolution.Order
	resolutionOrders := make([]resolution.Order, len(orders))
	for i, order := range orders {
		resolutionOrders[i] = convertToResolutionOrder(order)
	}

	// Create adjudicator and resolve
	adj := resolution.NewAdjudicator(resolutionOrders)
	adjResults := adj.ResolveAll()

	// Convert results to OrderOutcome for pipeline compatibility
	outcomes := make([]resolution.OrderOutcome, len(adjResults))
	for i, result := range adjResults {
		outcomes[i] = resolution.OrderOutcome{
			Result:      getResultString(orders[i], result.Success),
			Dislodged:   result.Dislodged,
			Destination: result.Destination,
		}
		fmt.Printf("  %d: %s\n", i, outcomes[i].Result)
	}

	return outcomes, nil
}

// convertToResolutionOrder converts a game.Order to resolution.Order
func convertToResolutionOrder(order *game.Order) resolution.Order {
	var orderType resolution.OrderType
	switch order.Type {
	case game.Move:
		orderType = resolution.Move
	case game.Support:
		orderType = resolution.Support
	case game.Convoy:
		orderType = resolution.Convoy
	case game.Hold:
		orderType = resolution.Hold
	default:
		orderType = resolution.Hold
	}

	// Set auxiliary field and destination based on order type
	var auxiliary string
	var destination string

	switch order.Type {
	case game.Move:
		destination = order.To
	case game.Support:
		if order.SupportDestination != "" {
			// Support move: "A Berlin S A Munich -> Silesia"
			auxiliary = order.SupportTarget + " -> " + order.SupportDestination
		} else {
			// Support hold: "A Berlin S A Munich"
			auxiliary = order.SupportTarget
		}
		// Support orders don't move, so no destination
	case game.Convoy:
		auxiliary = order.ConvoyTarget + " -> " + order.To
		// Convoy orders don't move, so no destination
	case game.Hold:
		// Hold orders don't move, so no destination
	}

	return resolution.Order{
		Unit:        string(order.UnitType) + " " + order.From,
		Type:        orderType,
		Source:      order.From,
		Destination: destination,
		Auxiliary:   auxiliary,
		Owner:       string(order.Owner),
	}
}

// getResultString converts success/failure to appropriate result string
func getResultString(order *game.Order, success bool) string {
	if success {
		switch order.Type {
		case game.Move:
			return resolution.MoveSuccess
		case game.Support:
			return resolution.SupportSuccess
		case game.Convoy:
			return resolution.ConvoySuccess
		case game.Hold:
			return resolution.HoldSuccess
		}
	} else {
		switch order.Type {
		case game.Move:
			return resolution.MoveBounced
		case game.Support:
			return resolution.SupportCut
		case game.Convoy:
			return resolution.ConvoyDisrupted
		case game.Hold:
			return resolution.HoldSuccess // Hold can't really fail
		}
	}
	return "unknown"
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

	// FIXED: Capture all units before applying moves to handle circular movement correctly
	originalUnits := make(map[string]*game.Unit)
	for province, unit := range oldState.Board.Units {
		if unit != nil {
			// Create a copy of the unit
			unitCopy := *unit
			originalUnits[province] = &unitCopy
		}
	}

	// FIXED: Apply moves in three phases to prevent order-dependent board corruption
	// Phase 0: Handle dislodgements from successful moves first
	for i, order := range orders {
		if i >= len(results) {
			continue
		}
		result := results[i]
		if order.Type == game.Move && result.Result == resolution.MoveSuccess {
			// Check if there's a unit at the destination that gets dislodged
			targetUnit := newState.Board.GetUnit(order.To)
			if targetUnit != nil {
				dislodgedUnits = append(dislodgedUnits, targetUnit)
				newState.Board.RemoveUnit(order.To)
			}
		}
	}

	// Phase 1: Collect all moves (successful and failed) for atomic processing
	var allMoves []struct {
		order  *game.Order
		result resolution.OrderOutcome
		unit   *game.Unit
	}

	// First pass: collect all move orders and their results
	for i, order := range orders {
		if i >= len(results) {
			continue
		}

		result := results[i]

		switch order.Type {
		case game.Move:
			// Collect ALL moves (successful and failed) for atomic processing
			unit := originalUnits[order.From]
			if unit != nil {
				allMoves = append(allMoves, struct {
					order  *game.Order
					result resolution.OrderOutcome
					unit   *game.Unit
				}{order, result, unit})
			}
		case game.Hold:
			if err := tp.applyHoldResult(newState, order, result, &dislodgedUnits); err != nil {
				return nil, err
			}
		case game.Support:
			// Support orders don't directly change unit positions, but can be dislodged
			if result.Dislodged {
				unit := newState.Board.GetUnit(order.From)
				if unit != nil {
					dislodgedUnits = append(dislodgedUnits, unit)
					newState.Board.RemoveUnit(order.From)
				}
			}
		case game.Convoy:
			// Convoy orders don't directly change unit positions, but can be dislodged
			if result.Dislodged {
				unit := newState.Board.GetUnit(order.From)
				if unit != nil {
					dislodgedUnits = append(dislodgedUnits, unit)
					newState.Board.RemoveUnit(order.From)
				}
			}
		}
	}

	// Phase 2: Remove all units that are moving successfully from their sources
	fmt.Printf("🔍 Removing units from sources\n")
	for _, move := range allMoves {
		if move.result.Result == resolution.MoveSuccess {
			fmt.Printf("  ✅ %s %s: %s -> %s\n", move.order.Owner, move.order.UnitType, move.order.From, move.order.To)
			newState.Board.RemoveUnit(move.order.From)
		}
	}

	// Phase 3: Place all successfully moving units at their destinations
	fmt.Printf("🔍 Placing units at destinations\n")
	for _, move := range allMoves {
		if move.result.Result == resolution.MoveSuccess {
			destination := move.order.To
			if move.result.Destination != "" {
				destination = move.result.Destination
			}

			// Update unit location
			move.unit.Province = destination
			move.unit.Coast = move.order.ToCoast

			// Place unit at destination
			if err := newState.Board.PlaceUnit(move.unit); err != nil {
				fmt.Printf("  ❌ Failed to place %s %s at %s: %v\n", move.unit.Owner, move.unit.Type, destination, err)
				return nil, fmt.Errorf("failed to place unit at %s: %w", destination, err)
			}
			fmt.Printf("  ✅ Placed %s %s at %s\n", move.unit.Owner, move.unit.Type, destination)
		} else {
			// Handle failed moves: unit stays in place
			// For failed moves, the unit should already be in its original position
			// We just need to ensure it's there and handle any dislodgement logic

			// Ensure the unit is at its original position (it should be unless removed by successful move)
			if newState.Board.GetUnit(move.order.From) == nil {
				// Unit was removed (shouldn't happen for failed moves), put it back
				unitCopy := *move.unit // Create a copy
				if err := newState.Board.PlaceUnit(&unitCopy); err != nil {
					return nil, fmt.Errorf("failed to restore unit at %s: %w", move.order.From, err)
				}
			}
		}
	}

	// Advance to next phase based on dislodged units
	newState.AdvanceToNextPhase(dislodgedUnits)

	return newState, nil
}

// applyMoveResultWithOriginalUnits applies the result of a move order using original unit positions
func (tp *TurnProcessor) applyMoveResultWithOriginalUnits(
	state *game.GameState,
	order *game.Order,
	result resolution.OrderOutcome,
	dislodgedUnits *[]*game.Unit,
	allOrders []*game.Order,
	originalUnits map[string]*game.Unit,
) error {
	// FIXED: Use original unit from before any moves were applied
	unit := originalUnits[order.From]
	if unit == nil {
		return fmt.Errorf("unit not found at %s for move order", order.From)
	}

	return tp.applyMoveResultInternal(state, order, result, dislodgedUnits, allOrders, unit)
}

// applyMoveResult applies the result of a move order (legacy function for non-circular moves)
func (tp *TurnProcessor) applyMoveResult(
	state *game.GameState,
	order *game.Order,
	result resolution.OrderOutcome,
	dislodgedUnits *[]*game.Unit,
	allOrders []*game.Order,
) error {
	unit := state.Board.GetUnit(order.From)
	if unit == nil {
		return fmt.Errorf("unit not found at %s for move order", order.From)
	}

	return tp.applyMoveResultInternal(state, order, result, dislodgedUnits, allOrders, unit)
}

// applyMoveResultInternal contains the actual move application logic
func (tp *TurnProcessor) applyMoveResultInternal(
	state *game.GameState,
	order *game.Order,
	result resolution.OrderOutcome,
	dislodgedUnits *[]*game.Unit,
	allOrders []*game.Order,
	unit *game.Unit,
) error {

	switch result.Result {
	case resolution.MoveSuccess:
		// Move succeeded - relocate unit
		fmt.Printf("🔍 Applying MoveSuccess: %s %s %s->%s, result.Destination=%s\n",
			unit.Owner, unit.Type, order.From, order.To, result.Destination)
		fmt.Printf("  Unit before move: %+v\n", unit)
		state.Board.RemoveUnit(order.From)
		// Use Destination from result if set (for circular movements), otherwise use original To
		destination := order.To
		if result.Destination != "" {
			destination = result.Destination
		}
		fmt.Printf("  Moving unit to: %s\n", destination)

		// In circular movement, destination might be occupied by a unit that's also moving
		// Check if the existing unit at destination is also moving this turn
		existingUnit := state.Board.GetUnit(destination)
		if existingUnit != nil {
			// Check if this unit has a move order in the current turn
			hasMovingOrder := false
			for _, checkOrder := range allOrders {
				if checkOrder.Type == game.Move && checkOrder.From == destination {
					hasMovingOrder = true
					break
				}
			}

			if hasMovingOrder {
				fmt.Printf("  🔄 Destination %s occupied by %s %s that's also moving, removing\n",
					destination, existingUnit.Owner, existingUnit.Type)
				state.Board.RemoveUnit(destination)
			} else {
				// Unit at destination is not moving, this should be a dislodgement
				fmt.Printf("  ⚔️ Destination %s occupied by %s %s that's not moving\n",
					destination, existingUnit.Owner, existingUnit.Type)
				// For now, still remove it - proper dislodgement logic would go here
				state.Board.RemoveUnit(destination)
			}
		}
		unit.Province = destination
		unit.Coast = order.ToCoast
		fmt.Printf("  Unit after move: %+v\n", unit)

		if err := state.Board.PlaceUnit(unit); err != nil {
			return fmt.Errorf("failed to place unit at %s: %w", destination, err)
		}
		fmt.Printf("  Unit placed. Board now has at %s: %+v\n", destination, state.Board.GetUnit(destination))
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
