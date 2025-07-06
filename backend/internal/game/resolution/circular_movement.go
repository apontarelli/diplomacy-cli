package resolution

import (
	"diplomacy-cli/backend/internal/game"
	"fmt"
)

// detectCircularMovements finds and validates circular movements in the current orders
func (re *ResolutionEngine) detectCircularMovements() {
	fmt.Printf("🔍 detectCircularMovements called\n")

	// Reset circular movement state
	re.circularMovements = make([]CircularMovement, 0)
	re.protectedMoves = make(map[int]bool)

	// Build movement graph from valid moves
	graph := re.buildMovementGraph()

	// Debug: Print graph structure
	fmt.Printf("🔍 Movement Graph - Nodes: %d, Edges: %d\n", len(graph.Nodes), len(graph.Edges))
	for territory, node := range graph.Nodes {
		fmt.Printf("  Node: %s (order %d)\n", territory, node.OrderIndex)
	}
	for _, edge := range graph.Edges {
		fmt.Printf("  Edge: %s -> %s (order %d, convoy: %t)\n", edge.From, edge.To, edge.OrderIndex, edge.IsConvoy)
	}

	// Find cycles in the movement graph
	cycles := re.findCycles(graph)

	fmt.Printf("🔍 Detected %d cycles\n", len(cycles))
	for i, cycle := range cycles {
		fmt.Printf("  Cycle %d: territories %v, orders %v, hasConvoy: %t\n", i, cycle.Territories, cycle.OrderIndices, cycle.HasConvoy)
	}

	// Validate each cycle and mark valid ones as protected
	for _, cycle := range cycles {
		fmt.Printf("🔍 Validating cycle with orders %v\n", cycle.OrderIndices)
		if re.isValidCircularMovement(cycle) {
			// Set correct destinations for circular movement
			re.setCircularMovementDestinations(cycle)

			for _, orderIndex := range cycle.OrderIndices {
				re.protectedMoves[orderIndex] = true
				re.outcomes[orderIndex].IsCircular = true
			}
			re.circularMovements = append(re.circularMovements, cycle)
			fmt.Printf("✅ Valid circular movement detected: orders %v\n", cycle.OrderIndices)
		} else {
			fmt.Printf("❌ Invalid circular movement: orders %v\n", cycle.OrderIndices)
		}
	}
}

// buildMovementGraph creates a graph representation of all valid moves
func (re *ResolutionEngine) buildMovementGraph() *MovementGraph {
	graph := &MovementGraph{
		Nodes: make(map[string]*MovementNode),
		Edges: make([]*MovementEdge, 0),
	}

	// Add nodes and edges for all valid moves
	for i, order := range re.orders {
		if order.Type == game.Move && order.NewTerritory != order.OrigTerritory {
			// Add nodes for origin and destination
			if _, exists := graph.Nodes[order.OrigTerritory]; !exists {
				graph.AddNode(order.OrigTerritory, i)
			}
			if _, exists := graph.Nodes[order.NewTerritory]; !exists {
				graph.AddNode(order.NewTerritory, -1) // Destination node doesn't have an order
			}

			// Determine if this move uses convoy
			isConvoy := false
			if order.UnitType == game.Army {
				origin := NewProvinceCoast(order.OrigTerritory, order.FromCoast)
				dest := NewProvinceCoast(order.NewTerritory, order.ToCoast)
				convoyKey := ConvoyKey{Origin: origin, Destination: dest}
				_, hasConvoyPath := re.convoys[convoyKey]
				isConvoy = hasConvoyPath
			}

			// Add edge for this move
			graph.AddEdge(order.OrigTerritory, order.NewTerritory, i, isConvoy)
		}
	}

	return graph
}

// findCycles uses a simpler approach to find cycles in the movement graph
func (re *ResolutionEngine) findCycles(graph *MovementGraph) []CircularMovement {
	var cycles []CircularMovement
	visited := make(map[string]bool)
	inStack := make(map[string]bool)

	// DFS-based cycle detection for cycles of any length
	var dfs func(territory string, path []string, orderPath []int) []CircularMovement
	dfs = func(territory string, path []string, orderPath []int) []CircularMovement {
		var foundCycles []CircularMovement

		if inStack[territory] {
			// Found a cycle - find where it starts
			cycleStart := -1
			for i, t := range path {
				if t == territory {
					cycleStart = i
					break
				}
			}

			if cycleStart >= 0 {
				cyclePath := path[cycleStart:]
				cycleOrders := orderPath[cycleStart:]

				// Check if any order in the cycle uses convoy
				hasConvoy := false
				for _, orderIndex := range cycleOrders {
					if orderIndex < len(re.orders) {
						order := re.orders[orderIndex]
						if order.UnitType == game.Army {
							origin := NewProvinceCoast(order.OrigTerritory, order.FromCoast)
							dest := NewProvinceCoast(order.NewTerritory, order.ToCoast)
							convoyKey := ConvoyKey{Origin: origin, Destination: dest}
							_, hasConvoyPath := re.convoys[convoyKey]
							if hasConvoyPath {
								hasConvoy = true
								break
							}
						}
					}
				}

				foundCycles = append(foundCycles, CircularMovement{
					Territories:  cyclePath,
					OrderIndices: cycleOrders,
					HasConvoy:    hasConvoy,
				})
			}
			return foundCycles
		}

		if visited[territory] {
			return foundCycles
		}

		visited[territory] = true
		inStack[territory] = true

		// Follow all outgoing edges
		for _, edge := range graph.Edges {
			if edge.From == territory {
				newPath := append(path, edge.To)
				newOrderPath := append(orderPath, edge.OrderIndex)
				foundCycles = append(foundCycles, dfs(edge.To, newPath, newOrderPath)...)
			}
		}

		inStack[territory] = false
		return foundCycles
	}

	// Start DFS from each unvisited node
	for territory := range graph.Nodes {
		if !visited[territory] {
			cycles = append(cycles, dfs(territory, []string{territory}, []int{})...)
		}
	}

	// Remove duplicate cycles
	var uniqueCycles []CircularMovement
	for _, cycle := range cycles {
		isDuplicate := false
		for _, existing := range uniqueCycles {
			if len(cycle.OrderIndices) == len(existing.OrderIndices) {
				// Check if it's the same cycle (possibly rotated)
				match := true
				for _, orderIndex := range cycle.OrderIndices {
					found := false
					for _, existingOrderIndex := range existing.OrderIndices {
						if orderIndex == existingOrderIndex {
							found = true
							break
						}
					}
					if !found {
						match = false
						break
					}
				}
				if match {
					isDuplicate = true
					break
				}
			}
		}
		if !isDuplicate {
			uniqueCycles = append(uniqueCycles, cycle)
		}
	}

	return uniqueCycles
}

// setCircularMovementDestinations calculates and sets the correct destinations for units in a circular movement
func (re *ResolutionEngine) setCircularMovementDestinations(cycle CircularMovement) {
	// In a circular movement, each unit moves to the position that the next unit in the cycle is vacating
	// For example, in cycle A->B->C->A: A goes to B, B goes to C, C goes to A

	for i, orderIndex := range cycle.OrderIndices {
		if orderIndex >= len(re.orders) {
			continue
		}

		// Find the next position in the cycle (wrap around to start if at end)
		nextIndex := (i + 1) % len(cycle.OrderIndices)
		nextOrderIndex := cycle.OrderIndices[nextIndex]

		if nextOrderIndex >= len(re.orders) {
			continue
		}

		// The destination for this unit is where the next unit in the cycle is currently located
		nextOrder := re.orders[nextOrderIndex]
		destination := nextOrder.OrigTerritory

		// Update the order's NewTerritory and the outcome's Destination
		re.orders[orderIndex].NewTerritory = destination
		re.outcomes[orderIndex].Destination = destination

		fmt.Printf("🔄 Circular movement: Order %d (%s->%s) now goes to %s\n",
			orderIndex, re.orders[orderIndex].OrigTerritory, re.orders[orderIndex].To, destination)
	}
}

// isValidCircularMovement checks if a detected circular movement is valid
func (re *ResolutionEngine) isValidCircularMovement(cycle CircularMovement) bool {
	fmt.Printf("🔍 isValidCircularMovement: checking cycle %v\n", cycle.OrderIndices)

	if len(cycle.OrderIndices) < 2 {
		fmt.Printf("  ❌ Cycle too small: %d orders\n", len(cycle.OrderIndices))
		return false
	}

	// Check that all moves in the cycle are valid
	for _, orderIndex := range cycle.OrderIndices {
		if orderIndex >= len(re.orders) {
			fmt.Printf("  ❌ Invalid order index: %d\n", orderIndex)
			return false
		}

		order := re.orders[orderIndex]
		if order.Type != game.Move || order.NewTerritory == order.OrigTerritory {
			fmt.Printf("  ❌ Invalid move: order %d type=%s, from=%s, to=%s\n", orderIndex, order.Type, order.OrigTerritory, order.NewTerritory)
			return false
		}
	}

	// Check for external conflicts (units not in the cycle attacking cycle territories)
	for _, territory := range cycle.Territories {
		if re.hasExternalConflict(territory, cycle.OrderIndices) {
			fmt.Printf("  ❌ External conflict at territory %s\n", territory)
			return false
		}
	}

	// If the cycle involves convoys, check that convoy paths are intact
	if cycle.HasConvoy {
		for _, orderIndex := range cycle.OrderIndices {
			order := re.orders[orderIndex]
			if order.UnitType == game.Army {
				origin := NewProvinceCoast(order.OrigTerritory, order.FromCoast)
				dest := NewProvinceCoast(order.NewTerritory, order.ToCoast)
				convoyKey := ConvoyKey{Origin: origin, Destination: dest}
				_, hasConvoyPath := re.convoys[convoyKey]

				if hasConvoyPath {
					// Check that convoy fleets are not dislodged
					// For now, assume convoy paths are intact in circular movements
					// This can be enhanced later with proper convoy validation
				}
			}
		}
	}

	// Check for "help in dislodgement of own unit" - a unit cannot help dislodge its own country's unit
	for _, orderIndex := range cycle.OrderIndices {
		// Check if this move would help dislodge a unit of the same country
		if re.wouldHelpDislodgeOwnUnit(orderIndex, cycle.OrderIndices) {
			return false
		}
	}

	// Also check if any support orders would help dislodge own units in the cycle
	if re.hasSupportHelpingDislodgeOwnUnit(cycle.OrderIndices) {
		return false
	}

	return true
}

// wouldHelpDislodgeOwnUnit checks if a move would help dislodge a unit of the same country
func (re *ResolutionEngine) wouldHelpDislodgeOwnUnit(orderIndex int, cycleOrderIndices []int) bool {
	// In a circular movement, a unit moving to a territory occupied by its own country's unit
	// would be helping to dislodge that unit, which is not allowed

	if orderIndex >= len(re.orders) {
		return false
	}

	order := re.orders[orderIndex]

	// Find what unit is currently at the destination territory
	for i, otherOrder := range re.orders {
		// Skip if this is the same order or not part of the cycle
		if i == orderIndex || !contains(cycleOrderIndices, i) {
			continue
		}

		// Check if the other order's unit is currently at our destination
		if otherOrder.OrigTerritory == order.NewTerritory && otherOrder.Owner == order.Owner {
			// This move would help dislodge a unit of the same country
			return true
		}
	}

	return false
}

// hasExternalConflict checks if there are units not in the cycle attacking cycle territories
func (re *ResolutionEngine) hasExternalConflict(territory string, cycleOrderIndices []int) bool {
	// Rebuild conflicts to get current state
	re.rebuildConflictMap()

	conflictingOrders, exists := re.conflicts[territory]
	if !exists {
		return false
	}

	// Check if any conflicting order is not part of the cycle
	for _, orderIndex := range conflictingOrders {
		if !contains(cycleOrderIndices, orderIndex) {
			// There's an external unit attacking this territory
			return true
		}
	}

	return false
}

// hasSupportHelpingDislodgeOwnUnit checks if any support orders help dislodge units of the same country
func (re *ResolutionEngine) hasSupportHelpingDislodgeOwnUnit(cycleOrderIndices []int) bool {
	fmt.Printf("🔍 Checking support helping dislodge own unit for cycle %v\n", cycleOrderIndices)

	// Check all support orders to see if they're helping moves that would dislodge own units
	for i, order := range re.orders {
		if order.Type != game.Support || order.SupportDestination == "" {
			continue
		}

		fmt.Printf("  Support order %d: %s %s supporting %s -> %s\n", i, order.Owner, order.From, order.SupportTarget, order.SupportDestination)

		// Find the order being supported
		supportedOrderIndex := -1
		for j, otherOrder := range re.orders {
			if otherOrder.Type == game.Move &&
				otherOrder.From == order.SupportTarget &&
				otherOrder.To == order.SupportDestination {
				supportedOrderIndex = j
				break
			}
		}

		if supportedOrderIndex == -1 {
			fmt.Printf("    No matching move order found\n")
			continue
		}

		fmt.Printf("    Supporting move order %d\n", supportedOrderIndex)

		// Check if the supported move is part of the cycle and would dislodge a unit of the same country
		if contains(cycleOrderIndices, supportedOrderIndex) {
			supportedOrder := re.orders[supportedOrderIndex]
			fmt.Printf("    Supported move is in cycle: %s -> %s\n", supportedOrder.From, supportedOrder.To)

			// Find what unit is currently at the destination
			for j, otherOrder := range re.orders {
				if contains(cycleOrderIndices, j) &&
					otherOrder.OrigTerritory == supportedOrder.To &&
					otherOrder.Owner == order.Owner {
					// This support is helping to dislodge a unit of the same country
					fmt.Printf("    ❌ Support helping dislodge own unit: %s supporting %s->%s, dislodging own unit at %s\n",
						order.From, supportedOrder.From, supportedOrder.To, otherOrder.OrigTerritory)
					return true
				}
			}
		}
	}

	fmt.Printf("  No support helping dislodge own unit found\n")
	return false
}
