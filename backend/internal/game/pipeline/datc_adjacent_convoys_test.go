package pipeline

import (
	"diplomacy-cli/backend/internal/game"
	"testing"
)

// DATC 6.G - CONVOYING TO ADJACENT PROVINCES Test Cases
// Tests for convoy mechanics when moving to adjacent provinces

func TestDATCG1_TwoUnitsCanSwapProvincesByConvoy(t *testing.T) {
	// 6.G.1 - TWO UNITS CAN SWAP PROVINCES BY CONVOY
	// The only way to swap two units, is by convoy.
	gameState := CreateDATCGameState(t)

	// Setup: Place units for the test scenario
	gameState.Board.PlaceUnit(&game.Unit{Type: game.Army, Owner: game.England, Province: "norway"})
	gameState.Board.PlaceUnit(&game.Unit{Type: game.Fleet, Owner: game.England, Province: "skagerrak"})
	gameState.Board.PlaceUnit(&game.Unit{Type: game.Army, Owner: game.Russia, Province: "sweden"})

	// Orders: England attempts to convoy army to Sweden, Russia moves to Norway
	gameState.AddRawOrder(game.England, "A Norway - Sweden")
	gameState.AddRawOrder(game.England, "F Skagerrak Convoys A Norway - Sweden")
	gameState.AddRawOrder(game.Russia, "A Sweden - Norway")

	// Process the turn
	result := ProcessDATCTest(t, gameState)
	if result.ProcessError != nil {
		t.Fatalf("❌ 6.G.1: Processing failed: %v", result.ProcessError)
	}

	// Verify the result using the new game state
	newState := result.NewGameState

	// Debug output for verification
	t.Logf("🔍 Final unit positions:")
	for province, unit := range newState.Board.Units {
		if unit != nil {
			t.Logf("  %s: %s %s", province, unit.Owner, unit.Type)
		}
	}

	// Expected outcome: Both armies successfully swap positions
	englishUnit := newState.Board.GetUnit("sweden")
	if englishUnit == nil || englishUnit.Owner != game.England {
		t.Errorf("❌ 6.G.1: English army should have moved to Sweden via convoy")
	}

	russianUnit := newState.Board.GetUnit("norway")
	if russianUnit == nil || russianUnit.Owner != game.Russia {
		t.Errorf("❌ 6.G.1: Russian army should have moved to Norway")
	}

	// Fleet should remain in place
	fleetUnit := newState.Board.GetUnit("skagerrak")
	if fleetUnit == nil || fleetUnit.Owner != game.England {
		t.Errorf("❌ 6.G.1: English fleet should remain in Skagerrak")
	}

	t.Logf("✅ 6.G.1: Two units can swap provinces by convoy - PASSED")
}
