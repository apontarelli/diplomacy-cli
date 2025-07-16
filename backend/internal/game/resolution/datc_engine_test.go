package resolution

import (
	"testing"
)

func TestDATCEngine_BasicMove(t *testing.T) {
	// Test case: Simple move with no opposition
	orders := []Order{
		{
			Type:        Move,
			Unit:        "A Munich",
			Source:      "Munich",
			Destination: "Tyrolia",
			Owner:       "Austria",
		},
	}

	engine := NewDATCEngine(orders)
	results := engine.ResolveAll()

	if len(results) != 1 {
		t.Fatalf("Expected 1 result, got %d", len(results))
	}

	result := results[0]
	if !result.Success {
		t.Errorf("Expected move to succeed, but it failed: %s", result.Reason)
	}
}

func TestDATCEngine_SupportedAttack(t *testing.T) {
	// Test case: Supported attack should succeed
	orders := []Order{
		{
			Type:        Move,
			Unit:        "A Munich",
			Source:      "Munich",
			Destination: "Tyrolia",
			Owner:       "Austria",
		},
		{
			Type:      Support,
			Unit:      "A Vienna",
			Source:    "Vienna",
			Owner:     "Austria",
			Auxiliary: "Munich -> Tyrolia",
		},
		{
			Type:   Hold,
			Unit:   "A Tyrolia",
			Source: "Tyrolia",
			Owner:  "Italy",
		},
	}

	engine := NewDATCEngine(orders)
	results := engine.ResolveAll()

	if len(results) != 3 {
		t.Fatalf("Expected 3 results, got %d", len(results))
	}

	// Check move succeeds
	moveResult := results[0]
	if !moveResult.Success {
		t.Errorf("Expected supported move to succeed, but it failed: %s", moveResult.Reason)
	}

	// Check support succeeds
	supportResult := results[1]
	if !supportResult.Success {
		t.Errorf("Expected support to succeed, but it failed: %s", supportResult.Reason)
	}
}

func TestDATCEngine_Standoff(t *testing.T) {
	// Test case: Equal strength moves should result in standoff
	orders := []Order{
		{
			Type:        Move,
			Unit:        "A Munich",
			Source:      "Munich",
			Destination: "Tyrolia",
			Owner:       "Austria",
		},
		{
			Type:        Move,
			Unit:        "A Venice",
			Source:      "Venice",
			Destination: "Tyrolia",
			Owner:       "Italy",
		},
	}

	engine := NewDATCEngine(orders)
	results := engine.ResolveAll()

	if len(results) != 2 {
		t.Fatalf("Expected 2 results, got %d", len(results))
	}

	// Both moves should fail due to standoff
	for i, result := range results {
		if result.Success {
			t.Errorf("Expected move %d to fail in standoff, but it succeeded: %s", i, result.Reason)
		}
	}
}

func TestDATCEngine_HeadToHeadBattle(t *testing.T) {
	// Test case: Head-to-head battle with equal strength
	orders := []Order{
		{
			Type:        Move,
			Unit:        "A Munich",
			Source:      "Munich",
			Destination: "Tyrolia",
			Owner:       "Austria",
		},
		{
			Type:        Move,
			Unit:        "A Tyrolia",
			Source:      "Tyrolia",
			Destination: "Munich",
			Owner:       "Italy",
		},
	}

	engine := NewDATCEngine(orders)
	results := engine.ResolveAll()

	if len(results) != 2 {
		t.Fatalf("Expected 2 results, got %d", len(results))
	}

	// Both moves should fail in head-to-head with equal strength
	for i, result := range results {
		if result.Success {
			t.Errorf("Expected head-to-head move %d to fail, but it succeeded: %s", i, result.Reason)
		}
	}
}

func TestDATCEngine_SupportCut(t *testing.T) {
	// Test case: Support should be cut by successful attack
	orders := []Order{
		{
			Type:        Move,
			Unit:        "A Munich",
			Source:      "Munich",
			Destination: "Tyrolia",
			Owner:       "Austria",
		},
		{
			Type:      Support,
			Unit:      "A Vienna",
			Source:    "Vienna",
			Owner:     "Austria",
			Auxiliary: "Munich -> Tyrolia",
		},
		{
			Type:        Move,
			Unit:        "A Trieste",
			Source:      "Trieste",
			Destination: "Vienna", // Attack the supporting unit
			Owner:       "Italy",
		},
		{
			Type:      Support,
			Unit:      "A Venice",
			Source:    "Venice",
			Owner:     "Italy",
			Auxiliary: "Trieste -> Vienna", // Support the attack to make it succeed
		},
		{
			Type:   Hold,
			Unit:   "A Tyrolia",
			Source: "Tyrolia",
			Owner:  "Italy",
		},
	}

	engine := NewDATCEngine(orders)
	results := engine.ResolveAll()

	if len(results) != 5 {
		t.Fatalf("Expected 5 results, got %d", len(results))
	}

	// Move should fail because support was cut
	moveResult := results[0]
	if moveResult.Success {
		t.Errorf("Expected move to fail due to cut support, but it succeeded: %s", moveResult.Reason)
	}

	// Support should fail because it was attacked successfully
	supportResult := results[1]
	if supportResult.Success {
		t.Errorf("Expected support to be cut, but it succeeded: %s", supportResult.Reason)
	}

	// Attack on support should succeed (2 vs 1)
	attackResult := results[2]
	if !attackResult.Success {
		t.Errorf("Expected supported attack on support to succeed, but it failed: %s", attackResult.Reason)
	}

	// Support for the attack should succeed
	attackSupportResult := results[3]
	if !attackSupportResult.Success {
		t.Errorf("Expected attack support to succeed, but it failed: %s", attackSupportResult.Reason)
	}
}

func TestDATCEngine_HoldStrengthCalculation(t *testing.T) {
	// Test case: Unit moving away provides no hold strength if move succeeds
	orders := []Order{
		{
			Type:        Move,
			Unit:        "A Munich",
			Source:      "Munich",
			Destination: "Tyrolia",
			Owner:       "Austria",
		},
		{
			Type:        Move,
			Unit:        "A Tyrolia",
			Source:      "Tyrolia",
			Destination: "Vienna", // Moving away
			Owner:       "Italy",
		},
	}

	engine := NewDATCEngine(orders)
	results := engine.ResolveAll()

	if len(results) != 2 {
		t.Fatalf("Expected 2 results, got %d", len(results))
	}

	// Munich -> Tyrolia should succeed because Tyrolia is moving away
	munichResult := results[0]
	if !munichResult.Success {
		t.Errorf("Expected Munich -> Tyrolia to succeed (no hold strength), but it failed: %s", munichResult.Reason)
	}

	// Tyrolia -> Vienna should succeed (no opposition)
	tyroliaResult := results[1]
	if !tyroliaResult.Success {
		t.Errorf("Expected Tyrolia -> Vienna to succeed, but it failed: %s", tyroliaResult.Reason)
	}
}

func TestDATCEngine_CyclicDependency(t *testing.T) {
	// Test case: Simple cycle that should trigger backup rule
	orders := []Order{
		{
			Type:        Move,
			Unit:        "A Munich",
			Source:      "Munich",
			Destination: "Tyrolia",
			Owner:       "Austria",
		},
		{
			Type:        Move,
			Unit:        "A Tyrolia",
			Source:      "Tyrolia",
			Destination: "Vienna",
			Owner:       "Italy",
		},
		{
			Type:        Move,
			Unit:        "A Vienna",
			Source:      "Vienna",
			Destination: "Munich",
			Owner:       "Russia",
		},
	}

	engine := NewDATCEngine(orders)
	results := engine.ResolveAll()

	if len(results) != 3 {
		t.Fatalf("Expected 3 results, got %d", len(results))
	}

	// All moves in cycle should fail due to backup rule
	for i, result := range results {
		if result.Success {
			t.Errorf("Expected cyclic move %d to fail due to backup rule, but it succeeded: %s", i, result.Reason)
		}
	}
}

func TestDATCEngine_PartialInformation(t *testing.T) {
	// Test case: Order that can be resolved with partial information
	orders := []Order{
		{
			Type:        Move,
			Unit:        "A Munich",
			Source:      "Munich",
			Destination: "Tyrolia",
			Owner:       "Austria",
		},
		{
			Type:      Support,
			Unit:      "A Vienna",
			Source:    "Vienna",
			Owner:     "Austria",
			Auxiliary: "Munich -> Tyrolia",
		},
		{
			Type:      Support,
			Unit:      "A Bohemia",
			Source:    "Bohemia",
			Owner:     "Austria",
			Auxiliary: "Munich -> Tyrolia",
		},
		{
			Type:   Hold,
			Unit:   "A Tyrolia",
			Source: "Tyrolia",
			Owner:  "Italy",
		},
	}

	engine := NewDATCEngine(orders)
	results := engine.ResolveAll()

	if len(results) != 4 {
		t.Fatalf("Expected 4 results, got %d", len(results))
	}

	// Move should succeed with overwhelming support
	moveResult := results[0]
	if !moveResult.Success {
		t.Errorf("Expected heavily supported move to succeed, but it failed: %s", moveResult.Reason)
	}
}

// Benchmark tests for performance validation

func BenchmarkDATCEngine_SimpleMove(b *testing.B) {
	orders := []Order{
		{
			Type:        Move,
			Unit:        "A Munich",
			Source:      "Munich",
			Destination: "Tyrolia",
			Owner:       "Austria",
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		engine := NewDATCEngine(orders)
		engine.ResolveAll()
	}
}

func BenchmarkDATCEngine_ComplexScenario(b *testing.B) {
	orders := []Order{
		{Type: Move, Unit: "A Munich", Source: "Munich", Destination: "Tyrolia", Owner: "Austria"},
		{Type: Support, Unit: "A Vienna", Source: "Vienna", Owner: "Austria", Auxiliary: "Munich -> Tyrolia"},
		{Type: Move, Unit: "A Venice", Source: "Venice", Destination: "Tyrolia", Owner: "Italy"},
		{Type: Support, Unit: "A Rome", Source: "Rome", Owner: "Italy", Auxiliary: "Venice -> Tyrolia"},
		{Type: Move, Unit: "A Trieste", Source: "Trieste", Destination: "Vienna", Owner: "Italy"},
		{Type: Hold, Unit: "A Tyrolia", Source: "Tyrolia", Owner: "Germany"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		engine := NewDATCEngine(orders)
		engine.ResolveAll()
	}
}
