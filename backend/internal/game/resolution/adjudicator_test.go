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

func TestSupportedVsUnsupported(t *testing.T) {
	// Test supported move vs unsupported move
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
		{
			Unit:      "A Kiel",
			Type:      Support,
			Source:    "Kiel",
			Auxiliary: "Berlin -> Munich", // Supporting Berlin
		},
	}

	adj := NewAdjudicator(orders)
	results := adj.ResolveAll()

	if len(results) != 3 {
		t.Fatalf("Expected 3 results, got %d", len(results))
	}

	// Berlin move should succeed (strength 2 vs Vienna's 1)
	berlinResult := results[0]
	if !berlinResult.Success {
		t.Errorf("Expected Berlin move to succeed with support, but it failed: %s", berlinResult.Reason)
	}

	// Vienna move should fail (strength 1 vs Berlin's 2)
	viennaResult := results[1]
	if viennaResult.Success {
		t.Errorf("Expected Vienna move to fail against supported move, but it succeeded")
	}

	// Support should succeed (not cut)
	supportResult := results[2]
	if !supportResult.Success {
		t.Errorf("Expected support to succeed, but it failed: %s", supportResult.Reason)
	}
}

func TestHoldWithSupport(t *testing.T) {
	// Test unit holding with support against attack
	orders := []Order{
		{
			Unit:        "A Munich",
			Type:        Hold,
			Source:      "Munich",
			Destination: "",
		},
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
			Auxiliary: "Munich", // Supporting Munich to hold
		},
	}

	adj := NewAdjudicator(orders)
	results := adj.ResolveAll()

	if len(results) != 3 {
		t.Fatalf("Expected 3 results, got %d", len(results))
	}

	// Hold should succeed (not dislodged)
	holdResult := results[0]
	if !holdResult.Success {
		t.Errorf("Expected hold to succeed with support, but it failed: %s", holdResult.Reason)
	}

	// Attack should fail (strength 1 vs hold strength 2)
	attackResult := results[1]
	if attackResult.Success {
		t.Errorf("Expected attack to fail against supported hold, but it succeeded")
	}

	// Support should succeed (not cut)
	supportResult := results[2]
	if !supportResult.Success {
		t.Errorf("Expected support to succeed, but it failed: %s", supportResult.Reason)
	}
}
