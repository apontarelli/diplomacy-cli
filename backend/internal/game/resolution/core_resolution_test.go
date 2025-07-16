package resolution

import (
	"diplomacy-cli/backend/internal/game"
	"testing"
)

func TestCoreResolutionAlgorithm_SimpleMove(t *testing.T) {
	config := EngineConfig{
		StrictDATCCompliance: false, // Allow validation errors for testing
		ParadoxResolution:    BackupRuleResolution,
		ValidationLevel:      StandardValidation,
		EnableDebugLogging:   false,
		MaxRecursionDepth:    100,
	}
	engine := NewDATCCompliantEngine(config)
	board := createTestBoard()

	// Simple successful move
	orders := []game.Order{
		{
			ID:       "1",
			UnitType: game.Army,
			From:     "munich",
			Type:     game.Move,
			To:       "berlin",
			Owner:    game.Germany,
		},
	}

	result, err := engine.Resolve(orders, board)
	if err != nil {
		t.Fatalf("Resolution failed: %v", err)
	}

	// Verify unit outcome
	unitOutcomes := result.UnitOutcomes()
	if len(unitOutcomes) != 1 {
		t.Errorf("Expected 1 unit outcome, got %d", len(unitOutcomes))
	}

	germanArmyID := UnitID{Type: game.Army, Owner: game.Germany, Province: "munich"}
	outcome, exists := unitOutcomes[germanArmyID]
	if !exists {
		t.Fatal("Expected outcome for German army not found")
	}

	if outcome.FinalStatus() != UnitMoved {
		t.Errorf("Expected unit to move, got status %v", outcome.FinalStatus())
	}

	if outcome.ToProvince() == nil || *outcome.ToProvince() != "berlin" {
		t.Errorf("Expected unit to move to berlin, got %v", outcome.ToProvince())
	}

	if outcome.Strength() == 0 {
		t.Error("Expected non-zero strength calculation")
	}

	// Verify resolution log
	resolutionLog := result.ResolutionLog()
	if len(resolutionLog) == 0 {
		t.Error("Expected resolution log entries")
	}
}

func TestCoreResolutionAlgorithm_HeadToHeadBattle(t *testing.T) {
	config := EngineConfig{
		StrictDATCCompliance: false, // Allow validation errors for testing
		ParadoxResolution:    BackupRuleResolution,
		ValidationLevel:      StandardValidation,
		EnableDebugLogging:   false,
		MaxRecursionDepth:    100,
	}
	engine := NewDATCCompliantEngine(config)
	board := createTestBoard()

	// Add units for head-to-head battle
	berlinUnit := &game.Unit{
		Type:     game.Army,
		Owner:    game.Germany,
		Province: "berlin",
	}
	board.PlaceUnit(berlinUnit)

	// Head-to-head battle: Munich -> Berlin, Berlin -> Munich
	orders := []game.Order{
		{
			ID:       "1",
			UnitType: game.Army,
			From:     "munich",
			Type:     game.Move,
			To:       "berlin",
			Owner:    game.Germany,
		},
		{
			ID:       "2",
			UnitType: game.Army,
			From:     "berlin",
			Type:     game.Move,
			To:       "munich",
			Owner:    game.Germany,
		},
	}

	result, err := engine.Resolve(orders, board)
	if err != nil {
		t.Fatalf("Resolution failed: %v", err)
	}

	// Verify conflicts were detected
	conflicts := result.Conflicts()
	if len(conflicts) == 0 {
		t.Error("Expected head-to-head conflict to be detected")
	}

	// Debug: Print all conflicts
	for i, conflict := range conflicts {
		t.Logf("Conflict %d: Province=%s, Type=%v, Competitors=%d",
			i, conflict.Province(), conflict.Resolution(), len(conflict.Competitors()))
	}

	// Check for head-to-head conflict type
	foundHeadToHead := false
	for _, conflict := range conflicts {
		if conflict.Resolution() == HeadToHead {
			foundHeadToHead = true
			if len(conflict.Competitors()) != 2 {
				t.Errorf("Expected 2 competitors in head-to-head, got %d", len(conflict.Competitors()))
			}
		}
	}

	if !foundHeadToHead {
		t.Error("Expected head-to-head conflict type")
	}

	// Verify unit outcomes have detailed reasoning
	unitOutcomes := result.UnitOutcomes()
	for _, outcome := range unitOutcomes {
		if outcome.Reason() == "" {
			t.Error("Expected detailed reasoning for unit outcome")
		}
		if outcome.Strength() == 0 {
			t.Error("Expected strength calculation for unit")
		}
	}
}

func TestCoreResolutionAlgorithm_SupportedAttack(t *testing.T) {
	config := EngineConfig{
		StrictDATCCompliance: false, // Allow validation errors for testing
		ParadoxResolution:    BackupRuleResolution,
		ValidationLevel:      StandardValidation,
		EnableDebugLogging:   false,
		MaxRecursionDepth:    100,
	}
	engine := NewDATCCompliantEngine(config)
	board := createTestBoard()

	// Add supporting unit
	viennaUnit := &game.Unit{
		Type:     game.Army,
		Owner:    game.Germany,
		Province: "vienna",
	}
	board.PlaceUnit(viennaUnit)

	// Supported attack: Munich -> Berlin with support from Vienna
	orders := []game.Order{
		{
			ID:       "1",
			UnitType: game.Army,
			From:     "munich",
			Type:     game.Move,
			To:       "berlin",
			Owner:    game.Germany,
		},
		{
			ID:                 "2",
			UnitType:           game.Army,
			From:               "vienna",
			Type:               game.Support,
			SupportTarget:      "munich",
			SupportDestination: "berlin",
			Owner:              game.Germany,
		},
	}

	result, err := engine.Resolve(orders, board)
	if err != nil {
		t.Fatalf("Resolution failed: %v", err)
	}

	// Verify attack succeeded with support
	unitOutcomes := result.UnitOutcomes()
	germanArmyID := UnitID{Type: game.Army, Owner: game.Germany, Province: "munich"}
	outcome, exists := unitOutcomes[germanArmyID]
	if !exists {
		t.Fatal("Expected outcome for German army not found")
	}

	if outcome.FinalStatus() != UnitMoved {
		t.Errorf("Expected supported attack to succeed, got status %v", outcome.FinalStatus())
	}

	if outcome.Strength() < 2 {
		t.Errorf("Expected attack strength >= 2 with support, got %d", outcome.Strength())
	}

	if outcome.SupportCount() != 1 {
		t.Errorf("Expected 1 support, got %d", outcome.SupportCount())
	}

	// Verify support outcomes
	supportOutcomes := result.SupportOutcomes()
	if len(supportOutcomes) == 0 {
		t.Error("Expected support outcome tracking")
	}

	for _, supportOutcome := range supportOutcomes {
		if !supportOutcome.IsSuccessful() {
			t.Error("Expected support to be successful")
		}
		if supportOutcome.Reason() == "" {
			t.Error("Expected detailed reasoning for support outcome")
		}
	}
}

func TestCoreResolutionAlgorithm_Standoff(t *testing.T) {
	config := EngineConfig{
		StrictDATCCompliance: false, // Allow validation errors for testing
		ParadoxResolution:    BackupRuleResolution,
		ValidationLevel:      StandardValidation,
		EnableDebugLogging:   false,
		MaxRecursionDepth:    100,
	}
	engine := NewDATCCompliantEngine(config)
	board := createTestBoard()

	// Add competing units
	viennaUnit := &game.Unit{
		Type:     game.Army,
		Owner:    game.Austria,
		Province: "vienna",
	}
	board.PlaceUnit(viennaUnit)

	// Standoff: Munich -> Berlin, Vienna -> Berlin (equal strength)
	orders := []game.Order{
		{
			ID:       "1",
			UnitType: game.Army,
			From:     "munich",
			Type:     game.Move,
			To:       "berlin",
			Owner:    game.Germany,
		},
		{
			ID:       "2",
			UnitType: game.Army,
			From:     "vienna",
			Type:     game.Move,
			To:       "berlin",
			Owner:    game.Austria,
		},
	}

	result, err := engine.Resolve(orders, board)
	if err != nil {
		t.Fatalf("Resolution failed: %v", err)
	}

	// Verify standoff occurred
	conflicts := result.Conflicts()
	foundStandoff := false
	for _, conflict := range conflicts {
		if conflict.Resolution() == Standoff {
			foundStandoff = true
			if conflict.Winner() != nil {
				t.Error("Expected no winner in standoff")
			}
			if len(conflict.Competitors()) != 2 {
				t.Errorf("Expected 2 competitors in standoff, got %d", len(conflict.Competitors()))
			}
		}
	}

	if !foundStandoff {
		t.Error("Expected standoff conflict type")
	}

	// Verify both moves failed
	unitOutcomes := result.UnitOutcomes()
	for _, outcome := range unitOutcomes {
		if outcome.OrderGiven().Type == game.Move {
			if outcome.FinalStatus() == UnitMoved {
				t.Error("Expected moves to fail in standoff")
			}
		}
	}
}

func TestStrengthCalculationDetailed(t *testing.T) {
	orders := []Order{
		{
			Type:        Move,
			Source:      "munich",
			Destination: "berlin",
			Unit:        "A Munich",
			Owner:       "germany",
		},
		{
			Type:      Support,
			Source:    "vienna",
			Unit:      "A Vienna",
			Owner:     "germany",
			Auxiliary: "munich -> berlin",
		},
	}

	adjudicator := NewAdjudicator(orders)

	// Test detailed attack strength calculation
	attackCalc := adjudicator.calculateAttackStrengthDetailed(&orders[0], true)

	if attackCalc.BaseStrength != 1 {
		t.Errorf("Expected base strength 1, got %d", attackCalc.BaseStrength)
	}

	if len(attackCalc.Reasoning) == 0 {
		t.Error("Expected detailed reasoning")
	}

	if !attackCalc.PathValid {
		t.Error("Expected valid path for direct move")
	}

	// Verify reasoning contains expected elements
	foundBaseReasoning := false
	for _, reason := range attackCalc.Reasoning {
		if reason == "Base attack strength: 1" {
			foundBaseReasoning = true
		}
	}

	if !foundBaseReasoning {
		t.Error("Expected base strength reasoning in calculation")
	}
}

func TestConflictAnalysisDetailed(t *testing.T) {
	orders := []Order{
		{
			Type:        Move,
			Source:      "munich",
			Destination: "berlin",
			Unit:        "A Munich",
			Owner:       "germany",
		},
		{
			Type:        Move,
			Source:      "vienna",
			Destination: "berlin",
			Unit:        "A Vienna",
			Owner:       "austria",
		},
	}

	adjudicator := NewAdjudicator(orders)
	resolver := NewDetailedConflictResolver(adjudicator)

	conflicts := resolver.AnalyzeConflicts(true)

	if len(conflicts) != 1 {
		t.Errorf("Expected 1 conflict, got %d", len(conflicts))
	}

	conflict := conflicts[0]
	if conflict.Province != "berlin" {
		t.Errorf("Expected conflict in berlin, got %s", conflict.Province)
	}

	if len(conflict.Competitors) != 2 {
		t.Errorf("Expected 2 competitors, got %d", len(conflict.Competitors))
	}

	if conflict.ConflictType != Standoff {
		t.Errorf("Expected standoff, got %v", conflict.ConflictType)
	}

	if len(conflict.Resolution.Reasoning) == 0 {
		t.Error("Expected detailed resolution reasoning")
	}

	if len(conflict.Strengths) == 0 {
		t.Error("Expected strength calculations for competitors")
	}
}
