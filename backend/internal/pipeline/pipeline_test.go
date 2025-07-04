package pipeline

import (
	"diplomacy-cli/backend/internal/game"
	"diplomacy-cli/backend/internal/game/loader"
	"testing"
)

// TestEndToEndPipeline tests the complete pipeline from raw orders to final game state
func TestEndToEndPipeline(t *testing.T) {
	// Load the classic map
	mapLoader := loader.NewJSONLoader("../../data/classic")
	board, err := mapLoader.LoadBoard()
	if err != nil {
		t.Fatalf("Failed to load board: %v", err)
	}

	// TODO: Complete this test implementation
	_ = board // Use board to avoid unused variable error
}

// TestProcessTurnWithConflicts tests the pipeline with conflicting orders
func TestProcessTurnWithConflicts(t *testing.T) {
	// Create a simple test board
	board := game.NewBoard()

	// Add test provinces
	board.AddProvince(&game.Province{
		Name:          "paris",
		ShortCode:     "par",
		DisplayName:   "Paris",
		Type:          game.Land,
		SupplyCenter:  true,
		ArmyNeighbors: []string{"burgundy", "picardy"},
	})

	board.AddProvince(&game.Province{
		Name:          "burgundy",
		ShortCode:     "bur",
		DisplayName:   "Burgundy",
		Type:          game.Land,
		SupplyCenter:  false,
		ArmyNeighbors: []string{"paris", "munich"},
	})

	board.AddProvince(&game.Province{
		Name:          "munich",
		ShortCode:     "mun",
		DisplayName:   "Munich",
		Type:          game.Land,
		SupplyCenter:  true,
		ArmyNeighbors: []string{"burgundy"},
	})

	// Place units
	board.PlaceUnit(&game.Unit{Type: game.Army, Owner: game.France, Province: "paris"})
	board.PlaceUnit(&game.Unit{Type: game.Army, Owner: game.Germany, Province: "munich"})

	// Create game state
	gameState := game.NewGameState(board, game.SpringMovement, 1901)

	// Add conflicting orders (both armies move to burgundy)
	gameState.AddRawOrder(game.France, "A Par-Bur")
	gameState.AddRawOrder(game.Germany, "A Mun-Bur")

	// Process the turn
	processor := NewTurnProcessor()
	newState, err := processor.ProcessTurn(gameState)
	if err != nil {
		t.Fatalf("Failed to process turn with conflicts: %v", err)
	}

	// Verify that both units bounced (stayed in original positions)
	frenchUnit := newState.Board.GetUnit("paris")
	germanUnit := newState.Board.GetUnit("munich")
	burgundyUnit := newState.Board.GetUnit("burgundy")

	if frenchUnit == nil {
		t.Error("French unit should still be in Paris after bounce")
	}

	if germanUnit == nil {
		t.Error("German unit should still be in Munich after bounce")
	}

	if burgundyUnit != nil {
		t.Error("No unit should be in Burgundy after bounce")
	}
}

// TestProcessTurnWithSupport tests the pipeline with support orders
func TestProcessTurnWithSupport(t *testing.T) {
	// Create a test board
	board := game.NewBoard()

	// Add test provinces
	board.AddProvince(&game.Province{
		Name:          "paris",
		ShortCode:     "par",
		DisplayName:   "Paris",
		Type:          game.Land,
		SupplyCenter:  true,
		ArmyNeighbors: []string{"burgundy", "picardy"},
	})

	board.AddProvince(&game.Province{
		Name:          "burgundy",
		ShortCode:     "bur",
		DisplayName:   "Burgundy",
		Type:          game.Land,
		SupplyCenter:  false,
		ArmyNeighbors: []string{"paris", "munich", "picardy"},
	})

	board.AddProvince(&game.Province{
		Name:          "picardy",
		ShortCode:     "pic",
		DisplayName:   "Picardy",
		Type:          game.Land,
		SupplyCenter:  false,
		ArmyNeighbors: []string{"paris", "burgundy"},
	})

	board.AddProvince(&game.Province{
		Name:          "munich",
		ShortCode:     "mun",
		DisplayName:   "Munich",
		Type:          game.Land,
		SupplyCenter:  true,
		ArmyNeighbors: []string{"burgundy"},
	})

	// Place units
	board.PlaceUnit(&game.Unit{Type: game.Army, Owner: game.France, Province: "paris"})
	board.PlaceUnit(&game.Unit{Type: game.Army, Owner: game.France, Province: "picardy"})
	board.PlaceUnit(&game.Unit{Type: game.Army, Owner: game.Germany, Province: "munich"})

	// Create game state
	gameState := game.NewGameState(board, game.SpringMovement, 1901)

	// Add orders: French army moves to Burgundy with support, German army also moves to Burgundy
	gameState.AddRawOrder(game.France, "A Par-Bur")
	gameState.AddRawOrder(game.France, "A Pic S Par-Bur")
	gameState.AddRawOrder(game.Germany, "A Mun-Bur")

	// Process the turn
	processor := NewTurnProcessor()
	newState, err := processor.ProcessTurn(gameState)
	if err != nil {
		t.Fatalf("Failed to process turn with support: %v", err)
	}

	// Verify that French army succeeded (with support) and German army bounced
	frenchUnit := newState.Board.GetUnit("burgundy")
	germanUnit := newState.Board.GetUnit("munich")
	supportUnit := newState.Board.GetUnit("picardy")

	if frenchUnit == nil || frenchUnit.Owner != game.France {
		t.Error("French unit should have moved to Burgundy with support")
	}

	if germanUnit == nil {
		t.Error("German unit should have bounced back to Munich")
	}

	if supportUnit == nil || supportUnit.Owner != game.France {
		t.Error("Supporting unit should remain in Picardy")
	}
}
