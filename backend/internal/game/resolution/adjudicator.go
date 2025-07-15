package resolution

import (
	"fmt"
	"strings"
)

// Adjudicator holds the state for a single turn's resolution using the
// recursive dependency resolution algorithm from the adjudication article.
type Adjudicator struct {
	orders           map[string]*Order // Map from unit location to its order
	orderedOrders    []*Order          // Orders in their original input order
	cycle            []*Order          // Tracks orders in a potential dependency cycle
	recursionHits    int               // Number of times we've hit recursion
	uncertain        bool              // Whether the current resolution path is uncertain
	recursionDepth   int               // Current recursion depth (safety limit)
	convoyPathFinder *ConvoyPathFinder // Graph-based convoy path finder
}

const maxRecursionDepth = 100 // Safety limit to prevent infinite recursion

// NewAdjudicator creates a new adjudicator for a set of orders.
func NewAdjudicator(orders []Order) *Adjudicator {
	orderMap := make(map[string]*Order)
	orderedOrders := make([]*Order, len(orders))

	for i := range orders {
		order := &orders[i]
		orderMap[order.Source] = order
		orderedOrders[i] = order
	}

	return &Adjudicator{
		orders:        orderMap,
		orderedOrders: orderedOrders,
		cycle:         make([]*Order, 0),
	}
}

// ResolveAll resolves all orders and returns the results.
func (adj *Adjudicator) ResolveAll() []AdjudicationResult {
	results := make([]AdjudicationResult, 0, len(adj.orderedOrders))

	// Reset all orders before resolution
	for _, order := range adj.orderedOrders {
		order.reset()
	}

	// Resolve each order in original input order
	for _, order := range adj.orderedOrders {
		if !order.IsResolved() {
			// Reset adjudicator state before each top-level resolution
			adj.cycle = adj.cycle[:0]
			adj.recursionHits = 0
			adj.uncertain = false
			adj.recursionDepth = 0 // Reset recursion depth

			adj.resolve(order, true) // Start with optimistic resolution
		}

		results = append(results, AdjudicationResult{
			Order:       *order,
			Success:     order.Resolution(),
			Reason:      adj.getResolutionReason(order),
			Destination: order.Destination,
			Dislodged:   false, // TODO: Detect dislodgements
		})
	}

	return results
}

// resolve is the core recursive resolution function implementing the
// partial information algorithm from the adjudication article.
func (adj *Adjudicator) resolve(order *Order, optimistic bool) bool {
	// Safety check: prevent infinite recursion
	adj.recursionDepth++
	defer func() { adj.recursionDepth-- }()

	if adj.recursionDepth > maxRecursionDepth {
		fmt.Printf("⚠️ Maximum recursion depth exceeded for %s, returning optimistic=%t\n", order.String(), optimistic)
		adj.uncertain = true
		return optimistic
	}

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

	// Check for cycle detection - FIXED: Prevent infinite recursion
	if order.isVisited {
		// We've found a cycle!
		fmt.Printf("🔄 Cycle detected at %s (cycle length: %d)\n", order.String(), len(adj.cycle))
		adj.cycle = append(adj.cycle, order)
		adj.recursionHits++
		adj.uncertain = true
		// CRITICAL FIX: Return immediately to prevent infinite recursion
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

	// Pessimistic run (only if optimistic was uncertain)
	pesResult := optResult
	if adj.uncertain {
		adj.uncertain = false // Reset for pessimistic run
		pesResult = adj.adjudicate(order, false)
		fmt.Printf("    Pessimistic result: %t\n", pesResult)
	}

	// Backtrack - ALWAYS reset visited flag
	order.isVisited = false

	// If both runs agree, we have a definitive result
	if optResult == pesResult {
		order.setResolution(optResult)
		fmt.Printf("🔍 Applying %s: %s (opt=%t, pes=%t)\n",
			map[bool]string{true: "MoveSuccess", false: "MoveFailed"}[optResult], order.String(), optResult, pesResult)
		// Clean up cycle data from this branch
		adj.cycle = adj.cycle[:oldCycleLen]
		adj.recursionHits = oldRecursionHits
		return order.resolution
	}

	// Results disagree - we have a paradox
	// Check if this order is in the cycle we detected
	orderInCycle := false
	for _, o := range adj.cycle {
		if o == order {
			orderInCycle = true
			adj.recursionHits--
			break
		}
	}

	// If we've retreated to the ancestor of the whole cycle, apply backup rules
	if adj.recursionHits == oldRecursionHits {
		fmt.Printf("🔧 Applying backup rules (recursionHits: %d, oldRecursionHits: %d)\n", adj.recursionHits, oldRecursionHits)
		// Apply backup rule on all orders in the cycle
		adj.applyBackupRule()
		adj.cycle = adj.cycle[:oldCycleLen]

		// FIXED: Don't recurse again - backup rule has resolved the orders
		// Return the resolution that was set by the backup rule
		if order.isResolved {
			fmt.Printf("🔧 Order %s resolved by backup rule: %t\n", order.String(), order.resolution)
			return order.resolution
		}
		// If backup rule didn't resolve this specific order, use optimistic default
		fmt.Printf("🔧 Order %s not resolved by backup rule, using optimistic: %t\n", order.String(), optimistic)
		return optimistic
	}

	// We're returning from recursion but not at the cycle ancestor yet
	if !orderInCycle {
		adj.cycle = append(adj.cycle, order)
	}

	// Default to optimistic for uncertain outcomes
	return optimistic
}

// hasValidPath checks if a move order has a valid path to its destination.
// According to DATC 5.B.4: PATH is successful when the unit can directly move
// to the destination OR is convoyed and there is a chain of adjacent fleets
// from origin to destination each with a matching and successful CONVOY order.
func (adj *Adjudicator) hasValidPath(order *Order, optimistic bool) bool {
	if order.Type != Move {
		return false
	}

	fmt.Printf("🛤️ hasValidPath: checking %s\n", order.String())

	// Check if this is a convoy move by looking for convoy orders
	convoyOrders := adj.findConvoyOrders(order.Source, order.Destination)

	if len(convoyOrders) == 0 {
		// No convoy orders - this should be a direct move
		// For now, assume direct moves are valid (adjacency was checked during parsing)
		fmt.Printf("  ✅ Direct move, PATH valid\n")
		return true
	}

	// This is a convoy move - validate the convoy chain
	result := adj.validateConvoyChain(order.Source, order.Destination, convoyOrders, optimistic)
	fmt.Printf("  🚢 Convoy move, PATH valid: %t\n", result)
	return result
}

// findConvoyOrders finds all convoy orders that could support a move from source to destination.
func (adj *Adjudicator) findConvoyOrders(source, destination string) []*Order {
	var convoyOrders []*Order

	fmt.Printf("🔍 findConvoyOrders: looking for convoy from %s to %s\n", source, destination)
	for _, order := range adj.orders {
		if order.Type == Convoy {
			fmt.Printf("  Found convoy order: %s (auxiliary: '%s')\n", order.String(), order.Auxiliary)
			// Parse convoy auxiliary field: "source -> destination"
			// Expected format: "bulgaria -> trieste" or similar
			if order.Auxiliary == source+" -> "+destination {
				fmt.Printf("  ✅ Convoy matches!\n")
				convoyOrders = append(convoyOrders, order)
			}
		}
	}

	fmt.Printf("🔍 Found %d convoy orders for %s -> %s\n", len(convoyOrders), source, destination)
	return convoyOrders
}

// validateConvoyChain validates that there is a valid chain of convoy fleets.
// Uses graph-based pathfinding to find valid convoy routes.
func (adj *Adjudicator) validateConvoyChain(source, destination string, convoyOrders []*Order, optimistic bool) bool {
	// Use enhanced convoy chain validation
	return adj.validateConvoyChainEnhanced(source, destination, convoyOrders, optimistic)
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
	// First check: Self-dislodgement prevention
	// A support order fails if it would help dislodge the supporting player's own unit
	if adj.wouldHelpDislodgeOwnUnit(order) {
		fmt.Printf("  ❌ Support cut: would help dislodge own unit\n")
		return false
	}

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

// wouldHelpDislodgeOwnUnit checks if a support order would help dislodge the supporting player's own unit
func (adj *Adjudicator) wouldHelpDislodgeOwnUnit(supportOrder *Order) bool {
	if supportOrder.Type != Support {
		return false
	}

	// Find the move being supported
	var supportedMove *Order

	// Parse the auxiliary field to extract the source location
	// Auxiliary format: "source -> destination" for support move, or just "source" for support hold
	var supportedSource string
	if strings.Contains(supportOrder.Auxiliary, " -> ") {
		// Support move: "bulgaria -> constantinople"
		parts := strings.Split(supportOrder.Auxiliary, " -> ")
		if len(parts) >= 1 {
			supportedSource = strings.TrimSpace(parts[0])
		}
	} else {
		// Support hold: "bulgaria"
		supportedSource = strings.TrimSpace(supportOrder.Auxiliary)
	}

	for _, order := range adj.orders {
		if order.Type == Move && order.Source == supportedSource {
			supportedMove = order
			break
		}
	}

	if supportedMove == nil {
		return false // No move to support
	}

	// Check if the supported move would dislodge a unit owned by the same player as the supporter
	targetLocation := supportedMove.Destination
	for _, order := range adj.orders {
		// Look for a unit at the target location that belongs to the same owner as the supporter
		if order.Source == targetLocation && order.Owner == supportOrder.Owner {
			fmt.Printf("  🚫 Support from %s would help dislodge own unit at %s\n",
				supportOrder.Source, targetLocation)
			return true
		}
	}

	return false
}
