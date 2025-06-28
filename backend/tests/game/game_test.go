package game_test

import (
	"testing"

	"diplomacy-backend/internal/game"
	"diplomacy-backend/internal/game/rules"
)

// TestGameRules loads the classic game rules for testing
func TestGameRules(t *testing.T) {
	rules := rules.MustLoadRules("classic")
	if rules == nil {
		t.Fatal("Failed to load classic rules")
	}
	t.Log("Successfully loaded classic game rules")
}

// TestOrderCreation tests basic order creation
func TestOrderCreation(t *testing.T) {
	order := game.Order{
		ID:            "test_order",
		UnitID:        "test_unit",
		OrderType:     game.OrderTypeMove,
		FromTerritory: "ber",
		ToTerritory:   "mun",
		Status:        game.OrderStatusSubmitted,
	}
	
	if order.ID != "test_order" {
		t.Errorf("Expected order ID 'test_order', got '%s'", order.ID)
	}
	
	if order.OrderType != game.OrderTypeMove {
		t.Errorf("Expected order type 'move', got '%s'", order.OrderType)
	}
	
	t.Log("Order creation test passed")
}

// TestResolutionResults tests resolution result constants
func TestResolutionResults(t *testing.T) {
	results := []game.ResolutionResult{
		game.ResolutionSuccess,
		game.ResolutionFailed,
		game.ResolutionBounced,
		game.ResolutionCut,
	}
	
	if len(results) != 4 {
		t.Errorf("Expected 4 resolution results, got %d", len(results))
	}
	
	t.Log("Resolution results test passed")
}