package resolution

import (
	"fmt"
)

// strength.go implements the four types of strength calculations
// from the adjudication article: Attack, Hold, Defend, and Prevent.
// Enhanced with detailed reasoning tracking for DATC compliance.

// StrengthCalculation captures the detailed breakdown of a strength calculation
type StrengthCalculation struct {
	BaseStrength    int                   `json:"base_strength"`
	SupportStrength int                   `json:"support_strength"`
	TotalStrength   int                   `json:"total_strength"`
	SupportDetails  []SupportContribution `json:"support_details"`
	Reasoning       []string              `json:"reasoning"`
	PathValid       bool                  `json:"path_valid"`
	PathReason      string                `json:"path_reason"`
}

// SupportContribution tracks individual support contributions
type SupportContribution struct {
	SupportingUnit string `json:"supporting_unit"`
	SupportType    string `json:"support_type"`
	IsSuccessful   bool   `json:"is_successful"`
	Reason         string `json:"reason"`
	Strength       int    `json:"strength"`
}

// calculateAttackStrength calculates the attack strength of a moving unit.
// Attack strength = 1 (base) + number of supporting units
// DATC 5.B.8: If PATH fails, attack strength is 0
func (adj *Adjudicator) calculateAttackStrength(order *Order, optimistic bool) int {
	calc := adj.calculateAttackStrengthDetailed(order, optimistic)
	return calc.TotalStrength
}

// calculateAttackStrengthDetailed provides detailed attack strength calculation with reasoning
func (adj *Adjudicator) calculateAttackStrengthDetailed(order *Order, optimistic bool) StrengthCalculation {
	calc := StrengthCalculation{
		BaseStrength:    0,
		SupportStrength: 0,
		SupportDetails:  make([]SupportContribution, 0),
		Reasoning:       make([]string, 0),
		PathValid:       true,
	}

	if order.Type != Move {
		calc.Reasoning = append(calc.Reasoning, "Non-move orders have no attack strength")
		calc.TotalStrength = 0
		return calc
	}

	// DATC 5.B.8: If the PATH of the move order fails, then the ATTACK STRENGTH is zero
	calc.PathValid = adj.hasValidPath(order, optimistic)
	if !calc.PathValid {
		calc.PathReason = "Move path is invalid or blocked"
		calc.Reasoning = append(calc.Reasoning, "DATC 5.B.8: Attack strength is 0 when PATH fails")
		calc.TotalStrength = 0
		return calc
	}

	calc.BaseStrength = 1
	calc.Reasoning = append(calc.Reasoning, "Base attack strength: 1")

	// Add support from other units
	for _, otherOrder := range adj.orders {
		if otherOrder.Type == Support && adj.isSupporting(otherOrder, order) {
			contribution := SupportContribution{
				SupportingUnit: otherOrder.Source,
				SupportType:    "attack",
				Strength:       0,
			}

			// Support succeeds if we resolve it optimistically (good for us)
			if adj.resolve(otherOrder, optimistic) {
				contribution.IsSuccessful = true
				contribution.Strength = 1
				contribution.Reason = "Support successful"
				calc.SupportStrength++
				calc.Reasoning = append(calc.Reasoning,
					fmt.Sprintf("Support from %s: +1 strength", otherOrder.Source))
			} else {
				contribution.IsSuccessful = false
				contribution.Reason = "Support cut or failed"
				calc.Reasoning = append(calc.Reasoning,
					fmt.Sprintf("Support from %s failed: no strength contribution", otherOrder.Source))
			}

			calc.SupportDetails = append(calc.SupportDetails, contribution)
		}
	}

	calc.TotalStrength = calc.BaseStrength + calc.SupportStrength
	calc.Reasoning = append(calc.Reasoning,
		fmt.Sprintf("Total attack strength: %d (base) + %d (support) = %d",
			calc.BaseStrength, calc.SupportStrength, calc.TotalStrength))

	return calc
}

// calculateHoldStrength calculates the hold strength of a territory.
// Hold strength = 1 (if occupied) + number of supporting units
func (adj *Adjudicator) calculateHoldStrength(territory string, optimistic bool) int {
	calc := adj.calculateHoldStrengthDetailed(territory, optimistic)
	return calc.TotalStrength
}

// calculateHoldStrengthDetailed provides detailed hold strength calculation with reasoning
func (adj *Adjudicator) calculateHoldStrengthDetailed(territory string, optimistic bool) StrengthCalculation {
	calc := StrengthCalculation{
		BaseStrength:    0,
		SupportStrength: 0,
		SupportDetails:  make([]SupportContribution, 0),
		Reasoning:       make([]string, 0),
		PathValid:       true,
	}

	// Check if territory is occupied
	occupyingOrder := adj.orders[territory]
	if occupyingOrder == nil {
		calc.Reasoning = append(calc.Reasoning, "Unoccupied territory has no hold strength")
		calc.TotalStrength = 0
		return calc
	}

	calc.Reasoning = append(calc.Reasoning, fmt.Sprintf("Territory %s is occupied by %s", territory, occupyingOrder.Unit))

	// CRITICAL: A unit that is moving away provides NO hold strength
	// BUT only if the move actually succeeds. We need to check move success first.
	if occupyingOrder.Type == Move && occupyingOrder.Destination != territory {
		calc.Reasoning = append(calc.Reasoning, "Unit is attempting to move away - checking move success")

		// Check if the move actually succeeds before assuming zero hold strength
		// Use opposite optimism for move success (pessimistic for hold strength calculation)
		moveSucceeds := false

		// FIXED: Prevent infinite recursion by checking if we're already resolving this order
		if !occupyingOrder.isVisited {
			moveSucceeds = adj.resolve(occupyingOrder, !optimistic)
		} else {
			// FIXED: If we're in a cycle, mark as uncertain and add to cycle if not already there
			adj.uncertain = true // Mark as uncertain due to cycle

			// Add to cycle if not already present
			alreadyInCycle := false
			for _, cycleOrder := range adj.cycle {
				if cycleOrder == occupyingOrder {
					alreadyInCycle = true
					break
				}
			}
			if !alreadyInCycle {
				adj.cycle = append(adj.cycle, occupyingOrder)
				adj.recursionHits++
			}

			moveSucceeds = optimistic
			calc.Reasoning = append(calc.Reasoning, "Move resolution uncertain due to cycle - using optimistic assumption")
		}

		if moveSucceeds {
			calc.BaseStrength = 0
			calc.Reasoning = append(calc.Reasoning, "Move succeeds - unit provides no hold strength")
		} else {
			calc.BaseStrength = 1
			calc.Reasoning = append(calc.Reasoning, "Move fails - unit stays and provides hold strength: 1")
		}
	} else {
		// Unit is holding (not moving), provides base strength of 1
		calc.BaseStrength = 1
		calc.Reasoning = append(calc.Reasoning, "Unit is holding - base hold strength: 1")
	}

	// Add support for holding (regardless of whether unit is moving or staying)
	for _, otherOrder := range adj.orders {
		if otherOrder.Type == Support && adj.isSupportingHold(otherOrder, territory) {
			contribution := SupportContribution{
				SupportingUnit: otherOrder.Source,
				SupportType:    "hold",
				Strength:       0,
			}

			// Support succeeds if we resolve it optimistically (good for hold)
			if adj.resolve(otherOrder, optimistic) {
				contribution.IsSuccessful = true
				contribution.Strength = 1
				contribution.Reason = "Hold support successful"
				calc.SupportStrength++
				calc.Reasoning = append(calc.Reasoning,
					fmt.Sprintf("Hold support from %s: +1 strength", otherOrder.Source))
			} else {
				contribution.IsSuccessful = false
				contribution.Reason = "Hold support cut or failed"
				calc.Reasoning = append(calc.Reasoning,
					fmt.Sprintf("Hold support from %s failed: no strength contribution", otherOrder.Source))
			}

			calc.SupportDetails = append(calc.SupportDetails, contribution)
		}
	}

	calc.TotalStrength = calc.BaseStrength + calc.SupportStrength
	calc.Reasoning = append(calc.Reasoning,
		fmt.Sprintf("Total hold strength: %d (base) + %d (support) = %d",
			calc.BaseStrength, calc.SupportStrength, calc.TotalStrength))

	return calc
}

// calculateDefendStrength calculates the defend strength for head-to-head battles.
// This is used when two units are moving to each other's territories.
func (adj *Adjudicator) calculateDefendStrength(order *Order, optimistic bool) int {
	calc := adj.calculateDefendStrengthDetailed(order, optimistic)
	return calc.TotalStrength
}

// calculateDefendStrengthDetailed provides detailed defend strength calculation with reasoning
func (adj *Adjudicator) calculateDefendStrengthDetailed(order *Order, optimistic bool) StrengthCalculation {
	if order.Type != Move {
		return StrengthCalculation{
			BaseStrength:    0,
			SupportStrength: 0,
			TotalStrength:   0,
			SupportDetails:  make([]SupportContribution, 0),
			Reasoning:       []string{"Non-move orders have no defend strength"},
			PathValid:       false,
		}
	}

	// Defend strength is like attack strength but for the defensive position
	calc := adj.calculateAttackStrengthDetailed(order, optimistic)

	// Update reasoning to reflect this is defend strength
	for i, reason := range calc.Reasoning {
		if reason == "Base attack strength: 1" {
			calc.Reasoning[i] = "Base defend strength: 1"
		}
		if reason == fmt.Sprintf("Total attack strength: %d (base) + %d (support) = %d",
			calc.BaseStrength, calc.SupportStrength, calc.TotalStrength) {
			calc.Reasoning[i] = fmt.Sprintf("Total defend strength: %d (base) + %d (support) = %d",
				calc.BaseStrength, calc.SupportStrength, calc.TotalStrength)
		}
	}

	// Update support type for defend context
	for i := range calc.SupportDetails {
		calc.SupportDetails[i].SupportType = "defend"
	}

	return calc
}

// calculatePreventStrength calculates the prevent strength of a competing move.
// This determines which move succeeds when multiple units move to the same destination.
// DATC 5.B.6: If PATH fails, prevent strength is 0
func (adj *Adjudicator) calculatePreventStrength(order *Order, optimistic bool) int {
	calc := adj.calculatePreventStrengthDetailed(order, optimistic)
	return calc.TotalStrength
}

// calculatePreventStrengthDetailed provides detailed prevent strength calculation with reasoning
func (adj *Adjudicator) calculatePreventStrengthDetailed(order *Order, optimistic bool) StrengthCalculation {
	calc := StrengthCalculation{
		BaseStrength:    0,
		SupportStrength: 0,
		SupportDetails:  make([]SupportContribution, 0),
		Reasoning:       make([]string, 0),
		PathValid:       true,
	}

	if order.Type != Move {
		calc.Reasoning = append(calc.Reasoning, "Non-move orders have no prevent strength")
		calc.TotalStrength = 0
		return calc
	}

	// DATC 5.B.6: If the PATH of the move order fails, then the PREVENT STRENGTH is 0
	calc.PathValid = adj.hasValidPath(order, optimistic)
	if !calc.PathValid {
		calc.PathReason = "Move path is invalid or blocked"
		calc.Reasoning = append(calc.Reasoning, "DATC 5.B.6: Prevent strength is 0 when PATH fails")
		calc.TotalStrength = 0
		return calc
	}

	// Prevent strength is the same as attack strength (when PATH is valid)
	calc.BaseStrength = 1
	calc.Reasoning = append(calc.Reasoning, "Base prevent strength: 1")

	// Add support from other units
	for _, otherOrder := range adj.orders {
		if otherOrder.Type == Support && adj.isSupporting(otherOrder, order) {
			contribution := SupportContribution{
				SupportingUnit: otherOrder.Source,
				SupportType:    "prevent",
				Strength:       0,
			}

			// Support succeeds if we resolve it optimistically (good for us)
			if adj.resolve(otherOrder, optimistic) {
				contribution.IsSuccessful = true
				contribution.Strength = 1
				contribution.Reason = "Support successful"
				calc.SupportStrength++
				calc.Reasoning = append(calc.Reasoning,
					fmt.Sprintf("Support from %s: +1 strength", otherOrder.Source))
			} else {
				contribution.IsSuccessful = false
				contribution.Reason = "Support cut or failed"
				calc.Reasoning = append(calc.Reasoning,
					fmt.Sprintf("Support from %s failed: no strength contribution", otherOrder.Source))
			}

			calc.SupportDetails = append(calc.SupportDetails, contribution)
		}
	}

	calc.TotalStrength = calc.BaseStrength + calc.SupportStrength
	calc.Reasoning = append(calc.Reasoning,
		fmt.Sprintf("Total prevent strength: %d (base) + %d (support) = %d",
			calc.BaseStrength, calc.SupportStrength, calc.TotalStrength))

	return calc
}

// isSupporting checks if a support order is supporting a specific move.
func (adj *Adjudicator) isSupporting(supportOrder *Order, moveOrder *Order) bool {
	if supportOrder.Type != Support || moveOrder.Type != Move {
		return false
	}

	// Support format: "A Berlin S A Munich -> Silesia"
	// supportOrder.Source = "Berlin" (supporter location)
	// supportOrder.Auxiliary = "Munich -> Silesia" (what's being supported)

	// For now, simplified: check if auxiliary matches the move
	// TODO: Parse auxiliary field properly to extract source and destination
	return supportOrder.Auxiliary == moveOrder.Source+" -> "+moveOrder.Destination
}

// isSupportingHold checks if a support order is supporting a unit holding in place.
func (adj *Adjudicator) isSupportingHold(supportOrder *Order, territory string) bool {
	if supportOrder.Type != Support {
		return false
	}

	// Support for hold: "A Berlin S A Munich"
	// supportOrder.Auxiliary = "Munich" (territory being supported to hold)
	return supportOrder.Auxiliary == territory
}

// wouldCreateHeadToHead checks if two moves create a head-to-head battle.
func (adj *Adjudicator) wouldCreateHeadToHead(order1, order2 *Order) bool {
	if order1.Type != Move || order2.Type != Move {
		return false
	}

	// Head-to-head: A->B and B->A
	return order1.Source == order2.Destination && order1.Destination == order2.Source
}
