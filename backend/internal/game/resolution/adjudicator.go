package resolution

import (
	"fmt"
)

// Adjudicator holds the state for a single turn's resolution using the
// recursive dependency resolution algorithm from the adjudication article.
type Adjudicator struct {
	orders        map[string]*Order // Map from unit location to its order
	cycle         []*Order          // Tracks orders in a potential dependency cycle
	recursionHits int               // Number of times we've hit recursion
	uncertain     bool              // Whether the current resolution path is uncertain
}

// NewAdjudicator creates a new adjudicator for a set of orders.
func NewAdjudicator(orders []Order) *Adjudicator {
	orderMap := make(map[string]*Order)
	for i := range orders {
		order := &orders[i]
		orderMap[order.Source] = order
	}

	return &Adjudicator{
		orders: orderMap,
		cycle:  make([]*Order, 0),
	}
}

// ResolveAll resolves all orders and returns the results.
func (adj *Adjudicator) ResolveAll() []InternalOrderResult {
	results := make([]InternalOrderResult, 0, len(adj.orders))

	// Reset all orders before resolution
	for _, order := range adj.orders {
		order.reset()
	}

	// Resolve each order
	for _, order := range adj.orders {
		if !order.IsResolved() {
			adj.resolve(order, true) // Start with optimistic resolution
		}

		results = append(results, InternalOrderResult{
			Order:   *order,
			Success: order.Resolution(),
			Reason:  adj.getResolutionReason(order),
		})
	}

	return results
}

// resolve is the core recursive resolution function implementing the
// partial information algorithm from the adjudication article.
func (adj *Adjudicator) resolve(order *Order, optimistic bool) bool {
	// If already resolved, return cached result
	if order.isResolved {
		return order.resolution
	}

	// Check if we've already determined this order is in an unresolvable cycle
	for _, cycleOrder := range adj.cycle {
		if cycleOrder == order {
			adj.uncertain = true
			return optimistic
		}
	}

	// Check for cycle detection
	if order.isVisited {
		// We've found a cycle!
		fmt.Printf("🔄 Cycle detected at %s\n", order.String())
		adj.cycle = append(adj.cycle, order)
		adj.recursionHits++
		adj.uncertain = true
		return optimistic
	}

	// Mark as visited for cycle detection
	order.isVisited = true

	// Store current state to restore later
	oldCycleLen := len(adj.cycle)
	oldRecursionHits := adj.recursionHits

	// Optimistic run
	adj.uncertain = false
	optResult := adj.adjudicate(order, true)

	// Pessimistic run (only if optimistic was uncertain and succeeded)
	pesResult := optResult
	if adj.uncertain && optResult {
		adj.uncertain = false // Reset for pessimistic run
		pesResult = adj.adjudicate(order, false)
	}

	// Backtrack
	order.isVisited = false

	// If both runs agree, we have a definitive result
	if optResult == pesResult {
		order.setResolution(optResult)

		// Clean up cycle data from this branch
		adj.cycle = adj.cycle[:oldCycleLen]
		adj.recursionHits = oldRecursionHits
		return order.resolution
	}

	// Results disagree - we have a paradox
	// Apply backup rules if we have a complete cycle
	if len(adj.cycle) > oldCycleLen {
		adj.applyBackupRule()
		return order.resolution
	}

	// Default to optimistic for uncertain outcomes
	return optimistic
}

// adjudicate contains the specific Diplomacy rules for each order type.
// This is where the game logic lives.
func (adj *Adjudicator) adjudicate(order *Order, optimistic bool) bool {
	switch order.Type {
	case Move:
		return adj.adjudicateMove(order, optimistic)
	case Support:
		return adj.adjudicateSupport(order, optimistic)
	case Convoy:
		return adj.adjudicateConvoy(order, optimistic)
	case Hold:
		return adj.adjudicateHold(order, optimistic)
	default:
		return false
	}
}

// adjudicateMove determines if a move order succeeds.
func (adj *Adjudicator) adjudicateMove(order *Order, optimistic bool) bool {
	// A move succeeds if:
	// 1. Attack strength > hold strength of destination
	// 2. Attack strength >= prevent strength of any competing moves

	attackStrength := adj.calculateAttackStrength(order, optimistic)
	holdStrength := adj.calculateHoldStrength(order.Destination, !optimistic)

	// Debug output for circular movement
	if len(adj.cycle) > 0 {
		fmt.Printf("🔍 Move %s: attack=%d, hold=%d, optimistic=%t\n",
			order.String(), attackStrength, holdStrength, optimistic)
	}

	// Check if we can overcome the hold strength
	if attackStrength <= holdStrength {
		return false
	}

	// Check against competing moves to the same destination
	for _, otherOrder := range adj.orders {
		if otherOrder != order && otherOrder.Type == Move && otherOrder.Destination == order.Destination {
			preventStrength := adj.calculatePreventStrength(otherOrder, !optimistic)
			if attackStrength <= preventStrength {
				return false
			}
		}
	}

	return true
}

// adjudicateSupport determines if a support order succeeds.
func (adj *Adjudicator) adjudicateSupport(order *Order, optimistic bool) bool {
	// A support succeeds if the supporting unit is not dislodged
	// Check if any unit is attacking this supporter
	for _, otherOrder := range adj.orders {
		if otherOrder.Type == Move && otherOrder.Destination == order.Source {
			// Someone is attacking our supporter
			// The attack succeeds if we resolve it pessimistically (bad for us)
			if adj.resolve(otherOrder, !optimistic) {
				return false // Support is cut
			}
		}
	}
	return true
}

// adjudicateConvoy determines if a convoy order succeeds.
func (adj *Adjudicator) adjudicateConvoy(order *Order, optimistic bool) bool {
	// A convoy succeeds if:
	// 1. The convoying fleet is not dislodged
	// 2. There is a valid convoy chain to the destination

	// Check if the convoying fleet is dislodged
	for _, otherOrder := range adj.orders {
		if otherOrder.Type == Move && otherOrder.Destination == order.Source {
			if adj.resolve(otherOrder, !optimistic) {
				return false // Convoy is disrupted
			}
		}
	}

	// TODO: Implement convoy chain validation
	return true
}

// adjudicateHold determines if a hold order succeeds.
func (adj *Adjudicator) adjudicateHold(order *Order, optimistic bool) bool {
	// A hold always succeeds (it's just staying in place)
	// The question is whether the unit gets dislodged by attacks
	return true
}

// getResolutionReason returns a human-readable explanation for the order result.
func (adj *Adjudicator) getResolutionReason(order *Order) string {
	if order.Resolution() {
		return "Success"
	}
	return "Failed"
}
