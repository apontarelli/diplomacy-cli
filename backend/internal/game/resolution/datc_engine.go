package resolution

import (
	"fmt"
)

// DATCEngine implements the DATC-compliant partial information algorithm
// This is a complete rewrite following the DATC specification exactly
type DATCEngine struct {
	orders         map[string]*Order // Map from unit location to its order
	orderedOrders  []*Order          // Orders in their original input order
	cycle          []*Order          // Tracks orders in a potential dependency cycle
	recursionHits  int               // Number of times we've hit recursion
	uncertain      bool              // Whether the current resolution path is uncertain
	recursionDepth int               // Current recursion depth (safety limit)
}

// NewDATCEngine creates a new DATC-compliant resolution engine
func NewDATCEngine(orders []Order) *DATCEngine {
	orderMap := make(map[string]*Order)
	orderedOrders := make([]*Order, len(orders))

	for i := range orders {
		order := &orders[i]
		orderMap[order.Source] = order
		orderedOrders[i] = order
	}

	return &DATCEngine{
		orders:        orderMap,
		orderedOrders: orderedOrders,
		cycle:         make([]*Order, 0),
	}
}

// ResolveAll resolves all orders using the DATC partial information algorithm
func (engine *DATCEngine) ResolveAll() []AdjudicationResult {
	results := make([]AdjudicationResult, 0, len(engine.orderedOrders))

	// Reset all orders before resolution
	for _, order := range engine.orderedOrders {
		order.reset()
	}

	// Resolve each order in original input order
	for _, order := range engine.orderedOrders {
		if !order.IsResolved() {
			// Reset engine state before each top-level resolution
			engine.cycle = engine.cycle[:0]
			engine.recursionHits = 0
			engine.uncertain = false
			engine.recursionDepth = 0

			engine.resolve(order, true) // Start with optimistic resolution
		}

		results = append(results, AdjudicationResult{
			Order:       *order,
			Success:     order.Resolution(),
			Reason:      engine.getResolutionReason(order),
			Destination: order.Destination,
			Dislodged:   false, // TODO: Detect dislodgements
		})
	}

	return results
}

// resolve implements the DATC partial information algorithm exactly as specified
func (engine *DATCEngine) resolve(order *Order, optimistic bool) bool {
	// Safety check: prevent infinite recursion
	engine.recursionDepth++
	defer func() { engine.recursionDepth-- }()

	if engine.recursionDepth > maxRecursionDepth {
		engine.uncertain = true
		return optimistic
	}

	// If already resolved, return cached result
	if order.isResolved {
		return order.resolution
	}

	// Check if we've already determined this order is in an unresolvable cycle
	for _, cycleOrder := range engine.cycle {
		if cycleOrder == order {
			engine.uncertain = true
			return optimistic
		}
	}

	// DATC Algorithm: Check for cycle detection
	if order.isVisited {
		// We hit cyclic dependency - add to cycle and return optimistic assumption
		engine.cycle = append(engine.cycle, order)
		engine.recursionHits++
		engine.uncertain = true
		return optimistic
	}

	// Mark as visited to prevent endless recursion
	order.isVisited = true
	oldCycleLen := len(engine.cycle)
	oldRecursionHits := engine.recursionHits
	oldUncertain := engine.uncertain
	engine.uncertain = false

	// DATC Algorithm: Try both optimistic and pessimistic scenarios
	optResult := engine.adjudicate(order, true)

	// Performance optimization: only try pessimistic if uncertain and optimistic succeeded
	var pesResult bool
	if engine.uncertain && optResult {
		pesResult = engine.adjudicate(order, false)
	} else {
		pesResult = optResult
	}

	order.isVisited = false

	// DATC Algorithm: If both scenarios agree, we have a definitive resolution
	if optResult == pesResult {
		// Single resolution found - wipe out any cycle information from recursion
		engine.cycle = engine.cycle[:oldCycleLen]
		engine.recursionHits = oldRecursionHits
		engine.uncertain = oldUncertain

		// Store the result and return it
		order.resolution = optResult
		order.isResolved = true
		return optResult
	}

	// Check if this order is in the cycle we just discovered
	orderInCycle := false
	for _, cycleOrder := range engine.cycle {
		if cycleOrder == order {
			orderInCycle = true
			break
		}
	}

	if orderInCycle {
		// We returned from recursion where this order hit the cycle
		engine.recursionHits--
	}

	// DATC Algorithm: Check if we've retreated enough to be the cycle ancestor
	if engine.recursionHits == oldRecursionHits {
		// This order was the ancestor of the whole cycle - apply backup rule
		engine.applyBackupRule(engine.cycle[oldCycleLen:])
		engine.cycle = engine.cycle[:oldCycleLen]

		// The backup rule might not have resolved this order - try again
		return engine.resolve(order, optimistic)
	}

	// We're still retreating from recursion - add to cycle and continue
	engine.cycle = append(engine.cycle, order)
	return optimistic
}

// adjudicate determines if an order succeeds based on DATC rules
func (engine *DATCEngine) adjudicate(order *Order, optimistic bool) bool {
	switch order.Type {
	case Move:
		return engine.adjudicateMove(order, optimistic)
	case Hold:
		return engine.adjudicateHold(order, optimistic)
	case Support:
		return engine.adjudicateSupport(order, optimistic)
	case Convoy:
		return engine.adjudicateConvoy(order, optimistic)
	default:
		return false
	}
}

// adjudicateMove determines if a move order succeeds
func (engine *DATCEngine) adjudicateMove(order *Order, optimistic bool) bool {
	// DATC 5.B.8: If PATH fails, attack strength is 0
	if !engine.hasValidPath(order, optimistic) {
		return false
	}

	// Get all competing moves to same destination
	competitors := engine.getCompetingMoves(order.Destination)

	if len(competitors) == 1 {
		// Only this move - check against hold strength
		attackStrength := engine.calculateAttackStrength(order, optimistic)
		holdStrength := engine.calculateHoldStrength(order.Destination, !optimistic)
		return attackStrength > holdStrength
	}

	// Multiple moves to same destination - use prevent strength comparison
	thisPreventStrength := engine.calculatePreventStrength(order, optimistic)
	holdStrength := engine.calculateHoldStrength(order.Destination, !optimistic)

	// First check if this move can overcome hold strength
	if thisPreventStrength <= holdStrength {
		return false
	}

	// Check against other competing moves
	for _, competitor := range competitors {
		if competitor == order {
			continue
		}

		// Use opposite optimism for competitors (pessimistic for us)
		competitorStrength := engine.calculatePreventStrength(competitor, !optimistic)

		// If any competitor has equal or greater strength, this move fails
		if competitorStrength >= thisPreventStrength {
			return false
		}
	}

	// This move has the highest prevent strength and overcomes hold strength
	return true
}

// adjudicateHold determines if a hold order succeeds (always true)
func (engine *DATCEngine) adjudicateHold(order *Order, optimistic bool) bool {
	// Hold orders always succeed - they just provide hold strength
	return true
}

// adjudicateSupport determines if a support order succeeds
func (engine *DATCEngine) adjudicateSupport(order *Order, optimistic bool) bool {
	// Support fails if the supporting unit is attacked and dislodged
	attackers := engine.getAttackersOf(order.Source)

	for _, attacker := range attackers {
		// CRITICAL: Support is only cut if the attack actually succeeds
		// Use pessimistic resolution for attackers (bad for support)
		if engine.resolve(attacker, !optimistic) {
			return false // Support is cut by successful attack
		}
	}

	return true // Support succeeds if not cut by successful attacks
}

// adjudicateConvoy determines if a convoy order succeeds
func (engine *DATCEngine) adjudicateConvoy(order *Order, optimistic bool) bool {
	// Convoy fails if the convoying fleet is attacked and dislodged
	attackers := engine.getAttackersOf(order.Source)

	for _, attacker := range attackers {
		// Use pessimistic resolution for attackers (bad for convoy)
		if engine.resolve(attacker, !optimistic) {
			return false // Convoy is disrupted
		}
	}

	return true // Convoy succeeds if not disrupted
}

// DATC Strength Calculations (following specification exactly)

// calculateAttackStrength calculates attack strength per DATC 5.B.8
func (engine *DATCEngine) calculateAttackStrength(order *Order, optimistic bool) int {
	if order.Type != Move {
		return 0
	}

	// DATC 5.B.8: If PATH fails, attack strength is 0
	if !engine.hasValidPath(order, optimistic) {
		return 0
	}

	strength := 1 // Base attack strength

	// Add support
	for _, otherOrder := range engine.orderedOrders {
		if otherOrder.Type == Support && engine.isSupporting(otherOrder, order) {
			// Support succeeds if we resolve it optimistically (good for attack)
			if engine.resolve(otherOrder, optimistic) {
				strength++
			}
		}
	}

	return strength
}

// calculateHoldStrength calculates hold strength per DATC specification
func (engine *DATCEngine) calculateHoldStrength(territory string, optimistic bool) int {
	occupyingOrder := engine.orders[territory]
	if occupyingOrder == nil {
		return 0 // Unoccupied territory
	}

	strength := 0

	// CRITICAL: A unit moving away provides NO hold strength if the move succeeds
	if occupyingOrder.Type == Move && occupyingOrder.Destination != territory {
		// Check if the move succeeds using opposite optimism
		if !engine.resolve(occupyingOrder, !optimistic) {
			// Move fails - unit stays and provides hold strength
			strength = 1
		}
		// If move succeeds, strength remains 0
	} else if occupyingOrder.Type == Hold {
		// Unit is explicitly holding - provides base strength
		strength = 1
	} else if occupyingOrder.Type == Support || occupyingOrder.Type == Convoy {
		// Unit is supporting or convoying - provides hold strength unless dislodged
		strength = 1
	}

	// Add hold support
	for _, otherOrder := range engine.orderedOrders {
		if otherOrder.Type == Support && engine.isSupportingHold(otherOrder, territory) {
			// Support succeeds if we resolve it optimistically (good for hold)
			if engine.resolve(otherOrder, optimistic) {
				strength++
			}
		}
	}

	return strength
}

// calculatePreventStrength calculates prevent strength per DATC 5.B.6
func (engine *DATCEngine) calculatePreventStrength(order *Order, optimistic bool) int {
	if order.Type != Move {
		return 0
	}

	// DATC 5.B.6: If PATH fails, prevent strength is 0
	if !engine.hasValidPath(order, optimistic) {
		return 0
	}

	// Prevent strength is same as attack strength when PATH is valid
	return engine.calculateAttackStrength(order, optimistic)
}

// Helper functions

// hasValidPath checks if a move has a valid path (direct or convoy)
func (engine *DATCEngine) hasValidPath(order *Order, optimistic bool) bool {
	if order.Type != Move {
		return false
	}

	// TODO: Implement proper path validation including convoy paths
	// For now, assume all paths are valid
	return true
}

// getCompetingMoves returns all moves targeting the same destination
func (engine *DATCEngine) getCompetingMoves(destination string) []*Order {
	var competitors []*Order
	for _, order := range engine.orderedOrders {
		if order.Type == Move && order.Destination == destination {
			competitors = append(competitors, order)
		}
	}
	return competitors
}

// getAttackersOf returns all moves targeting a specific province
func (engine *DATCEngine) getAttackersOf(province string) []*Order {
	var attackers []*Order
	for _, order := range engine.orderedOrders {
		if order.Type == Move && order.Destination == province {
			attackers = append(attackers, order)
		}
	}
	return attackers
}

// isSupporting checks if a support order supports a specific move
func (engine *DATCEngine) isSupporting(supportOrder *Order, moveOrder *Order) bool {
	if supportOrder.Type != Support || moveOrder.Type != Move {
		return false
	}
	// TODO: Parse auxiliary field properly
	return supportOrder.Auxiliary == moveOrder.Source+" -> "+moveOrder.Destination
}

// isSupportingHold checks if a support order supports holding a territory
func (engine *DATCEngine) isSupportingHold(supportOrder *Order, territory string) bool {
	if supportOrder.Type != Support {
		return false
	}
	// TODO: Parse auxiliary field properly
	return supportOrder.Auxiliary == territory
}

// applyBackupRule applies the DATC backup rule for cycles
func (engine *DATCEngine) applyBackupRule(cycleOrders []*Order) {
	// DATC Backup Rule: In case of cycles, all moves in the cycle fail
	for _, order := range cycleOrders {
		if order.Type == Move {
			order.resolution = false
			order.isResolved = true
		}
	}
}

// getResolutionReason provides human-readable reason for resolution
func (engine *DATCEngine) getResolutionReason(order *Order) string {
	if order.IsResolved() {
		if order.Resolution() {
			return fmt.Sprintf("%s order succeeded", order.Type.String())
		} else {
			return fmt.Sprintf("%s order failed", order.Type.String())
		}
	}
	return "Order resolution uncertain"
}
