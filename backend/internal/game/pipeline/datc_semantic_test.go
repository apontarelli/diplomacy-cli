package pipeline

import (
	"diplomacy-cli/backend/internal/game"
	"testing"
)

// Test DATC 6.A.1: MOVING TO AN AREA THAT IS NOT A NEIGHBOUR
func TestDATCSemanticA1_MovingToNonNeighbour(t *testing.T) {
	gameState := CreateDATCGameState(t)

	// Add unit: F North Sea
	gameState.Board.Units["north_sea"] = &game.Unit{
		Type:     game.Fleet,
		Owner:    game.England,
		Province: "north_sea",
	}

	// Add order: F North Sea - Picardy (not adjacent)
	gameState.RawOrders[game.England] = []string{"F North Sea - Western Mediterranean"}

	// Process turn with new semantic validation
	processor := NewTurnProcessor()
	_, err := processor.ProcessTurn(gameState)

	// Should fail due to semantic validation
	if err == nil {
		t.Errorf("Expected semantic error for non-adjacent move, but got none")
	}

	if err != nil && !containsError(err.Error(), "not adjacent") {
		t.Errorf("Expected adjacency error, got: %v", err)
	}
}

// Test DATC 6.A.2: MOVE ARMY TO SEA
func TestDATCSemanticA2_MoveArmyToSea(t *testing.T) {
	gameState := CreateDATCGameState(t)

	// Add unit: A Liverpool
	gameState.Board.Units["liverpool"] = &game.Unit{
		Type:     game.Army,
		Owner:    game.England,
		Province: "liverpool",
	}

	// Add order: A Liverpool - Irish Sea
	gameState.RawOrders[game.England] = []string{"A Liverpool - Irish Sea"}

	// Process turn with new semantic validation
	processor := NewTurnProcessor()
	_, err := processor.ProcessTurn(gameState)

	// Should fail due to semantic validation
	if err == nil {
		t.Errorf("Expected semantic error for army to sea, but got none")
	}

	if err != nil && !containsError(err.Error(), "cannot move to sea") {
		t.Errorf("Expected invalid destination error, got: %v", err)
	}
}

// Test DATC 6.A.3: MOVE FLEET TO LAND
func TestDATCSemanticA3_MoveFleetToLand(t *testing.T) {
	gameState := CreateDATCGameState(t)

	// Add unit: F Kiel
	gameState.Board.Units["kiel"] = &game.Unit{
		Type:     game.Fleet,
		Owner:    game.Germany,
		Province: "kiel",
	}

	// Add order: F Kiel - Munich
	gameState.RawOrders[game.Germany] = []string{"F Kiel - Munich"}

	// Process turn with new semantic validation
	processor := NewTurnProcessor()
	_, err := processor.ProcessTurn(gameState)

	// Should fail due to semantic validation
	if err == nil {
		t.Errorf("Expected semantic error for fleet to land, but got none")
	}

	if err != nil && !containsError(err.Error(), "cannot move to land") {
		t.Errorf("Expected invalid destination error, got: %v", err)
	}
}

// Test DATC 6.A.4: MOVE TO OWN SECTOR
func TestDATCSemanticA4_MoveToOwnSector(t *testing.T) {
	gameState := CreateDATCGameState(t)

	// Add unit: F Kiel
	gameState.Board.Units["kiel"] = &game.Unit{
		Type:     game.Fleet,
		Owner:    game.Germany,
		Province: "kiel",
	}

	// Add order: F Kiel - Kiel
	gameState.RawOrders[game.Germany] = []string{"F Kiel - Kiel"}

	// Process turn with new semantic validation
	processor := NewTurnProcessor()
	_, err := processor.ProcessTurn(gameState)

	// Should fail due to semantic validation
	if err == nil {
		t.Errorf("Expected semantic error for move to same location, but got none")
	}

	if err != nil && !containsError(err.Error(), "cannot move to same location") {
		t.Errorf("Expected same location error, got: %v", err)
	}
}

// Test DATC 6.A.6: ORDERING A UNIT OF ANOTHER COUNTRY
func TestDATCSemanticA6_OrderingUnitOfAnotherCountry(t *testing.T) {
	gameState := CreateDATCGameState(t)

	// Add unit: F London (owned by England, not Germany)
	gameState.Board.Units["london"] = &game.Unit{
		Type:     game.Fleet,
		Owner:    game.England,
		Province: "london",
	}

	// Add order: Germany tries to order England's fleet
	gameState.RawOrders[game.Germany] = []string{"F London - North Sea"}

	// Process turn with new semantic validation
	processor := NewTurnProcessor()
	_, err := processor.ProcessTurn(gameState)

	// Should fail due to ownership validation
	if err == nil {
		t.Errorf("Expected ownership error, but got none")
	}

	if err != nil && !containsError(err.Error(), "belongs to") {
		t.Errorf("Expected ownership error, got: %v", err)
	}
}

// Test DATC 6.A.8: SUPPORT TO HOLD YOURSELF IS NOT POSSIBLE
func TestDATCSemanticA8_SupportToHoldYourselfNotPossible(t *testing.T) {
	gameState := CreateDATCGameState(t)

	// Add units
	gameState.Board.Units["trieste"] = &game.Unit{
		Type:     game.Fleet,
		Owner:    game.Austria,
		Province: "trieste",
	}

	// Add orders: Self-support (should be invalid)
	gameState.RawOrders[game.Austria] = []string{
		"F Trieste S Trieste", // Self-support
	}

	// Process turn with new semantic validation
	processor := NewTurnProcessor()
	_, err := processor.ProcessTurn(gameState)

	// Should fail due to semantic validation
	if err == nil {
		t.Errorf("Expected semantic error for self-support, but got none")
	}

	if err != nil && !containsError(err.Error(), "cannot reach destination") {
		t.Errorf("Expected self-support error, got: %v", err)
	}
}

// Test valid orders pass through
func TestDATCSemanticValidOrders(t *testing.T) {
	gameState := CreateDATCGameState(t)

	// Add units for valid orders
	gameState.Board.Units["paris"] = &game.Unit{
		Type:     game.Army,
		Owner:    game.France,
		Province: "paris",
	}
	gameState.Board.Units["burgundy"] = &game.Unit{
		Type:     game.Army,
		Owner:    game.France,
		Province: "burgundy",
	}
	gameState.Board.Units["london"] = &game.Unit{
		Type:     game.Fleet,
		Owner:    game.England,
		Province: "london",
	}

	// Add valid orders
	gameState.RawOrders[game.France] = []string{
		"A Paris - Burgundy", // Valid move (but will bounce)
		"A Burgundy S Paris", // Valid support hold
	}
	gameState.RawOrders[game.England] = []string{
		"F London Hold", // Valid hold
	}

	// Process turn with new semantic validation
	processor := NewTurnProcessor()
	newState, err := processor.ProcessTurn(gameState)

	// Should succeed (semantic validation passes, resolution may have bounces)
	if err != nil {
		t.Errorf("Expected valid orders to pass semantic validation, got error: %v", err)
	}

	if newState == nil {
		t.Errorf("Expected new game state, got nil")
	}
}

// Helper function to check if error contains substring
func containsError(errorStr, substring string) bool {
	return len(errorStr) > 0 && len(substring) > 0 &&
		(errorStr == substring ||
			len(errorStr) >= len(substring) &&
				errorStr[len(errorStr)-len(substring):] == substring ||
			findSubstring(errorStr, substring))
}

func findSubstring(str, substr string) bool {
	for i := 0; i <= len(str)-len(substr); i++ {
		if str[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
