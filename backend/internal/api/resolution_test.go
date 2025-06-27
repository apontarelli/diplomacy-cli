package api

import (
	"testing"

	"diplomacy-backend/internal/game"
)

func TestResolveMovementOrders_SingleMove_Success(t *testing.T) {
	scenario := TestScenario{
		Name: "Army in Berlin moves to Munich (unoccupied)",
		MoveOrders: []game.Order{
			CreateMoveOrder("order1", "germany_army_berlin", "ber", "mun"),
		},
		UnitPositions: CreateUnitPositions("germany_army_berlin", "ber"),
		Expectations: []OrderExpectation{
			{OrderID: "order1", ExpectedResult: game.ResolutionSuccess, ExpectedPosition: "mun"},
		},
	}
	
	RunTestScenario(t, scenario)
}

func TestResolveMovementOrders_MultipleMovesToSameTerritory_Bounce(t *testing.T) {
	scenario := TestScenario{
		Name: "Two armies try to move to the same territory",
		MoveOrders: []game.Order{
			CreateMoveOrder("order1", "germany_army_berlin", "ber", "mun"),
			CreateMoveOrder("order2", "austria_army_bohemia", "boh", "mun"),
		},
		UnitPositions: CreateUnitPositions("germany_army_berlin", "ber", "austria_army_bohemia", "boh"),
		Expectations: []OrderExpectation{
			{OrderID: "order1", ExpectedResult: game.ResolutionBounced},
			{OrderID: "order2", ExpectedResult: game.ResolutionBounced},
		},
	}
	
	RunTestScenario(t, scenario)
}

func TestResolveMovementOrders_SupportedMove_Success(t *testing.T) {
	server := TestServer()
	
	// Setup: Army in Berlin moves to Munich with support from Bohemia
	moveOrders := []game.Order{
		{
			ID:            "order1",
			UnitID:        "germany_army_berlin",
			OrderType:     game.OrderTypeMove,
			FromTerritory: "ber",
			ToTerritory:   "mun",
		},
	}
	
	supportOrders := []game.Order{
		{
			ID:            "support1",
			UnitID:        "austria_army_bohemia",
			OrderType:     game.OrderTypeSupport,
			SupportUnit:   "germany_army_berlin",
			ToTerritory:   "mun",
		},
	}
	
	unitPositions := map[string]string{
		"germany_army_berlin": "ber",
		"austria_army_bohemia": "boh",
		"france_army_munich": "mun", // Defending unit
	}
	
	results := server.resolveMovementOrders(moveOrders, supportOrders, unitPositions)
	
	if len(results) != 1 {
		t.Fatalf("Expected 1 result, got %d", len(results))
	}
	
	result := results[0]
	if result.Result != game.ResolutionSuccess {
		t.Errorf("Expected success with support, got %v", result.Result)
	}
}

func TestCalculateSupportStrength(t *testing.T) {
	server := &Server{}
	
	moveOrders := []game.Order{
		{
			ID:            "move1",
			UnitID:        "unit1",
			ToTerritory:   "mun",
		},
	}
	
	supportOrders := []game.Order{
		{
			SupportUnit: "unit1",
			ToTerritory: "mun",
		},
		{
			SupportUnit: "unit1",
			ToTerritory: "mun",
		},
	}
	
	strength := server.calculateSupportStrength(moveOrders, supportOrders)
	
	if strength["move1"] != 2 {
		t.Errorf("Expected support strength 2, got %d", strength["move1"])
	}
}

func TestGetDefendingStrength(t *testing.T) {
	server := &Server{}
	
	supportOrders := []game.Order{
		{
			SupportUnit: "defender1",
		},
	}
	
	unitPositions := map[string]string{
		"defender1": "mun",
	}
	
	strength := server.getDefendingStrength("mun", supportOrders, unitPositions)
	
	if strength != 2 { // 1 base + 1 support
		t.Errorf("Expected defending strength 2, got %d", strength)
	}
}

func TestResolveSupportOrders(t *testing.T) {
	server := &Server{}
	
	supportOrders := []game.Order{
		{
			ID:        "support1",
			OrderType: game.OrderTypeSupport,
		},
		{
			ID:        "support2",
			OrderType: game.OrderTypeSupport,
		},
	}
	
	moveResults := []game.OrderResult{} // Empty for this test
	
	results := server.resolveSupportOrders(supportOrders, moveResults)
	
	if len(results) != 2 {
		t.Fatalf("Expected 2 results, got %d", len(results))
	}
	
	for _, result := range results {
		if result.Result != game.ResolutionSuccess {
			t.Errorf("Expected support success, got %v", result.Result)
		}
	}
}

func TestResolveHoldOrders(t *testing.T) {
	server := &Server{}
	
	holdOrders := []game.Order{
		{
			ID:        "hold1",
			OrderType: game.OrderTypeHold,
		},
	}
	
	moveResults := []game.OrderResult{} // Empty for this test
	
	results := server.resolveHoldOrders(holdOrders, moveResults)
	
	if len(results) != 1 {
		t.Fatalf("Expected 1 result, got %d", len(results))
	}
	
	result := results[0]
	if result.Result != game.ResolutionSuccess {
		t.Errorf("Expected hold success, got %v", result.Result)
	}
}

func TestResolveConvoyOrders(t *testing.T) {
	server := &Server{}
	
	convoyOrders := []game.Order{
		{
			ID:        "convoy1",
			OrderType: game.OrderTypeConvoy,
		},
	}
	
	results := server.resolveConvoyOrders(convoyOrders)
	
	if len(results) != 1 {
		t.Fatalf("Expected 1 result, got %d", len(results))
	}
	
	result := results[0]
	if result.Result != game.ResolutionSuccess {
		t.Errorf("Expected convoy success, got %v", result.Result)
	}
}

// Integration test for complex scenario
func TestComplexResolutionScenario(t *testing.T) {
	server := &Server{}
	
	// Scenario: Germany attacks France in Burgundy with support from Munich
	// France defends Burgundy with support from Paris
	// Expected: Attack fails due to equal strength (2 vs 2)
	
	orders := []game.Order{
		// German attack
		{
			ID:            "ger_move",
			UnitID:        "ger_army_ruh",
			OrderType:     game.OrderTypeMove,
			FromTerritory: "ruh",
			ToTerritory:   "bur",
		},
		// German support
		{
			ID:          "ger_support",
			UnitID:      "ger_army_mun",
			OrderType:   game.OrderTypeSupport,
			SupportUnit: "ger_army_ruh",
			ToTerritory: "bur",
		},
		// French defense (hold)
		{
			ID:            "fra_hold",
			UnitID:        "fra_army_bur",
			OrderType:     game.OrderTypeHold,
			FromTerritory: "bur",
		},
		// French support
		{
			ID:          "fra_support",
			UnitID:      "fra_army_par",
			OrderType:   game.OrderTypeSupport,
			SupportUnit: "fra_army_bur",
		},
	}
	
	unitPositions := map[string]string{
		"ger_army_ruh": "ruh",
		"ger_army_mun": "mun",
		"fra_army_bur": "bur",
		"fra_army_par": "par",
	}
	
	// Separate orders by type
	var moveOrders, supportOrders, holdOrders []game.Order
	for _, order := range orders {
		switch order.OrderType {
		case game.OrderTypeMove:
			moveOrders = append(moveOrders, order)
		case game.OrderTypeSupport:
			supportOrders = append(supportOrders, order)
		case game.OrderTypeHold:
			holdOrders = append(holdOrders, order)
		}
	}
	
	// Resolve movement orders
	moveResults := server.resolveMovementOrders(moveOrders, supportOrders, unitPositions)
	
	if len(moveResults) != 1 {
		t.Fatalf("Expected 1 move result, got %d", len(moveResults))
	}
	
	// Attack should fail due to equal strength
	if moveResults[0].Result != game.ResolutionFailed {
		t.Errorf("Expected attack to fail, got %v", moveResults[0].Result)
	}
}

func TestDetectOrderConflicts(t *testing.T) {
	server := &Server{}
	
	orders := []game.Order{
		{
			ID:          "order1",
			OrderType:   game.OrderTypeMove,
			ToTerritory: "mun",
			UnitID:      "unit1",
			Status:      game.OrderStatusSubmitted,
		},
		{
			ID:          "order2",
			OrderType:   game.OrderTypeMove,
			ToTerritory: "mun",
			UnitID:      "unit2",
			Status:      game.OrderStatusSubmitted,
		},
		{
			ID:          "order3",
			OrderType:   game.OrderTypeMove,
			ToTerritory: "ber",
			UnitID:      "unit3",
			Status:      game.OrderStatusSubmitted,
		},
	}
	
	conflicts := server.detectOrderConflicts(orders)
	
	if len(conflicts) != 1 {
		t.Fatalf("Expected 1 conflict, got %d", len(conflicts))
	}
	
	conflict := conflicts[0]
	if conflict["type"] != "movement_conflict" {
		t.Errorf("Expected movement_conflict, got %v", conflict["type"])
	}
	if conflict["territory"] != "mun" {
		t.Errorf("Expected territory 'mun', got %v", conflict["territory"])
	}
}