package resolution

import (
	"fmt"
	"strings"

	"diplomacy-cli/backend/internal/game"
)

// RetreatProcessor handles retreat phase logic
type RetreatProcessor struct {
	board *game.Board
}

// NewRetreatProcessor creates a new retreat processor
func NewRetreatProcessor(board *game.Board) *RetreatProcessor {
	return &RetreatProcessor{
		board: board,
	}
}

// ProcessRetreats handles the retreat phase for dislodged units
func (rp *RetreatProcessor) ProcessRetreats(dislodgedUnits []DislodgedUnit, retreatOrders []RetreatOrder) []RetreatResult {

	results := make([]RetreatResult, 0)

	// Create map of retreat orders by unit for quick lookup
	ordersByUnit := make(map[string]RetreatOrder)
	for _, order := range retreatOrders {
		ordersByUnit[order.Unit] = order
	}

	// Track retreat destinations to detect conflicts
	retreatDestinations := make(map[string][]string) // destination -> list of units retreating there

	// First pass: validate all retreat orders and collect destinations
	validRetreats := make(map[string]RetreatOrder)
	for _, dislodged := range dislodgedUnits {
		order, hasOrder := ordersByUnit[dislodged.Unit]

		if !hasOrder {
			// No retreat order given - unit is disbanded
			results = append(results, RetreatResult{
				Order: RetreatOrder{
					Unit:  dislodged.Unit,
					From:  dislodged.DislodgedFrom,
					Owner: dislodged.Owner,
				},
				Success:   false,
				Reason:    "No retreat order given",
				Disbanded: true,
			})
			continue
		}

		// Validate the retreat order
		if err := rp.ValidateRetreatOrder(order, dislodged); err != nil {
			results = append(results, RetreatResult{
				Order:     order,
				Success:   false,
				Reason:    err.Error(),
				Disbanded: true,
			})
			continue
		}

		// Valid retreat - add to tracking
		validRetreats[dislodged.Unit] = order
		retreatDestinations[order.Destination] = append(retreatDestinations[order.Destination], dislodged.Unit)
	}

	// Second pass: resolve conflicts and apply retreats
	for _, order := range validRetreats {
		unitsToDestination := retreatDestinations[order.Destination]

		if len(unitsToDestination) > 1 {
			// Conflict - all units retreating to same destination are disbanded
			results = append(results, RetreatResult{
				Order:     order,
				Success:   false,
				Reason:    fmt.Sprintf("Retreat conflict: %d units attempting to retreat to %s", len(unitsToDestination), order.Destination),
				Disbanded: true,
			})
		} else {
			// Successful retreat
			results = append(results, RetreatResult{
				Order:     order,
				Success:   true,
				Reason:    "Retreat successful",
				Disbanded: false,
			})
		}
	}

	return results
}

// ValidateRetreatOrder validates a single retreat order against DATC rules
func (rp *RetreatProcessor) ValidateRetreatOrder(order RetreatOrder, dislodged DislodgedUnit) error {
	// Rule 1: Must retreat from the province where dislodged
	if order.From != dislodged.DislodgedFrom {
		return fmt.Errorf("unit must retreat from %s, not %s", dislodged.DislodgedFrom, order.From)
	}

	// Rule 2: Destination must be in the list of possible retreats
	validDestination := false
	for _, possible := range dislodged.PossibleRetreats {
		if possible == order.Destination {
			validDestination = true
			break
		}
	}

	if !validDestination {
		return fmt.Errorf("invalid retreat destination %s, valid options: %v", order.Destination, dislodged.PossibleRetreats)
	}

	// Rule 3: Cannot retreat to attacker's origin
	if order.Destination == dislodged.AttackerOrigin {
		return fmt.Errorf("cannot retreat to attacker's origin province %s", dislodged.AttackerOrigin)
	}

	// Rule 4: Destination must be vacant (checked during possible retreat calculation)
	// This is already handled in CalculatePossibleRetreats

	return nil
}

// CalculatePossibleRetreats determines valid retreat destinations for a dislodged unit
func (rp *RetreatProcessor) CalculatePossibleRetreats(unit *game.Unit, attackerOrigin string, gameState *game.GameState) []string {
	possibleRetreats := make([]string, 0)

	// Get adjacent provinces based on unit type
	var adjacentProvinces []string
	province := rp.board.Provinces[unit.Province]

	if unit.Type == game.Army {
		adjacentProvinces = province.ArmyNeighbors
	} else if unit.Type == game.Fleet {
		// For fleets, check coast-specific neighbors if on a coast
		if unit.Coast != "" && len(province.CoastNeighbors[unit.Coast]) > 0 {
			adjacentProvinces = province.CoastNeighbors[unit.Coast]
		} else {
			adjacentProvinces = province.FleetNeighbors
		}
	}

	for _, adjacent := range adjacentProvinces {
		// Rule 1: Cannot retreat to attacker's origin
		if adjacent == attackerOrigin {
			continue
		}

		// Rule 2: Destination must be vacant
		if rp.board.Units[adjacent] != nil {
			continue
		}

		// Rule 3: Cannot retreat to province vacated by standoff
		if rp.wasVacatedByStandoff(adjacent, gameState) {
			continue
		}

		possibleRetreats = append(possibleRetreats, adjacent)
	}

	return possibleRetreats
}

// wasVacatedByStandoff checks if a province was vacated due to a standoff in the movement phase
func (rp *RetreatProcessor) wasVacatedByStandoff(province string, gameState *game.GameState) bool {
	// This would need to track standoff information from the movement phase
	// For now, return false - this can be enhanced later with proper standoff tracking
	return false
}

// CreateDislodgedUnit creates a DislodgedUnit from adjudication results
func (rp *RetreatProcessor) CreateDislodgedUnit(unit *game.Unit, attackerOrigin string, gameState *game.GameState) DislodgedUnit {
	possibleRetreats := rp.CalculatePossibleRetreats(unit, attackerOrigin, gameState)

	return DislodgedUnit{
		Unit:             fmt.Sprintf("%s %s", unit.Type, unit.Province),
		Owner:            string(unit.Owner),
		DislodgedFrom:    unit.Province,
		AttackerOrigin:   attackerOrigin,
		PossibleRetreats: possibleRetreats,
	}
}

// ApplyRetreatResults updates the game board based on retreat results
func (rp *RetreatProcessor) ApplyRetreatResults(results []RetreatResult) {
	for _, result := range results {
		if result.Success && !result.Disbanded {
			// Move unit to retreat destination
			if unit := rp.board.Units[result.Order.From]; unit != nil {
				// Remove from old position
				delete(rp.board.Units, result.Order.From)

				// Add to new position
				unit.Province = result.Order.Destination
				if result.Order.Coast != "" {
					unit.Coast = result.Order.Coast
				}
				rp.board.Units[result.Order.Destination] = unit
			}
		} else {
			// Unit is disbanded - remove from board
			delete(rp.board.Units, result.Order.From)
		}
	}
}

// parseUnitFromString extracts unit info from string like "A Berlin"
func (rp *RetreatProcessor) parseUnitFromString(unitStr string) (game.UnitType, string) {
	parts := strings.Split(unitStr, " ")
	if len(parts) >= 2 {
		unitType := game.Army
		if parts[0] == "F" {
			unitType = game.Fleet
		}
		province := strings.Join(parts[1:], " ")
		return unitType, province
	}
	return game.Army, unitStr
}
