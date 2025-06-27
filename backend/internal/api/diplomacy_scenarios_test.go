package api

import (
	"testing"

	"diplomacy-backend/internal/game"
)

// Test classic Diplomacy opening scenarios

func TestClassicOpening_Austria_Italy_Conflict(t *testing.T) {
	server := &Server{}
	
	// Classic scenario: Austria A(Vie) -> Tri, Italy A(Ven) -> Tri
	// Both should bounce
	
	moveOrders := []game.Order{
		{
			ID:            "austria_vie_tri",
			UnitID:        "austria_army_vie",
			OrderType:     game.OrderTypeMove,
			FromTerritory: "vie",
			ToTerritory:   "tri",
		},
		{
			ID:            "italy_ven_tri",
			UnitID:        "italy_army_ven",
			OrderType:     game.OrderTypeMove,
			FromTerritory: "ven",
			ToTerritory:   "tri",
		},
	}
	
	supportOrders := []game.Order{}
	unitPositions := map[string]string{
		"austria_army_vie": "vie",
		"italy_army_ven":   "ven",
	}
	
	results := server.resolveMovementOrders(moveOrders, supportOrders, unitPositions)
	
	if len(results) != 2 {
		t.Fatalf("Expected 2 results, got %d", len(results))
	}
	
	// Both moves should bounce
	for _, result := range results {
		if result.Result != game.ResolutionBounced {
			t.Errorf("Expected bounce in Austria-Italy Trieste conflict, got %v for order %s", result.Result, result.OrderID)
		}
	}
}

func TestSupportedAttack_Overcomes_Defense(t *testing.T) {
	server := &Server{}
	
	// Germany attacks France in Burgundy with support
	// F(Ruh) -> Bur supported by A(Mun)
	// France holds in Burgundy: A(Bur) H
	// Attack should succeed (2 vs 1)
	
	moveOrders := []game.Order{
		{
			ID:            "ger_ruh_bur",
			UnitID:        "ger_army_ruh",
			OrderType:     game.OrderTypeMove,
			FromTerritory: "ruh",
			ToTerritory:   "bur",
		},
	}
	
	supportOrders := []game.Order{
		{
			ID:          "ger_mun_support",
			UnitID:      "ger_army_mun",
			OrderType:   game.OrderTypeSupport,
			SupportUnit: "ger_army_ruh",
			ToTerritory: "bur",
		},
	}
	
	unitPositions := map[string]string{
		"ger_army_ruh": "ruh",
		"ger_army_mun": "mun",
		"fra_army_bur": "bur", // Defending unit
	}
	
	results := server.resolveMovementOrders(moveOrders, supportOrders, unitPositions)
	
	if len(results) != 1 {
		t.Fatalf("Expected 1 result, got %d", len(results))
	}
	
	result := results[0]
	if result.Result != game.ResolutionSuccess {
		t.Errorf("Expected supported attack to succeed, got %v", result.Result)
	}
	if result.NewPosition != "bur" {
		t.Errorf("Expected new position 'bur', got %v", result.NewPosition)
	}
}

func TestEqualStrength_Attack_Fails(t *testing.T) {
	server := &Server{}
	
	// Germany attacks France with equal strength
	// A(Ruh) -> Bur supported by A(Mun) (strength 2)
	// France defends A(Bur) H supported by A(Par) (strength 2)
	// Attack should fail
	
	moveOrders := []game.Order{
		{
			ID:            "ger_ruh_bur",
			UnitID:        "ger_army_ruh",
			OrderType:     game.OrderTypeMove,
			FromTerritory: "ruh",
			ToTerritory:   "bur",
		},
	}
	
	supportOrders := []game.Order{
		{
			ID:          "ger_mun_support",
			UnitID:      "ger_army_mun",
			OrderType:   game.OrderTypeSupport,
			SupportUnit: "ger_army_ruh",
			ToTerritory: "bur",
		},
		{
			ID:          "fra_par_support",
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
	
	results := server.resolveMovementOrders(moveOrders, supportOrders, unitPositions)
	
	if len(results) != 1 {
		t.Fatalf("Expected 1 result, got %d", len(results))
	}
	
	result := results[0]
	if result.Result != game.ResolutionFailed {
		t.Errorf("Expected equal strength attack to fail, got %v", result.Result)
	}
}

func TestThreeWayBounce(t *testing.T) {
	server := &Server{}
	
	// Three units try to move to the same territory
	// All should bounce
	
	moveOrders := []game.Order{
		{
			ID:            "unit1_move",
			UnitID:        "unit1",
			OrderType:     game.OrderTypeMove,
			FromTerritory: "ber",
			ToTerritory:   "sil",
		},
		{
			ID:            "unit2_move",
			UnitID:        "unit2",
			OrderType:     game.OrderTypeMove,
			FromTerritory: "war",
			ToTerritory:   "sil",
		},
		{
			ID:            "unit3_move",
			UnitID:        "unit3",
			OrderType:     game.OrderTypeMove,
			FromTerritory: "gal",
			ToTerritory:   "sil",
		},
	}
	
	supportOrders := []game.Order{}
	unitPositions := map[string]string{
		"unit1": "ber",
		"unit2": "war",
		"unit3": "gal",
	}
	
	results := server.resolveMovementOrders(moveOrders, supportOrders, unitPositions)
	
	if len(results) != 3 {
		t.Fatalf("Expected 3 results, got %d", len(results))
	}
	
	// All moves should bounce
	for _, result := range results {
		if result.Result != game.ResolutionBounced {
			t.Errorf("Expected bounce in three-way conflict, got %v for order %s", result.Result, result.OrderID)
		}
	}
}

func TestSupportedMove_Against_SupportedDefense_Higher_Strength_Wins(t *testing.T) {
	server := &Server{}
	
	// Attack with 3 strength vs defense with 2 strength
	// A(Ruh) -> Bur supported by A(Mun) and A(Kie) (strength 3)
	// France defends A(Bur) H supported by A(Par) (strength 2)
	// Attack should succeed
	
	moveOrders := []game.Order{
		{
			ID:            "ger_ruh_bur",
			UnitID:        "ger_army_ruh",
			OrderType:     game.OrderTypeMove,
			FromTerritory: "ruh",
			ToTerritory:   "bur",
		},
	}
	
	supportOrders := []game.Order{
		{
			ID:          "ger_mun_support",
			UnitID:      "ger_army_mun",
			OrderType:   game.OrderTypeSupport,
			SupportUnit: "ger_army_ruh",
			ToTerritory: "bur",
		},
		{
			ID:          "ger_kie_support",
			UnitID:      "ger_army_kie",
			OrderType:   game.OrderTypeSupport,
			SupportUnit: "ger_army_ruh",
			ToTerritory: "bur",
		},
		{
			ID:          "fra_par_support",
			UnitID:      "fra_army_par",
			OrderType:   game.OrderTypeSupport,
			SupportUnit: "fra_army_bur",
		},
	}
	
	unitPositions := map[string]string{
		"ger_army_ruh": "ruh",
		"ger_army_mun": "mun",
		"ger_army_kie": "kie",
		"fra_army_bur": "bur",
		"fra_army_par": "par",
	}
	
	results := server.resolveMovementOrders(moveOrders, supportOrders, unitPositions)
	
	if len(results) != 1 {
		t.Fatalf("Expected 1 result, got %d", len(results))
	}
	
	result := results[0]
	if result.Result != game.ResolutionSuccess {
		t.Errorf("Expected higher strength attack to succeed, got %v", result.Result)
	}
}

func TestNoSupportForInvalidMove(t *testing.T) {
	server := &Server{}
	
	// Test that support is only counted for the correct move
	moveOrders := []game.Order{
		{
			ID:            "move_to_mun",
			UnitID:        "unit1",
			OrderType:     game.OrderTypeMove,
			FromTerritory: "ber",
			ToTerritory:   "mun",
		},
	}
	
	supportOrders := []game.Order{
		{
			ID:          "wrong_support",
			UnitID:      "unit2",
			OrderType:   game.OrderTypeSupport,
			SupportUnit: "unit1",
			ToTerritory: "vie", // Supporting move to different territory
		},
	}
	
	strength := server.calculateSupportStrength(moveOrders, supportOrders)
	
	if strength["move_to_mun"] != 0 {
		t.Errorf("Expected no support for wrong territory, got %d", strength["move_to_mun"])
	}
}

func TestEmptyTerritoryMove_AlwaysSucceeds(t *testing.T) {
	server := &Server{}
	
	// Move to empty territory should always succeed
	moveOrders := []game.Order{
		{
			ID:            "move_to_empty",
			UnitID:        "unit1",
			OrderType:     game.OrderTypeMove,
			FromTerritory: "ber",
			ToTerritory:   "sil", // Empty territory
		},
	}
	
	supportOrders := []game.Order{}
	unitPositions := map[string]string{
		"unit1": "ber",
		// sil is empty
	}
	
	results := server.resolveMovementOrders(moveOrders, supportOrders, unitPositions)
	
	if len(results) != 1 {
		t.Fatalf("Expected 1 result, got %d", len(results))
	}
	
	result := results[0]
	if result.Result != game.ResolutionSuccess {
		t.Errorf("Expected move to empty territory to succeed, got %v", result.Result)
	}
	if result.NewPosition != "sil" {
		t.Errorf("Expected new position 'sil', got %v", result.NewPosition)
	}
}

func TestCancelledOrdersIgnored(t *testing.T) {
	server := &Server{}
	
	orders := []game.Order{
		{
			ID:          "active_order",
			OrderType:   game.OrderTypeMove,
			ToTerritory: "mun",
			UnitID:      "unit1",
			Status:      game.OrderStatusSubmitted,
		},
		{
			ID:          "cancelled_order",
			OrderType:   game.OrderTypeMove,
			ToTerritory: "mun",
			UnitID:      "unit2",
			Status:      game.OrderStatusCancelled,
		},
	}
	
	conflicts := server.detectOrderConflicts(orders)
	
	// Should be no conflicts since one order is cancelled
	if len(conflicts) != 0 {
		t.Errorf("Expected no conflicts with cancelled order, got %d", len(conflicts))
	}
}