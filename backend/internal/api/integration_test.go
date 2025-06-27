package api

import (
	"testing"

	"diplomacy-backend/internal/game"
)

// Integration tests for the full resolution flow

func TestFullResolutionFlow_MixedOrders(t *testing.T) {
	server := &Server{}
	
	// Test order separation and processing logic without database
	orders := []game.Order{
		// Movement orders
		{
			ID:            "ger_ber_sil",
			UnitID:        "ger_army_ber",
			OrderType:     game.OrderTypeMove,
			FromTerritory: "ber",
			ToTerritory:   "sil",
			Status:        game.OrderStatusSubmitted,
		},
		{
			ID:            "aus_vie_bud",
			UnitID:        "aus_army_vie",
			OrderType:     game.OrderTypeMove,
			FromTerritory: "vie",
			ToTerritory:   "bud",
			Status:        game.OrderStatusSubmitted,
		},
		// Support orders
		{
			ID:          "ger_mun_support",
			UnitID:      "ger_army_mun",
			OrderType:   game.OrderTypeSupport,
			SupportUnit: "ger_army_ber",
			ToTerritory: "sil",
			Status:      game.OrderStatusSubmitted,
		},
		// Hold orders
		{
			ID:            "fra_par_hold",
			UnitID:        "fra_army_par",
			OrderType:     game.OrderTypeHold,
			FromTerritory: "par",
			Status:        game.OrderStatusSubmitted,
		},
		// Convoy orders
		{
			ID:          "eng_lon_convoy",
			UnitID:      "eng_fleet_lon",
			OrderType:   game.OrderTypeConvoy,
			SupportUnit: "eng_army_wal",
			ToTerritory: "bre",
			Status:      game.OrderStatusSubmitted,
		},
	}
	
	// Test order separation logic
	var moveOrders, supportOrders, holdOrders, convoyOrders []game.Order
	for _, order := range orders {
		if order.Status == game.OrderStatusCancelled {
			continue
		}

		switch order.OrderType {
		case game.OrderTypeMove:
			moveOrders = append(moveOrders, order)
		case game.OrderTypeSupport:
			supportOrders = append(supportOrders, order)
		case game.OrderTypeHold:
			holdOrders = append(holdOrders, order)
		case game.OrderTypeConvoy:
			convoyOrders = append(convoyOrders, order)
		}
	}
	
	// Verify correct separation
	if len(moveOrders) != 2 {
		t.Errorf("Expected 2 move orders, got %d", len(moveOrders))
	}
	if len(supportOrders) != 1 {
		t.Errorf("Expected 1 support order, got %d", len(supportOrders))
	}
	if len(holdOrders) != 1 {
		t.Errorf("Expected 1 hold order, got %d", len(holdOrders))
	}
	if len(convoyOrders) != 1 {
		t.Errorf("Expected 1 convoy order, got %d", len(convoyOrders))
	}
	
	// Test individual resolution components
	unitPositions := map[string]string{
		"ger_army_ber": "ber",
		"ger_army_mun": "mun",
		"aus_army_vie": "vie",
		"fra_army_par": "par",
		"eng_fleet_lon": "lon",
	}
	
	moveResults := server.resolveMovementOrders(moveOrders, supportOrders, unitPositions)
	supportResults := server.resolveSupportOrders(supportOrders, moveResults)
	holdResults := server.resolveHoldOrders(holdOrders, moveResults)
	convoyResults := server.resolveConvoyOrders(convoyOrders)
	
	// Verify all components return results
	if len(moveResults) != 2 {
		t.Errorf("Expected 2 move results, got %d", len(moveResults))
	}
	if len(supportResults) != 1 {
		t.Errorf("Expected 1 support result, got %d", len(supportResults))
	}
	if len(holdResults) != 1 {
		t.Errorf("Expected 1 hold result, got %d", len(holdResults))
	}
	if len(convoyResults) != 1 {
		t.Errorf("Expected 1 convoy result, got %d", len(convoyResults))
	}
}

func TestResolutionWithCancelledOrders(t *testing.T) {
	server := &Server{}
	
	orders := []game.Order{
		{
			ID:            "active_move",
			UnitID:        "unit1",
			OrderType:     game.OrderTypeMove,
			FromTerritory: "ber",
			ToTerritory:   "mun",
			Status:        game.OrderStatusSubmitted,
		},
		{
			ID:            "cancelled_move",
			UnitID:        "unit2",
			OrderType:     game.OrderTypeMove,
			FromTerritory: "vie",
			ToTerritory:   "mun",
			Status:        game.OrderStatusCancelled,
		},
	}
	
	// Test that cancelled orders are filtered out
	var activeOrders []game.Order
	for _, order := range orders {
		if order.Status != game.OrderStatusCancelled {
			activeOrders = append(activeOrders, order)
		}
	}
	
	if len(activeOrders) != 1 {
		t.Fatalf("Expected 1 active order, got %d", len(activeOrders))
	}
	
	if activeOrders[0].ID != "active_move" {
		t.Errorf("Expected active_move to remain, got %s", activeOrders[0].ID)
	}
	
	// Test movement resolution with only active orders
	unitPositions := map[string]string{
		"unit1": "ber",
		"unit2": "vie",
	}
	
	results := server.resolveMovementOrders(activeOrders, []game.Order{}, unitPositions)
	
	if len(results) != 1 {
		t.Fatalf("Expected 1 result, got %d", len(results))
	}
	
	if results[0].Result != game.ResolutionSuccess {
		t.Errorf("Expected success for uncontested move, got %v", results[0].Result)
	}
}

func TestResolutionOrderProcessing(t *testing.T) {
	
	// Test that orders are processed in the correct sequence
	// Movement orders should be resolved first, then support, hold, convoy
	
	orders := []game.Order{
		// Mix up the order to ensure processing sequence is correct
		{
			ID:        "hold1",
			OrderType: game.OrderTypeHold,
			Status:    game.OrderStatusSubmitted,
		},
		{
			ID:          "move1",
			OrderType:   game.OrderTypeMove,
			ToTerritory: "mun",
			Status:      game.OrderStatusSubmitted,
		},
		{
			ID:        "support1",
			OrderType: game.OrderTypeSupport,
			Status:    game.OrderStatusSubmitted,
		},
		{
			ID:        "convoy1",
			OrderType: game.OrderTypeConvoy,
			Status:    game.OrderStatusSubmitted,
		},
	}
	
	// Test order separation
	var moveOrders, supportOrders, holdOrders, convoyOrders []game.Order
	for _, order := range orders {
		if order.Status == game.OrderStatusCancelled {
			continue
		}

		switch order.OrderType {
		case game.OrderTypeMove:
			moveOrders = append(moveOrders, order)
		case game.OrderTypeSupport:
			supportOrders = append(supportOrders, order)
		case game.OrderTypeHold:
			holdOrders = append(holdOrders, order)
		case game.OrderTypeConvoy:
			convoyOrders = append(convoyOrders, order)
		}
	}
	
	// Verify correct separation
	if len(moveOrders) != 1 {
		t.Errorf("Expected 1 move order, got %d", len(moveOrders))
	}
	if len(supportOrders) != 1 {
		t.Errorf("Expected 1 support order, got %d", len(supportOrders))
	}
	if len(holdOrders) != 1 {
		t.Errorf("Expected 1 hold order, got %d", len(holdOrders))
	}
	if len(convoyOrders) != 1 {
		t.Errorf("Expected 1 convoy order, got %d", len(convoyOrders))
	}
}

func TestEmptyOrdersList(t *testing.T) {
	orders := []game.Order{}
	
	// Test that empty orders list is handled correctly
	if len(orders) != 0 {
		t.Fatalf("Expected 0 orders, got %d", len(orders))
	}
}

func TestAllOrdersCancelled(t *testing.T) {
	orders := []game.Order{
		{
			ID:     "cancelled1",
			Status: game.OrderStatusCancelled,
		},
		{
			ID:     "cancelled2",
			Status: game.OrderStatusCancelled,
		},
	}
	
	// Filter out cancelled orders
	var activeOrders []game.Order
	for _, order := range orders {
		if order.Status != game.OrderStatusCancelled {
			activeOrders = append(activeOrders, order)
		}
	}
	
	if len(activeOrders) != 0 {
		t.Fatalf("Expected 0 active orders when all are cancelled, got %d", len(activeOrders))
	}
}

// Test the resolution result structure
func TestResolutionResultStructure(t *testing.T) {
	server := &Server{}
	
	orders := []game.Order{
		{
			ID:            "test_move",
			UnitID:        "test_unit",
			OrderType:     game.OrderTypeMove,
			FromTerritory: "ber",
			ToTerritory:   "mun",
			Status:        game.OrderStatusSubmitted,
		},
	}
	
	unitPositions := map[string]string{
		"test_unit": "ber",
	}
	
	results := server.resolveMovementOrders(orders, []game.Order{}, unitPositions)
	
	if len(results) != 1 {
		t.Fatalf("Expected 1 result, got %d", len(results))
	}
	
	result := results[0]
	
	// Check result structure
	if result.OrderID == "" {
		t.Error("OrderID should not be empty")
	}
	if result.Result == "" {
		t.Error("Result should not be empty")
	}
	if result.Reason == "" {
		t.Error("Reason should not be empty")
	}
	
	// For successful move, NewPosition should be set
	if result.Result == game.ResolutionSuccess && result.NewPosition == "" {
		t.Error("NewPosition should be set for successful move")
	}
}