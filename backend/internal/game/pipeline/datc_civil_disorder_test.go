package pipeline

import (
	"testing"

	"diplomacy-cli/backend/internal/game"
)

// DATC 6.J - CIVIL DISORDER AND DISBANDS Test Cases
// Tests for automatic unit disbanding when players don't provide proper orders

// Test 6.J.1: TOO MANY DISBAND ORDERS
// Original DATC test - check how program reacts when someone orders too many disbands
func TestDATCJ1_TooManyDisbandOrders(t *testing.T) {
	gameState := CreateDATCGameState(t)

	// Change to winter build phase for disband orders
	gameState.Phase = game.WinterBuild

	// Add units for France - more than they can support
	gameState.Board.Units["gulf_of_lyon"] = &game.Unit{
		Type:     game.Fleet,
		Owner:    game.France,
		Province: "gulf_of_lyon",
	}
	gameState.Board.Units["picardy"] = &game.Unit{
		Type:     game.Army,
		Owner:    game.France,
		Province: "picardy",
	}
	gameState.Board.Units["paris"] = &game.Unit{
		Type:     game.Army,
		Owner:    game.France,
		Province: "paris",
	}

	// France orders too many disbands
	gameState.RawOrders[game.France] = []string{
		"remove f gulf_of_lyon",
		"remove a picardy",
		"remove a paris",
	}

	// Process turn
	result := ProcessDATCTest(t, gameState)

	// Expected: Program should handle too many disband orders gracefully
	ValidateExpectedOutcome(t, result, "Check how program reacts when someone orders too many disbands.", "6.J.1")
}

// Test 6.J.2: REMOVING THE SAME UNIT TWICE
// Original DATC test - check reaction to duplicate disband orders
func TestDATCJ2_RemovingTheSameUnitTwice(t *testing.T) {
	gameState := CreateDATCGameState(t)

	// Change to winter build phase
	gameState.Phase = game.WinterBuild

	// Add units for France
	gameState.Board.Units["gulf_of_lyon"] = &game.Unit{
		Type:     game.Fleet,
		Owner:    game.France,
		Province: "gulf_of_lyon",
	}
	gameState.Board.Units["picardy"] = &game.Unit{
		Type:     game.Army,
		Owner:    game.France,
		Province: "picardy",
	}

	// France orders the same unit to be removed twice
	gameState.RawOrders[game.France] = []string{
		"remove f gulf_of_lyon",
		"remove f gulf_of_lyon",
	}

	// Process turn
	result := ProcessDATCTest(t, gameState)

	// Expected: Program should handle duplicate disband orders
	ValidateExpectedOutcome(t, result, "Check reaction to removing the same unit twice.", "6.J.2")
}

// Test 6.J.3: CIVIL DISORDER TWO ARMIES WITH DIFFERENT DISTANCE
// Original DATC test - army with greater distance should be removed
func TestDATCJ3_CivilDisorderTwoArmiesWithDifferentDistance(t *testing.T) {
	gameState := CreateDATCGameState(t)

	// Change to winter build phase
	gameState.Phase = game.WinterBuild

	// Set up scenario: Russia has to remove one unit
	// Russia owns St Petersburg (home supply center)
	gameState.Board.Units["moscow"] = &game.Unit{
		Type:     game.Army,
		Owner:    game.Russia,
		Province: "moscow",
	}
	gameState.Board.Units["warsaw"] = &game.Unit{
		Type:     game.Army,
		Owner:    game.Russia,
		Province: "warsaw",
	}

	// Russia does not order a disband (civil disorder)
	gameState.RawOrders[game.Russia] = []string{}

	// Process turn
	result := ProcessDATCTest(t, gameState)

	// Expected: Army with greater distance from home supply center is removed
	ValidateExpectedOutcome(t, result, "If two armies have different distance from the home supply centers, then the army with the greatest distance has to be removed.", "6.J.3")
}

// Test 6.J.4: CIVIL DISORDER TWO ARMIES WITH EQUAL DISTANCE
// Original DATC test - alphabetical order when armies have equal distance
func TestDATCJ4_CivilDisorderTwoArmiesWithEqualDistance(t *testing.T) {
	gameState := CreateDATCGameState(t)

	// Change to winter build phase
	gameState.Phase = game.WinterBuild

	// Set up scenario: Russia has armies with equal distance
	gameState.Board.Units["livonia"] = &game.Unit{
		Type:     game.Army,
		Owner:    game.Russia,
		Province: "livonia",
	}
	gameState.Board.Units["finland"] = &game.Unit{
		Type:     game.Army,
		Owner:    game.Russia,
		Province: "finland",
	}

	// Russia does not order a disband (civil disorder)
	gameState.RawOrders[game.Russia] = []string{}

	// Process turn
	result := ProcessDATCTest(t, gameState)

	// Expected: Alphabetical order is used when armies have equal distance
	ValidateExpectedOutcome(t, result, "Alphabetical order is used, when two armies have equal distance to the home supply centers.", "6.J.4")
}

// Test 6.J.5: CIVIL DISORDER TWO FLEETS WITH DIFFERENT DISTANCE
// Original DATC test - fleet with greater distance should be removed
func TestDATCJ5_CivilDisorderTwoFleetsWithDifferentDistance(t *testing.T) {
	gameState := CreateDATCGameState(t)

	// Change to winter build phase
	gameState.Phase = game.WinterBuild

	// Set up scenario: Russia has fleets with different distances
	gameState.Board.Units["skagerrak"] = &game.Unit{
		Type:     game.Fleet,
		Owner:    game.Russia,
		Province: "skagerrak",
	}
	gameState.Board.Units["berlin"] = &game.Unit{
		Type:     game.Fleet,
		Owner:    game.Russia,
		Province: "berlin",
	}

	// Russia does not order a disband (civil disorder)
	gameState.RawOrders[game.Russia] = []string{}

	// Process turn
	result := ProcessDATCTest(t, gameState)

	// Expected: Fleet with greater distance is removed (Berlin distance 3, Skagerrak distance 2)
	ValidateExpectedOutcome(t, result, "If two fleets have different distance from the home supply centers, then the fleet with the greatest distance has to be removed. Note that fleets cannot go over land.", "6.J.5")
}

// Test 6.J.6: CIVIL DISORDER TWO FLEETS WITH EQUAL DISTANCE
// Original DATC test - alphabetical order when fleets have equal distance
func TestDATCJ6_CivilDisorderTwoFleetsWithEqualDistance(t *testing.T) {
	gameState := CreateDATCGameState(t)

	// Change to winter build phase
	gameState.Phase = game.WinterBuild

	// Set up scenario: Russia has fleets with equal distance
	gameState.Board.Units["barents_sea"] = &game.Unit{
		Type:     game.Fleet,
		Owner:    game.Russia,
		Province: "barents_sea",
	}
	gameState.Board.Units["norway"] = &game.Unit{
		Type:     game.Fleet,
		Owner:    game.Russia,
		Province: "norway",
	}

	// Russia does not order a disband (civil disorder)
	gameState.RawOrders[game.Russia] = []string{}

	// Process turn
	result := ProcessDATCTest(t, gameState)

	// Expected: Alphabetical order is used when fleets have equal distance
	ValidateExpectedOutcome(t, result, "Alphabetical order is used, when two fleets have equal distance to the home supply centers.", "6.J.6")
}

// Test 6.J.7: CIVIL DISORDER TWO FLEETS AND ONE ARMY WITH EQUAL DISTANCE
// Original DATC test - unit type priority when distances are equal
func TestDATCJ7_CivilDisorderTwoFleetsAndOneArmyWithEqualDistance(t *testing.T) {
	gameState := CreateDATCGameState(t)

	// Change to winter build phase
	gameState.Phase = game.WinterBuild

	// Set up scenario: Russia has mixed units with equal distance
	gameState.Board.Units["livonia"] = &game.Unit{
		Type:     game.Army,
		Owner:    game.Russia,
		Province: "livonia",
	}
	gameState.Board.Units["barents_sea"] = &game.Unit{
		Type:     game.Fleet,
		Owner:    game.Russia,
		Province: "barents_sea",
	}
	gameState.Board.Units["norway"] = &game.Unit{
		Type:     game.Fleet,
		Owner:    game.Russia,
		Province: "norway",
	}

	// Russia does not order a disband (civil disorder)
	gameState.RawOrders[game.Russia] = []string{}

	// Process turn
	result := ProcessDATCTest(t, gameState)

	// Expected: Unit type and alphabetical order determine removal
	ValidateExpectedOutcome(t, result, "Unit type priority and alphabetical order when distances are equal.", "6.J.7")
}

// Test 6.J.8: CIVIL DISORDER A FLEET WITH TWO COASTS
// Original DATC test - fleet on coast with distance calculation
func TestDATCJ8_CivilDisorderFleetWithTwoCoasts(t *testing.T) {
	gameState := CreateDATCGameState(t)

	// Change to winter build phase
	gameState.Phase = game.WinterBuild

	// Set up scenario: Russia has fleet on coast
	gameState.Board.Units["st_petersburg"] = &game.Unit{
		Type:     game.Fleet,
		Owner:    game.Russia,
		Province: "st_petersburg",
		Coast:    "sc", // South coast
	}
	gameState.Board.Units["moscow"] = &game.Unit{
		Type:     game.Army,
		Owner:    game.Russia,
		Province: "moscow",
	}

	// Russia does not order a disband (civil disorder)
	gameState.RawOrders[game.Russia] = []string{}

	// Process turn
	result := ProcessDATCTest(t, gameState)

	// Expected: Fleet on coast distance calculation
	ValidateExpectedOutcome(t, result, "Fleet with two coasts distance calculation scenario.", "6.J.8")
}

// Test 6.J.9: CIVIL DISORDER MUST RETREAT
// Original DATC test - civil disorder during retreat phase
func TestDATCJ9_CivilDisorderMustRetreat(t *testing.T) {
	gameState := CreateDATCGameState(t)

	// Change to retreat phase
	gameState.Phase = game.FallRetreat

	// Set up scenario: unit must retreat but no orders given
	gameState.Board.Units["moscow"] = &game.Unit{
		Type:     game.Army,
		Owner:    game.Russia,
		Province: "moscow",
	}

	// Add dislodged unit that needs to retreat
	gameState.DislodgedUnits = append(gameState.DislodgedUnits, &game.Unit{
		Type:     game.Army,
		Owner:    game.Russia,
		Province: "warsaw",
	})

	// Russia does not order a retreat (civil disorder)
	gameState.RawOrders[game.Russia] = []string{}

	// Process turn
	result := ProcessDATCTest(t, gameState)

	// Expected: Unit is disbanded when no retreat order given
	ValidateExpectedOutcome(t, result, "Civil disorder during retreat phase - unit disbanded.", "6.J.9")
}

// Test 6.J.10: CIVIL DISORDER COUNTING CONVOYING DISTANCE
// Original DATC test - distance calculation considering convoy routes
func TestDATCJ10_CivilDisorderCountingConvoyingDistance(t *testing.T) {
	gameState := CreateDATCGameState(t)

	// Change to winter build phase
	gameState.Phase = game.WinterBuild

	// Set up scenario: Italy has armies with different convoy distances
	gameState.Board.Units["greece"] = &game.Unit{
		Type:     game.Army,
		Owner:    game.Italy,
		Province: "greece",
	}
	gameState.Board.Units["piedmont"] = &game.Unit{
		Type:     game.Army,
		Owner:    game.Italy,
		Province: "piedmont",
	}

	// Italy does not order a disband (civil disorder)
	gameState.RawOrders[game.Italy] = []string{}

	// Process turn
	result := ProcessDATCTest(t, gameState)

	// Expected: Distance calculation considers convoy routes (Greece can reach Naples in 2 via convoy, Piedmont farther)
	ValidateExpectedOutcome(t, result, "For calculating the distance for armies all areas must be considered.", "6.J.10")
}

// Test 6.J.11: DISTANCE TO OWNED SUPPLY CENTER
// Original DATC test - distance calculated to owned supply center (2023 rules)
func TestDATCJ11_DistanceToOwnedSupplyCenter(t *testing.T) {
	gameState := CreateDATCGameState(t)

	// Change to winter build phase
	gameState.Phase = game.WinterBuild

	// Set up scenario: distance to owned supply center vs home supply center
	gameState.Board.Units["moscow"] = &game.Unit{
		Type:     game.Army,
		Owner:    game.Russia,
		Province: "moscow",
	}
	gameState.Board.Units["warsaw"] = &game.Unit{
		Type:     game.Army,
		Owner:    game.Russia,
		Province: "warsaw",
	}

	// Russia does not order a disband (civil disorder)
	gameState.RawOrders[game.Russia] = []string{}

	// Process turn
	result := ProcessDATCTest(t, gameState)

	// Expected: Distance calculated to owned supply center (2023 rules)
	ValidateExpectedOutcome(t, result, "The 2023 rules say that distance must be calculated to owned supply center instead of home supply center (as it was in the older rulebooks).", "6.J.11")
}
