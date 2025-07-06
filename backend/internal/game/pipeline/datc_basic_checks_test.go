package pipeline

import (
	"testing"

	"diplomacy-cli/backend/internal/game"
)

// Test 6.A.1: MOVING TO AN AREA THAT IS NOT A NEIGHBOUR
func TestDATCA1_MovingToNonNeighbour(t *testing.T) {
	gameState := CreateDATCGameState(t)

	// Add unit: F North Sea (inferred from order)
	gameState.Board.Units["North Sea"] = &game.Unit{
		Type:     game.Fleet,
		Owner:    game.England,
		Province: "North Sea",
	}

	// Add order: F North Sea - Picardy
	gameState.RawOrders[game.England] = []string{"F North Sea - Picardy"}

	// Process turn
	result := ProcessDATCTest(t, gameState)

	// Validate: Order should fail
	ValidateExpectedOutcome(t, result, "Order should fail.", "6.A.1")
}

// Test 6.A.2: MOVE ARMY TO SEA
func TestDATCA2_MoveArmyToSea(t *testing.T) {
	gameState := CreateDATCGameState(t)

	// Add unit: A Liverpool
	gameState.Board.Units["liverpool"] = &game.Unit{
		Type:     game.Army,
		Owner:    game.England,
		Province: "liverpool",
	}

	// Add order: A Liverpool - Irish Sea
	gameState.RawOrders[game.England] = []string{"A Liverpool - Irish Sea"}

	// Process turn
	result := ProcessDATCTest(t, gameState)

	// Validate: Order should fail
	ValidateExpectedOutcome(t, result, "Order should fail.", "6.A.2")
}

// Test 6.A.3: MOVE FLEET TO LAND
func TestDATCA3_MoveFleetToLand(t *testing.T) {
	gameState := CreateDATCGameState(t)

	// Add unit: F Kiel
	gameState.Board.Units["kiel"] = &game.Unit{
		Type:     game.Fleet,
		Owner:    game.Germany,
		Province: "kiel",
	}

	// Add order: F Kiel - Munich
	gameState.RawOrders[game.Germany] = []string{"F Kiel - Munich"}

	// Process turn
	result := ProcessDATCTest(t, gameState)

	// Validate: Order should fail
	ValidateExpectedOutcome(t, result, "Order should fail.", "6.A.3")
}

// Test 6.A.4: MOVE TO OWN SECTOR
func TestDATCA4_MoveToOwnSector(t *testing.T) {
	gameState := CreateDATCGameState(t)

	// Add unit: F Kiel
	gameState.Board.Units["kiel"] = &game.Unit{
		Type:     game.Fleet,
		Owner:    game.Germany,
		Province: "kiel",
	}

	// Add order: F Kiel - Kiel
	gameState.RawOrders[game.Germany] = []string{"F Kiel - Kiel"}

	// Process turn
	result := ProcessDATCTest(t, gameState)

	// Validate: Order should fail (moving to same sector is illegal)
	ValidateExpectedOutcome(t, result, "Order should fail.", "6.A.4")
}

// Test 6.A.5: MOVE TO OWN SECTOR WITH CONVOY
func TestDATCA5_MoveToOwnSectorWithConvoy(t *testing.T) {
	gameState := CreateDATCGameState(t)

	// Add units for convoy scenario
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
	gameState.Board.Units["liverpool"] = &game.Unit{
		Type:     game.Army,
		Owner:    game.England,
		Province: "liverpool",
	}
	gameState.Board.Units["london"] = &game.Unit{
		Type:     game.Fleet,
		Owner:    game.Germany,
		Province: "london",
	}
	gameState.Board.Units["wales"] = &game.Unit{
		Type:     game.Army,
		Owner:    game.Germany,
		Province: "wales",
	}

	// Add orders
	gameState.RawOrders[game.England] = []string{
		"f north_sea c york - york",
		"a york - york",
		"a liverpool s york",
	}
	gameState.RawOrders[game.Germany] = []string{
		"f london - york",
		"a wales s london - york",
	}

	// Process turn
	result := ProcessDATCTest(t, gameState)

	// Validate: Order should fail (moving to same sector is illegal even with convoy)
	ValidateExpectedOutcome(t, result, "Order should fail.", "6.A.5")
}

// Test 6.A.6: ORDERING A UNIT OF ANOTHER COUNTRY
func TestDATCA6_OrderingUnitOfAnotherCountry(t *testing.T) {
	gameState := CreateDATCGameState(t)

	// Add unit: F London (owned by England, not Germany)
	gameState.Board.Units["london"] = &game.Unit{
		Type:     game.Fleet,
		Owner:    game.England,
		Province: "london",
	}

	// Add order: Germany tries to order England's fleet
	gameState.RawOrders[game.Germany] = []string{"F London - North Sea"}

	// Process turn
	result := ProcessDATCTest(t, gameState)

	// Validate: Order should fail
	ValidateExpectedOutcome(t, result, "Order should fail.", "6.A.6")
}

// Test 6.A.7: ONLY ARMIES CAN BE CONVOYED
func TestDATCA7_OnlyArmiesCanBeConvoyed(t *testing.T) {
	gameState := CreateDATCGameState(t)

	// Add units
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

	// Add orders: Try to convoy a fleet (should fail)
	gameState.RawOrders[game.England] = []string{
		"f london - belgium",
		"f north_sea c london - belgium",
	}

	// Process turn
	result := ProcessDATCTest(t, gameState)

	// Validate: Move from London to Belgium should fail
	ValidateExpectedOutcome(t, result, "Move from London to Belgium should fail.", "6.A.7")
}

// Test 6.A.8: SUPPORT TO HOLD YOURSELF IS NOT POSSIBLE
func TestDATCA8_SupportToHoldYourselfNotPossible(t *testing.T) {
	gameState := CreateDATCGameState(t)

	// Add units
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
	gameState.Board.Units["trieste"] = &game.Unit{
		Type:     game.Fleet,
		Owner:    game.Austria,
		Province: "trieste",
	}

	// Add orders
	gameState.RawOrders[game.Italy] = []string{
		"a venice - trieste",
		"a tyrolia s venice - trieste",
	}
	gameState.RawOrders[game.Austria] = []string{
		"f trieste s trieste", // Self-support (should be invalid)
	}

	// Process turn
	result := ProcessDATCTest(t, gameState)

	// Validate: The army in Trieste should be dislodged
	ValidateExpectedOutcome(t, result, "The army in Trieste should be dislodged.", "6.A.8")
}

// Test 6.A.9: FLEETS MUST FOLLOW COAST IF NOT ON SEA
func TestDATCA9_FleetsMustFollowCoast(t *testing.T) {
	gameState := CreateDATCGameState(t)

	// Add unit: F Rome
	gameState.Board.Units["rome"] = &game.Unit{
		Type:     game.Fleet,
		Owner:    game.Italy,
		Province: "rome",
	}

	// Add order: F Rome - Venice (should fail - fleet cannot move from Rome to Venice)
	gameState.RawOrders[game.Italy] = []string{"f rome - venice"}

	// Process turn
	result := ProcessDATCTest(t, gameState)

	// Validate: Move fails
	ValidateExpectedOutcome(t, result, "Move fails. An army can go from Rome to Venice, but a fleet cannot.", "6.A.9")
}

// Test 6.A.10: SUPPORT ON UNREACHABLE DESTINATION NOT POSSIBLE
func TestDATCA10_SupportOnUnreachableDestination(t *testing.T) {
	gameState := CreateDATCGameState(t)

	// Add units
	gameState.Board.Units["venice"] = &game.Unit{
		Type:     game.Army,
		Owner:    game.Austria,
		Province: "venice",
	}
	gameState.Board.Units["rome"] = &game.Unit{
		Type:     game.Fleet,
		Owner:    game.Italy,
		Province: "rome",
	}
	gameState.Board.Units["apulia"] = &game.Unit{
		Type:     game.Army,
		Owner:    game.Italy,
		Province: "apulia",
	}

	// Add orders
	gameState.RawOrders[game.Austria] = []string{
		"a venice hold",
	}
	gameState.RawOrders[game.Italy] = []string{
		"f rome s apulia - venice", // Fleet cannot reach Venice, so support is invalid
		"a apulia - venice",
	}

	// Process turn
	result := ProcessDATCTest(t, gameState)

	// Validate: Support is illegal, Venice is not dislodged
	ValidateExpectedOutcome(t, result, "The support of Rome is illegal, because Venice cannot be reached from Rome by a fleet. Venice is not dislodged.", "6.A.10")
}

// Test 6.A.11: SIMPLE BOUNCE
func TestDATCA11_SimpleBounce(t *testing.T) {
	gameState := CreateDATCGameState(t)

	// Add units
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

	// Add orders: Both armies move to Tyrolia (should bounce)
	gameState.RawOrders[game.Austria] = []string{
		"a vienna - tyrolia",
	}
	gameState.RawOrders[game.Italy] = []string{
		"a venice - tyrolia",
	}

	// Process turn
	result := ProcessDATCTest(t, gameState)

	// Validate: Both armies bounce and stay in place
	ValidateExpectedOutcome(t, result, "No outcome specified", "6.A.11")
}

// Test 6.A.12: BOUNCE OF THREE UNITS
func TestDATCA12_BounceOfThreeUnits(t *testing.T) {
	gameState := CreateDATCGameState(t)

	// Add units
	gameState.Board.Units["vienna"] = &game.Unit{
		Type:     game.Army,
		Owner:    game.Austria,
		Province: "vienna",
	}
	gameState.Board.Units["munich"] = &game.Unit{
		Type:     game.Army,
		Owner:    game.Germany,
		Province: "munich",
	}
	gameState.Board.Units["venice"] = &game.Unit{
		Type:     game.Army,
		Owner:    game.Italy,
		Province: "venice",
	}

	// Add orders: All three armies move to Tyrolia (should all bounce)
	gameState.RawOrders[game.Austria] = []string{
		"a vienna - tyrolia",
	}
	gameState.RawOrders[game.Germany] = []string{
		"a munich - tyrolia",
	}
	gameState.RawOrders[game.Italy] = []string{
		"a venice - tyrolia",
	}

	// Process turn
	result := ProcessDATCTest(t, gameState)

	// Validate: All three armies bounce and stay in place
	ValidateExpectedOutcome(t, result, "No outcome specified", "6.A.12")
}
