package resolution

import (
	"diplomacy-cli/backend/internal/game"
	"testing"
)

func TestCircularMovementDetection(t *testing.T) {
	// Create a simple test case with two armies swapping places
	orders := []*game.Order{
		{
			Type:     game.Move,
			Owner:    game.England,
			UnitType: game.Army,
			From:     "london",
			To:       "belgium",
		},
		{
			Type:     game.Move,
			Owner:    game.France,
			UnitType: game.Army,
			From:     "belgium",
			To:       "london",
		},
	}

	// Create a mock board
	board := &game.Board{
		Provinces: make(map[string]*game.Province),
	}

	// Create resolution engine
	engine := NewResolutionEngine(orders, board)

	// Test movement graph building
	graph := engine.buildMovementGraph()

	t.Logf("Graph has %d nodes and %d edges", len(graph.Nodes), len(graph.Edges))

	if len(graph.Nodes) != 2 {
		t.Errorf("Expected 2 nodes, got %d", len(graph.Nodes))
	}

	if len(graph.Edges) != 2 {
		t.Errorf("Expected 2 edges, got %d", len(graph.Edges))
	}

	// Test cycle detection
	cycles := engine.findCycles(graph)

	t.Logf("Found %d cycles", len(cycles))

	if len(cycles) != 1 {
		t.Errorf("Expected 1 cycle, got %d", len(cycles))
	}

	if len(cycles) > 0 {
		cycle := cycles[0]
		t.Logf("Cycle: orders %v, territories %v", cycle.OrderIndices, cycle.Territories)

		if len(cycle.OrderIndices) != 2 {
			t.Errorf("Expected cycle with 2 orders, got %d", len(cycle.OrderIndices))
		}
	}
}
