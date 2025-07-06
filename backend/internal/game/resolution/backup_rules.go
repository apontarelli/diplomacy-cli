package resolution

import (
	"fmt"
)

// backup_rules.go implements the backup rules for paradoxical situations
// as described in the adjudication article.

// applyBackupRule analyzes the cycle and applies the appropriate backup rule.
func (adj *Adjudicator) applyBackupRule() {
	if len(adj.cycle) == 0 {
		return
	}

	fmt.Printf("🔄 Applying backup rule to cycle of %d orders\n", len(adj.cycle))
	for i, order := range adj.cycle {
		fmt.Printf("  [%d]: %s\n", i, order.String())
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
			fmt.Printf("⚠️ Unknown cycle type, applying default resolution\n")
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
	fmt.Printf("✅ Applying circular movement rule: all moves succeed\n")

	for _, order := range adj.cycle {
		if order.Type == Move {
			order.setResolution(true)
			fmt.Printf("  ✓ %s succeeds (circular movement)\n", order.String())
		}
	}
}

// applyConvoyParadoxRule implements the Szykman rule:
// In a convoy paradox, the convoy fails.
func (adj *Adjudicator) applyConvoyParadoxRule() {
	fmt.Printf("❌ Applying convoy paradox rule (Szykman): convoy fails\n")

	for _, order := range adj.cycle {
		if order.Type == Convoy {
			order.setResolution(false)
			fmt.Printf("  ✗ %s fails (convoy paradox)\n", order.String())
		} else if order.Type == Move {
			// Move that depends on the failed convoy also fails
			order.setResolution(false)
			fmt.Printf("  ✗ %s fails (convoy disrupted)\n", order.String())
		}
	}
}

// applyDefaultResolution applies a default resolution when the cycle type is unclear
func (adj *Adjudicator) applyDefaultResolution() {
	fmt.Printf("🔧 Applying default resolution: all orders fail\n")

	for _, order := range adj.cycle {
		order.setResolution(false)
		fmt.Printf("  ✗ %s fails (default)\n", order.String())
	}
}
