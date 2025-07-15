package resolution

// strength.go implements the four types of strength calculations
// from the adjudication article: Attack, Hold, Defend, and Prevent.

// calculateAttackStrength calculates the attack strength of a moving unit.
// Attack strength = 1 (base) + number of supporting units
// DATC 5.B.8: If PATH fails, attack strength is 0
func (adj *Adjudicator) calculateAttackStrength(order *Order, optimistic bool) int {
	if order.Type != Move {
		return 0
	}

	// DATC 5.B.8: If the PATH of the move order fails, then the ATTACK STRENGTH is zero
	if !adj.hasValidPath(order, optimistic) {
		return 0
	}

	strength := 1 // Base strength

	// Add support from other units
	for _, otherOrder := range adj.orders {
		if otherOrder.Type == Support && adj.isSupporting(otherOrder, order) {
			// Support succeeds if we resolve it optimistically (good for us)
			if adj.resolve(otherOrder, optimistic) {
				strength++
			}
		}
	}

	return strength
}

// calculateHoldStrength calculates the hold strength of a territory.
// Hold strength = 1 (if occupied) + number of supporting units
func (adj *Adjudicator) calculateHoldStrength(territory string, optimistic bool) int {
	// Check if territory is occupied
	occupyingOrder := adj.orders[territory]
	if occupyingOrder == nil {
		return 0 // Unoccupied territory has no hold strength
	}

	// CRITICAL: A unit that is moving away provides NO hold strength
	// BUT only if the move actually succeeds. We need to check move success first.
	if occupyingOrder.Type == Move && occupyingOrder.Destination != territory {
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
		}

		var strength int
		if moveSucceeds {
			strength = 0 // Moving unit provides no hold strength
		} else {
			strength = 1 // Failed move means unit stays and provides hold strength
		}
		// Add support for holding (regardless of whether unit is moving or staying)
		for _, otherOrder := range adj.orders {
			if otherOrder.Type == Support && adj.isSupportingHold(otherOrder, territory) {
				// Support succeeds if we resolve it optimistically (good for hold)
				if adj.resolve(otherOrder, optimistic) {
					strength++
				}
			}
		}
		return strength
	}

	// Unit is holding (not moving), provides base strength of 1
	strength := 1 // Base strength for holding unit

	// Add support for the holding unit
	for _, otherOrder := range adj.orders {
		if otherOrder.Type == Support && adj.isSupportingHold(otherOrder, territory) {
			// Support succeeds if we resolve it optimistically (good for hold)
			if adj.resolve(otherOrder, optimistic) {
				strength++
			}
		}
	}

	return strength
}

// calculateDefendStrength calculates the defend strength for head-to-head battles.
// This is used when two units are moving to each other's territories.
func (adj *Adjudicator) calculateDefendStrength(order *Order, optimistic bool) int {
	if order.Type != Move {
		return 0
	}

	// Defend strength is like attack strength but for the defensive position
	return adj.calculateAttackStrength(order, optimistic)
}

// calculatePreventStrength calculates the prevent strength of a competing move.
// This determines which move succeeds when multiple units move to the same destination.
// DATC 5.B.6: If PATH fails, prevent strength is 0
func (adj *Adjudicator) calculatePreventStrength(order *Order, optimistic bool) int {
	if order.Type != Move {
		return 0
	}

	// DATC 5.B.6: If the PATH of the move order fails, then the PREVENT STRENGTH is 0
	if !adj.hasValidPath(order, optimistic) {
		return 0
	}

	// Prevent strength is the same as attack strength (when PATH is valid)
	strength := 1 // Base strength

	// Add support from other units
	for _, otherOrder := range adj.orders {
		if otherOrder.Type == Support && adj.isSupporting(otherOrder, order) {
			// Support succeeds if we resolve it optimistically (good for us)
			if adj.resolve(otherOrder, optimistic) {
				strength++
			}
		}
	}

	return strength
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
