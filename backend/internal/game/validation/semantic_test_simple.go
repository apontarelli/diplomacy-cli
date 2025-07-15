package validation

import (
	"diplomacy-cli/backend/internal/game"
	"diplomacy-cli/backend/internal/game/loader"
	"testing"
)

func createSimpleTestGameState(t *testing.T) *game.GameState {
	jsonLoader := loader.NewJSONLoader("../../../data/classic")
	board, err := jsonLoader.LoadBoard()
	if err != nil {
		t.Fatalf("Failed to load board: %v", err)
	}

	gameState := &game.GameState{
		Board: board,
		Phase: game.SpringMovement,
		Year:  1901,
	}

	return gameState
}

func TestSemanticValidatorSimple(t *testing.T) {
	gameState := createSimpleTestGameState(t)
	validator := NewSemanticValidator(gameState.Board)

	// Test valid army move
	gameState.Board.Units = map[string]*game.Unit{
		"paris": {Type: game.Army, Owner: game.France, Province: "paris"},
	}

	order := &game.Order{
		Type:     game.Move,
		UnitType: game.Army,
		From:     "paris",
		To:       "burgundy",
	}

	err := validator.ValidateOrder(order, "A Par - Bur", gameState)
	if err != nil {
		t.Errorf("Expected valid move to pass, got error: %v", err)
	}

	// Test unit not found
	gameState.Board.Units = map[string]*game.Unit{}

	err = validator.ValidateOrder(order, "A Par - Bur", gameState)
	if err == nil {
		t.Errorf("Expected error for missing unit, got none")
	}

	if semanticErr, ok := err.(SemanticError); ok {
		if semanticErr.Type != "unit_not_found" {
			t.Errorf("Expected error type unit_not_found, got %s", semanticErr.Type)
		}
	}

	// Test army to sea
	gameState.Board.Units = map[string]*game.Unit{
		"liverpool": {Type: game.Army, Owner: game.England, Province: "liverpool"},
	}

	order = &game.Order{
		Type:     game.Move,
		UnitType: game.Army,
		From:     "liverpool",
		To:       "irish_sea",
	}

	err = validator.ValidateOrder(order, "A Lvp - IRI", gameState)
	if err == nil {
		t.Errorf("Expected error for army to sea, got none")
	}

	if semanticErr, ok := err.(SemanticError); ok {
		if semanticErr.Type != "invalid_destination_for_unit_type" {
			t.Errorf("Expected error type invalid_destination_for_unit_type, got %s", semanticErr.Type)
		}
	}

	// Test valid hold
	gameState.Board.Units = map[string]*game.Unit{
		"paris": {Type: game.Army, Owner: game.France, Province: "paris"},
	}

	order = &game.Order{
		Type:     game.Hold,
		UnitType: game.Army,
		From:     "paris",
	}

	err = validator.ValidateOrder(order, "A Par H", gameState)
	if err != nil {
		t.Errorf("Expected valid hold to pass, got error: %v", err)
	}

	// Test valid support
	gameState.Board.Units = map[string]*game.Unit{
		"burgundy": {Type: game.Army, Owner: game.France, Province: "burgundy"},
		"paris":    {Type: game.Army, Owner: game.France, Province: "paris"},
	}

	order = &game.Order{
		Type:          game.Support,
		UnitType:      game.Army,
		From:          "burgundy",
		SupportTarget: "paris",
	}

	err = validator.ValidateOrder(order, "A Bur S Par", gameState)
	if err != nil {
		t.Errorf("Expected valid support to pass, got error: %v", err)
	}
}
