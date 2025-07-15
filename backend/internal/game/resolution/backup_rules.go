package resolution

// backup_rules.go implements the backup rules for paradoxical situations
// as described in the adjudication article.

// applyBackupRule analyzes the cycle and applies the appropriate backup rule.
func (adj *Adjudicator) applyBackupRule() {
	if len(adj.cycle) == 0 {
		return
	}

	// Analyze the cycle to determine which backup rule to apply
	if adj.isCircularMovement() {
		adj.applyCircularMovementRule()
	} else if adj.isConvoyParadox() {
		adj.applyConvoyParadoxRule()
	} else {
		// Default: treat as circular movement if all moves
		if adj.allMovesInCycle() {
			adj.applyCircularMovementRule()
		} else {
			adj.applyDefaultResolution()
		}
	}
}

// isCircularMovement checks if the cycle represents a circular movement.
// A circular movement is when units move in a circle: A→B, B→C, C→A
func (adj *Adjudicator) isCircularMovement() bool {
	if len(adj.cycle) < 2 {
		return false
	}

	// Check if all orders in cycle are moves
	for _, order := range adj.cycle {
		if order.Type != Move {
			return false
		}
	}

	// Check if moves form a circle
	// For each move, check if its destination is the source of another move in the cycle
	for _, order := range adj.cycle {
		foundNext := false
		for _, otherOrder := range adj.cycle {
			if otherOrder != order && otherOrder.Source == order.Destination {
				foundNext = true
				break
			}
		}
		if !foundNext {
			return false
		}
	}

	return true
}

// isConvoyParadox checks if the cycle involves a convoy paradox.
// A convoy paradox occurs when a convoy's success depends on itself.
func (adj *Adjudicator) isConvoyParadox() bool {
	// Check if any order in the cycle is a convoy
	hasConvoy := false
	for _, order := range adj.cycle {
		if order.Type == Convoy {
			hasConvoy = true
			break
		}
	}

	return hasConvoy
}

// allMovesInCycle checks if all orders in the cycle are moves
func (adj *Adjudicator) allMovesInCycle() bool {
	for _, order := range adj.cycle {
		if order.Type != Move {
			return false
		}
	}
	return true
}

// applyCircularMovementRule implements the circular movement backup rule:
// All moves in the circular movement succeed.
func (adj *Adjudicator) applyCircularMovementRule() {
	for _, order := range adj.cycle {
		if order.Type == Move {
			order.setResolution(true)
		}
	}
}

// applyConvoyParadoxRule implements the Szykman rule:
// In a convoy paradox, the convoy fails.
func (adj *Adjudicator) applyConvoyParadoxRule() {
	for _, order := range adj.cycle {
		if order.Type == Convoy {
			order.setResolution(false)
		} else if order.Type == Move {
			// Move that depends on the failed convoy also fails
			order.setResolution(false)
		}
	}
}

// applyDefaultResolution applies a default resolution when the cycle type is unclear
func (adj *Adjudicator) applyDefaultResolution() {
	for _, order := range adj.cycle {
		order.setResolution(false)
	}
}
