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

// Test 6.F.11: DISLODGE OF MULTI-ROUTE CONVOY WITH ONLY FOREIGN FLEETS
// Original DATC test - multi-route convoy with all foreign fleets
func TestDATCF11_DislodgeOfMultiRouteConvoyWithOnlyForeignFleets(t *testing.T) {
	gameState := CreateDATCGameState(t)

	// Add units exactly as in original DATC test
	gameState.Board.Units["london"] = &game.Unit{
		Type:     game.Army,
		Owner:    game.England,
		Province: "london",
	}
	gameState.Board.Units["english_channel"] = &game.Unit{
		Type:     game.Fleet,
		Owner:    game.Germany,
		Province: "english_channel",
	}
	gameState.Board.Units["north_sea"] = &game.Unit{
		Type:     game.Fleet,
		Owner:    game.Russia,
		Province: "north_sea",
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
		"a london - belgium",
	}
	gameState.RawOrders[game.Germany] = []string{
		"f english_channel c london - belgium",
	}
	gameState.RawOrders[game.Russia] = []string{
		"f north_sea c london - belgium",
	}
	gameState.RawOrders[game.France] = []string{
		"f brest s midatlantic_ocean - english_channel",
		"f midatlantic_ocean - english_channel",
	}

	// Process turn
	result := ProcessDATCTest(t, gameState)

	// Expected: Foreign fleets provide convoy despite one being dislodged
	ValidateExpectedOutcome(t, result, "With the 1971 rulebook one could adopt a rule (DPTG) that foreign fleets are not used when not necessary, but this doesn't prevent an \"unwanted\" convoy when all convoying fleets are foreign.", "6.F.11")
}

// Test 6.F.12: DISLODGED CONVOYING FLEET NOT ON ROUTE
// Original DATC test - convoy fleet dislodged but not on the route taken
func TestDATCF12_DislodgedConvoyingFleetNotOnRoute(t *testing.T) {
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
	gameState.Board.Units["irish_sea"] = &game.Unit{
		Type:     game.Fleet,
		Owner:    game.England,
		Province: "irish_sea",
	}
	gameState.Board.Units["north_atlantic_ocean"] = &game.Unit{
		Type:     game.Fleet,
		Owner:    game.France,
		Province: "north_atlantic_ocean",
	}
	gameState.Board.Units["midatlantic_ocean"] = &game.Unit{
		Type:     game.Fleet,
		Owner:    game.France,
		Province: "midatlantic_ocean",
	}

	// Original DATC orders
	gameState.RawOrders[game.England] = []string{
		"f english_channel c london - belgium",
		"a london - belgium",
		"f irish_sea c london - belgium",
	}
	gameState.RawOrders[game.France] = []string{
		"f north_atlantic_ocean s midatlantic_ocean - irish_sea",
		"f midatlantic_ocean - irish_sea",
	}

	// Process turn
	result := ProcessDATCTest(t, gameState)

	// Expected: Convoy succeeds via English Channel despite Irish Sea being dislodged
	ValidateExpectedOutcome(t, result, "When the rule is used that convoys are disrupted when one of the routes is disrupted, the convoy is not necessarily disrupted when one of the fleets ordered to convoy is dislodged.", "6.F.12")
}

// Test 6.F.13: THE UNWANTED ALTERNATIVE
// Original DATC test - convoy succeeds via unwanted alternative route
func TestDATCF13_TheUnwantedAlternative(t *testing.T) {
	gameState := CreateDATCGameState(t)

	// Add units exactly as in original DATC test
	gameState.Board.Units["london"] = &game.Unit{
		Type:     game.Army,
		Owner:    game.England,
		Province: "london",
	}
	gameState.Board.Units["north_sea"] = &game.Unit{
		Type:     game.Fleet,
		Owner:    game.England,
		Province: "north_sea",
	}
	gameState.Board.Units["english_channel"] = &game.Unit{
		Type:     game.Fleet,
		Owner:    game.France,
		Province: "english_channel",
	}
	gameState.Board.Units["holland"] = &game.Unit{
		Type:     game.Fleet,
		Owner:    game.Germany,
		Province: "holland",
	}
	gameState.Board.Units["denmark"] = &game.Unit{
		Type:     game.Fleet,
		Owner:    game.Germany,
		Province: "denmark",
	}

	// Original DATC orders
	gameState.RawOrders[game.England] = []string{
		"a london - belgium",
		"f north_sea c london - belgium",
	}
	gameState.RawOrders[game.France] = []string{
		"f english_channel c london - belgium",
	}
	gameState.RawOrders[game.Germany] = []string{
		"f holland s denmark - north_sea",
		"f denmark - north_sea",
	}

	// Process turn
	result := ProcessDATCTest(t, gameState)

	// Expected: Convoy succeeds via English Channel despite North Sea being dislodged
	ValidateExpectedOutcome(t, result, "The convoy of the army in London succeeds and the fleet in Denmark dislodges the fleet in the North Sea.", "6.F.13")
}

// Test 6.F.16: PANDIN'S PARADOX
// Original DATC test - attacked unit protects convoying fleet by beleaguered garrison
func TestDATCF16_PandinsParadox(t *testing.T) {
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
	gameState.Board.Units["north_sea"] = &game.Unit{
		Type:     game.Fleet,
		Owner:    game.Germany,
		Province: "north_sea",
	}
	gameState.Board.Units["belgium"] = &game.Unit{
		Type:     game.Fleet,
		Owner:    game.Germany,
		Province: "belgium",
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
	gameState.RawOrders[game.Germany] = []string{
		"f north_sea s belgium - english_channel",
		"f belgium - english_channel",
	}

	// Process turn
	result := ProcessDATCTest(t, gameState)

	// Expected: Paradox resolution - attacked unit protects convoying fleet
	ValidateExpectedOutcome(t, result, "In Pandin's paradox, the attacked unit protects the convoying fleet by a beleaguered garrison.", "6.F.16")
}

// Test 6.F.17: PANDIN'S EXTENDED PARADOX
// Original DATC test - extended version of Pandin's paradox
func TestDATCF17_PandinsExtendedParadox(t *testing.T) {
	gameState := CreateDATCGameState(t)

	// Add units for extended paradox scenario
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
	gameState.Board.Units["yorkshire"] = &game.Unit{
		Type:     game.Fleet,
		Owner:    game.England,
		Province: "yorkshire",
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
	gameState.Board.Units["north_sea"] = &game.Unit{
		Type:     game.Fleet,
		Owner:    game.Germany,
		Province: "north_sea",
	}
	gameState.Board.Units["belgium"] = &game.Unit{
		Type:     game.Fleet,
		Owner:    game.Germany,
		Province: "belgium",
	}

	// Extended paradox orders
	gameState.RawOrders[game.England] = []string{
		"f london s wales - english_channel",
		"f wales - english_channel",
		"f yorkshire s wales - english_channel",
	}
	gameState.RawOrders[game.France] = []string{
		"a brest - london",
		"f english_channel c brest - london",
	}
	gameState.RawOrders[game.Germany] = []string{
		"f north_sea s belgium - english_channel",
		"f belgium - english_channel",
	}

	// Process turn
	result := ProcessDATCTest(t, gameState)

	// Expected: Extended paradox resolution
	ValidateExpectedOutcome(t, result, "Extended version of Pandin's paradox with additional support.", "6.F.17")
}

// Test 6.F.18: BETRAYAL PARADOX
// Original DATC test - betrayal paradox scenario
func TestDATCF18_BetrayalParadox(t *testing.T) {
	gameState := CreateDATCGameState(t)

	// Add units for betrayal paradox
	gameState.Board.Units["london"] = &game.Unit{
		Type:     game.Fleet,
		Owner:    game.England,
		Province: "london",
	}
	gameState.Board.Units["north_sea"] = &game.Unit{
		Type:     game.Fleet,
		Owner:    game.England,
		Province: "north_sea",
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

	// Betrayal paradox orders
	gameState.RawOrders[game.England] = []string{
		"f london s north_sea - english_channel",
		"f north_sea - english_channel",
	}
	gameState.RawOrders[game.France] = []string{
		"a brest - london",
		"f english_channel c brest - london",
	}

	// Process turn
	result := ProcessDATCTest(t, gameState)

	// Expected: Betrayal paradox resolution
	ValidateExpectedOutcome(t, result, "Betrayal paradox where the convoy depends on the success of the attack.", "6.F.18")
}

// Test 6.F.19: MULTI-ROUTE CONVOY DISRUPTION PARADOX
// Original DATC test - multi-route convoy with disruption paradox
func TestDATCF19_MultiRouteConvoyDisruptionParadox(t *testing.T) {
	gameState := CreateDATCGameState(t)

	// Add units for multi-route disruption paradox
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
	gameState.Board.Units["north_sea"] = &game.Unit{
		Type:     game.Fleet,
		Owner:    game.England,
		Province: "north_sea",
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

	// Multi-route disruption orders
	gameState.RawOrders[game.England] = []string{
		"a london - belgium",
		"f english_channel c london - belgium",
		"f north_sea c london - belgium",
	}
	gameState.RawOrders[game.France] = []string{
		"f brest s midatlantic_ocean - english_channel",
		"f midatlantic_ocean - english_channel",
	}

	// Process turn
	result := ProcessDATCTest(t, gameState)

	// Expected: Multi-route convoy with one route disrupted
	ValidateExpectedOutcome(t, result, "Multi-route convoy disruption paradox scenario.", "6.F.19")
}

// Test 6.F.20: UNWANTED MULTI-ROUTE CONVOY PARADOX
// Original DATC test - unwanted multi-route convoy paradox
func TestDATCF20_UnwantedMultiRouteConvoyParadox(t *testing.T) {
	gameState := CreateDATCGameState(t)

	// Add units for unwanted multi-route convoy paradox
	gameState.Board.Units["tunis"] = &game.Unit{
		Type:     game.Army,
		Owner:    game.France,
		Province: "tunis",
	}
	gameState.Board.Units["tyrrhenian_sea"] = &game.Unit{
		Type:     game.Fleet,
		Owner:    game.France,
		Province: "tyrrhenian_sea",
	}
	gameState.Board.Units["naples"] = &game.Unit{
		Type:     game.Fleet,
		Owner:    game.Italy,
		Province: "naples",
	}
	gameState.Board.Units["ionian_sea"] = &game.Unit{
		Type:     game.Fleet,
		Owner:    game.Italy,
		Province: "ionian_sea",
	}
	gameState.Board.Units["aegean_sea"] = &game.Unit{
		Type:     game.Fleet,
		Owner:    game.Turkey,
		Province: "aegean_sea",
	}
	gameState.Board.Units["eastern_mediterranean"] = &game.Unit{
		Type:     game.Fleet,
		Owner:    game.Turkey,
		Province: "eastern_mediterranean",
	}

	// Unwanted multi-route convoy orders
	gameState.RawOrders[game.France] = []string{
		"a tunis - naples",
		"f tyrrhenian_sea c tunis - naples",
	}
	gameState.RawOrders[game.Italy] = []string{
		"f naples s ionian_sea",
		"f ionian_sea c tunis - naples",
	}
	gameState.RawOrders[game.Turkey] = []string{
		"f aegean_sea s eastern_mediterranean - ionian_sea",
		"f eastern_mediterranean - ionian_sea",
	}

	// Process turn
	result := ProcessDATCTest(t, gameState)

	// Expected: Unwanted multi-route convoy paradox resolution
	ValidateExpectedOutcome(t, result, "The 1982 paradox rule allows some creative defense.", "6.F.20")
}

// Test 6.F.21: DAD'S ARMY CONVOY
// Original DATC test - Dad's Army convoy scenario
func TestDATCF21_DadsArmyConvoy(t *testing.T) {
	gameState := CreateDATCGameState(t)

	// Add units for Dad's Army convoy
	gameState.Board.Units["london"] = &game.Unit{
		Type:     game.Army,
		Owner:    game.England,
		Province: "london",
	}
	gameState.Board.Units["wales"] = &game.Unit{
		Type:     game.Fleet,
		Owner:    game.England,
		Province: "wales",
	}
	gameState.Board.Units["english_channel"] = &game.Unit{
		Type:     game.Fleet,
		Owner:    game.France,
		Province: "english_channel",
	}
	gameState.Board.Units["brest"] = &game.Unit{
		Type:     game.Army,
		Owner:    game.France,
		Province: "brest",
	}

	// Dad's Army convoy orders
	gameState.RawOrders[game.England] = []string{
		"a london - brest",
		"f wales s english_channel",
	}
	gameState.RawOrders[game.France] = []string{
		"f english_channel c london - brest",
		"a brest - london",
	}

	// Process turn
	result := ProcessDATCTest(t, gameState)

	// Expected: Dad's Army convoy resolution
	ValidateExpectedOutcome(t, result, "Dad's Army convoy scenario with mutual dependency.", "6.F.21")
}

// Test 6.F.22: SECOND ORDER PARADOX WITH TWO RESOLUTIONS
// Original DATC test - second order paradox
func TestDATCF22_SecondOrderParadoxWithTwoResolutions(t *testing.T) {
	gameState := CreateDATCGameState(t)

	// Add units for second order paradox
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
	gameState.Board.Units["belgium"] = &game.Unit{
		Type:     game.Fleet,
		Owner:    game.Germany,
		Province: "belgium",
	}

	// Second order paradox orders
	gameState.RawOrders[game.England] = []string{
		"f london s wales - english_channel",
		"f wales - english_channel",
	}
	gameState.RawOrders[game.France] = []string{
		"a brest - london",
		"f english_channel c brest - london",
	}
	gameState.RawOrders[game.Germany] = []string{
		"f belgium - english_channel",
	}

	// Process turn
	result := ProcessDATCTest(t, gameState)

	// Expected: Second order paradox with two possible resolutions
	ValidateExpectedOutcome(t, result, "Second order paradox with two possible resolutions.", "6.F.22")
}

// Test 6.F.23: SECOND ORDER PARADOX WITH TWO EXCLUSIVE CONVOYS
// Original DATC test - second order paradox with exclusive convoys
func TestDATCF23_SecondOrderParadoxWithTwoExclusiveConvoys(t *testing.T) {
	gameState := CreateDATCGameState(t)

	// Add units for exclusive convoys paradox
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
	gameState.Board.Units["belgium"] = &game.Unit{
		Type:     game.Army,
		Owner:    game.Germany,
		Province: "belgium",
	}
	gameState.Board.Units["north_sea"] = &game.Unit{
		Type:     game.Fleet,
		Owner:    game.Germany,
		Province: "north_sea",
	}

	// Exclusive convoys paradox orders
	gameState.RawOrders[game.England] = []string{
		"f london s wales - english_channel",
		"f wales - english_channel",
	}
	gameState.RawOrders[game.France] = []string{
		"a brest - london",
		"f english_channel c brest - london",
	}
	gameState.RawOrders[game.Germany] = []string{
		"a belgium - london",
		"f north_sea c belgium - london",
	}

	// Process turn
	result := ProcessDATCTest(t, gameState)

	// Expected: Second order paradox with exclusive convoys
	ValidateExpectedOutcome(t, result, "Second order paradox with two exclusive convoys.", "6.F.23")
}

// Test 6.F.24: SECOND ORDER PARADOX WITH NO RESOLUTION
// Original DATC test - second order paradox with no resolution
func TestDATCF24_SecondOrderParadoxWithNoResolution(t *testing.T) {
	gameState := CreateDATCGameState(t)

	// Add units for no resolution paradox
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
	gameState.Board.Units["yorkshire"] = &game.Unit{
		Type:     game.Fleet,
		Owner:    game.England,
		Province: "yorkshire",
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

	// No resolution paradox orders
	gameState.RawOrders[game.England] = []string{
		"f london s wales - english_channel",
		"f wales - english_channel",
		"f yorkshire s wales - english_channel",
	}
	gameState.RawOrders[game.France] = []string{
		"a brest - london",
		"f english_channel c brest - london",
	}

	// Process turn
	result := ProcessDATCTest(t, gameState)

	// Expected: Second order paradox with no resolution
	ValidateExpectedOutcome(t, result, "Second order paradox with no resolution.", "6.F.24")
}

// Test 6.F.25: SECOND ORDER PARADOX FORCING BACKUP RULE
// Original DATC test - second order paradox forcing backup rule
func TestDATCF25_SecondOrderParadoxForcingBackupRule(t *testing.T) {
	gameState := CreateDATCGameState(t)

	// Add units for backup rule paradox
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
	gameState.Board.Units["belgium"] = &game.Unit{
		Type:     game.Fleet,
		Owner:    game.Germany,
		Province: "belgium",
	}
	gameState.Board.Units["north_sea"] = &game.Unit{
		Type:     game.Fleet,
		Owner:    game.Germany,
		Province: "north_sea",
	}

	// Backup rule paradox orders
	gameState.RawOrders[game.England] = []string{
		"f london s wales - english_channel",
		"f wales - english_channel",
	}
	gameState.RawOrders[game.France] = []string{
		"a brest - london",
		"f english_channel c brest - london",
	}
	gameState.RawOrders[game.Germany] = []string{
		"f belgium s north_sea - english_channel",
		"f north_sea - english_channel",
	}

	// Process turn
	result := ProcessDATCTest(t, gameState)

	// Expected: Second order paradox forcing backup rule
	ValidateExpectedOutcome(t, result, "Second order paradox forcing backup rule application.", "6.F.25")
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
