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

// Test 6.D.6: SUPPORT TO HOLD ON CONVOYING UNIT ALLOWED
func TestDATCD6_SupportToHoldOnConvoyingUnitAllowed(t *testing.T) {
	gameState := CreateDATCGameState(t)

	// Add units based on orders
	gameState.Board.Units["berlin"] = &game.Unit{
		Type:     game.Army,
		Owner:    game.Germany,
		Province: "berlin",
	}
	gameState.Board.Units["baltic_sea"] = &game.Unit{
		Type:     game.Fleet,
		Owner:    game.Germany,
		Province: "baltic_sea",
	}
	gameState.Board.Units["prussia"] = &game.Unit{
		Type:     game.Fleet,
		Owner:    game.Germany,
		Province: "prussia",
	}
	gameState.Board.Units["livonia"] = &game.Unit{
		Type:     game.Fleet,
		Owner:    game.Russia,
		Province: "livonia",
	}
	gameState.Board.Units["gulf_of_bothnia"] = &game.Unit{
		Type:     game.Fleet,
		Owner:    game.Russia,
		Province: "gulf_of_bothnia",
	}

	// Add orders
	gameState.RawOrders[game.Germany] = []string{
		"a berlin - sweden",
		"f baltic_sea c berlin - sweden",
		"f prussia s baltic_sea",
	}
	gameState.RawOrders[game.Russia] = []string{
		"f livonia - baltic_sea",
		"f gulf_of_bothnia s livonia - baltic_sea",
	}

	// Process turn
	result := ProcessDATCTest(t, gameState)

	// Expected: Russian move fails, convoy succeeds
	ValidateExpectedOutcome(t, result, "The Russian move from Livonia to the Baltic Sea fails. The convoy from Berlin to Sweden succeeds.", "6.D.6")
}

// Test 6.D.7: SUPPORT TO HOLD ON MOVING UNIT NOT ALLOWED
func TestDATCD7_SupportToHoldOnMovingUnitNotAllowed(t *testing.T) {
	gameState := CreateDATCGameState(t)

	// Add units based on orders
	gameState.Board.Units["baltic_sea"] = &game.Unit{
		Type:     game.Fleet,
		Owner:    game.Germany,
		Province: "baltic_sea",
	}
	gameState.Board.Units["prussia"] = &game.Unit{
		Type:     game.Fleet,
		Owner:    game.Germany,
		Province: "prussia",
	}
	gameState.Board.Units["livonia"] = &game.Unit{
		Type:     game.Fleet,
		Owner:    game.Russia,
		Province: "livonia",
	}
	gameState.Board.Units["gulf_of_bothnia"] = &game.Unit{
		Type:     game.Fleet,
		Owner:    game.Russia,
		Province: "gulf_of_bothnia",
	}
	gameState.Board.Units["finland"] = &game.Unit{
		Type:     game.Army,
		Owner:    game.Russia,
		Province: "finland",
	}

	// Add orders
	gameState.RawOrders[game.Germany] = []string{
		"f baltic_sea - sweden",
		"f prussia s baltic_sea",
	}
	gameState.RawOrders[game.Russia] = []string{
		"f livonia - baltic_sea",
		"f gulf_of_bothnia s livonia - baltic_sea",
		"a finland - sweden",
	}

	// Process turn
	result := ProcessDATCTest(t, gameState)

	// Expected: Support fails, Baltic Sea bounces and is dislodged
	ValidateExpectedOutcome(t, result, "The support of the fleet in Prussia fails. The fleet in Baltic Sea will bounce on the Russian army in Finland and will be dislodged by the Russian fleet from Livonia.", "6.D.7")
}

// Test 6.D.8: FAILED CONVOY CANNOT RECEIVE HOLD SUPPORT
func TestDATCD8_FailedConvoyCannotReceiveHoldSupport(t *testing.T) {
	gameState := CreateDATCGameState(t)

	// Add units based on orders
	gameState.Board.Units["ionian_sea"] = &game.Unit{
		Type:     game.Fleet,
		Owner:    game.Austria,
		Province: "ionian_sea",
	}
	gameState.Board.Units["serbia"] = &game.Unit{
		Type:     game.Army,
		Owner:    game.Austria,
		Province: "serbia",
	}
	gameState.Board.Units["albania"] = &game.Unit{
		Type:     game.Army,
		Owner:    game.Austria,
		Province: "albania",
	}
	gameState.Board.Units["greece"] = &game.Unit{
		Type:     game.Army,
		Owner:    game.Turkey,
		Province: "greece",
	}
	gameState.Board.Units["bulgaria"] = &game.Unit{
		Type:     game.Army,
		Owner:    game.Turkey,
		Province: "bulgaria",
	}

	// Add orders
	gameState.RawOrders[game.Austria] = []string{
		"f ionian_sea hold",
		"a serbia s albania - greece",
		"a albania - greece",
	}
	gameState.RawOrders[game.Turkey] = []string{
		"a greece - naples",
		"a bulgaria s greece",
	}

	// Process turn
	result := ProcessDATCTest(t, gameState)

	// Expected: Greece is dislodged (cannot receive hold support while trying to move)
	ValidateExpectedOutcome(t, result, "There was a possible convoy from Greece to Naples, before the orders were processed. However, the convoy cannot be executed, because the fleet in the Ionian Sea does not convoy. Since the army in Greece tried to move, it cannot receive support in hold. The army in Greece is dislodged by the army from Albania.", "6.D.8")
}

// Test 6.D.9: SUPPORT TO MOVE ON HOLDING UNIT NOT ALLOWED
func TestDATCD9_SupportToMoveOnHoldingUnitNotAllowed(t *testing.T) {
	gameState := CreateDATCGameState(t)

	// Add units based on orders
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
	gameState.Board.Units["albania"] = &game.Unit{
		Type:     game.Army,
		Owner:    game.Austria,
		Province: "albania",
	}
	gameState.Board.Units["trieste"] = &game.Unit{
		Type:     game.Army,
		Owner:    game.Austria,
		Province: "trieste",
	}

	// Add orders
	gameState.RawOrders[game.Italy] = []string{
		"a venice - trieste",
		"a tyrolia s venice - trieste",
	}
	gameState.RawOrders[game.Austria] = []string{
		"a albania s trieste - serbia",
		"a trieste hold",
	}

	// Process turn
	result := ProcessDATCTest(t, gameState)

	// Expected: Support fails, Trieste is dislodged
	ValidateExpectedOutcome(t, result, "The support of the army in Albania fails and the army in Trieste is dislodged by the army from Venice.", "6.D.9")
}

// Test 6.D.10: SELF DISLODGMENT PROHIBITED
func TestDATCD10_SelfDislodgmentProhibited(t *testing.T) {
	gameState := CreateDATCGameState(t)

	// Add units based on orders
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
		"a berlin hold",
		"f kiel - berlin",
		"a munich s kiel - berlin",
	}

	// Process turn
	result := ProcessDATCTest(t, gameState)

	// Expected: Move to Berlin fails (self-dislodgement prohibited)
	ValidateExpectedOutcome(t, result, "Move to Berlin fails.", "6.D.10")
}

// Test 6.D.11: NO SELF DISLODGMENT OF RETURNING UNIT
func TestDATCD11_NoSelfDislodgmentOfReturningUnit(t *testing.T) {
	gameState := CreateDATCGameState(t)

	// Add units based on orders
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
	gameState.Board.Units["warsaw"] = &game.Unit{
		Type:     game.Army,
		Owner:    game.Russia,
		Province: "warsaw",
	}

	// Add orders
	gameState.RawOrders[game.Germany] = []string{
		"a berlin - prussia",
		"f kiel - berlin",
		"a munich s kiel - berlin",
	}
	gameState.RawOrders[game.Russia] = []string{
		"a warsaw - prussia",
	}

	// Process turn
	result := ProcessDATCTest(t, gameState)

	// Expected: Army in Berlin bounces, but is not dislodged by own unit
	ValidateExpectedOutcome(t, result, "Army in Berlin bounces, but is not dislodged by own unit.", "6.D.11")
}

// Test 6.D.12: SUPPORTING A FOREIGN UNIT TO DISLODGE OWN UNIT PROHIBITED
func TestDATCD12_SupportingForeignUnitToDislodgeOwnUnitProhibited(t *testing.T) {
	gameState := CreateDATCGameState(t)

	// Add units based on orders
	gameState.Board.Units["trieste"] = &game.Unit{
		Type:     game.Fleet,
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

	// Add orders
	gameState.RawOrders[game.Austria] = []string{
		"f trieste hold",
		"a vienna s venice - trieste",
	}
	gameState.RawOrders[game.Italy] = []string{
		"a venice - trieste",
	}

	// Process turn
	result := ProcessDATCTest(t, gameState)

	// Expected: No dislodgment of fleet in Trieste
	ValidateExpectedOutcome(t, result, "No dislodgment of fleet in Trieste.", "6.D.12")
}

// Test 6.D.13: SUPPORTING A FOREIGN UNIT TO DISLODGE A RETURNING OWN UNIT PROHIBITED
func TestDATCD13_SupportingForeignUnitToDislodgeReturningOwnUnitProhibited(t *testing.T) {
	gameState := CreateDATCGameState(t)

	// Add units based on orders
	gameState.Board.Units["trieste"] = &game.Unit{
		Type:     game.Fleet,
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
	gameState.Board.Units["apulia"] = &game.Unit{
		Type:     game.Fleet,
		Owner:    game.Italy,
		Province: "apulia",
	}

	// Add orders
	gameState.RawOrders[game.Austria] = []string{
		"f trieste - adriatic_sea",
		"a vienna s venice - trieste",
	}
	gameState.RawOrders[game.Italy] = []string{
		"a venice - trieste",
		"f apulia - adriatic_sea",
	}

	// Process turn
	result := ProcessDATCTest(t, gameState)

	// Expected: No dislodgment of fleet in Trieste
	ValidateExpectedOutcome(t, result, "No dislodgment of fleet in Trieste.", "6.D.13")
}

// Test 6.D.14: SUPPORTING A FOREIGN UNIT IS NOT ENOUGH TO PREVENT DISLODGEMENT
func TestDATCD14_SupportingForeignUnitNotEnoughToPreventDislodgement(t *testing.T) {
	gameState := CreateDATCGameState(t)

	// Add units based on orders
	gameState.Board.Units["trieste"] = &game.Unit{
		Type:     game.Fleet,
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
	gameState.Board.Units["adriatic_sea"] = &game.Unit{
		Type:     game.Fleet,
		Owner:    game.Italy,
		Province: "adriatic_sea",
	}

	// Add orders
	gameState.RawOrders[game.Austria] = []string{
		"f trieste hold",
		"a vienna s venice - trieste",
	}
	gameState.RawOrders[game.Italy] = []string{
		"a venice - trieste",
		"a tyrolia s venice - trieste",
		"f adriatic_sea s venice - trieste",
	}

	// Process turn
	result := ProcessDATCTest(t, gameState)

	// Expected: The fleet in Trieste is dislodged
	ValidateExpectedOutcome(t, result, "The fleet in Trieste is dislodged.", "6.D.14")
}

// Test 6.D.15: DEFENDER CANNOT CUT SUPPORT FOR ATTACK ON ITSELF
func TestDATCD15_DefenderCannotCutSupportForAttackOnItself(t *testing.T) {
	gameState := CreateDATCGameState(t)

	// Add units based on orders
	gameState.Board.Units["constantinople"] = &game.Unit{
		Type:     game.Fleet,
		Owner:    game.Russia,
		Province: "constantinople",
	}
	gameState.Board.Units["black_sea"] = &game.Unit{
		Type:     game.Fleet,
		Owner:    game.Russia,
		Province: "black_sea",
	}
	gameState.Board.Units["ankara"] = &game.Unit{
		Type:     game.Fleet,
		Owner:    game.Turkey,
		Province: "ankara",
	}

	// Add orders
	gameState.RawOrders[game.Russia] = []string{
		"f constantinople s black_sea - ankara",
		"f black_sea - ankara",
	}
	gameState.RawOrders[game.Turkey] = []string{
		"f ankara - constantinople",
	}

	// Process turn
	result := ProcessDATCTest(t, gameState)

	// Expected: Support is not cut, fleet in Ankara is dislodged
	ValidateExpectedOutcome(t, result, "The support of Constantinople is not cut and the fleet in Ankara is dislodged by the fleet in the Black Sea.", "6.D.15")
}

// Test 6.D.16: CONVOYING A UNIT DISLODGING A UNIT OF SAME POWER IS ALLOWED
func TestDATCD16_ConvoyingUnitDislodgingUnitOfSamePowerAllowed(t *testing.T) {
	gameState := CreateDATCGameState(t)

	// Add units based on orders
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
	gameState.Board.Units["belgium"] = &game.Unit{
		Type:     game.Army,
		Owner:    game.France,
		Province: "belgium",
	}

	// Add orders
	gameState.RawOrders[game.England] = []string{
		"a london hold",
		"f north_sea c belgium - london",
	}
	gameState.RawOrders[game.France] = []string{
		"f english_channel s belgium - london",
		"a belgium - london",
	}

	// Process turn
	result := ProcessDATCTest(t, gameState)

	// Expected: English army in London is dislodged by French army from Belgium
	ValidateExpectedOutcome(t, result, "The English army in London is dislodged by the French army coming from Belgium.", "6.D.16")
}

// Test 6.D.17: DISLODGEMENT CUTS SUPPORTS
func TestDATCD17_DislodgementCutsSupports(t *testing.T) {
	gameState := CreateDATCGameState(t)

	// Add units based on orders
	gameState.Board.Units["constantinople"] = &game.Unit{
		Type:     game.Fleet,
		Owner:    game.Russia,
		Province: "constantinople",
	}
	gameState.Board.Units["black_sea"] = &game.Unit{
		Type:     game.Fleet,
		Owner:    game.Russia,
		Province: "black_sea",
	}
	gameState.Board.Units["ankara"] = &game.Unit{
		Type:     game.Fleet,
		Owner:    game.Turkey,
		Province: "ankara",
	}
	gameState.Board.Units["smyrna"] = &game.Unit{
		Type:     game.Army,
		Owner:    game.Turkey,
		Province: "smyrna",
	}
	gameState.Board.Units["armenia"] = &game.Unit{
		Type:     game.Army,
		Owner:    game.Turkey,
		Province: "armenia",
	}

	// Add orders
	gameState.RawOrders[game.Russia] = []string{
		"f constantinople s black_sea - ankara",
		"f black_sea - ankara",
	}
	gameState.RawOrders[game.Turkey] = []string{
		"f ankara - constantinople",
		"a smyrna s ankara - constantinople",
		"a armenia - ankara",
	}

	// Process turn
	result := ProcessDATCTest(t, gameState)

	// Expected: Constantinople is dislodged, cutting support; Black Sea bounces with Armenia
	ValidateExpectedOutcome(t, result, "The Russian fleet in Constantinople is dislodged. This cuts the support to from Black Sea to Ankara. Black Sea will bounce with the army from Armenia.", "6.D.17")
}

// Test 6.D.18: A SURVIVING UNIT WILL SUSTAIN SUPPORT
func TestDATCD18_SurvivingUnitWillSustainSupport(t *testing.T) {
	gameState := CreateDATCGameState(t)

	// Add units based on orders
	gameState.Board.Units["constantinople"] = &game.Unit{
		Type:     game.Fleet,
		Owner:    game.Russia,
		Province: "constantinople",
	}
	gameState.Board.Units["black_sea"] = &game.Unit{
		Type:     game.Fleet,
		Owner:    game.Russia,
		Province: "black_sea",
	}
	gameState.Board.Units["bulgaria"] = &game.Unit{
		Type:     game.Army,
		Owner:    game.Russia,
		Province: "bulgaria",
	}
	gameState.Board.Units["ankara"] = &game.Unit{
		Type:     game.Fleet,
		Owner:    game.Turkey,
		Province: "ankara",
	}
	gameState.Board.Units["smyrna"] = &game.Unit{
		Type:     game.Army,
		Owner:    game.Turkey,
		Province: "smyrna",
	}
	gameState.Board.Units["armenia"] = &game.Unit{
		Type:     game.Army,
		Owner:    game.Turkey,
		Province: "armenia",
	}

	// Add orders
	gameState.RawOrders[game.Russia] = []string{
		"f constantinople s black_sea - ankara",
		"f black_sea - ankara",
		"a bulgaria s constantinople",
	}
	gameState.RawOrders[game.Turkey] = []string{
		"f ankara - constantinople",
		"a smyrna s ankara - constantinople",
		"a armenia - ankara",
	}

	// Process turn
	result := ProcessDATCTest(t, gameState)

	// Expected: Russian fleet in Black Sea dislodges Turkish fleet in Ankara
	ValidateExpectedOutcome(t, result, "The Russian fleet in the Black Sea will dislodge the Turkish fleet in Ankara.", "6.D.18")
}

// Test 6.D.19: EVEN WHEN SURVIVING IS IN ALTERNATIVE WAY
func TestDATCD19_EvenWhenSurvivingIsInAlternativeWay(t *testing.T) {
	gameState := CreateDATCGameState(t)

	// Add units based on orders
	gameState.Board.Units["constantinople"] = &game.Unit{
		Type:     game.Fleet,
		Owner:    game.Russia,
		Province: "constantinople",
	}
	gameState.Board.Units["black_sea"] = &game.Unit{
		Type:     game.Fleet,
		Owner:    game.Russia,
		Province: "black_sea",
	}
	gameState.Board.Units["smyrna"] = &game.Unit{
		Type:     game.Army,
		Owner:    game.Russia,
		Province: "smyrna",
	}
	gameState.Board.Units["ankara"] = &game.Unit{
		Type:     game.Fleet,
		Owner:    game.Turkey,
		Province: "ankara",
	}

	// Add orders
	gameState.RawOrders[game.Russia] = []string{
		"f constantinople s black_sea - ankara",
		"f black_sea - ankara",
		"a smyrna s ankara - constantinople",
	}
	gameState.RawOrders[game.Turkey] = []string{
		"f ankara - constantinople",
	}

	// Process turn
	result := ProcessDATCTest(t, gameState)

	// Expected: Constantinople is not dislodged, support sustains
	ValidateExpectedOutcome(t, result, "The Russian fleet in Constantinople is not dislodged, because one of the supports is of Russian origin. The support from Black Sea to Ankara will sustain and the fleet in Ankara is dislodged.", "6.D.19")
}

// Test 6.D.20: UNIT CANNOT CUT SUPPORT OF ITS OWN COUNTRY
func TestDATCD20_UnitCannotCutSupportOfItsOwnCountry(t *testing.T) {
	gameState := CreateDATCGameState(t)

	// Add units based on orders
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
	gameState.Board.Units["york"] = &game.Unit{
		Type:     game.Army,
		Owner:    game.England,
		Province: "york",
	}
	gameState.Board.Units["english_channel"] = &game.Unit{
		Type:     game.Fleet,
		Owner:    game.France,
		Province: "english_channel",
	}

	// Add orders
	gameState.RawOrders[game.England] = []string{
		"f london s north_sea - english_channel",
		"f north_sea - english_channel",
		"a york - london",
	}
	gameState.RawOrders[game.France] = []string{
		"f english_channel hold",
	}

	// Process turn
	result := ProcessDATCTest(t, gameState)

	// Expected: Army in Yorkshire does not cut support, English Channel is dislodged
	ValidateExpectedOutcome(t, result, "The army in York does not cut support. This means that the fleet in the English Channel is dislodged by the fleet in the North Sea.", "6.D.20")
}

// Test 6.D.21: DISLODGING DOES NOT CANCEL A SUPPORT CUT
func TestDATCD21_DislodgingDoesNotCancelSupportCut(t *testing.T) {
	gameState := CreateDATCGameState(t)

	// Add units based on orders
	gameState.Board.Units["trieste"] = &game.Unit{
		Type:     game.Fleet,
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
	gameState.Board.Units["munich"] = &game.Unit{
		Type:     game.Army,
		Owner:    game.Germany,
		Province: "munich",
	}
	gameState.Board.Units["silesia"] = &game.Unit{
		Type:     game.Army,
		Owner:    game.Russia,
		Province: "silesia",
	}
	gameState.Board.Units["berlin"] = &game.Unit{
		Type:     game.Army,
		Owner:    game.Russia,
		Province: "berlin",
	}

	// Add orders
	gameState.RawOrders[game.Austria] = []string{
		"f trieste hold",
	}
	gameState.RawOrders[game.Italy] = []string{
		"a venice - trieste",
		"a tyrolia s venice - trieste",
	}
	gameState.RawOrders[game.Germany] = []string{
		"a munich - tyrolia",
	}
	gameState.RawOrders[game.Russia] = []string{
		"a silesia - munich",
		"a berlin s silesia - munich",
	}

	// Process turn
	result := ProcessDATCTest(t, gameState)

	// Expected: German army is dislodged but still cuts support
	ValidateExpectedOutcome(t, result, "Although the German army is dislodged, it still cuts the support from Tyrolia to Venice. Venice will not advance to Trieste.", "6.D.21")
}

// Test 6.D.22: IMPOSSIBLE FLEET MOVE CANNOT BE SUPPORTED
func TestDATCD22_ImpossibleFleetMoveCannotBeSupported(t *testing.T) {
	gameState := CreateDATCGameState(t)

	// Add units based on orders
	gameState.Board.Units["kiel"] = &game.Unit{
		Type:     game.Fleet,
		Owner:    game.Germany,
		Province: "kiel",
	}
	gameState.Board.Units["burgundy"] = &game.Unit{
		Type:     game.Army,
		Owner:    game.Germany,
		Province: "burgundy",
	}
	gameState.Board.Units["munich"] = &game.Unit{
		Type:     game.Army,
		Owner:    game.Russia,
		Province: "munich",
	}
	gameState.Board.Units["berlin"] = &game.Unit{
		Type:     game.Army,
		Owner:    game.Russia,
		Province: "berlin",
	}

	// Add orders
	gameState.RawOrders[game.Germany] = []string{
		"f kiel - munich",
		"a burgundy s kiel - munich",
	}
	gameState.RawOrders[game.Russia] = []string{
		"a munich - kiel",
		"a berlin s munich - kiel",
	}

	// Process turn
	result := ProcessDATCTest(t, gameState)

	// Expected: German move is illegal, Russian army succeeds
	ValidateExpectedOutcome(t, result, "The German move from Kiel to Munich is illegal (fleets cannot go to Munich). Illegal moves cannot be supported. The Russian army in Munich will dislodge the German fleet in Kiel.", "6.D.22")
}

// Test 6.D.23: IMPOSSIBLE COAST MOVE CANNOT BE SUPPORTED
func TestDATCD23_ImpossibleCoastMoveCannotBeSupported(t *testing.T) {
	gameState := CreateDATCGameState(t)

	// Add units based on orders
	gameState.Board.Units["gulf_of_lyon"] = &game.Unit{
		Type:     game.Fleet,
		Owner:    game.Italy,
		Province: "gulf_of_lyon",
	}
	gameState.Board.Units["western_mediterranean"] = &game.Unit{
		Type:     game.Fleet,
		Owner:    game.Italy,
		Province: "western_mediterranean",
	}
	gameState.Board.Units["spain_sc"] = &game.Unit{
		Type:     game.Fleet,
		Owner:    game.France,
		Province: "spain_sc",
	}
	gameState.Board.Units["marseilles"] = &game.Unit{
		Type:     game.Fleet,
		Owner:    game.France,
		Province: "marseilles",
	}

	// Add orders
	gameState.RawOrders[game.Italy] = []string{
		"f gulf_of_lyon - spain_sc",
		"f western_mediterranean s gulf_of_lyon - spain_sc",
	}
	gameState.RawOrders[game.France] = []string{
		"f spain_sc - gulf_of_lyon",
		"f marseilles s spain_sc - gulf_of_lyon",
	}

	// Process turn
	result := ProcessDATCTest(t, gameState)

	// Expected: French move is illegal, Italian fleet succeeds
	ValidateExpectedOutcome(t, result, "The French move from Spain North Coast to Gulf of Lyon is illegal (wrong coast). Therefore, the support from Marseilles is also illegal. The Italian fleet will dislodge the French fleet in Spain.", "6.D.23")
}

// Test 6.D.24: IMPOSSIBLE ARMY MOVE CANNOT BE SUPPORTED
func TestDATCD24_ImpossibleArmyMoveCannotBeSupported(t *testing.T) {
	gameState := CreateDATCGameState(t)

	// Add units based on orders
	gameState.Board.Units["marseilles"] = &game.Unit{
		Type:     game.Army,
		Owner:    game.France,
		Province: "marseilles",
	}
	gameState.Board.Units["spain_sc"] = &game.Unit{
		Type:     game.Fleet,
		Owner:    game.France,
		Province: "spain_sc",
	}
	gameState.Board.Units["gulf_of_lyon"] = &game.Unit{
		Type:     game.Fleet,
		Owner:    game.Italy,
		Province: "gulf_of_lyon",
	}
	gameState.Board.Units["tyrrhenian_sea"] = &game.Unit{
		Type:     game.Fleet,
		Owner:    game.Turkey,
		Province: "tyrrhenian_sea",
	}
	gameState.Board.Units["western_mediterranean"] = &game.Unit{
		Type:     game.Fleet,
		Owner:    game.Turkey,
		Province: "western_mediterranean",
	}

	// Add orders
	gameState.RawOrders[game.France] = []string{
		"a marseilles - gulf_of_lyon",
		"f spain_sc s marseilles - gulf_of_lyon",
	}
	gameState.RawOrders[game.Italy] = []string{
		"f gulf_of_lyon hold",
	}
	gameState.RawOrders[game.Turkey] = []string{
		"f tyrrhenian_sea s western_mediterranean - gulf_of_lyon",
		"f western_mediterranean - gulf_of_lyon",
	}

	// Process turn
	result := ProcessDATCTest(t, gameState)

	// Expected: French move is illegal, Turkish fleet succeeds
	ValidateExpectedOutcome(t, result, "The French move from Marseilles to Gulf of Lyon is illegal (armies cannot move to sea). Therefore, the support from Spain is also illegal. The Turkish fleet will dislodge the Italian fleet in Gulf of Lyon.", "6.D.24")
}

// Test 6.D.25: FAILING HOLD SUPPORT CAN BE SUPPORTED
func TestDATCD25_FailingHoldSupportCanBeSupported(t *testing.T) {
	gameState := CreateDATCGameState(t)

	// Add units based on orders
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

	// Add orders
	gameState.RawOrders[game.Germany] = []string{
		"a berlin s prussia",
		"f kiel s berlin",
	}
	gameState.RawOrders[game.Russia] = []string{
		"f baltic_sea s prussia - berlin",
		"a prussia - berlin",
	}

	// Process turn
	result := ProcessDATCTest(t, gameState)

	// Expected: Berlin will not be dislodged
	ValidateExpectedOutcome(t, result, "Although the support order from Berlin to Prussia is unmatched (Prussia is not holding), the support from Kiel to Berlin is still valid. Berlin will not be dislodged.", "6.D.25")
}

// Test 6.D.26: FAILING MOVE SUPPORT CAN BE SUPPORTED
func TestDATCD26_FailingMoveSupportCanBeSupported(t *testing.T) {
	gameState := CreateDATCGameState(t)

	// Add units based on orders
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

	// Add orders
	gameState.RawOrders[game.Germany] = []string{
		"a berlin s prussia - silesia",
		"f kiel s berlin",
	}
	gameState.RawOrders[game.Russia] = []string{
		"f baltic_sea s prussia - berlin",
		"a prussia - berlin",
	}

	// Process turn
	result := ProcessDATCTest(t, gameState)

	// Expected: Berlin will not be dislodged
	ValidateExpectedOutcome(t, result, "Again, Berlin will not be dislodged.", "6.D.26")
}

// Test 6.D.27: FAILING CONVOY CAN BE SUPPORTED
func TestDATCD27_FailingConvoyCanBeSupported(t *testing.T) {
	gameState := CreateDATCGameState(t)

	// Add units based on orders
	gameState.Board.Units["sweden"] = &game.Unit{
		Type:     game.Fleet,
		Owner:    game.England,
		Province: "sweden",
	}
	gameState.Board.Units["denmark"] = &game.Unit{
		Type:     game.Fleet,
		Owner:    game.England,
		Province: "denmark",
	}
	gameState.Board.Units["berlin"] = &game.Unit{
		Type:     game.Army,
		Owner:    game.Germany,
		Province: "berlin",
	}
	gameState.Board.Units["baltic_sea"] = &game.Unit{
		Type:     game.Fleet,
		Owner:    game.Russia,
		Province: "baltic_sea",
	}
	gameState.Board.Units["prussia"] = &game.Unit{
		Type:     game.Fleet,
		Owner:    game.Russia,
		Province: "prussia",
	}

	// Add orders
	gameState.RawOrders[game.England] = []string{
		"f sweden - baltic_sea",
		"f denmark s sweden - baltic_sea",
	}
	gameState.RawOrders[game.Germany] = []string{
		"a berlin hold",
	}
	gameState.RawOrders[game.Russia] = []string{
		"f baltic_sea c berlin - livonia",
		"f prussia s baltic_sea",
	}

	// Process turn
	result := ProcessDATCTest(t, gameState)

	// Expected: Baltic Sea is not dislodged
	ValidateExpectedOutcome(t, result, "The convoy order in the Baltic Sea is unmatched and fails. However, the support of Prussia on the Baltic Sea is still valid and the fleet in the Baltic Sea is not dislodged.", "6.D.27")
}

// Test 6.D.28: IMPOSSIBLE MOVE AND SUPPORT
func TestDATCD28_ImpossibleMoveAndSupport(t *testing.T) {
	gameState := CreateDATCGameState(t)

	// Add units based on orders
	gameState.Board.Units["budapest"] = &game.Unit{
		Type:     game.Army,
		Owner:    game.Austria,
		Province: "budapest",
	}
	gameState.Board.Units["rumania"] = &game.Unit{
		Type:     game.Fleet,
		Owner:    game.Russia,
		Province: "rumania",
	}
	gameState.Board.Units["black_sea"] = &game.Unit{
		Type:     game.Fleet,
		Owner:    game.Turkey,
		Province: "black_sea",
	}
	gameState.Board.Units["bulgaria"] = &game.Unit{
		Type:     game.Army,
		Owner:    game.Turkey,
		Province: "bulgaria",
	}

	// Add orders
	gameState.RawOrders[game.Austria] = []string{
		"a budapest s rumania",
	}
	gameState.RawOrders[game.Russia] = []string{
		"f rumania - holland",
	}
	gameState.RawOrders[game.Turkey] = []string{
		"f black_sea - rumania",
		"a bulgaria s black_sea - rumania",
	}

	// Process turn
	result := ProcessDATCTest(t, gameState)

	// Expected: Fleet in Rumania is not dislodged
	ValidateExpectedOutcome(t, result, "Illegal orders are ignored. Without an order, Rumania holds and receives support. The fleet in Rumania is not dislodged.", "6.D.28")
}

// Test 6.D.29: MOVE TO IMPOSSIBLE COAST AND SUPPORT
func TestDATCD29_MoveToImpossibleCoastAndSupport(t *testing.T) {
	gameState := CreateDATCGameState(t)

	// Add units based on orders
	gameState.Board.Units["budapest"] = &game.Unit{
		Type:     game.Army,
		Owner:    game.Austria,
		Province: "budapest",
	}
	gameState.Board.Units["rumania"] = &game.Unit{
		Type:     game.Fleet,
		Owner:    game.Russia,
		Province: "rumania",
	}
	gameState.Board.Units["black_sea"] = &game.Unit{
		Type:     game.Fleet,
		Owner:    game.Turkey,
		Province: "black_sea",
	}
	gameState.Board.Units["bulgaria"] = &game.Unit{
		Type:     game.Army,
		Owner:    game.Turkey,
		Province: "bulgaria",
	}

	// Add orders
	gameState.RawOrders[game.Austria] = []string{
		"a budapest s rumania",
	}
	gameState.RawOrders[game.Russia] = []string{
		"f rumania - bulgaria_sc",
	}
	gameState.RawOrders[game.Turkey] = []string{
		"f black_sea - rumania",
		"a bulgaria s black_sea - rumania",
	}

	// Process turn
	result := ProcessDATCTest(t, gameState)

	// Expected: Fleet in Rumania is not dislodged
	ValidateExpectedOutcome(t, result, "Illegal orders are ignored. Without an order, Rumania holds and receives support. The fleet in Rumania is not dislodged.", "6.D.29")
}

// Test 6.D.30: MOVE WITHOUT COAST AND SUPPORT
func TestDATCD30_MoveWithoutCoastAndSupport(t *testing.T) {
	gameState := CreateDATCGameState(t)

	// Add units based on orders
	gameState.Board.Units["aegean_sea"] = &game.Unit{
		Type:     game.Fleet,
		Owner:    game.Italy,
		Province: "aegean_sea",
	}
	gameState.Board.Units["constantinople"] = &game.Unit{
		Type:     game.Fleet,
		Owner:    game.Russia,
		Province: "constantinople",
	}
	gameState.Board.Units["black_sea"] = &game.Unit{
		Type:     game.Fleet,
		Owner:    game.Turkey,
		Province: "black_sea",
	}
	gameState.Board.Units["bulgaria"] = &game.Unit{
		Type:     game.Army,
		Owner:    game.Turkey,
		Province: "bulgaria",
	}

	// Add orders
	gameState.RawOrders[game.Italy] = []string{
		"f aegean_sea s constantinople",
	}
	gameState.RawOrders[game.Russia] = []string{
		"f constantinople - bulgaria",
	}
	gameState.RawOrders[game.Turkey] = []string{
		"f black_sea - constantinople",
		"a bulgaria s black_sea - constantinople",
	}

	// Process turn
	result := ProcessDATCTest(t, gameState)

	// Expected: Fleet in Constantinople is not dislodged
	ValidateExpectedOutcome(t, result, "Illegal orders are ignored. Without an order, Constantinople holds and receives support. The fleet in Constantinople is not dislodged.", "6.D.30")
}

// Test 6.D.31: A TRICKY IMPOSSIBLE SUPPORT
func TestDATCD31_TrickyImpossibleSupport(t *testing.T) {
	gameState := CreateDATCGameState(t)

	// Add units based on orders
	gameState.Board.Units["rumania"] = &game.Unit{
		Type:     game.Army,
		Owner:    game.Austria,
		Province: "rumania",
	}
	gameState.Board.Units["black_sea"] = &game.Unit{
		Type:     game.Fleet,
		Owner:    game.Turkey,
		Province: "black_sea",
	}

	// Add orders
	gameState.RawOrders[game.Austria] = []string{
		"a rumania - armenia",
	}
	gameState.RawOrders[game.Turkey] = []string{
		"f black_sea s rumania - armenia",
	}

	// Process turn
	result := ProcessDATCTest(t, gameState)

	// Expected: Support is impossible, army bounces
	ValidateExpectedOutcome(t, result, "Although the army in Rumania can move to Armenia and the fleet in the Black Sea can also go to Armenia, the support is still not possible. The reason is that the only possible convoy is through the Black Sea and a fleet cannot convoy and support at the same time.", "6.D.31")
}

// Test 6.D.32: A MISSING FLEET
func TestDATCD32_MissingFleet(t *testing.T) {
	gameState := CreateDATCGameState(t)

	// Add units based on orders
	gameState.Board.Units["edinburgh"] = &game.Unit{
		Type:     game.Fleet,
		Owner:    game.England,
		Province: "edinburgh",
	}
	gameState.Board.Units["liverpool"] = &game.Unit{
		Type:     game.Army,
		Owner:    game.England,
		Province: "liverpool",
	}
	gameState.Board.Units["london"] = &game.Unit{
		Type:     game.Fleet,
		Owner:    game.France,
		Province: "london",
	}
	gameState.Board.Units["york"] = &game.Unit{
		Type:     game.Army,
		Owner:    game.Germany,
		Province: "york",
	}

	// Add orders
	gameState.RawOrders[game.England] = []string{
		"f edinburgh s liverpool - york",
		"a liverpool - york",
	}
	gameState.RawOrders[game.France] = []string{
		"f london s york",
	}
	gameState.RawOrders[game.Germany] = []string{
		"a york - holland",
	}

	// Process turn
	result := ProcessDATCTest(t, gameState)

	// Expected: German order cannot be executed
	ValidateExpectedOutcome(t, result, "The German order to Yorkshire cannot be executed, because there is no unit in Yorkshire. Therefore, the support of London fails and the army in Liverpool successfully moves to Yorkshire.", "6.D.32")
}

// Test 6.D.33: UNWANTED SUPPORT ALLOWED
func TestDATCD33_UnwantedSupportAllowed(t *testing.T) {
	gameState := CreateDATCGameState(t)

	// Add units based on orders
	gameState.Board.Units["serbia"] = &game.Unit{
		Type:     game.Army,
		Owner:    game.Austria,
		Province: "serbia",
	}
	gameState.Board.Units["vienna"] = &game.Unit{
		Type:     game.Army,
		Owner:    game.Austria,
		Province: "vienna",
	}
	gameState.Board.Units["galicia"] = &game.Unit{
		Type:     game.Army,
		Owner:    game.Russia,
		Province: "galicia",
	}
	gameState.Board.Units["bulgaria"] = &game.Unit{
		Type:     game.Army,
		Owner:    game.Turkey,
		Province: "bulgaria",
	}

	// Add orders
	gameState.RawOrders[game.Austria] = []string{
		"a serbia - budapest",
		"a vienna - budapest",
	}
	gameState.RawOrders[game.Russia] = []string{
		"a galicia s serbia - budapest",
	}
	gameState.RawOrders[game.Turkey] = []string{
		"a bulgaria - serbia",
	}

	// Process turn
	result := ProcessDATCTest(t, gameState)

	// Expected: Serbia advances to Budapest, Turkey captures Serbia
	ValidateExpectedOutcome(t, result, "Due to the Russian support, the army in Serbia advances to Budapest. This enables Turkey to capture Serbia with the army in Bulgaria.", "6.D.33")
}

// Test 6.D.34: SUPPORT TARGETING OWN AREA NOT ALLOWED
func TestDATCD34_SupportTargetingOwnAreaNotAllowed(t *testing.T) {
	gameState := CreateDATCGameState(t)

	// Add units based on orders
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
		"a silesia s berlin - silesia",
	}
	gameState.RawOrders[game.Russia] = []string{
		"a prussia - berlin",
	}

	// Process turn
	result := ProcessDATCTest(t, gameState)

	// Expected: Support is illegal, units bounce
	ValidateExpectedOutcome(t, result, "Support targeting the area where the supporting unit is standing, is illegal. The support from Silesia is ignored and the two armies bounce.", "6.D.34")
}
