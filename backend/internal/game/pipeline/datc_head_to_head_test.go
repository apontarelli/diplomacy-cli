package pipeline

import (
	"testing"

	"diplomacy-cli/backend/internal/game"
)

// Test 6.E.1: DISLOGED UNIT HAS NO EFFECT ON ATTACKERS AREA
func TestDATCE1_DislogedUnitHasNoEffectOnAttackersArea(t *testing.T) {
	gameState := CreateDATCGameState(t)

	// Add units based on test case
	gameState.Board.Units["berlin"] = &game.Unit{
		Type:     game.Army,
		Owner:    game.Germany,
		Province: "berlin",
	}
	gameState.Board.Units["silesia"] = &game.Unit{
		Type:     game.Army,
		Owner:    game.Germany,
		Province: "silesia",
	}
	gameState.Board.Units["prussia"] = &game.Unit{
		Type:     game.Army,
		Owner:    game.Russia,
		Province: "prussia",
	}

	// Add orders
	gameState.RawOrders[game.Germany] = []string{
		"a berlin - prussia",
		"a silesia s berlin - prussia",
	}
	gameState.RawOrders[game.Russia] = []string{
		"a prussia - berlin",
	}

	// Process turn
	result := ProcessDATCTest(t, gameState)

	// Expected: German army succeeds with support, Russian army is dislodged
	ValidateExpectedOutcome(t, result, "German army succeeds with support, Russian army is dislodged.", "6.E.1")
}

// Test 6.E.2: NO SELF DISLODGEMENT
func TestDATCE2_NoSelfDislodgement(t *testing.T) {
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
		Owner:    game.Germany,
		Province: "silesia",
	}

	// Add orders - Germany tries to dislodge its own unit
	gameState.RawOrders[game.Germany] = []string{
		"a berlin - munich",
		"a munich - berlin",
		"a silesia s berlin - munich",
	}

	// Process turn
	result := ProcessDATCTest(t, gameState)

	// Expected: No units move (self-dislodgement not allowed)
	ValidateExpectedOutcome(t, result, "No units move (self-dislodgement not allowed).", "6.E.2")
}

// Test 6.E.3: NO HELP IN DISLODGING OWN UNIT
func TestDATCE3_NoHelpInDislodgingOwnUnit(t *testing.T) {
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
	ValidateExpectedOutcome(t, result, "Austrian attack fails (German unit cannot help dislodge its own unit).", "6.E.3")
}

// Test 6.E.4: NO SELF DISLODGEMENT IN HEAD-TO-HEAD BATTLE (E.2)
func TestDATCE4_NoSelfDislodgementInHeadToHeadBattle(t *testing.T) {
	gameState := CreateDATCGameState(t)

	// Add units based on test case
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

	// Add orders
	gameState.RawOrders[game.Germany] = []string{
		"a berlin - kiel",
		"f kiel - berlin",
		"a munich s berlin - kiel",
	}

	// Process turn
	result := ProcessDATCTest(t, gameState)

	// Expected: No unit will move
	ValidateExpectedOutcome(t, result, "No unit will move.", "6.E.4")
}

// Test 6.E.5: NO HELP IN DISLODGING OWN UNIT IN HEAD-TO-HEAD (E.3)
func TestDATCE5_NoHelpInDislodgingOwnUnitInHeadToHead(t *testing.T) {
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
		Owner:    game.England,
		Province: "kiel",
	}

	// Add orders
	gameState.RawOrders[game.Germany] = []string{
		"a berlin - kiel",
		"a munich s kiel - berlin",
	}
	gameState.RawOrders[game.England] = []string{
		"f kiel - berlin",
	}

	// Process turn
	result := ProcessDATCTest(t, gameState)

	// Expected: No unit will move
	ValidateExpectedOutcome(t, result, "No unit will move.", "6.E.5")
}

// Test 6.E.6: NON-DISLODGED LOSER STILL HAS EFFECT (E.4)
func TestDATCE6_NonDislodgedLoserStillHasEffect(t *testing.T) {
	gameState := CreateDATCGameState(t)

	// Add units based on test case
	gameState.Board.Units["holland"] = &game.Unit{
		Type:     game.Fleet,
		Owner:    game.Germany,
		Province: "holland",
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
	gameState.Board.Units["north_sea"] = &game.Unit{
		Type:     game.Fleet,
		Owner:    game.France,
		Province: "north_sea",
	}
	gameState.Board.Units["belgium"] = &game.Unit{
		Type:     game.Fleet,
		Owner:    game.France,
		Province: "belgium",
	}
	gameState.Board.Units["edinburgh"] = &game.Unit{
		Type:     game.Fleet,
		Owner:    game.England,
		Province: "edinburgh",
	}
	gameState.Board.Units["york"] = &game.Unit{
		Type:     game.Fleet,
		Owner:    game.England,
		Province: "york",
	}
	gameState.Board.Units["norwegian_sea"] = &game.Unit{
		Type:     game.Fleet,
		Owner:    game.England,
		Province: "norwegian_sea",
	}

	// Add orders
	gameState.RawOrders[game.Germany] = []string{
		"f holland - north_sea",
		"f heligoland_bight s holland - north_sea",
		"f skagerrak s holland - north_sea",
	}
	gameState.RawOrders[game.France] = []string{
		"f north_sea - holland",
		"f belgium s north_sea - holland",
	}
	gameState.RawOrders[game.England] = []string{
		"f edinburgh s norwegian_sea - north_sea",
		"f york s norwegian_sea - north_sea",
		"f norwegian_sea - north_sea",
	}

	// Process turn
	result := ProcessDATCTest(t, gameState)

	// Expected: German fleet succeeds, French fleet bounces but prevents English fleet
	ValidateExpectedOutcome(t, result, "The German fleet in Holland will dislodge the French fleet in North Sea. The French fleet will not advance to Holland. The English fleet in Norwegian Sea will bounce on the French fleet in North Sea and will not advance to North Sea.", "6.E.6")
}

// Test 6.E.7: LOSER DISLODGED BY ANOTHER ARMY STILL HAS EFFECT (E.5)
func TestDATCE7_LoserDislodgedByAnotherArmyStillHasEffect(t *testing.T) {
	gameState := CreateDATCGameState(t)

	// Add units based on test case
	gameState.Board.Units["holland"] = &game.Unit{
		Type:     game.Fleet,
		Owner:    game.Germany,
		Province: "holland",
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
	gameState.Board.Units["north_sea"] = &game.Unit{
		Type:     game.Fleet,
		Owner:    game.France,
		Province: "north_sea",
	}
	gameState.Board.Units["belgium"] = &game.Unit{
		Type:     game.Fleet,
		Owner:    game.France,
		Province: "belgium",
	}
	gameState.Board.Units["edinburgh"] = &game.Unit{
		Type:     game.Fleet,
		Owner:    game.England,
		Province: "edinburgh",
	}
	gameState.Board.Units["york"] = &game.Unit{
		Type:     game.Fleet,
		Owner:    game.England,
		Province: "york",
	}
	gameState.Board.Units["norwegian_sea"] = &game.Unit{
		Type:     game.Fleet,
		Owner:    game.England,
		Province: "norwegian_sea",
	}
	gameState.Board.Units["london"] = &game.Unit{
		Type:     game.Fleet,
		Owner:    game.England,
		Province: "london",
	}

	// Add orders
	gameState.RawOrders[game.Germany] = []string{
		"f holland - north_sea",
		"f heligoland_bight s holland - north_sea",
		"f skagerrak s holland - north_sea",
	}
	gameState.RawOrders[game.France] = []string{
		"f north_sea - holland",
		"f belgium s north_sea - holland",
	}
	gameState.RawOrders[game.England] = []string{
		"f edinburgh s norwegian_sea - north_sea",
		"f york s norwegian_sea - north_sea",
		"f norwegian_sea - north_sea",
		"f london s norwegian_sea - north_sea",
	}

	// Process turn
	result := ProcessDATCTest(t, gameState)

	// Expected: German fleet succeeds, French fleet is dislodged, English fleet bounces
	ValidateExpectedOutcome(t, result, "The German fleet in Holland will dislodge the French fleet in North Sea. The French fleet will not advance to Holland. The English fleet in Norwegian Sea will bounce on the French fleet in North Sea and will not advance to North Sea.", "6.E.7")
}

// Test 6.E.8: NOT DISLODGE BECAUSE OF OWN SUPPORT STILL HAS EFFECT (E.6)
func TestDATCE8_NotDislodgeBecauseOfOwnSupportStillHasEffect(t *testing.T) {
	gameState := CreateDATCGameState(t)

	// Add units based on test case
	gameState.Board.Units["holland"] = &game.Unit{
		Type:     game.Fleet,
		Owner:    game.Germany,
		Province: "holland",
	}
	gameState.Board.Units["heligoland_bight"] = &game.Unit{
		Type:     game.Fleet,
		Owner:    game.Germany,
		Province: "heligoland_bight",
	}
	gameState.Board.Units["north_sea"] = &game.Unit{
		Type:     game.Fleet,
		Owner:    game.France,
		Province: "north_sea",
	}
	gameState.Board.Units["belgium"] = &game.Unit{
		Type:     game.Fleet,
		Owner:    game.France,
		Province: "belgium",
	}
	gameState.Board.Units["english_channel"] = &game.Unit{
		Type:     game.Fleet,
		Owner:    game.France,
		Province: "english_channel",
	}
	gameState.Board.Units["vienna"] = &game.Unit{
		Type:     game.Army,
		Owner:    game.Austria,
		Province: "vienna",
	}

	// Add orders
	gameState.RawOrders[game.Germany] = []string{
		"f holland - north_sea",
		"f heligoland_bight s holland - north_sea",
	}
	gameState.RawOrders[game.France] = []string{
		"f north_sea - holland",
		"f belgium s north_sea - holland",
		"f english_channel s holland - north_sea",
	}
	gameState.RawOrders[game.Austria] = []string{
		"a vienna - tyrolia",
	}

	// Process turn
	result := ProcessDATCTest(t, gameState)

	// Expected: No unit moves due to self-support prevention
	ValidateExpectedOutcome(t, result, "The French fleet in North Sea is not dislodged because the support from English Channel is not allowed (supporting attack on own unit). The German fleet in Holland will not advance to North Sea. The Austrian army in Vienna will move to Tyrolia.", "6.E.8")
}

// Test 6.E.9: NO SELF DISLODGEMENT WITH BELEAGUERED GARRISON (E.7)
func TestDATCE9_NoSelfDislodgementWithBeleagueredGarrison(t *testing.T) {
	gameState := CreateDATCGameState(t)

	// Add units based on test case
	gameState.Board.Units["north_sea"] = &game.Unit{
		Type:     game.Fleet,
		Owner:    game.England,
		Province: "north_sea",
	}
	gameState.Board.Units["york"] = &game.Unit{
		Type:     game.Fleet,
		Owner:    game.England,
		Province: "york",
	}
	gameState.Board.Units["holland"] = &game.Unit{
		Type:     game.Fleet,
		Owner:    game.Germany,
		Province: "holland",
	}
	gameState.Board.Units["heligoland_bight"] = &game.Unit{
		Type:     game.Fleet,
		Owner:    game.Germany,
		Province: "heligoland_bight",
	}
	gameState.Board.Units["skagerrak"] = &game.Unit{
		Type:     game.Fleet,
		Owner:    game.Russia,
		Province: "skagerrak",
	}
	gameState.Board.Units["norway"] = &game.Unit{
		Type:     game.Fleet,
		Owner:    game.Russia,
		Province: "norway",
	}

	// Add orders
	gameState.RawOrders[game.England] = []string{
		"f north_sea hold",
		"f york s norway - north_sea",
	}
	gameState.RawOrders[game.Germany] = []string{
		"f holland s heligoland_bight - north_sea",
		"f heligoland_bight - north_sea",
	}
	gameState.RawOrders[game.Russia] = []string{
		"f skagerrak s norway - north_sea",
		"f norway - north_sea",
	}

	// Process turn
	result := ProcessDATCTest(t, gameState)

	// Expected: Russian fleet succeeds, English fleet holds
	ValidateExpectedOutcome(t, result, "Although the Russians beat the Germans with a score of 2 to 1, the Russians will not dislodge the English fleet in the North Sea. This is because the English fleet supports the Russian attack. Since the English fleet is not dislodged, the Germans will not advance to the North Sea either.", "6.E.9")
}

// Test 6.E.10: NO SELF DISLODGEMENT WITH BELEAGUERED GARRISON AND HEAD-TO-HEAD BATTLE (E.8)
func TestDATCE10_NoSelfDislodgementWithBeleagueredGarrisonAndHeadToHeadBattle(t *testing.T) {
	gameState := CreateDATCGameState(t)

	// Add units based on test case
	gameState.Board.Units["north_sea"] = &game.Unit{
		Type:     game.Fleet,
		Owner:    game.England,
		Province: "north_sea",
	}
	gameState.Board.Units["york"] = &game.Unit{
		Type:     game.Fleet,
		Owner:    game.England,
		Province: "york",
	}
	gameState.Board.Units["holland"] = &game.Unit{
		Type:     game.Fleet,
		Owner:    game.Germany,
		Province: "holland",
	}
	gameState.Board.Units["heligoland_bight"] = &game.Unit{
		Type:     game.Fleet,
		Owner:    game.Germany,
		Province: "heligoland_bight",
	}
	gameState.Board.Units["skagerrak"] = &game.Unit{
		Type:     game.Fleet,
		Owner:    game.Russia,
		Province: "skagerrak",
	}
	gameState.Board.Units["norway"] = &game.Unit{
		Type:     game.Fleet,
		Owner:    game.Russia,
		Province: "norway",
	}

	// Add orders
	gameState.RawOrders[game.England] = []string{
		"f north_sea - norway",
		"f york s norway - north_sea",
	}
	gameState.RawOrders[game.Germany] = []string{
		"f holland s heligoland_bight - north_sea",
		"f heligoland_bight - north_sea",
	}
	gameState.RawOrders[game.Russia] = []string{
		"f skagerrak s norway - north_sea",
		"f norway - north_sea",
	}

	// Process turn
	result := ProcessDATCTest(t, gameState)

	// Expected: None of the units move
	ValidateExpectedOutcome(t, result, "Again, none of the units move.", "6.E.10")
}

// Test 6.E.11: ALMOST SELF DISLODGEMENT WITH BELEAGUERED GARRISON (E.9)
func TestDATCE11_AlmostSelfDislodgementWithBeleagueredGarrison(t *testing.T) {
	gameState := CreateDATCGameState(t)

	// Add units based on test case
	gameState.Board.Units["north_sea"] = &game.Unit{
		Type:     game.Fleet,
		Owner:    game.England,
		Province: "north_sea",
	}
	gameState.Board.Units["york"] = &game.Unit{
		Type:     game.Fleet,
		Owner:    game.England,
		Province: "york",
	}
	gameState.Board.Units["holland"] = &game.Unit{
		Type:     game.Fleet,
		Owner:    game.Germany,
		Province: "holland",
	}
	gameState.Board.Units["heligoland_bight"] = &game.Unit{
		Type:     game.Fleet,
		Owner:    game.Germany,
		Province: "heligoland_bight",
	}
	gameState.Board.Units["skagerrak"] = &game.Unit{
		Type:     game.Fleet,
		Owner:    game.Russia,
		Province: "skagerrak",
	}
	gameState.Board.Units["norway"] = &game.Unit{
		Type:     game.Fleet,
		Owner:    game.Russia,
		Province: "norway",
	}

	// Add orders
	gameState.RawOrders[game.England] = []string{
		"f north_sea - norwegian_sea",
		"f york s norway - north_sea",
	}
	gameState.RawOrders[game.Germany] = []string{
		"f holland s heligoland_bight - north_sea",
		"f heligoland_bight - north_sea",
	}
	gameState.RawOrders[game.Russia] = []string{
		"f skagerrak s norway - north_sea",
		"f norway - north_sea",
	}

	// Process turn
	result := ProcessDATCTest(t, gameState)

	// Expected: Both fleets move successfully
	ValidateExpectedOutcome(t, result, "Both the fleet in the North Sea and the fleet in Norway will move.", "6.E.11")
}

// Test 6.E.12: ALMOST CIRCULAR MOVEMENT WITH NO SELF DISLODGEMENT WITH BELEAGUERED GARRISON (E.10)
func TestDATCE12_AlmostCircularMovementWithNoSelfDislodgementWithBeleagueredGarrison(t *testing.T) {
	gameState := CreateDATCGameState(t)

	// Add units based on test case
	gameState.Board.Units["north_sea"] = &game.Unit{
		Type:     game.Fleet,
		Owner:    game.England,
		Province: "north_sea",
	}
	gameState.Board.Units["york"] = &game.Unit{
		Type:     game.Fleet,
		Owner:    game.England,
		Province: "york",
	}
	gameState.Board.Units["holland"] = &game.Unit{
		Type:     game.Fleet,
		Owner:    game.Germany,
		Province: "holland",
	}
	gameState.Board.Units["heligoland_bight"] = &game.Unit{
		Type:     game.Fleet,
		Owner:    game.Germany,
		Province: "heligoland_bight",
	}
	gameState.Board.Units["denmark"] = &game.Unit{
		Type:     game.Fleet,
		Owner:    game.Germany,
		Province: "denmark",
	}
	gameState.Board.Units["skagerrak"] = &game.Unit{
		Type:     game.Fleet,
		Owner:    game.Russia,
		Province: "skagerrak",
	}
	gameState.Board.Units["norway"] = &game.Unit{
		Type:     game.Fleet,
		Owner:    game.Russia,
		Province: "norway",
	}

	// Add orders
	gameState.RawOrders[game.England] = []string{
		"f north_sea - denmark",
		"f york s norway - north_sea",
	}
	gameState.RawOrders[game.Germany] = []string{
		"f holland s heligoland_bight - north_sea",
		"f heligoland_bight - north_sea",
		"f denmark - heligoland_bight",
	}
	gameState.RawOrders[game.Russia] = []string{
		"f skagerrak s norway - north_sea",
		"f norway - north_sea",
	}

	// Process turn
	result := ProcessDATCTest(t, gameState)

	// Expected: No unit moves
	ValidateExpectedOutcome(t, result, "No unit will move.", "6.E.12")
}

// Test 6.E.13: SUPPORT ON ATTACK ON OWN UNIT CAN BE USED FOR OTHER MEANS (E.12)
func TestDATCE13_SupportOnAttackOnOwnUnitCanBeUsedForOtherMeans(t *testing.T) {
	gameState := CreateDATCGameState(t)

	// Add units based on test case
	gameState.Board.Units["budapest"] = &game.Unit{
		Type:     game.Army,
		Owner:    game.Austria,
		Province: "budapest",
	}
	gameState.Board.Units["serbia"] = &game.Unit{
		Type:     game.Army,
		Owner:    game.Austria,
		Province: "serbia",
	}
	gameState.Board.Units["vienna"] = &game.Unit{
		Type:     game.Army,
		Owner:    game.Italy,
		Province: "vienna",
	}
	gameState.Board.Units["galicia"] = &game.Unit{
		Type:     game.Army,
		Owner:    game.Russia,
		Province: "galicia",
	}
	gameState.Board.Units["rumania"] = &game.Unit{
		Type:     game.Army,
		Owner:    game.Russia,
		Province: "rumania",
	}

	// Add orders
	gameState.RawOrders[game.Austria] = []string{
		"a budapest - rumania",
		"a serbia s vienna - budapest",
	}
	gameState.RawOrders[game.Italy] = []string{
		"a vienna - budapest",
	}
	gameState.RawOrders[game.Russia] = []string{
		"a galicia - budapest",
		"a rumania s galicia - budapest",
	}

	// Process turn
	result := ProcessDATCTest(t, gameState)

	// Expected: Support prevents Russian attack
	ValidateExpectedOutcome(t, result, "The support of Serbia on the Italian army prevents that the Russian army in Galicia will dislodge the Austrian army in Budapest. The Austrian army in Budapest will not move to Rumania.", "6.E.13")
}

// Test 6.E.14: THREE WAY BELEAGUERED GARRISON (E.13)
func TestDATCE14_ThreeWayBeleagueredGarrison(t *testing.T) {
	gameState := CreateDATCGameState(t)

	// Add units based on test case
	gameState.Board.Units["edinburgh"] = &game.Unit{
		Type:     game.Fleet,
		Owner:    game.England,
		Province: "edinburgh",
	}
	gameState.Board.Units["york"] = &game.Unit{
		Type:     game.Fleet,
		Owner:    game.England,
		Province: "york",
	}
	gameState.Board.Units["belgium"] = &game.Unit{
		Type:     game.Fleet,
		Owner:    game.France,
		Province: "belgium",
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
	gameState.Board.Units["norwegian_sea"] = &game.Unit{
		Type:     game.Fleet,
		Owner:    game.Russia,
		Province: "norwegian_sea",
	}
	gameState.Board.Units["norway"] = &game.Unit{
		Type:     game.Fleet,
		Owner:    game.Russia,
		Province: "norway",
	}

	// Add orders
	gameState.RawOrders[game.England] = []string{
		"f edinburgh s york - north_sea",
		"f york - north_sea",
	}
	gameState.RawOrders[game.France] = []string{
		"f belgium - north_sea",
		"f english_channel s belgium - north_sea",
	}
	gameState.RawOrders[game.Germany] = []string{
		"f north_sea hold",
	}
	gameState.RawOrders[game.Russia] = []string{
		"f norwegian_sea - north_sea",
		"f norway s norwegian_sea - north_sea",
	}

	// Process turn
	result := ProcessDATCTest(t, gameState)

	// Expected: None of the attacks succeed
	ValidateExpectedOutcome(t, result, "None of the attacks succeed. The fleet in the North Sea is not dislodged.", "6.E.14")
}
