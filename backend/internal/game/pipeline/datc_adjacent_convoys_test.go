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

// Test 6.G.2: KIDNAPPING AN ARMY
// Original DATC test - Germany provides convoy instead of expected support
func TestDATCG2_KidnappingAnArmy(t *testing.T) {
	gameState := CreateDATCGameState(t)

	// Add units exactly as in original DATC test
	gameState.Board.Units["norway"] = &game.Unit{
		Type:     game.Army,
		Owner:    game.England,
		Province: "norway",
	}
	gameState.Board.Units["sweden"] = &game.Unit{
		Type:     game.Fleet,
		Owner:    game.Russia,
		Province: "sweden",
	}
	gameState.Board.Units["skagerrak"] = &game.Unit{
		Type:     game.Fleet,
		Owner:    game.Germany,
		Province: "skagerrak",
	}

	// Original DATC orders
	gameState.RawOrders[game.England] = []string{
		"a norway - sweden",
	}
	gameState.RawOrders[game.Russia] = []string{
		"f sweden - norway",
	}
	gameState.RawOrders[game.Germany] = []string{
		"f skagerrak c norway - sweden",
	}

	// Process turn
	result := ProcessDATCTest(t, gameState)

	// Expected: Germany provides convoy, enabling the swap
	ValidateExpectedOutcome(t, result, "Germany promised England to support to dislodge the Russian fleet in Sweden and it promised Russia to support to dislodge the English army in Norway. Instead, the joking German orders a convoy.", "6.G.2")
}

// Test 6.G.3: SWAPPING WITH UNINTENDED INTENT
// Original DATC test - unintended convoy swap
func TestDATCG3_SwappingWithUnintendedIntent(t *testing.T) {
	gameState := CreateDATCGameState(t)

	// Add units for unintended convoy swap
	gameState.Board.Units["norway"] = &game.Unit{
		Type:     game.Army,
		Owner:    game.England,
		Province: "norway",
	}
	gameState.Board.Units["sweden"] = &game.Unit{
		Type:     game.Army,
		Owner:    game.Russia,
		Province: "sweden",
	}
	gameState.Board.Units["skagerrak"] = &game.Unit{
		Type:     game.Fleet,
		Owner:    game.Germany,
		Province: "skagerrak",
	}

	// Orders for unintended convoy
	gameState.RawOrders[game.England] = []string{
		"a norway - sweden",
	}
	gameState.RawOrders[game.Russia] = []string{
		"a sweden - norway",
	}
	gameState.RawOrders[game.Germany] = []string{
		"f skagerrak c norway - sweden",
	}

	// Process turn
	result := ProcessDATCTest(t, gameState)

	// Expected: Convoy enables swap despite unintended nature
	ValidateExpectedOutcome(t, result, "Swapping with unintended intent via convoy.", "6.G.3")
}

// Test 6.G.4: SWAPPING WITH INTENDED INTENT
// Original DATC test - intended convoy swap
func TestDATCG4_SwappingWithIntendedIntent(t *testing.T) {
	gameState := CreateDATCGameState(t)

	// Add units for intended convoy swap
	gameState.Board.Units["norway"] = &game.Unit{
		Type:     game.Army,
		Owner:    game.England,
		Province: "norway",
	}
	gameState.Board.Units["sweden"] = &game.Unit{
		Type:     game.Army,
		Owner:    game.Russia,
		Province: "sweden",
	}
	gameState.Board.Units["skagerrak"] = &game.Unit{
		Type:     game.Fleet,
		Owner:    game.England,
		Province: "skagerrak",
	}

	// Orders for intended convoy swap
	gameState.RawOrders[game.England] = []string{
		"a norway - sweden",
		"f skagerrak c norway - sweden",
	}
	gameState.RawOrders[game.Russia] = []string{
		"a sweden - norway",
	}

	// Process turn
	result := ProcessDATCTest(t, gameState)

	// Expected: Intended convoy swap succeeds
	ValidateExpectedOutcome(t, result, "Swapping with intended intent via convoy.", "6.G.4")
}

// Test 6.G.5: SWAPPING WITH MULTIPLE FLEETS WITH ONE OWN FLEET
// Original DATC test - multiple fleets with one own fleet
func TestDATCG5_SwappingWithMultipleFleetsWithOneOwnFleet(t *testing.T) {
	gameState := CreateDATCGameState(t)

	// Add units for multiple fleet convoy
	gameState.Board.Units["rome"] = &game.Unit{
		Type:     game.Army,
		Owner:    game.Italy,
		Province: "rome",
	}
	gameState.Board.Units["tyrrhenian_sea"] = &game.Unit{
		Type:     game.Fleet,
		Owner:    game.Italy,
		Province: "tyrrhenian_sea",
	}
	gameState.Board.Units["apulia"] = &game.Unit{
		Type:     game.Army,
		Owner:    game.Turkey,
		Province: "apulia",
	}
	gameState.Board.Units["ionian_sea"] = &game.Unit{
		Type:     game.Fleet,
		Owner:    game.Turkey,
		Province: "ionian_sea",
	}

	// Orders for multiple fleet convoy
	gameState.RawOrders[game.Italy] = []string{
		"a rome - apulia",
		"f tyrrhenian_sea c apulia - rome",
	}
	gameState.RawOrders[game.Turkey] = []string{
		"a apulia - rome",
		"f ionian_sea c apulia - rome",
	}

	// Process turn
	result := ProcessDATCTest(t, gameState)

	// Expected: One fleet is sufficient to show intent to convoy
	ValidateExpectedOutcome(t, result, "One fleet is sufficient to show the intent to convoy.", "6.G.5")
}

// Test 6.G.6: SWAPPING WITH MULTIPLE FLEETS WITH MULTIPLE OWN FLEETS
// Original DATC test - multiple fleets with multiple own fleets
func TestDATCG6_SwappingWithMultipleFleetsWithMultipleOwnFleets(t *testing.T) {
	gameState := CreateDATCGameState(t)

	// Add units for multiple own fleet convoy
	gameState.Board.Units["rome"] = &game.Unit{
		Type:     game.Army,
		Owner:    game.Italy,
		Province: "rome",
	}
	gameState.Board.Units["tyrrhenian_sea"] = &game.Unit{
		Type:     game.Fleet,
		Owner:    game.Italy,
		Province: "tyrrhenian_sea",
	}
	gameState.Board.Units["western_mediterranean"] = &game.Unit{
		Type:     game.Fleet,
		Owner:    game.Italy,
		Province: "western_mediterranean",
	}
	gameState.Board.Units["apulia"] = &game.Unit{
		Type:     game.Army,
		Owner:    game.Turkey,
		Province: "apulia",
	}
	gameState.Board.Units["ionian_sea"] = &game.Unit{
		Type:     game.Fleet,
		Owner:    game.Turkey,
		Province: "ionian_sea",
	}

	// Orders for multiple own fleet convoy
	gameState.RawOrders[game.Italy] = []string{
		"a rome - apulia",
		"f tyrrhenian_sea c rome - apulia",
		"f western_mediterranean c rome - apulia",
	}
	gameState.RawOrders[game.Turkey] = []string{
		"a apulia - rome",
		"f ionian_sea c apulia - rome",
	}

	// Process turn
	result := ProcessDATCTest(t, gameState)

	// Expected: Multiple own fleets provide convoy
	ValidateExpectedOutcome(t, result, "Multiple own fleets provide convoy for swapping.", "6.G.6")
}

// Test 6.G.7: SWAPPING WITH FOREIGN FLEET
// Original DATC test - swapping with foreign fleet convoy
func TestDATCG7_SwappingWithForeignFleet(t *testing.T) {
	gameState := CreateDATCGameState(t)

	// Add units for foreign fleet convoy
	gameState.Board.Units["rome"] = &game.Unit{
		Type:     game.Army,
		Owner:    game.Italy,
		Province: "rome",
	}
	gameState.Board.Units["apulia"] = &game.Unit{
		Type:     game.Army,
		Owner:    game.Turkey,
		Province: "apulia",
	}
	gameState.Board.Units["ionian_sea"] = &game.Unit{
		Type:     game.Fleet,
		Owner:    game.Austria,
		Province: "ionian_sea",
	}

	// Orders for foreign fleet convoy
	gameState.RawOrders[game.Italy] = []string{
		"a rome - apulia",
	}
	gameState.RawOrders[game.Turkey] = []string{
		"a apulia - rome",
	}
	gameState.RawOrders[game.Austria] = []string{
		"f ionian_sea c rome - apulia",
	}

	// Process turn
	result := ProcessDATCTest(t, gameState)

	// Expected: Foreign fleet enables convoy swap
	ValidateExpectedOutcome(t, result, "Foreign fleet enables convoy swap.", "6.G.7")
}

// Test 6.G.8: SWAPPING WITH UNINTENDED FOREIGN FLEET
// Original DATC test - unintended foreign fleet convoy
func TestDATCG8_SwappingWithUnintendedForeignFleet(t *testing.T) {
	gameState := CreateDATCGameState(t)

	// Add units for unintended foreign fleet convoy
	gameState.Board.Units["rome"] = &game.Unit{
		Type:     game.Army,
		Owner:    game.Italy,
		Province: "rome",
	}
	gameState.Board.Units["apulia"] = &game.Unit{
		Type:     game.Army,
		Owner:    game.Turkey,
		Province: "apulia",
	}
	gameState.Board.Units["ionian_sea"] = &game.Unit{
		Type:     game.Fleet,
		Owner:    game.Austria,
		Province: "ionian_sea",
	}

	// Orders for unintended foreign fleet convoy
	gameState.RawOrders[game.Italy] = []string{
		"a rome - apulia",
	}
	gameState.RawOrders[game.Turkey] = []string{
		"a apulia - rome",
	}
	gameState.RawOrders[game.Austria] = []string{
		"f ionian_sea c apulia - rome", // Convoy for Turkey's army
	}

	// Process turn
	result := ProcessDATCTest(t, gameState)

	// Expected: Unintended foreign fleet convoy
	ValidateExpectedOutcome(t, result, "Unintended foreign fleet convoy scenario.", "6.G.8")
}

// Test 6.G.9: SWAPPING WITH MULTIPLE FOREIGN FLEETS
// Original DATC test - multiple foreign fleet convoy
func TestDATCG9_SwappingWithMultipleForeignFleets(t *testing.T) {
	gameState := CreateDATCGameState(t)

	// Add units for multiple foreign fleet convoy
	gameState.Board.Units["rome"] = &game.Unit{
		Type:     game.Army,
		Owner:    game.Italy,
		Province: "rome",
	}
	gameState.Board.Units["apulia"] = &game.Unit{
		Type:     game.Army,
		Owner:    game.Turkey,
		Province: "apulia",
	}
	gameState.Board.Units["ionian_sea"] = &game.Unit{
		Type:     game.Fleet,
		Owner:    game.Austria,
		Province: "ionian_sea",
	}
	gameState.Board.Units["tyrrhenian_sea"] = &game.Unit{
		Type:     game.Fleet,
		Owner:    game.France,
		Province: "tyrrhenian_sea",
	}

	// Orders for multiple foreign fleet convoy
	gameState.RawOrders[game.Italy] = []string{
		"a rome - apulia",
	}
	gameState.RawOrders[game.Turkey] = []string{
		"a apulia - rome",
	}
	gameState.RawOrders[game.Austria] = []string{
		"f ionian_sea c rome - apulia",
	}
	gameState.RawOrders[game.France] = []string{
		"f tyrrhenian_sea c apulia - rome",
	}

	// Process turn
	result := ProcessDATCTest(t, gameState)

	// Expected: Multiple foreign fleets provide convoy
	ValidateExpectedOutcome(t, result, "Multiple foreign fleets provide convoy for swapping.", "6.G.9")
}

// Test 6.G.10: SWAPPED OR AN HEAD-TO-HEAD BATTLE?
// Original DATC test - convoy vs head-to-head battle
func TestDATCG10_SwappedOrHeadToHeadBattle(t *testing.T) {
	gameState := CreateDATCGameState(t)

	// Add units for convoy vs head-to-head scenario
	gameState.Board.Units["norway"] = &game.Unit{
		Type:     game.Army,
		Owner:    game.England,
		Province: "norway",
	}
	gameState.Board.Units["denmark"] = &game.Unit{
		Type:     game.Fleet,
		Owner:    game.England,
		Province: "denmark",
	}
	gameState.Board.Units["finland"] = &game.Unit{
		Type:     game.Fleet,
		Owner:    game.England,
		Province: "finland",
	}
	gameState.Board.Units["skagerrak"] = &game.Unit{
		Type:     game.Fleet,
		Owner:    game.Germany,
		Province: "skagerrak",
	}
	gameState.Board.Units["sweden"] = &game.Unit{
		Type:     game.Army,
		Owner:    game.Russia,
		Province: "sweden",
	}

	// Orders for convoy vs head-to-head
	gameState.RawOrders[game.England] = []string{
		"a norway - sweden",
		"f denmark s norway - sweden",
		"f finland s norway - sweden",
	}
	gameState.RawOrders[game.Germany] = []string{
		"f skagerrak c norway - sweden",
	}
	gameState.RawOrders[game.Russia] = []string{
		"a sweden - norway",
	}

	// Process turn
	result := ProcessDATCTest(t, gameState)

	// Expected: Convoy enables swap despite support
	ValidateExpectedOutcome(t, result, "Can a dislodged unit have effect on the attacker's area, when the attacker moved by convoy?", "6.G.10")
}

// Test 6.G.11: A CONVOY TO AN ADJACENT PROVINCE WITH A PARADOX
// Original DATC test - adjacent convoy with paradox
func TestDATCG11_ConvoyToAdjacentProvinceWithParadox(t *testing.T) {
	gameState := CreateDATCGameState(t)

	// Add units for adjacent convoy paradox
	gameState.Board.Units["norway"] = &game.Unit{
		Type:     game.Army,
		Owner:    game.England,
		Province: "norway",
	}
	gameState.Board.Units["skagerrak"] = &game.Unit{
		Type:     game.Fleet,
		Owner:    game.England,
		Province: "skagerrak",
	}
	gameState.Board.Units["sweden"] = &game.Unit{
		Type:     game.Fleet,
		Owner:    game.Russia,
		Province: "sweden",
	}

	// Orders for adjacent convoy paradox
	gameState.RawOrders[game.England] = []string{
		"a norway - sweden",
		"f skagerrak c norway - sweden",
	}
	gameState.RawOrders[game.Russia] = []string{
		"f sweden - skagerrak",
	}

	// Process turn
	result := ProcessDATCTest(t, gameState)

	// Expected: Paradox resolution for adjacent convoy
	ValidateExpectedOutcome(t, result, "Adjacent convoy with paradox scenario.", "6.G.11")
}

// Test 6.G.12: A CONVOY TO AN ADJACENT PROVINCE WITH A PARADOX AND SUPPORT
// Original DATC test - adjacent convoy paradox with support
func TestDATCG12_ConvoyToAdjacentProvinceWithParadoxAndSupport(t *testing.T) {
	gameState := CreateDATCGameState(t)

	// Add units for adjacent convoy paradox with support
	gameState.Board.Units["norway"] = &game.Unit{
		Type:     game.Army,
		Owner:    game.England,
		Province: "norway",
	}
	gameState.Board.Units["skagerrak"] = &game.Unit{
		Type:     game.Fleet,
		Owner:    game.England,
		Province: "skagerrak",
	}
	gameState.Board.Units["finland"] = &game.Unit{
		Type:     game.Fleet,
		Owner:    game.England,
		Province: "finland",
	}
	gameState.Board.Units["sweden"] = &game.Unit{
		Type:     game.Fleet,
		Owner:    game.Russia,
		Province: "sweden",
	}

	// Orders for adjacent convoy paradox with support
	gameState.RawOrders[game.England] = []string{
		"a norway - sweden",
		"f skagerrak c norway - sweden",
		"f finland s norway - sweden",
	}
	gameState.RawOrders[game.Russia] = []string{
		"f sweden - skagerrak",
	}

	// Process turn
	result := ProcessDATCTest(t, gameState)

	// Expected: Paradox resolution with support
	ValidateExpectedOutcome(t, result, "Adjacent convoy paradox with support scenario.", "6.G.12")
}

// Test 6.G.13: A CONVOY TO AN ADJACENT PROVINCE WITH A PARADOX AND SUPPORT CUTTING
// Original DATC test - adjacent convoy paradox with support cutting
func TestDATCG13_ConvoyToAdjacentProvinceWithParadoxAndSupportCutting(t *testing.T) {
	gameState := CreateDATCGameState(t)

	// Add units for adjacent convoy paradox with support cutting
	gameState.Board.Units["norway"] = &game.Unit{
		Type:     game.Army,
		Owner:    game.England,
		Province: "norway",
	}
	gameState.Board.Units["skagerrak"] = &game.Unit{
		Type:     game.Fleet,
		Owner:    game.England,
		Province: "skagerrak",
	}
	gameState.Board.Units["finland"] = &game.Unit{
		Type:     game.Fleet,
		Owner:    game.England,
		Province: "finland",
	}
	gameState.Board.Units["sweden"] = &game.Unit{
		Type:     game.Fleet,
		Owner:    game.Russia,
		Province: "sweden",
	}
	gameState.Board.Units["baltic_sea"] = &game.Unit{
		Type:     game.Fleet,
		Owner:    game.Russia,
		Province: "baltic_sea",
	}

	// Orders for adjacent convoy paradox with support cutting
	gameState.RawOrders[game.England] = []string{
		"a norway - sweden",
		"f skagerrak c norway - sweden",
		"f finland s norway - sweden",
	}
	gameState.RawOrders[game.Russia] = []string{
		"f sweden - skagerrak",
		"f baltic_sea - finland",
	}

	// Process turn
	result := ProcessDATCTest(t, gameState)

	// Expected: Paradox resolution with support cutting
	ValidateExpectedOutcome(t, result, "Adjacent convoy paradox with support cutting scenario.", "6.G.13")
}

// Test 6.G.14: A CONVOY TO AN ADJACENT PROVINCE WITH A PARADOX AND DISLODGED CONVOY
// Original DATC test - adjacent convoy paradox with dislodged convoy
func TestDATCG14_ConvoyToAdjacentProvinceWithParadoxAndDislodgedConvoy(t *testing.T) {
	gameState := CreateDATCGameState(t)

	// Add units for adjacent convoy paradox with dislodged convoy
	gameState.Board.Units["norway"] = &game.Unit{
		Type:     game.Army,
		Owner:    game.England,
		Province: "norway",
	}
	gameState.Board.Units["skagerrak"] = &game.Unit{
		Type:     game.Fleet,
		Owner:    game.England,
		Province: "skagerrak",
	}
	gameState.Board.Units["sweden"] = &game.Unit{
		Type:     game.Fleet,
		Owner:    game.Russia,
		Province: "sweden",
	}
	gameState.Board.Units["baltic_sea"] = &game.Unit{
		Type:     game.Fleet,
		Owner:    game.Russia,
		Province: "baltic_sea",
	}

	// Orders for adjacent convoy paradox with dislodged convoy
	gameState.RawOrders[game.England] = []string{
		"a norway - sweden",
		"f skagerrak c norway - sweden",
	}
	gameState.RawOrders[game.Russia] = []string{
		"f sweden - skagerrak",
		"f baltic_sea s sweden - skagerrak",
	}

	// Process turn
	result := ProcessDATCTest(t, gameState)

	// Expected: Paradox resolution with dislodged convoy
	ValidateExpectedOutcome(t, result, "Adjacent convoy paradox with dislodged convoy scenario.", "6.G.14")
}

// Test 6.G.15: A CONVOY TO AN ADJACENT PROVINCE WITH A PARADOX AND FAILED CONVOY
// Original DATC test - adjacent convoy paradox with failed convoy
func TestDATCG15_ConvoyToAdjacentProvinceWithParadoxAndFailedConvoy(t *testing.T) {
	gameState := CreateDATCGameState(t)

	// Add units for adjacent convoy paradox with failed convoy
	gameState.Board.Units["norway"] = &game.Unit{
		Type:     game.Army,
		Owner:    game.England,
		Province: "norway",
	}
	gameState.Board.Units["skagerrak"] = &game.Unit{
		Type:     game.Fleet,
		Owner:    game.England,
		Province: "skagerrak",
	}
	gameState.Board.Units["sweden"] = &game.Unit{
		Type:     game.Fleet,
		Owner:    game.Russia,
		Province: "sweden",
	}
	gameState.Board.Units["denmark"] = &game.Unit{
		Type:     game.Fleet,
		Owner:    game.Germany,
		Province: "denmark",
	}

	// Orders for adjacent convoy paradox with failed convoy
	gameState.RawOrders[game.England] = []string{
		"a norway - sweden",
		"f skagerrak c norway - sweden",
	}
	gameState.RawOrders[game.Russia] = []string{
		"f sweden - skagerrak",
	}
	gameState.RawOrders[game.Germany] = []string{
		"f denmark - skagerrak",
	}

	// Process turn
	result := ProcessDATCTest(t, gameState)

	// Expected: Paradox resolution with failed convoy
	ValidateExpectedOutcome(t, result, "Adjacent convoy paradox with failed convoy scenario.", "6.G.15")
}

// Test 6.G.16: A CONVOY TO AN ADJACENT PROVINCE WITH A PARADOX AND MULTIPLE ROUTES
// Original DATC test - adjacent convoy paradox with multiple routes
func TestDATCG16_ConvoyToAdjacentProvinceWithParadoxAndMultipleRoutes(t *testing.T) {
	gameState := CreateDATCGameState(t)

	// Add units for adjacent convoy paradox with multiple routes
	gameState.Board.Units["norway"] = &game.Unit{
		Type:     game.Army,
		Owner:    game.England,
		Province: "norway",
	}
	gameState.Board.Units["skagerrak"] = &game.Unit{
		Type:     game.Fleet,
		Owner:    game.England,
		Province: "skagerrak",
	}
	gameState.Board.Units["north_sea"] = &game.Unit{
		Type:     game.Fleet,
		Owner:    game.England,
		Province: "north_sea",
	}
	gameState.Board.Units["sweden"] = &game.Unit{
		Type:     game.Fleet,
		Owner:    game.Russia,
		Province: "sweden",
	}

	// Orders for adjacent convoy paradox with multiple routes
	gameState.RawOrders[game.England] = []string{
		"a norway - sweden",
		"f skagerrak c norway - sweden",
		"f north_sea c norway - sweden",
	}
	gameState.RawOrders[game.Russia] = []string{
		"f sweden - skagerrak",
	}

	// Process turn
	result := ProcessDATCTest(t, gameState)

	// Expected: Paradox resolution with multiple routes
	ValidateExpectedOutcome(t, result, "Adjacent convoy paradox with multiple routes scenario.", "6.G.16")
}

// Test 6.G.17: A CONVOY TO AN ADJACENT PROVINCE WITH A PARADOX AND DISRUPTED ROUTE
// Original DATC test - adjacent convoy paradox with disrupted route
func TestDATCG17_ConvoyToAdjacentProvinceWithParadoxAndDisruptedRoute(t *testing.T) {
	gameState := CreateDATCGameState(t)

	// Add units for adjacent convoy paradox with disrupted route
	gameState.Board.Units["norway"] = &game.Unit{
		Type:     game.Army,
		Owner:    game.England,
		Province: "norway",
	}
	gameState.Board.Units["skagerrak"] = &game.Unit{
		Type:     game.Fleet,
		Owner:    game.England,
		Province: "skagerrak",
	}
	gameState.Board.Units["north_sea"] = &game.Unit{
		Type:     game.Fleet,
		Owner:    game.England,
		Province: "north_sea",
	}
	gameState.Board.Units["sweden"] = &game.Unit{
		Type:     game.Fleet,
		Owner:    game.Russia,
		Province: "sweden",
	}
	gameState.Board.Units["denmark"] = &game.Unit{
		Type:     game.Fleet,
		Owner:    game.Germany,
		Province: "denmark",
	}

	// Orders for adjacent convoy paradox with disrupted route
	gameState.RawOrders[game.England] = []string{
		"a norway - sweden",
		"f skagerrak c norway - sweden",
		"f north_sea c norway - sweden",
	}
	gameState.RawOrders[game.Russia] = []string{
		"f sweden - skagerrak",
	}
	gameState.RawOrders[game.Germany] = []string{
		"f denmark - north_sea",
	}

	// Process turn
	result := ProcessDATCTest(t, gameState)

	// Expected: Paradox resolution with disrupted route
	ValidateExpectedOutcome(t, result, "Adjacent convoy paradox with disrupted route scenario.", "6.G.17")
}

// Test 6.G.18: A CONVOY TO AN ADJACENT PROVINCE WITH A PARADOX AND BELEAGUERED CONVOY
// Original DATC test - adjacent convoy paradox with beleaguered convoy
func TestDATCG18_ConvoyToAdjacentProvinceWithParadoxAndBeleagueredConvoy(t *testing.T) {
	gameState := CreateDATCGameState(t)

	// Add units for adjacent convoy paradox with beleaguered convoy
	gameState.Board.Units["norway"] = &game.Unit{
		Type:     game.Army,
		Owner:    game.England,
		Province: "norway",
	}
	gameState.Board.Units["skagerrak"] = &game.Unit{
		Type:     game.Fleet,
		Owner:    game.England,
		Province: "skagerrak",
	}
	gameState.Board.Units["sweden"] = &game.Unit{
		Type:     game.Fleet,
		Owner:    game.Russia,
		Province: "sweden",
	}
	gameState.Board.Units["denmark"] = &game.Unit{
		Type:     game.Fleet,
		Owner:    game.Germany,
		Province: "denmark",
	}
	gameState.Board.Units["baltic_sea"] = &game.Unit{
		Type:     game.Fleet,
		Owner:    game.France,
		Province: "baltic_sea",
	}

	// Orders for adjacent convoy paradox with beleaguered convoy
	gameState.RawOrders[game.England] = []string{
		"a norway - sweden",
		"f skagerrak c norway - sweden",
	}
	gameState.RawOrders[game.Russia] = []string{
		"f sweden - skagerrak",
	}
	gameState.RawOrders[game.Germany] = []string{
		"f denmark - skagerrak",
	}
	gameState.RawOrders[game.France] = []string{
		"f baltic_sea - skagerrak",
	}

	// Process turn
	result := ProcessDATCTest(t, gameState)

	// Expected: Paradox resolution with beleaguered convoy
	ValidateExpectedOutcome(t, result, "Adjacent convoy paradox with beleaguered convoy scenario.", "6.G.18")
}

// Test 6.G.19: A CONVOY TO AN ADJACENT PROVINCE WITH A PARADOX AND SELF-DISLODGEMENT
// Original DATC test - adjacent convoy paradox with self-dislodgement
func TestDATCG19_ConvoyToAdjacentProvinceWithParadoxAndSelfDislodgement(t *testing.T) {
	gameState := CreateDATCGameState(t)

	// Add units for adjacent convoy paradox with self-dislodgement
	gameState.Board.Units["norway"] = &game.Unit{
		Type:     game.Army,
		Owner:    game.England,
		Province: "norway",
	}
	gameState.Board.Units["skagerrak"] = &game.Unit{
		Type:     game.Fleet,
		Owner:    game.England,
		Province: "skagerrak",
	}
	gameState.Board.Units["sweden"] = &game.Unit{
		Type:     game.Fleet,
		Owner:    game.England,
		Province: "sweden",
	}

	// Orders for adjacent convoy paradox with self-dislodgement
	gameState.RawOrders[game.England] = []string{
		"a norway - sweden",
		"f skagerrak c norway - sweden",
		"f sweden - skagerrak",
	}

	// Process turn
	result := ProcessDATCTest(t, gameState)

	// Expected: Self-dislodgement prevention in paradox
	ValidateExpectedOutcome(t, result, "Adjacent convoy paradox with self-dislodgement scenario.", "6.G.19")
}

// Test 6.G.20: A CONVOY TO AN ADJACENT PROVINCE WITH A PARADOX AND CIRCULAR MOVEMENT
// Original DATC test - adjacent convoy paradox with circular movement
func TestDATCG20_ConvoyToAdjacentProvinceWithParadoxAndCircularMovement(t *testing.T) {
	gameState := CreateDATCGameState(t)

	// Add units for adjacent convoy paradox with circular movement
	gameState.Board.Units["norway"] = &game.Unit{
		Type:     game.Army,
		Owner:    game.England,
		Province: "norway",
	}
	gameState.Board.Units["skagerrak"] = &game.Unit{
		Type:     game.Fleet,
		Owner:    game.England,
		Province: "skagerrak",
	}
	gameState.Board.Units["sweden"] = &game.Unit{
		Type:     game.Fleet,
		Owner:    game.Russia,
		Province: "sweden",
	}
	gameState.Board.Units["finland"] = &game.Unit{
		Type:     game.Fleet,
		Owner:    game.Russia,
		Province: "finland",
	}

	// Orders for adjacent convoy paradox with circular movement
	gameState.RawOrders[game.England] = []string{
		"a norway - sweden",
		"f skagerrak c norway - sweden",
	}
	gameState.RawOrders[game.Russia] = []string{
		"f sweden - finland",
		"f finland - skagerrak",
	}

	// Process turn
	result := ProcessDATCTest(t, gameState)

	// Expected: Paradox resolution with circular movement
	ValidateExpectedOutcome(t, result, "Adjacent convoy paradox with circular movement scenario.", "6.G.20")
}
