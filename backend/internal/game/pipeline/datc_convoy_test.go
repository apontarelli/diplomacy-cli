package pipeline

import (
	"testing"

	"diplomacy-cli/backend/internal/game"
)

// Test 6.F.1: NO CONVOY IN COASTAL AREAS
// Original DATC test - Constantinople cannot convoy (coastal area)
func TestDATCF1_NoConvoyInCoastalAreas(t *testing.T) {
	gameState := CreateDATCGameState(t)

	// Add units exactly as in original DATC test
	gameState.Board.Units["greece"] = &game.Unit{
		Type:     game.Army,
		Owner:    game.Turkey,
		Province: "greece",
	}
	gameState.Board.Units["aegean_sea"] = &game.Unit{
		Type:     game.Fleet,
		Owner:    game.Turkey,
		Province: "aegean_sea",
	}
	gameState.Board.Units["constantinople"] = &game.Unit{
		Type:     game.Fleet,
		Owner:    game.Turkey,
		Province: "constantinople",
	}
	gameState.Board.Units["black_sea"] = &game.Unit{
		Type:     game.Fleet,
		Owner:    game.Turkey,
		Province: "black_sea",
	}

	// Original DATC orders
	gameState.RawOrders[game.Turkey] = []string{
		"a greece - sevastopol",
		"f aegean_sea c greece - sevastopol",
		"f constantinople c greece - sevastopol",
		"f black_sea c greece - sevastopol",
	}

	// Process turn
	result := ProcessDATCTest(t, gameState)

	// Expected: Constantinople convoy fails (coastal), so army doesn't move
	ValidateExpectedOutcome(t, result, "The convoy in Constantinople is not possible. So, the army in Greece will not move to Sevastopol.", "6.F.1")
}

// Test 6.F.2: AN ARMY BEING CONVOYED CAN BOUNCE AS NORMAL
// Original DATC test - convoy vs direct move bounce
func TestDATCF2_ArmyBeingConvoyedCanBounce(t *testing.T) {
	gameState := CreateDATCGameState(t)

	// Add units exactly as in original DATC test
	gameState.Board.Units["london"] = &game.Unit{
		Type:     game.Army,
		Owner:    game.England,
		Province: "london",
	}
	gameState.Board.Units["english_channel"] = &game.Unit{
		Type:     game.Fleet,
		Owner:    game.England,
		Province: "english_channel",
	}
	gameState.Board.Units["paris"] = &game.Unit{
		Type:     game.Army,
		Owner:    game.France,
		Province: "paris",
	}

	// Original DATC orders - both armies target Brest
	gameState.RawOrders[game.England] = []string{
		"f english_channel c london - brest",
		"a london - brest",
	}
	gameState.RawOrders[game.France] = []string{
		"a paris - brest",
	}

	// Process turn
	result := ProcessDATCTest(t, gameState)

	// Expected: Both armies bounce, neither moves
	ValidateExpectedOutcome(t, result, "The English army in London bounces on the French army in Paris. Both units do not move.", "6.F.2")
}

// Test 6.F.3: AN ARMY BEING CONVOYED CAN RECEIVE SUPPORT
// Original DATC test - convoy with support beats direct move
func TestDATCF3_ArmyBeingConvoyedCanReceiveSupport(t *testing.T) {
	gameState := CreateDATCGameState(t)

	// Add units exactly as in original DATC test
	gameState.Board.Units["london"] = &game.Unit{
		Type:     game.Army,
		Owner:    game.England,
		Province: "london",
	}
	gameState.Board.Units["english_channel"] = &game.Unit{
		Type:     game.Fleet,
		Owner:    game.England,
		Province: "english_channel",
	}
	gameState.Board.Units["midatlantic_ocean"] = &game.Unit{
		Type:     game.Fleet,
		Owner:    game.England,
		Province: "midatlantic_ocean",
	}
	gameState.Board.Units["paris"] = &game.Unit{
		Type:     game.Army,
		Owner:    game.France,
		Province: "paris",
	}

	// Original DATC orders - convoy with support vs direct move
	gameState.RawOrders[game.England] = []string{
		"f english_channel c london - brest",
		"a london - brest",
		"f midatlantic_ocean s london - brest",
	}
	gameState.RawOrders[game.France] = []string{
		"a paris - brest",
	}

	// Process turn
	result := ProcessDATCTest(t, gameState)

	// Expected: London wins due to support, Paris stays
	ValidateExpectedOutcome(t, result, "The army in London receives support and beats the army in Paris. This means that the army London will end in Brest and the French army in Paris stays in Paris.", "6.F.3")
}

// Test: Simple convoy success case (for basic convoy validation)
func TestDATCF_SimpleConvoySuccess(t *testing.T) {
	gameState := CreateDATCGameState(t)

	// Add units: Army in London, Fleet in English Channel
	gameState.Board.Units["london"] = &game.Unit{
		Type:     game.Army,
		Owner:    game.England,
		Province: "london",
	}
	gameState.Board.Units["english_channel"] = &game.Unit{
		Type:     game.Fleet,
		Owner:    game.England,
		Province: "english_channel",
	}

	// Add orders: Army move + Fleet convoy (no opposition)
	gameState.RawOrders[game.England] = []string{
		"a london - picardy",
		"f english_channel c london - picardy",
	}

	// Process turn
	result := ProcessDATCTest(t, gameState)

	// Expected: Army successfully moves via convoy
	ValidateExpectedOutcome(t, result, "Army successfully moves from London to Picardy via convoy.", "Simple Convoy")
}
