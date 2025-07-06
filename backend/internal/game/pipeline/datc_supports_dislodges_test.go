package pipeline

import (
	"testing"

	"diplomacy-cli/backend/internal/game"
	"diplomacy-cli/backend/internal/game/loader"
)

// Test 6.D.1: THE SIMPLEST SUPPORT TO HOLD ORDER
func TestDATCD1_SimplestSupportToHold(t *testing.T) {
	// Load board directly without cache to ensure we get the fixed loader
	mapLoader := loader.NewJSONLoader("../../../../backend/data/classic")
	board, err := mapLoader.LoadBoard()
	if err != nil {
		t.Fatalf("Failed to load board: %v", err)
	}
	gameState := game.NewGameState(board, game.SpringMovement, 1901)

	// Add units based on orders (using normalized province names)
	gameState.Board.Units["adriatic_sea"] = &game.Unit{
		Type:     game.Fleet,
		Owner:    game.Austria,
		Province: "adriatic_sea",
	}
	gameState.Board.Units["trieste"] = &game.Unit{
		Type:     game.Army,
		Owner:    game.Austria,
		Province: "trieste",
	}
	gameState.Board.Units["venice"] = &game.Unit{
		Type:     game.Army,
		Owner:    game.Italy,
		Province: "venice",
	}
	gameState.Board.Units["tyrolia"] = &game.Unit{
		Type:     game.Army,
		Owner:    game.Italy,
		Province: "tyrolia",
	}

	// Add orders (using simple format that parser supports)
	gameState.RawOrders[game.Austria] = []string{
		"adriatic_sea s trieste - venice", // Fleet supports army move
		"a trieste - venice",              // Army move
	}
	gameState.RawOrders[game.Italy] = []string{
		"a venice hold",    // Army hold
		"tyrolia s venice", // Army supports army hold
	}

	// Process turn
	result := ProcessDATCTest(t, gameState)
	// Expected: Austria's supported attack should succeed, dislodging Venice
	ValidateExpectedOutcome(t, result, "Austria's supported attack succeeds, Venice is dislodged.", "6.D.1")
}

// Test 6.D.2: THE SIMPLEST SUPPORT ON HOLD CUT
func TestDATCD2_SimplestSupportOnHoldCut(t *testing.T) {
	gameState := CreateDATCGameState(t)

	// Add units based on orders (using normalized province names)
	gameState.Board.Units["adriatic_sea"] = &game.Unit{
		Type:     game.Fleet,
		Owner:    game.Austria,
		Province: "adriatic_sea",
	}
	gameState.Board.Units["trieste"] = &game.Unit{
		Type:     game.Army,
		Owner:    game.Austria,
		Province: "trieste",
	}
	gameState.Board.Units["vienna"] = &game.Unit{
		Type:     game.Army,
		Owner:    game.Austria,
		Province: "vienna",
	}
	gameState.Board.Units["venice"] = &game.Unit{
		Type:     game.Army,
		Owner:    game.Italy,
		Province: "venice",
	}
	gameState.Board.Units["tyrolia"] = &game.Unit{
		Type:     game.Army,
		Owner:    game.Italy,
		Province: "tyrolia",
	}

	// Add orders (using parser-compatible format with unit types)
	gameState.RawOrders[game.Austria] = []string{
		"f adriatic_sea s trieste - venice",
		"a trieste - venice",
		"a vienna - tyrolia",
	}
	gameState.RawOrders[game.Italy] = []string{
		"a venice hold",
		"a tyrolia s venice",
	}

	// Process turn
	result := ProcessDATCTest(t, gameState)

	// Expected: Vienna attacks Tyrolia, cutting support. Austria's attack succeeds.
	ValidateExpectedOutcome(t, result, "Vienna cuts support from Tyrolia, Austria's attack succeeds.", "6.D.2")
}

// Test 6.D.3: A MOVE CUTS SUPPORT ON MOVE
func TestDATCD3_MoveCutsSupportOnMove(t *testing.T) {
	gameState := CreateDATCGameState(t)

	// Add units based on orders (using normalized province names)
	gameState.Board.Units["adriatic_sea"] = &game.Unit{
		Type:     game.Fleet,
		Owner:    game.Austria,
		Province: "adriatic_sea",
	}
	gameState.Board.Units["trieste"] = &game.Unit{
		Type:     game.Army,
		Owner:    game.Austria,
		Province: "trieste",
	}
	gameState.Board.Units["venice"] = &game.Unit{
		Type:     game.Army,
		Owner:    game.Italy,
		Province: "venice",
	}
	gameState.Board.Units["ionian_sea"] = &game.Unit{
		Type:     game.Fleet,
		Owner:    game.Italy,
		Province: "ionian_sea",
	}

	// Add orders (using parser-compatible format with unit types)
	gameState.RawOrders[game.Austria] = []string{
		"f adriatic_sea s trieste - venice",
		"a trieste - venice",
	}
	gameState.RawOrders[game.Italy] = []string{
		"a venice hold",
		"f ionian_sea - adriatic_sea",
	}

	// Process turn
	result := ProcessDATCTest(t, gameState)

	// Expected: Italian fleet cuts Austrian fleet's support, so Venice holds
	ValidateExpectedOutcome(t, result, "Support is cut, Venice holds.", "6.D.3")
}

// Test 6.D.4: SUPPORT TO HOLD ON UNIT SUPPORTING A HOLD ALLOWED
func TestDATCD4_SupportToHoldOnUnitSupportingHoldAllowed(t *testing.T) {
	gameState := CreateDATCGameState(t)

	// Add units based on orders (using normalized province names)
	gameState.Board.Units["berlin"] = &game.Unit{
		Type:     game.Army,
		Owner:    game.Germany,
		Province: "berlin",
	}
	gameState.Board.Units["kiel"] = &game.Unit{
		Type:     game.Fleet,
		Owner:    game.Germany,
		Province: "kiel",
	}
	gameState.Board.Units["baltic_sea"] = &game.Unit{
		Type:     game.Fleet,
		Owner:    game.Russia,
		Province: "baltic_sea",
	}
	gameState.Board.Units["prussia"] = &game.Unit{
		Type:     game.Army,
		Owner:    game.Russia,
		Province: "prussia",
	}

	// Add orders (using parser-compatible format with unit types)
	gameState.RawOrders[game.Germany] = []string{
		"a berlin s kiel",
		"f kiel s berlin",
	}
	gameState.RawOrders[game.Russia] = []string{
		"f baltic_sea s prussia - berlin",
		"a prussia - berlin",
	}

	// Process turn
	result := ProcessDATCTest(t, gameState)

	// Expected: German mutual support holds, Russian attack fails
	ValidateExpectedOutcome(t, result, "The Russian move from Prussia to Berlin fails.", "6.D.4")
}

// Test 6.D.5: SUPPORT TO HOLD ON UNIT SUPPORTING A MOVE ALLOWED
func TestDATCD5_SupportToHoldOnUnitSupportingMoveAllowed(t *testing.T) {
	gameState := CreateDATCGameState(t)

	// Add units based on orders (using normalized province names)
	gameState.Board.Units["berlin"] = &game.Unit{
		Type:     game.Army,
		Owner:    game.Germany,
		Province: "berlin",
	}
	gameState.Board.Units["kiel"] = &game.Unit{
		Type:     game.Fleet,
		Owner:    game.Germany,
		Province: "kiel",
	}
	gameState.Board.Units["munich"] = &game.Unit{
		Type:     game.Army,
		Owner:    game.Germany,
		Province: "munich",
	}
	gameState.Board.Units["baltic_sea"] = &game.Unit{
		Type:     game.Fleet,
		Owner:    game.Russia,
		Province: "baltic_sea",
	}
	gameState.Board.Units["prussia"] = &game.Unit{
		Type:     game.Army,
		Owner:    game.Russia,
		Province: "prussia",
	}

	// Add orders (using parser-compatible format with unit types)
	gameState.RawOrders[game.Germany] = []string{
		"a berlin s munich - silesia",
		"f kiel s berlin",
		"a munich - silesia",
	}
	gameState.RawOrders[game.Russia] = []string{
		"f baltic_sea s prussia - berlin",
		"a prussia - berlin",
	}

	// Process turn
	result := ProcessDATCTest(t, gameState)

	// Expected: German Berlin holds with support, Munich moves to Silesia, Russian attack fails
	ValidateExpectedOutcome(t, result, "The Russian move from Prussia to Berlin fails.", "6.D.5")
}
