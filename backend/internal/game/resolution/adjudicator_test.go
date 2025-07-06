package resolution

import (
	"testing"
)

func TestBasicMoveResolution(t *testing.T) {
	// Test a simple move with no conflicts
	orders := []Order{
		{
			Unit:        "A Berlin",
			Type:        Move,
			Source:      "Berlin",
			Destination: "Munich",
		},
	}

	adj := NewAdjudicator(orders)
	results := adj.ResolveAll()

	if len(results) != 1 {
		t.Fatalf("Expected 1 result, got %d", len(results))
	}

	result := results[0]
	if !result.Success {
		t.Errorf("Expected move to succeed, but it failed: %s", result.Reason)
	}
}

func TestBasicConflict(t *testing.T) {
	// Test two units moving to the same destination
	orders := []Order{
		{
			Unit:        "A Berlin",
			Type:        Move,
			Source:      "Berlin",
			Destination: "Munich",
		},
		{
			Unit:        "A Vienna",
			Type:        Move,
			Source:      "Vienna",
			Destination: "Munich",
		},
	}

	adj := NewAdjudicator(orders)
	results := adj.ResolveAll()

	if len(results) != 2 {
		t.Fatalf("Expected 2 results, got %d", len(results))
	}

	// Both moves should fail due to equal strength
	for i, result := range results {
		if result.Success {
			t.Errorf("Expected move %d to fail due to conflict, but it succeeded", i)
		}
	}
}

func TestSupportedMove(t *testing.T) {
	// Test a move with support
	orders := []Order{
		{
			Unit:        "A Berlin",
			Type:        Move,
			Source:      "Berlin",
			Destination: "Munich",
		},
		{
			Unit:      "A Kiel",
			Type:      Support,
			Source:    "Kiel",
			Auxiliary: "Berlin -> Munich", // Supporting the move
		},
	}

	adj := NewAdjudicator(orders)
	results := adj.ResolveAll()

	if len(results) != 2 {
		t.Fatalf("Expected 2 results, got %d", len(results))
	}

	// Move should succeed with support
	moveResult := results[0]
	if !moveResult.Success {
		t.Errorf("Expected supported move to succeed, but it failed: %s", moveResult.Reason)
	}

	// Support should succeed (not cut)
	supportResult := results[1]
	if !supportResult.Success {
		t.Errorf("Expected support to succeed, but it failed: %s", supportResult.Reason)
	}
}
