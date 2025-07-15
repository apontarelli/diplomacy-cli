package pipeline

import (
	"testing"

	"diplomacy-cli/backend/internal/game"
)

// Test 6.I.1: TOO MANY BUILD ORDERS
// Too many build orders.
func TestDATCI1_TooManyBuildOrders(t *testing.T) {
	gameState := CreateDATCGameState(t)

	// Set phase to WinterBuild for build orders
	gameState.Phase = game.WinterBuild

	// Set up a scenario where a country has more build orders than allowed
	// Remove some units to create build opportunities
	delete(gameState.Board.Units, "london")
	delete(gameState.Board.Units, "liverpool")

	// Add build orders (more than allowed)
	gameState.RawOrders[game.England] = []string{
		"build a london",
		"build f liverpool",
		"build a edinburgh", // This should be too many
	}

	// Process turn
	result := ProcessDATCTest(t, gameState)

	// Expected: Too many build orders should be rejected
	ValidateExpectedOutcome(t, result, "Too many build orders.", "6.I.1")
}

// Test 6.I.2: FLEETS CANNOT BE BUILD IN LAND AREAS
// Fleets cannot be built in land areas.
func TestDATCI2_FleetsCannotBeBuildInLandAreas(t *testing.T) {
	gameState := CreateDATCGameState(t)

	// Set phase to WinterBuild for build orders
	gameState.Phase = game.WinterBuild

	// Set up scenario where fleet is built in land area
	delete(gameState.Board.Units, "moscow")

	// Add invalid build order (fleet in land area)
	gameState.RawOrders[game.Russia] = []string{
		"build f moscow", // Invalid - Moscow is a land area
	}

	// Process turn
	result := ProcessDATCTest(t, gameState)

	// Expected: Fleet cannot be built in land areas
	ValidateExpectedOutcome(t, result, "Fleets cannot be build in land areas.", "6.I.2")
}

// Test 6.I.3: SUPPLY CENTER MUST BE EMPTY FOR BUILDING
// Supply center must be empty for building.
func TestDATCI3_SupplyCenterMustBeEmptyForBuilding(t *testing.T) {
	gameState := CreateDATCGameState(t)

	// Set phase to WinterBuild for build orders
	gameState.Phase = game.WinterBuild

	// London already has a unit, try to build there
	gameState.RawOrders[game.England] = []string{
		"build a london", // Invalid - London already occupied
	}

	// Process turn
	result := ProcessDATCTest(t, gameState)

	// Expected: Supply center must be empty for building
	ValidateExpectedOutcome(t, result, "Supply center must be empty for building.", "6.I.3")
}

// Test 6.I.4: BOTH COASTS MUST BE EMPTY FOR BUILDING
// Both coasts must be empty for building.
func TestDATCI4_BothCoastsMustBeEmptyForBuilding(t *testing.T) {
	gameState := CreateDATCGameState(t)

	// Set phase to WinterBuild for build orders
	gameState.Phase = game.WinterBuild

	// Set up scenario with one coast occupied
	gameState.Board.Units["spain_north_coast"] = &game.Unit{
		Type:     game.Fleet,
		Owner:    game.France,
		Province: "spain_north_coast",
	}

	// Try to build on the other coast
	gameState.RawOrders[game.France] = []string{
		"build f spain_south_coast", // Invalid - other coast occupied
	}

	// Process turn
	result := ProcessDATCTest(t, gameState)

	// Expected: Both coasts must be empty for building
	ValidateExpectedOutcome(t, result, "Both coasts must be empty for building.", "6.I.4")
}

// Test 6.I.5: BUILDING IN HOME SUPPLY CENTER THAT IS NOT OWNED
// Building in home supply center that is not owned.
func TestDATCI5_BuildingInHomeSupplyCenterThatIsNotOwned(t *testing.T) {
	gameState := CreateDATCGameState(t)

	// Set phase to WinterBuild for build orders
	gameState.Phase = game.WinterBuild

	// Set up scenario where home supply center is not owned
	// Assume London is captured by Germany but England tries to build there
	delete(gameState.Board.Units, "london")

	// Change ownership of London to Germany (this would need proper supply center logic)
	gameState.RawOrders[game.England] = []string{
		"build a london", // Invalid - London not owned by England
	}

	// Process turn
	result := ProcessDATCTest(t, gameState)

	// Expected: Building in home supply center that is not owned
	ValidateExpectedOutcome(t, result, "Building in home supply center that is not owned.", "6.I.5")
}

// Test 6.I.6: BUILDING IN OWNED SUPPLY CENTER THAT IS NOT A HOME SUPPLY CENTER
// Building in owned supply center that is not a home supply center.
func TestDATCI6_BuildingInOwnedSupplyCenterThatIsNotAHomeSupplyCenter(t *testing.T) {
	gameState := CreateDATCGameState(t)

	// Set phase to WinterBuild for build orders
	gameState.Phase = game.WinterBuild

	// Set up scenario where England owns a non-home supply center
	delete(gameState.Board.Units, "paris") // Remove French unit from Paris

	// England tries to build in Paris (not a home supply center for England)
	gameState.RawOrders[game.England] = []string{
		"build a paris", // Invalid - Paris is not English home supply center
	}

	// Process turn
	result := ProcessDATCTest(t, gameState)

	// Expected: Building in owned supply center that is not a home supply center
	ValidateExpectedOutcome(t, result, "Building in owned supply center that is not a home supply center.", "6.I.6")
}

// Test 6.I.7: ONLY ONE BUILD IN A HOME SUPPLY CENTER
// Only one build in a home supply center.
func TestDATCI7_OnlyOneBuildInAHomeSupplyCenter(t *testing.T) {
	gameState := CreateDATCGameState(t)

	// Set phase to WinterBuild for build orders
	gameState.Phase = game.WinterBuild

	// Set up scenario where multiple builds are attempted in same supply center
	delete(gameState.Board.Units, "london")

	// Try to build multiple units in same supply center
	gameState.RawOrders[game.England] = []string{
		"build a london",
		"build f london", // Invalid - second build in same center
	}

	// Process turn
	result := ProcessDATCTest(t, gameState)

	// Expected: Only one build in a home supply center
	ValidateExpectedOutcome(t, result, "Only one build in a home supply center.", "6.I.7")
}
