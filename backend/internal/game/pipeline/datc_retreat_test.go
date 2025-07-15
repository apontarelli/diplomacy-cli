package pipeline

import (
	"testing"

	"diplomacy-cli/backend/internal/game"
)

// Test 6.H.1: NO SUPPORTS DURING RETREAT
// Supports are not allowed in the retreat phase.
func TestDATCH1_NoSupportsDuringRetreat(t *testing.T) {
	gameState := CreateDATCGameState(t)

	// Set up units for retreat scenario
	gameState.Board.Units["trieste"] = &game.Unit{
		Type:     game.Fleet,
		Owner:    game.Austria,
		Province: "trieste",
	}
	gameState.Board.Units["serbia"] = &game.Unit{
		Type:     game.Army,
		Owner:    game.Austria,
		Province: "serbia",
	}
	gameState.Board.Units["albania"] = &game.Unit{
		Type:     game.Army,
		Owner:    game.Turkey,
		Province: "albania",
	}

	// First turn - create a dislodgement situation
	gameState.RawOrders[game.Austria] = []string{
		"f trieste h",
		"a serbia h",
	}
	gameState.RawOrders[game.Turkey] = []string{
		"a albania - trieste",
	}

	// Process turn to create retreat situation
	result := ProcessDATCTest(t, gameState)

	// Expected: Support orders in retreat phase should be invalid
	ValidateExpectedOutcome(t, result, "Supports are not allowed in the retreat phase.", "6.H.1")
}

// Test 6.H.2: NO SUPPORTS FROM RETREATING UNIT
// Even a retreating unit cannot give support.
func TestDATCH2_NoSupportsFromRetreatingUnit(t *testing.T) {
	gameState := CreateDATCGameState(t)

	// Set up units for retreat scenario
	gameState.Board.Units["trieste"] = &game.Unit{
		Type:     game.Fleet,
		Owner:    game.Austria,
		Province: "trieste",
	}
	gameState.Board.Units["serbia"] = &game.Unit{
		Type:     game.Army,
		Owner:    game.Austria,
		Province: "serbia",
	}
	gameState.Board.Units["albania"] = &game.Unit{
		Type:     game.Army,
		Owner:    game.Turkey,
		Province: "albania",
	}

	// Orders that would create retreat situation
	gameState.RawOrders[game.Austria] = []string{
		"f trieste h",
		"a serbia h",
	}
	gameState.RawOrders[game.Turkey] = []string{
		"a albania - trieste",
	}

	// Process turn
	result := ProcessDATCTest(t, gameState)

	// Expected: Retreating unit cannot give support
	ValidateExpectedOutcome(t, result, "Even a retreating unit cannot give support.", "6.H.2")
}

// Test 6.H.3: NO CONVOY DURING RETREAT
// Convoys are not allowed during retreat.
func TestDATCH3_NoConvoyDuringRetreat(t *testing.T) {
	gameState := CreateDATCGameState(t)

	// Set up units for convoy retreat scenario
	gameState.Board.Units["trieste"] = &game.Unit{
		Type:     game.Army,
		Owner:    game.Austria,
		Province: "trieste",
	}
	gameState.Board.Units["adriatic_sea"] = &game.Unit{
		Type:     game.Fleet,
		Owner:    game.Austria,
		Province: "adriatic_sea",
	}
	gameState.Board.Units["albania"] = &game.Unit{
		Type:     game.Army,
		Owner:    game.Turkey,
		Province: "albania",
	}

	// Orders to create retreat situation
	gameState.RawOrders[game.Austria] = []string{
		"a trieste h",
		"f adriatic_sea h",
	}
	gameState.RawOrders[game.Turkey] = []string{
		"a albania - trieste",
	}

	// Process turn
	result := ProcessDATCTest(t, gameState)

	// Expected: Convoy orders not allowed in retreat phase
	ValidateExpectedOutcome(t, result, "Convoys are not allowed during retreat.", "6.H.3")
}

// Test 6.H.4: UNIT MAY NOT RETREAT TO THE AREA FROM WHICH IT IS ATTACKED
// A unit may not retreat to the area from which it is attacked.
func TestDATCH4_UnitMayNotRetreatToAttackingArea(t *testing.T) {
	gameState := CreateDATCGameState(t)

	// Set up units
	gameState.Board.Units["trieste"] = &game.Unit{
		Type:     game.Fleet,
		Owner:    game.Austria,
		Province: "trieste",
	}
	gameState.Board.Units["albania"] = &game.Unit{
		Type:     game.Army,
		Owner:    game.Turkey,
		Province: "albania",
	}

	// Orders to create attack from Albania to Trieste
	gameState.RawOrders[game.Austria] = []string{
		"f trieste h",
	}
	gameState.RawOrders[game.Turkey] = []string{
		"a albania - trieste",
	}

	// Process turn
	result := ProcessDATCTest(t, gameState)

	// Expected: Unit cannot retreat to attacking area
	ValidateExpectedOutcome(t, result, "A unit may not retreat to the area from which it is attacked.", "6.H.4")
}

// Test 6.H.5: UNIT MAY NOT RETREAT TO A CONTESTED AREA
// A unit may not retreat to a contested area.
func TestDATCH5_UnitMayNotRetreatToContestedArea(t *testing.T) {
	gameState := CreateDATCGameState(t)

	// Set up units for contested area scenario
	gameState.Board.Units["trieste"] = &game.Unit{
		Type:     game.Fleet,
		Owner:    game.Austria,
		Province: "trieste",
	}
	gameState.Board.Units["albania"] = &game.Unit{
		Type:     game.Army,
		Owner:    game.Turkey,
		Province: "albania",
	}
	gameState.Board.Units["serbia"] = &game.Unit{
		Type:     game.Army,
		Owner:    game.Russia,
		Province: "serbia",
	}
	gameState.Board.Units["budapest"] = &game.Unit{
		Type:     game.Army,
		Owner:    game.Austria,
		Province: "budapest",
	}

	// Orders to create contested area
	gameState.RawOrders[game.Austria] = []string{
		"f trieste h",
		"a budapest - serbia",
	}
	gameState.RawOrders[game.Turkey] = []string{
		"a albania - trieste",
	}
	gameState.RawOrders[game.Russia] = []string{
		"a serbia - budapest",
	}

	// Process turn
	result := ProcessDATCTest(t, gameState)

	// Expected: Unit cannot retreat to contested area
	ValidateExpectedOutcome(t, result, "A unit may not retreat to a contested area.", "6.H.5")
}

// Test 6.H.6: UNIT MAY NOT RETREAT TO A CONTESTED AREA
// A unit may not retreat to a contested area.
func TestDATCH6_UnitMayNotRetreatToContestedArea(t *testing.T) {
	gameState := CreateDATCGameState(t)

	// Set up units for contested area scenario
	gameState.Board.Units["trieste"] = &game.Unit{
		Type:     game.Fleet,
		Owner:    game.Austria,
		Province: "trieste",
	}
	gameState.Board.Units["albania"] = &game.Unit{
		Type:     game.Army,
		Owner:    game.Turkey,
		Province: "albania",
	}
	gameState.Board.Units["serbia"] = &game.Unit{
		Type:     game.Army,
		Owner:    game.Russia,
		Province: "serbia",
	}
	gameState.Board.Units["budapest"] = &game.Unit{
		Type:     game.Army,
		Owner:    game.Austria,
		Province: "budapest",
	}

	// Orders to create contested area
	gameState.RawOrders[game.Austria] = []string{
		"f trieste h",
		"a budapest - serbia",
	}
	gameState.RawOrders[game.Turkey] = []string{
		"a albania - trieste",
	}
	gameState.RawOrders[game.Russia] = []string{
		"a serbia - budapest",
	}

	// Process turn
	result := ProcessDATCTest(t, gameState)

	// Expected: Unit cannot retreat to contested area
	ValidateExpectedOutcome(t, result, "Unit may not retreat to a contested area.", "6.H.6")
}

// Test 6.H.7: MULTIPLE RETREAT TO SAME AREA WILL DISBAND UNITS
// Multiple units retreating to the same area will disband.
func TestDATCH7_MultipleRetreatToSameAreaWillDisbandUnits(t *testing.T) {
	gameState := CreateDATCGameState(t)

	// Set up units for multiple retreat scenario
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
	gameState.Board.Units["albania"] = &game.Unit{
		Type:     game.Army,
		Owner:    game.Turkey,
		Province: "albania",
	}
	gameState.Board.Units["serbia"] = &game.Unit{
		Type:     game.Army,
		Owner:    game.Russia,
		Province: "serbia",
	}

	// Orders to create multiple retreats to same area
	gameState.RawOrders[game.Austria] = []string{
		"f trieste h",
		"a vienna h",
	}
	gameState.RawOrders[game.Turkey] = []string{
		"a albania - trieste",
	}
	gameState.RawOrders[game.Russia] = []string{
		"a serbia - vienna",
	}

	// Process turn
	result := ProcessDATCTest(t, gameState)

	// Expected: Multiple retreats to same area cause disbanding
	ValidateExpectedOutcome(t, result, "Multiple retreat to same area will disband units.", "6.H.7")
}

// Test 6.H.8: TRIPLE RETREAT TO SAME AREA WILL DISBAND UNITS
// Triple units retreating to the same area will disband.
func TestDATCH8_TripleRetreatToSameAreaWillDisbandUnits(t *testing.T) {
	gameState := CreateDATCGameState(t)

	// Set up units for triple retreat scenario
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
	gameState.Board.Units["budapest"] = &game.Unit{
		Type:     game.Army,
		Owner:    game.Austria,
		Province: "budapest",
	}
	gameState.Board.Units["albania"] = &game.Unit{
		Type:     game.Army,
		Owner:    game.Turkey,
		Province: "albania",
	}
	gameState.Board.Units["serbia"] = &game.Unit{
		Type:     game.Army,
		Owner:    game.Russia,
		Province: "serbia",
	}
	gameState.Board.Units["rumania"] = &game.Unit{
		Type:     game.Army,
		Owner:    game.Turkey,
		Province: "rumania",
	}

	// Orders to create triple retreats to same area
	gameState.RawOrders[game.Austria] = []string{
		"f trieste h",
		"a vienna h",
		"a budapest h",
	}
	gameState.RawOrders[game.Turkey] = []string{
		"a albania - trieste",
		"a rumania - budapest",
	}
	gameState.RawOrders[game.Russia] = []string{
		"a serbia - vienna",
	}

	// Process turn
	result := ProcessDATCTest(t, gameState)

	// Expected: Triple retreats to same area cause disbanding
	ValidateExpectedOutcome(t, result, "Triple retreat to same area will disband units.", "6.H.8")
}

// Test 6.H.9: DISLODGED UNIT WILL NOT MAKE ATTACKERS AREA CONTESTED
// A dislodged unit will not make the attacker's area contested.
func TestDATCH9_DislodgedUnitWillNotMakeAttackersAreaContested(t *testing.T) {
	gameState := CreateDATCGameState(t)

	// Set up units
	gameState.Board.Units["trieste"] = &game.Unit{
		Type:     game.Fleet,
		Owner:    game.Austria,
		Province: "trieste",
	}
	gameState.Board.Units["albania"] = &game.Unit{
		Type:     game.Army,
		Owner:    game.Turkey,
		Province: "albania",
	}
	gameState.Board.Units["serbia"] = &game.Unit{
		Type:     game.Army,
		Owner:    game.Russia,
		Province: "serbia",
	}

	// Orders
	gameState.RawOrders[game.Austria] = []string{
		"f trieste h",
	}
	gameState.RawOrders[game.Turkey] = []string{
		"a albania - trieste",
	}
	gameState.RawOrders[game.Russia] = []string{
		"a serbia - albania",
	}

	// Process turn
	result := ProcessDATCTest(t, gameState)

	// Expected: Dislodged unit does not make attacker's area contested
	ValidateExpectedOutcome(t, result, "Dislodged unit will not make attackers area contested.", "6.H.9")
}

// Test 6.H.10: NOT RETREATING TO ATTACKER DOES NOT MEAN CONTESTED
// Not retreating to attacker does not mean contested.
func TestDATCH10_NotRetreatingToAttackerDoesNotMeanContested(t *testing.T) {
	gameState := CreateDATCGameState(t)

	// Set up units
	gameState.Board.Units["trieste"] = &game.Unit{
		Type:     game.Fleet,
		Owner:    game.Austria,
		Province: "trieste",
	}
	gameState.Board.Units["albania"] = &game.Unit{
		Type:     game.Army,
		Owner:    game.Turkey,
		Province: "albania",
	}

	// Orders
	gameState.RawOrders[game.Austria] = []string{
		"f trieste h",
	}
	gameState.RawOrders[game.Turkey] = []string{
		"a albania - trieste",
	}

	// Process turn
	result := ProcessDATCTest(t, gameState)

	// Expected: Not retreating to attacker does not mean contested
	ValidateExpectedOutcome(t, result, "Not retreating to attacker does not mean contested.", "6.H.10")
}

// Test 6.H.11: RETREAT WHEN DISLODGED BY ADJACENT CONVOY
// Retreat when dislodged by adjacent convoy.
func TestDATCH11_RetreatWhenDislodgedByAdjacentConvoy(t *testing.T) {
	gameState := CreateDATCGameState(t)

	// Set up units for adjacent convoy retreat scenario
	gameState.Board.Units["trieste"] = &game.Unit{
		Type:     game.Army,
		Owner:    game.Austria,
		Province: "trieste",
	}
	gameState.Board.Units["adriatic_sea"] = &game.Unit{
		Type:     game.Fleet,
		Owner:    game.Italy,
		Province: "adriatic_sea",
	}
	gameState.Board.Units["albania"] = &game.Unit{
		Type:     game.Army,
		Owner:    game.Italy,
		Province: "albania",
	}

	// Orders for adjacent convoy scenario
	gameState.RawOrders[game.Austria] = []string{
		"a trieste h",
	}
	gameState.RawOrders[game.Italy] = []string{
		"f adriatic_sea c albania - trieste",
		"a albania - trieste",
	}

	// Process turn
	result := ProcessDATCTest(t, gameState)

	// Expected: Retreat when dislodged by adjacent convoy
	ValidateExpectedOutcome(t, result, "Retreat when dislodged by adjacent convoy.", "6.H.11")
}

// Test 6.H.12: RETREAT WHEN DISLODGED BY ADJACENT CONVOY WHILE TRYING TO DO THE SAME
// Retreat when dislodged by adjacent convoy while trying to do the same.
func TestDATCH12_RetreatWhenDislodgedByAdjacentConvoyWhileTryingToDoTheSame(t *testing.T) {
	gameState := CreateDATCGameState(t)

	// Set up units for complex adjacent convoy scenario
	gameState.Board.Units["trieste"] = &game.Unit{
		Type:     game.Army,
		Owner:    game.Austria,
		Province: "trieste",
	}
	gameState.Board.Units["albania"] = &game.Unit{
		Type:     game.Army,
		Owner:    game.Italy,
		Province: "albania",
	}
	gameState.Board.Units["adriatic_sea"] = &game.Unit{
		Type:     game.Fleet,
		Owner:    game.Italy,
		Province: "adriatic_sea",
	}
	gameState.Board.Units["ionian_sea"] = &game.Unit{
		Type:     game.Fleet,
		Owner:    game.Austria,
		Province: "ionian_sea",
	}

	// Orders for complex adjacent convoy scenario
	gameState.RawOrders[game.Austria] = []string{
		"a trieste - albania",
		"f ionian_sea c trieste - albania",
	}
	gameState.RawOrders[game.Italy] = []string{
		"a albania - trieste",
		"f adriatic_sea c albania - trieste",
	}

	// Process turn
	result := ProcessDATCTest(t, gameState)

	// Expected: Retreat when dislodged by adjacent convoy while trying to do the same
	ValidateExpectedOutcome(t, result, "Retreat when dislodged by adjacent convoy while trying to do the same.", "6.H.12")
}

// Test 6.H.13: NO RETREAT WITH CONVOY IN MOVEMENT PHASE
// No retreat with convoy in movement phase.
func TestDATCH13_NoRetreatWithConvoyInMovementPhase(t *testing.T) {
	gameState := CreateDATCGameState(t)

	// Set up units
	gameState.Board.Units["trieste"] = &game.Unit{
		Type:     game.Army,
		Owner:    game.Austria,
		Province: "trieste",
	}
	gameState.Board.Units["adriatic_sea"] = &game.Unit{
		Type:     game.Fleet,
		Owner:    game.Austria,
		Province: "adriatic_sea",
	}
	gameState.Board.Units["albania"] = &game.Unit{
		Type:     game.Army,
		Owner:    game.Turkey,
		Province: "albania",
	}

	// Orders
	gameState.RawOrders[game.Austria] = []string{
		"a trieste h",
		"f adriatic_sea h",
	}
	gameState.RawOrders[game.Turkey] = []string{
		"a albania - trieste",
	}

	// Process turn
	result := ProcessDATCTest(t, gameState)

	// Expected: No retreat with convoy in movement phase
	ValidateExpectedOutcome(t, result, "No retreat with convoy in movement phase.", "6.H.13")
}

// Test 6.H.14: NO RETREAT WITH SUPPORT IN MOVEMENT PHASE
// No retreat with support in movement phase.
func TestDATCH14_NoRetreatWithSupportInMovementPhase(t *testing.T) {
	gameState := CreateDATCGameState(t)

	// Set up units
	gameState.Board.Units["trieste"] = &game.Unit{
		Type:     game.Fleet,
		Owner:    game.Austria,
		Province: "trieste",
	}
	gameState.Board.Units["serbia"] = &game.Unit{
		Type:     game.Army,
		Owner:    game.Austria,
		Province: "serbia",
	}
	gameState.Board.Units["albania"] = &game.Unit{
		Type:     game.Army,
		Owner:    game.Turkey,
		Province: "albania",
	}

	// Orders
	gameState.RawOrders[game.Austria] = []string{
		"f trieste h",
		"a serbia h",
	}
	gameState.RawOrders[game.Turkey] = []string{
		"a albania - trieste",
	}

	// Process turn
	result := ProcessDATCTest(t, gameState)

	// Expected: No retreat with support in movement phase
	ValidateExpectedOutcome(t, result, "No retreat with support in movement phase.", "6.H.14")
}

// Test 6.H.15: NO COASTAL CRAWL IN RETREAT
// No coastal crawl in retreat.
func TestDATCH15_NoCoastalCrawlInRetreat(t *testing.T) {
	gameState := CreateDATCGameState(t)

	// Set up units for coastal crawl scenario
	gameState.Board.Units["spain_north_coast"] = &game.Unit{
		Type:     game.Fleet,
		Owner:    game.France,
		Province: "spain_north_coast",
	}
	gameState.Board.Units["gascony"] = &game.Unit{
		Type:     game.Army,
		Owner:    game.Germany,
		Province: "gascony",
	}

	// Orders
	gameState.RawOrders[game.France] = []string{
		"f spain_north_coast h",
	}
	gameState.RawOrders[game.Germany] = []string{
		"a gascony - spain",
	}

	// Process turn
	result := ProcessDATCTest(t, gameState)

	// Expected: No coastal crawl in retreat
	ValidateExpectedOutcome(t, result, "No coastal crawl in retreat.", "6.H.15")
}

// Test 6.H.16: CONTESTED FOR BOTH COASTS
// Contested for both coasts.
func TestDATCH16_ContestedForBothCoasts(t *testing.T) {
	gameState := CreateDATCGameState(t)

	// Set up units for both coasts scenario
	gameState.Board.Units["spain_north_coast"] = &game.Unit{
		Type:     game.Fleet,
		Owner:    game.France,
		Province: "spain_north_coast",
	}
	gameState.Board.Units["spain_south_coast"] = &game.Unit{
		Type:     game.Fleet,
		Owner:    game.Italy,
		Province: "spain_south_coast",
	}
	gameState.Board.Units["gascony"] = &game.Unit{
		Type:     game.Army,
		Owner:    game.Germany,
		Province: "gascony",
	}
	gameState.Board.Units["western_mediterranean"] = &game.Unit{
		Type:     game.Fleet,
		Owner:    game.Germany,
		Province: "western_mediterranean",
	}

	// Orders
	gameState.RawOrders[game.France] = []string{
		"f spain_north_coast h",
	}
	gameState.RawOrders[game.Italy] = []string{
		"f spain_south_coast h",
	}
	gameState.RawOrders[game.Germany] = []string{
		"a gascony - spain",
		"f western_mediterranean - spain_south_coast",
	}

	// Process turn
	result := ProcessDATCTest(t, gameState)

	// Expected: Contested for both coasts
	ValidateExpectedOutcome(t, result, "Contested for both coasts.", "6.H.16")
}
