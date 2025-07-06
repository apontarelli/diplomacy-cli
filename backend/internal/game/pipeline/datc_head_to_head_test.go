package pipeline

import (
	"testing"

	"diplomacy-cli/backend/internal/game"
)

// Test 6.H.1: DISLOGED UNIT HAS NO EFFECT ON ATTACKERS AREA
func TestDATCH1_DislogedUnitHasNoEffectOnAttackersArea(t *testing.T) {
	gameState := CreateDATCGameState(t)

	// Add units based on test case
	gameState.Board.Units["berlin"] = &game.Unit{
		Type:     game.Army,
		Owner:    game.Germany,
		Province: "berlin",
	}
	gameState.Board.Units["munich"] = &game.Unit{
		Type:     game.Army,
		Owner:    game.Germany,
		Province: "munich",
	}
	gameState.Board.Units["prussia"] = &game.Unit{
		Type:     game.Army,
		Owner:    game.Russia,
		Province: "prussia",
	}

	// Add orders
	gameState.RawOrders[game.Germany] = []string{
		"a berlin - prussia",
		"a munich s berlin - prussia",
	}
	gameState.RawOrders[game.Russia] = []string{
		"a prussia - berlin",
	}

	// Process turn
	result := ProcessDATCTest(t, gameState)

	// Expected: German army succeeds with support, Russian army is dislodged
	ValidateExpectedOutcome(t, result, "German army succeeds with support, Russian army is dislodged.", "6.H.1")
}

// Test 6.H.2: NO SELF DISLODGEMENT
func TestDATCH2_NoSelfDislodgement(t *testing.T) {
	gameState := CreateDATCGameState(t)

	// Add units based on test case
	gameState.Board.Units["berlin"] = &game.Unit{
		Type:     game.Army,
		Owner:    game.Germany,
		Province: "berlin",
	}
	gameState.Board.Units["munich"] = &game.Unit{
		Type:     game.Army,
		Owner:    game.Germany,
		Province: "munich",
	}
	gameState.Board.Units["kiel"] = &game.Unit{
		Type:     game.Fleet,
		Owner:    game.Germany,
		Province: "kiel",
	}

	// Add orders - Germany tries to dislodge its own unit
	gameState.RawOrders[game.Germany] = []string{
		"a berlin - munich",
		"a munich - berlin",
		"f kiel s berlin - munich",
	}

	// Process turn
	result := ProcessDATCTest(t, gameState)

	// Expected: No units move (self-dislodgement not allowed)
	ValidateExpectedOutcome(t, result, "No units move (self-dislodgement not allowed).", "6.H.2")
}

// Test 6.H.3: NO HELP IN DISLODGING OWN UNIT
func TestDATCH3_NoHelpInDislodgingOwnUnit(t *testing.T) {
	gameState := CreateDATCGameState(t)

	// Add units based on test case
	gameState.Board.Units["berlin"] = &game.Unit{
		Type:     game.Army,
		Owner:    game.Germany,
		Province: "berlin",
	}
	gameState.Board.Units["munich"] = &game.Unit{
		Type:     game.Army,
		Owner:    game.Germany,
		Province: "munich",
	}
	gameState.Board.Units["silesia"] = &game.Unit{
		Type:     game.Army,
		Owner:    game.Austria,
		Province: "silesia",
	}

	// Add orders
	gameState.RawOrders[game.Germany] = []string{
		"a berlin hold",
		"a munich s silesia - berlin",
	}
	gameState.RawOrders[game.Austria] = []string{
		"a silesia - berlin",
	}

	// Process turn
	result := ProcessDATCTest(t, gameState)

	// Expected: Austrian attack fails (German unit cannot help dislodge its own unit)
	ValidateExpectedOutcome(t, result, "Austrian attack fails (German unit cannot help dislodge its own unit).", "6.H.3")
}
