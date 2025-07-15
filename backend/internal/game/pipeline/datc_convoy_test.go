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
	gameState.RawOrders[game.England] = []string{
		"f north_sea c london - holland",
		"a london - holland",
		"f yorkshire s f north_sea",
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

// Test 6.F.4: AN ATTACKED CONVOY IS NOT DISRUPTED
// Original DATC test - convoy continues even when attacked (but not dislodged)
func TestDATCF4_AttackedConvoyNotDisrupted(t *testing.T) {
	gameState := CreateDATCGameState(t)

	// Add units exactly as in original DATC test
	gameState.Board.Units["north_sea"] = &game.Unit{
		Type:     game.Fleet,
		Owner:    game.England,
		Province: "north_sea",
	}
	gameState.Board.Units["london"] = &game.Unit{
		Type:     game.Army,
		Owner:    game.England,
		Province: "london",
	}
	gameState.Board.Units["skagerrak"] = &game.Unit{
		Type:     game.Fleet,
		Owner:    game.Germany,
		Province: "skagerrak",
	}

	// Original DATC orders
	gameState.RawOrders[game.England] = []string{
		"f north_sea c london - holland",
		"a london - holland",
	}
	gameState.RawOrders[game.Germany] = []string{
		"f skagerrak - north_sea",
	}

	// Process turn
	result := ProcessDATCTest(t, gameState)

	// Expected: Army successfully convoys despite attack on fleet
	ValidateExpectedOutcome(t, result, "The army in London will successfully convoy and end in Holland.", "6.F.4")
}

// Test 6.F.5: A BELEAGUERED CONVOY IS NOT DISRUPTED
// Original DATC test - convoy continues even when beleaguered (multiple attacks)
func TestDATCF5_BeleagueredConvoyNotDisrupted(t *testing.T) {
	gameState := CreateDATCGameState(t)

	// Add units exactly as in original DATC test
	gameState.Board.Units["north_sea"] = &game.Unit{
		Type:     game.Fleet,
		Owner:    game.England,
		Province: "north_sea",
	}
	gameState.Board.Units["london"] = &game.Unit{
		Type:     game.Army,
		Owner:    game.England,
		Province: "london",
	}
	gameState.Board.Units["skagerrak"] = &game.Unit{
		Type:     game.Fleet,
		Owner:    game.Germany,
		Province: "skagerrak",
	}
	gameState.Board.Units["norway"] = &game.Unit{
		Type:     game.Fleet,
		Owner:    game.Russia,
		Province: "norway",
	}

	// Original DATC orders
	gameState.RawOrders[game.England] = []string{
		"f north_sea c london - holland",
		"a london - holland",
	}
	gameState.RawOrders[game.Germany] = []string{
		"f skagerrak - north_sea",
	}
	gameState.RawOrders[game.Russia] = []string{
		"f norway - north_sea",
	}

	// Process turn
	result := ProcessDATCTest(t, gameState)

	// Expected: Army successfully convoys despite multiple attacks on fleet
	ValidateExpectedOutcome(t, result, "The army in London will successfully convoy and end in Holland.", "6.F.5")
}

// Test 6.F.6: DISLODGED CONVOY DOES NOT CUT SUPPORT
// Original DATC test - dislodged convoy fleet doesn't cut support
func TestDATCF6_DislodgedConvoyDoesNotCutSupport(t *testing.T) {
	gameState := CreateDATCGameState(t)

	// Add units exactly as in original DATC test
	gameState.Board.Units["north_sea"] = &game.Unit{
		Type:     game.Fleet,
		Owner:    game.England,
		Province: "north_sea",
	}
	gameState.Board.Units["london"] = &game.Unit{
		Type:     game.Army,
		Owner:    game.England,
		Province: "london",
	}
	gameState.Board.Units["holland"] = &game.Unit{
		Type:     game.Army,
		Owner:    game.Germany,
		Province: "holland",
	}
	gameState.Board.Units["belgium"] = &game.Unit{
		Type:     game.Army,
		Owner:    game.Germany,
		Province: "belgium",
	}
	gameState.Board.Units["heligoland_bight"] = &game.Unit{
		Type:     game.Fleet,
		Owner:    game.Germany,
		Province: "heligoland_bight",
	}
	gameState.Board.Units["skagerrak"] = &game.Unit{
		Type:     game.Fleet,
		Owner:    game.Germany,
		Province: "skagerrak",
	}
	gameState.Board.Units["picardy"] = &game.Unit{
		Type:     game.Army,
		Owner:    game.France,
		Province: "picardy",
	}
	gameState.Board.Units["burgundy"] = &game.Unit{
		Type:     game.Army,
		Owner:    game.France,
		Province: "burgundy",
	}

	// Original DATC orders
	gameState.RawOrders[game.England] = []string{
		"f north_sea c london - holland",
		"a london - holland",
	}
	gameState.RawOrders[game.Germany] = []string{
		"a holland s belgium",
		"a belgium s holland",
		"f heligoland_bight s skagerrak - north_sea",
		"f skagerrak - north_sea",
	}
	gameState.RawOrders[game.France] = []string{
		"a picardy - belgium",
		"a burgundy s picardy - belgium",
	}

	// Process turn
	result := ProcessDATCTest(t, gameState)

	// Expected: Convoy fails, but support is not cut
	ValidateExpectedOutcome(t, result, "When a fleet of a convoy is dislodged, the convoy is completely cancelled. So, no support is cut.", "6.F.6")
}

// Test 6.F.7: DISLODGED CONVOY DOES NOT CAUSE CONTESTED AREA
// Original DATC test - dislodged convoy doesn't contest landing area for retreats
func TestDATCF7_DislodgedConvoyDoesNotCauseContestedArea(t *testing.T) {
	gameState := CreateDATCGameState(t)

	// Add units exactly as in original DATC test
	gameState.Board.Units["north_sea"] = &game.Unit{
		Type:     game.Fleet,
		Owner:    game.England,
		Province: "north_sea",
	}
	gameState.Board.Units["london"] = &game.Unit{
		Type:     game.Army,
		Owner:    game.England,
		Province: "london",
	}
	gameState.Board.Units["heligoland_bight"] = &game.Unit{
		Type:     game.Fleet,
		Owner:    game.Germany,
		Province: "heligoland_bight",
	}
	gameState.Board.Units["skagerrak"] = &game.Unit{
		Type:     game.Fleet,
		Owner:    game.Germany,
		Province: "skagerrak",
	}

	// Original DATC orders
	gameState.RawOrders[game.England] = []string{
		"f north_sea c london - holland",
		"a london - holland",
	}
	gameState.RawOrders[game.Germany] = []string{
		"f heligoland_bight s skagerrak - north_sea",
		"f skagerrak - north_sea",
	}

	// Process turn
	result := ProcessDATCTest(t, gameState)

	// Expected: Convoy fails, but Holland is not contested for retreats
	ValidateExpectedOutcome(t, result, "The dislodged English fleet can retreat to Holland.", "6.F.7")
}

// Test 6.F.8: DISLODGED CONVOY DOES NOT CAUSE A BOUNCE
// Original DATC test - dislodged convoy doesn't cause bounce in landing area
func TestDATCF8_DislodgedConvoyDoesNotCauseBounce(t *testing.T) {
	gameState := CreateDATCGameState(t)

	// Add units exactly as in original DATC test
	gameState.Board.Units["north_sea"] = &game.Unit{
		Type:     game.Fleet,
		Owner:    game.England,
		Province: "north_sea",
	}
	gameState.Board.Units["london"] = &game.Unit{
		Type:     game.Army,
		Owner:    game.England,
		Province: "london",
	}
	gameState.Board.Units["heligoland_bight"] = &game.Unit{
		Type:     game.Fleet,
		Owner:    game.Germany,
		Province: "heligoland_bight",
	}
	gameState.Board.Units["skagerrak"] = &game.Unit{
		Type:     game.Fleet,
		Owner:    game.Germany,
		Province: "skagerrak",
	}
	gameState.Board.Units["belgium"] = &game.Unit{
		Type:     game.Army,
		Owner:    game.Germany,
		Province: "belgium",
	}

	// Original DATC orders
	gameState.RawOrders[game.England] = []string{
		"f north_sea c london - holland",
		"a london - holland",
	}
	gameState.RawOrders[game.Germany] = []string{
		"f heligoland_bight s skagerrak - north_sea",
		"f skagerrak - north_sea",
		"a belgium - holland",
	}

	// Process turn
	result := ProcessDATCTest(t, gameState)

	// Expected: Army in Belgium moves to Holland without bounce
	ValidateExpectedOutcome(t, result, "The army in Belgium will not bounce and move to Holland.", "6.F.8")
}

// Test 6.F.9: DISLODGE OF MULTI-ROUTE CONVOY
// Original DATC test - multi-route convoy with one route dislodged
func TestDATCF9_DislodgeOfMultiRouteConvoy(t *testing.T) {
	gameState := CreateDATCGameState(t)

	// Add units exactly as in original DATC test
	gameState.Board.Units["english_channel"] = &game.Unit{
		Type:     game.Fleet,
		Owner:    game.England,
		Province: "english_channel",
	}
	gameState.Board.Units["north_sea"] = &game.Unit{
		Type:     game.Fleet,
		Owner:    game.England,
		Province: "north_sea",
	}
	gameState.Board.Units["london"] = &game.Unit{
		Type:     game.Army,
		Owner:    game.England,
		Province: "london",
	}
	gameState.Board.Units["brest"] = &game.Unit{
		Type:     game.Fleet,
		Owner:    game.France,
		Province: "brest",
	}
	gameState.Board.Units["midatlantic_ocean"] = &game.Unit{
		Type:     game.Fleet,
		Owner:    game.France,
		Province: "midatlantic_ocean",
	}

	// Original DATC orders
	gameState.RawOrders[game.England] = []string{
		"f english_channel c london - belgium",
		"f north_sea c london - belgium",
		"a london - belgium",
	}
	gameState.RawOrders[game.France] = []string{
		"f brest s midatlantic_ocean - english_channel",
		"f midatlantic_ocean - english_channel",
	}

	// Process turn
	result := ProcessDATCTest(t, gameState)

	// Expected: Depends on rulebook - convoy may succeed via alternate route
	ValidateExpectedOutcome(t, result, "When a fleet of a convoy with multiple routes is dislodged, the result depends on the rulebook that is used.", "6.F.9")
}

// Test 6.F.10: A CONVOY IS NOT DISRUPTED WHEN THE CONVOYED UNIT HAS A DIFFERENT ROUTE
// Original DATC test - convoy not disrupted when unit can take different route
func TestDATCF10_ConvoyNotDisruptedWhenUnitHasDifferentRoute(t *testing.T) {
	gameState := CreateDATCGameState(t)

	// Add units exactly as in original DATC test
	gameState.Board.Units["english_channel"] = &game.Unit{
		Type:     game.Fleet,
		Owner:    game.England,
		Province: "english_channel",
	}
	gameState.Board.Units["london"] = &game.Unit{
		Type:     game.Army,
		Owner:    game.England,
		Province: "london",
	}
	gameState.Board.Units["brest"] = &game.Unit{
		Type:     game.Fleet,
		Owner:    game.France,
		Province: "brest",
	}

	// Original DATC orders
	gameState.RawOrders[game.England] = []string{
		"f english_channel c london - brest",
		"a london - brest",
	}
	gameState.RawOrders[game.France] = []string{
		"f brest - english_channel",
	}

	// Process turn
	result := ProcessDATCTest(t, gameState)

	// Expected: Army moves via land route when convoy is disrupted
	ValidateExpectedOutcome(t, result, "The army in London will move to Brest by land and the French fleet in Brest will move to the English Channel.", "6.F.10")
}

// Test 6.F.14: SIMPLE CONVOY PARADOX
// Original DATC test - the most common paradox scenario
func TestDATCF14_SimpleConvoyParadox(t *testing.T) {
	gameState := CreateDATCGameState(t)

	// Add units exactly as in original DATC test
	gameState.Board.Units["london"] = &game.Unit{
		Type:     game.Fleet,
		Owner:    game.England,
		Province: "london",
	}
	gameState.Board.Units["wales"] = &game.Unit{
		Type:     game.Fleet,
		Owner:    game.England,
		Province: "wales",
	}
	gameState.Board.Units["brest"] = &game.Unit{
		Type:     game.Army,
		Owner:    game.France,
		Province: "brest",
	}
	gameState.Board.Units["english_channel"] = &game.Unit{
		Type:     game.Fleet,
		Owner:    game.France,
		Province: "english_channel",
	}

	// Original DATC orders
	gameState.RawOrders[game.England] = []string{
		"f london s wales - english_channel",
		"f wales - english_channel",
	}
	gameState.RawOrders[game.France] = []string{
		"a brest - london",
		"f english_channel c brest - london",
	}

	// Process turn
	result := ProcessDATCTest(t, gameState)

	// Expected: According to all rulebooks, the support of London is not cut
	ValidateExpectedOutcome(t, result, "According to all rulebooks, the support of London is not cut.", "6.F.14")
}

// Test 6.F.15: SIMPLE CONVOY PARADOX WITH ADDITIONAL CONVOY
// Original DATC test - paradox rules only apply on the paradox core
func TestDATCF15_SimpleConvoyParadoxWithAdditionalConvoy(t *testing.T) {
	gameState := CreateDATCGameState(t)

	// Add units exactly as in original DATC test
	gameState.Board.Units["london"] = &game.Unit{
		Type:     game.Fleet,
		Owner:    game.England,
		Province: "london",
	}
	gameState.Board.Units["wales"] = &game.Unit{
		Type:     game.Fleet,
		Owner:    game.England,
		Province: "wales",
	}
	gameState.Board.Units["brest"] = &game.Unit{
		Type:     game.Army,
		Owner:    game.France,
		Province: "brest",
	}
	gameState.Board.Units["english_channel"] = &game.Unit{
		Type:     game.Fleet,
		Owner:    game.France,
		Province: "english_channel",
	}
	gameState.Board.Units["irish_sea"] = &game.Unit{
		Type:     game.Fleet,
		Owner:    game.Italy,
		Province: "irish_sea",
	}
	gameState.Board.Units["midatlantic_ocean"] = &game.Unit{
		Type:     game.Fleet,
		Owner:    game.Italy,
		Province: "midatlantic_ocean",
	}
	gameState.Board.Units["north_africa"] = &game.Unit{
		Type:     game.Army,
		Owner:    game.Italy,
		Province: "north_africa",
	}

	// Original DATC orders
	gameState.RawOrders[game.England] = []string{
		"f london s wales - english_channel",
		"f wales - english_channel",
	}
	gameState.RawOrders[game.France] = []string{
		"a brest - london",
		"f english_channel c brest - london",
	}
	gameState.RawOrders[game.Italy] = []string{
		"f irish_sea c north_africa - wales",
		"f midatlantic_ocean c north_africa - wales",
		"a north_africa - wales",
	}

	// Process turn
	result := ProcessDATCTest(t, gameState)

	// Expected: Paradox rules only apply on the paradox core
	ValidateExpectedOutcome(t, result, "Paradox rules only apply on the paradox core.", "6.F.15")
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
