package validation

import (
	"diplomacy-cli/backend/internal/game"
	"fmt"
)

type SemanticValidator struct {
	board    *game.Board
	resolver *ProvinceResolver
}

func NewSemanticValidator(board *game.Board) *SemanticValidator {
	return &SemanticValidator{
		board:    board,
		resolver: NewProvinceResolver(board),
	}
}

type SemanticError struct {
	OrderText string
	Message   string
	Type      string
}

func (e SemanticError) Error() string {
	return fmt.Sprintf("semantic error in order '%s': %s", e.OrderText, e.Message)
}

func (sv *SemanticValidator) ValidateOrder(order *game.Order, orderText string, gameState *game.GameState) error {
	switch order.Type {
	case game.Move:
		return sv.validateMove(order, orderText, gameState)
	case game.Hold:
		return sv.validateHold(order, orderText, gameState)
	case game.Support:
		return sv.validateSupport(order, orderText, gameState)
	case game.Convoy:
		return sv.validateConvoy(order, orderText, gameState)
	default:
		return SemanticError{
			OrderText: orderText,
			Message:   fmt.Sprintf("unsupported order type: %v", order.Type),
			Type:      "unsupported_order_type",
		}
	}
}

func (sv *SemanticValidator) validateMove(order *game.Order, orderText string, gameState *game.GameState) error {
	// Check if unit exists at source location
	unit, exists := gameState.Board.Units[order.From]
	if !exists {
		return SemanticError{
			OrderText: orderText,
			Message:   fmt.Sprintf("no unit found at %s", order.From),
			Type:      "unit_not_found",
		}
	}

	// Check unit type matches if specified
	if order.UnitType != "" && unit.Type != order.UnitType {
		return SemanticError{
			OrderText: orderText,
			Message:   fmt.Sprintf("unit type mismatch: expected %v, found %v", order.UnitType, unit.Type),
			Type:      "unit_type_mismatch",
		}
	}

	// Check if moving to same location
	if order.From == order.To {
		return SemanticError{
			OrderText: orderText,
			Message:   "cannot move to same location",
			Type:      "move_to_same_location",
		}
	}

	// Check unit type can move to destination type
	if err := sv.validateUnitCanMoveToDestination(unit.Type, order.To); err != nil {
		return SemanticError{
			OrderText: orderText,
			Message:   err.Error(),
			Type:      "invalid_destination_for_unit_type",
		}
	}

	// For now, skip adjacency validation for armies (convoy-aware validation is complex)
	// The resolution engine will handle convoy validation
	if unit.Type == game.Fleet {
		if err := sv.ValidateAdjacency(order.From, order.To, order.FromCoast, order.ToCoast, unit.Type); err != nil {
			return SemanticError{
				OrderText: orderText,
				Message:   err.Error(),
				Type:      "adjacency_violation",
			}
		}
	}

	return nil
}

func (sv *SemanticValidator) validateHold(order *game.Order, orderText string, gameState *game.GameState) error {
	// Check if unit exists at location
	unit, exists := gameState.Board.Units[order.From]
	if !exists {
		return SemanticError{
			OrderText: orderText,
			Message:   fmt.Sprintf("no unit found at %s", order.From),
			Type:      "unit_not_found",
		}
	}

	// Check unit type matches if specified
	if order.UnitType != "" && unit.Type != order.UnitType {
		return SemanticError{
			OrderText: orderText,
			Message:   fmt.Sprintf("unit type mismatch: expected %v, found %v", order.UnitType, unit.Type),
			Type:      "unit_type_mismatch",
		}
	}

	return nil
}

func (sv *SemanticValidator) validateSupport(order *game.Order, orderText string, gameState *game.GameState) error {
	// Check if unit exists at location
	unit, exists := gameState.Board.Units[order.From]
	if !exists {
		return SemanticError{
			OrderText: orderText,
			Message:   fmt.Sprintf("no unit found at %s", order.From),
			Type:      "unit_not_found",
		}
	}

	// Check unit type matches if specified
	if order.UnitType != "" && unit.Type != order.UnitType {
		return SemanticError{
			OrderText: orderText,
			Message:   fmt.Sprintf("unit type mismatch: expected %v, found %v", order.UnitType, unit.Type),
			Type:      "unit_type_mismatch",
		}
	}

	// For support move orders, validate that the supporting unit can reach the destination
	if order.SupportDestination != "" {
		if err := sv.ValidateAdjacency(order.From, order.SupportDestination, order.FromCoast, order.ToCoast, unit.Type); err != nil {
			return SemanticError{
				OrderText: orderText,
				Message:   fmt.Sprintf("supporting unit cannot reach destination: %v", err),
				Type:      "support_unreachable",
			}
		}
	} else {
		// For support hold orders (no destination), check if supporting itself
		if order.From == order.SupportTarget {
			// Check if unit can "reach" its own position (should fail for same position)
			if err := sv.ValidateAdjacency(order.From, order.SupportTarget, order.FromCoast, "", unit.Type); err != nil {
				return SemanticError{
					OrderText: orderText,
					Message:   fmt.Sprintf("supporting unit cannot reach destination: %v", err),
					Type:      "support_unreachable",
				}
			}
		}
	}

	return nil
}

func (sv *SemanticValidator) validateConvoy(order *game.Order, orderText string, gameState *game.GameState) error {
	// Check if unit exists at location
	unit, exists := gameState.Board.Units[order.From]
	if !exists {
		return SemanticError{
			OrderText: orderText,
			Message:   fmt.Sprintf("no unit found at %s", order.From),
			Type:      "unit_not_found",
		}
	}

	// Check unit type matches if specified
	if order.UnitType != "" && unit.Type != order.UnitType {
		return SemanticError{
			OrderText: orderText,
			Message:   fmt.Sprintf("unit type mismatch: expected %v, found %v", order.UnitType, unit.Type),
			Type:      "unit_type_mismatch",
		}
	}

	// Only fleets can convoy
	if unit.Type != game.Fleet {
		return SemanticError{
			OrderText: orderText,
			Message:   "only fleets can convoy",
			Type:      "invalid_convoy_unit",
		}
	}

	// Convoy orders should have a target (the unit being convoyed)
	if order.ConvoyTarget == "" {
		return SemanticError{
			OrderText: orderText,
			Message:   "convoy order missing target",
			Type:      "missing_convoy_target",
		}
	}

	return nil
}

func (sv *SemanticValidator) validateUnitCanMoveToDestination(unitType game.UnitType, destination string) error {
	destProv, exists := sv.resolver.ValidateProvince(destination)
	if !exists {
		return fmt.Errorf("destination province %s does not exist", destination)
	}

	switch unitType {
	case game.Army:
		if destProv.Type == game.Sea {
			return fmt.Errorf("army cannot move to sea province %s", destination)
		}
	case game.Fleet:
		if destProv.Type == game.Land {
			// Check if the land province has a coast (fleets can move to coastal land provinces)
			boardProv := sv.board.GetProvince(destination)
			if boardProv == nil {
				return fmt.Errorf("province %s not found in board", destination)
			}
			// Check if province has fleet neighbors (indicating it's coastal)
			if len(boardProv.FleetNeighbors) == 0 && len(boardProv.CoastNeighbors) == 0 {
				return fmt.Errorf("fleet cannot move to land province %s", destination)
			}
		}
	}

	return nil
}

func (sv *SemanticValidator) ValidateAdjacency(from, to, fromCoast, toCoast string, unitType game.UnitType) error {
	fromProv := sv.board.GetProvince(from)
	if fromProv == nil {
		return fmt.Errorf("province %s not found", from)
	}

	// Handle coast-specific adjacency for provinces with multiple coasts
	if fromCoast != "" {
		coastNeighbors, hasCoast := fromProv.CoastNeighbors[fromCoast]
		if !hasCoast {
			return fmt.Errorf("province %s does not have coast %s", from, fromCoast)
		}

		// Check if destination is reachable from this coast
		targetWithCoast := to
		if toCoast != "" {
			targetWithCoast = to + "/" + toCoast
		}

		for _, neighbor := range coastNeighbors {
			if neighbor == targetWithCoast || neighbor == to {
				return nil
			}
		}
		return fmt.Errorf("provinces %s/%s and %s are not adjacent", from, fromCoast, to)
	}

	// Check unit-type specific adjacency
	var neighbors []string
	switch unitType {
	case game.Army:
		neighbors = fromProv.ArmyNeighbors
	case game.Fleet:
		neighbors = fromProv.FleetNeighbors
	}

	for _, neighbor := range neighbors {
		if neighbor == to {
			return nil
		}
	}

	return fmt.Errorf("provinces %s and %s are not adjacent for %s", from, to, unitType)
}
