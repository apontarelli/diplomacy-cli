package game

import (
	"testing"
	"time"
)

func TestNewGameState(t *testing.T) {
	board := NewBoard()
	state := NewGameState(board, SpringMovement, 1901)

	if state == nil {
		t.Fatal("NewGameState() returned nil")
	}

	if state.Board != board {
		t.Error("Board not set correctly")
	}

	if state.Phase != SpringMovement {
		t.Errorf("Expected phase %s, got %s", SpringMovement, state.Phase)
	}

	if state.Year != 1901 {
		t.Errorf("Expected year 1901, got %d", state.Year)
	}

	if state.RawOrders == nil {
		t.Error("RawOrders map not initialized")
	}

	if state.SupplyCenters == nil {
		t.Error("SupplyCenters map not initialized")
	}

	if len(state.RawOrders) != 0 {
		t.Error("Expected empty RawOrders map")
	}

	if state.CreatedAt.IsZero() {
		t.Error("CreatedAt not set")
	}
}

func TestNewGame(t *testing.T) {
	game := NewGame("test-id", "Test Game")

	if game == nil {
		t.Fatal("NewGame() returned nil")
	}

	if game.ID != "test-id" {
		t.Errorf("Expected ID 'test-id', got '%s'", game.ID)
	}

	if game.Name != "Test Game" {
		t.Errorf("Expected name 'Test Game', got '%s'", game.Name)
	}

	if game.Players == nil {
		t.Error("Players map not initialized")
	}

	if game.History == nil {
		t.Error("History slice not initialized")
	}

	if game.Status != WaitingForPlayers {
		t.Errorf("Expected status %s, got %s", WaitingForPlayers, game.Status)
	}

	if len(game.Players) != 0 {
		t.Error("Expected empty players map")
	}

	if len(game.History) != 0 {
		t.Error("Expected empty history")
	}
}

func TestAddRawOrder(t *testing.T) {
	board := NewBoard()
	state := NewGameState(board, SpringMovement, 1901)

	err := state.AddRawOrder(France, "A Par-Bur")
	if err != nil {
		t.Errorf("AddRawOrder() failed: %v", err)
	}

	orders := state.GetRawOrdersForNation(France)
	if len(orders) != 1 {
		t.Errorf("Expected 1 order, got %d", len(orders))
	}

	if orders[0] != "A Par-Bur" {
		t.Errorf("Expected 'A Par-Bur', got '%s'", orders[0])
	}

	err = state.AddRawOrder(France, "F Bre-MAO")
	if err != nil {
		t.Errorf("AddRawOrder() failed: %v", err)
	}

	orders = state.GetRawOrdersForNation(France)
	if len(orders) != 2 {
		t.Errorf("Expected 2 orders, got %d", len(orders))
	}

	err = state.AddRawOrder(Germany, "A Ber-Kie")
	if err != nil {
		t.Errorf("AddRawOrder() failed: %v", err)
	}

	germanOrders := state.GetRawOrdersForNation(Germany)
	if len(germanOrders) != 1 {
		t.Errorf("Expected 1 German order, got %d", len(germanOrders))
	}

	frenchOrders := state.GetRawOrdersForNation(France)
	if len(frenchOrders) != 2 {
		t.Errorf("Expected 2 French orders, got %d", len(frenchOrders))
	}
}

func TestGetRawOrdersForNation_EmptyNation(t *testing.T) {
	board := NewBoard()
	state := NewGameState(board, SpringMovement, 1901)

	orders := state.GetRawOrdersForNation(Italy)
	if len(orders) != 0 {
		t.Errorf("Expected 0 orders, got %d", len(orders))
	}
}

func TestAddPlayer(t *testing.T) {
	game := NewGame("test-id", "Test Game")

	err := game.AddPlayer(France, "player1")
	if err != nil {
		t.Errorf("AddPlayer() failed: %v", err)
	}

	if game.Players[France] != "player1" {
		t.Errorf("Expected player1 for France, got '%s'", game.Players[France])
	}

	err = game.AddPlayer(Germany, "player2")
	if err != nil {
		t.Errorf("AddPlayer() failed: %v", err)
	}

	if len(game.Players) != 2 {
		t.Errorf("Expected 2 players, got %d", len(game.Players))
	}
}

func TestAddPlayer_DuplicateNation(t *testing.T) {
	game := NewGame("test-id", "Test Game")

	err := game.AddPlayer(France, "player1")
	if err != nil {
		t.Fatalf("First AddPlayer() failed: %v", err)
	}

	err = game.AddPlayer(France, "player2")
	if err == nil {
		t.Error("Expected error when adding duplicate nation")
	}

	if game.Players[France] != "player1" {
		t.Error("Original player should remain")
	}
}

func TestAddPlayer_WrongStatus(t *testing.T) {
	game := NewGame("test-id", "Test Game")
	game.Status = InProgress

	err := game.AddPlayer(France, "player1")
	if err == nil {
		t.Error("Expected error when adding player to in-progress game")
	}

	if len(game.Players) != 0 {
		t.Error("No players should be added")
	}
}

func TestIsPlayerInGame(t *testing.T) {
	game := NewGame("test-id", "Test Game")
	game.AddPlayer(France, "player1")
	game.AddPlayer(Germany, "player2")

	if !game.IsPlayerInGame("player1") {
		t.Error("player1 should be in game")
	}

	if !game.IsPlayerInGame("player2") {
		t.Error("player2 should be in game")
	}

	if game.IsPlayerInGame("player3") {
		t.Error("player3 should not be in game")
	}
}

func TestGetPlayerNation(t *testing.T) {
	game := NewGame("test-id", "Test Game")
	game.AddPlayer(France, "player1")
	game.AddPlayer(Germany, "player2")

	nation := game.GetPlayerNation("player1")
	if nation != France {
		t.Errorf("Expected France, got %s", nation)
	}

	nation = game.GetPlayerNation("player2")
	if nation != Germany {
		t.Errorf("Expected Germany, got %s", nation)
	}

	nation = game.GetPlayerNation("player3")
	if nation != "" {
		t.Errorf("Expected empty string, got %s", nation)
	}
}

func TestAdvanceToNextPhase(t *testing.T) {
	board := NewBoard()
	state := NewGameState(board, SpringMovement, 1901)

	state.AddRawOrder(France, "A Par-Bur")
	state.AddRawOrder(Germany, "A Ber-Kie")

	tests := []struct {
		name           string
		currentPhase   Phase
		currentYear    int
		dislodgedUnits []*Unit
		expectedPhase  Phase
		expectedYear   int
	}{
		{
			name:           "Spring Movement to Fall Movement (no dislodged)",
			currentPhase:   SpringMovement,
			currentYear:    1901,
			dislodgedUnits: []*Unit{},
			expectedPhase:  FallMovement,
			expectedYear:   1901,
		},
		{
			name:           "Spring Movement to Spring Retreat (with dislodged)",
			currentPhase:   SpringMovement,
			currentYear:    1901,
			dislodgedUnits: []*Unit{{Type: Army, Owner: France, Province: "par"}},
			expectedPhase:  SpringRetreat,
			expectedYear:   1901,
		},
		{
			name:           "Fall Movement to Winter Build (no dislodged)",
			currentPhase:   FallMovement,
			currentYear:    1901,
			dislodgedUnits: []*Unit{},
			expectedPhase:  WinterBuild,
			expectedYear:   1902,
		},
		{
			name:           "Fall Movement to Fall Retreat (with dislodged)",
			currentPhase:   FallMovement,
			currentYear:    1901,
			dislodgedUnits: []*Unit{{Type: Army, Owner: France, Province: "par"}},
			expectedPhase:  FallRetreat,
			expectedYear:   1901,
		},
		{
			name:           "Spring Retreat to Fall Movement",
			currentPhase:   SpringRetreat,
			currentYear:    1901,
			dislodgedUnits: []*Unit{},
			expectedPhase:  FallMovement,
			expectedYear:   1901,
		},
		{
			name:           "Fall Retreat to Winter Build",
			currentPhase:   FallRetreat,
			currentYear:    1901,
			dislodgedUnits: []*Unit{},
			expectedPhase:  WinterBuild,
			expectedYear:   1902,
		},
		{
			name:           "Winter Build to Spring Movement",
			currentPhase:   WinterBuild,
			currentYear:    1901,
			dislodgedUnits: []*Unit{},
			expectedPhase:  SpringMovement,
			expectedYear:   1901,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			state.Phase = tt.currentPhase
			state.Year = tt.currentYear
			state.AddRawOrder(France, "test order")

			state.AdvanceToNextPhase(tt.dislodgedUnits)

			if state.Phase != tt.expectedPhase {
				t.Errorf("Expected phase %s, got %s", tt.expectedPhase, state.Phase)
			}

			if state.Year != tt.expectedYear {
				t.Errorf("Expected year %d, got %d", tt.expectedYear, state.Year)
			}

			if len(state.RawOrders) != 0 {
				t.Error("RawOrders should be cleared after phase advance")
			}
		})
	}
}

func TestGameStateClone(t *testing.T) {
	board := NewBoard()

	province := &Province{
		Name:           "paris",
		ShortCode:      "par",
		DisplayName:    "Paris",
		Type:           Land,
		SupplyCenter:   true,
		CoastNeighbors: map[string][]string{"nc": {"english_channel"}},
		ArmyNeighbors:  []string{"burgundy", "picardy"},
		FleetNeighbors: []string{},
	}
	board.AddProvince(province)

	unit := &Unit{Type: Army, Owner: France, Province: "paris", Coast: ""}
	board.PlaceUnit(unit)

	originalState := NewGameState(board, SpringMovement, 1901)
	originalState.AddRawOrder(France, "A Par-Bur")
	originalState.AddRawOrder(Germany, "A Ber-Kie")
	originalState.SupplyCenters[France] = []string{"paris", "marseilles"}
	originalState.SupplyCenters[Germany] = []string{"berlin", "munich"}

	testTime := time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC)
	originalState.CreatedAt = testTime

	cloned := originalState.Clone()

	if cloned.Phase != originalState.Phase {
		t.Errorf("Phase not cloned correctly: expected %s, got %s", originalState.Phase, cloned.Phase)
	}

	if cloned.Year != originalState.Year {
		t.Errorf("Year not cloned correctly: expected %d, got %d", originalState.Year, cloned.Year)
	}

	if !cloned.CreatedAt.Equal(originalState.CreatedAt) {
		t.Error("CreatedAt not cloned correctly")
	}

	if len(cloned.RawOrders) != len(originalState.RawOrders) {
		t.Error("RawOrders length mismatch")
	}

	frenchOrders := cloned.GetRawOrdersForNation(France)
	if len(frenchOrders) != 1 || frenchOrders[0] != "A Par-Bur" {
		t.Error("French orders not cloned correctly")
	}

	cloned.AddRawOrder(France, "F Bre-MAO")
	if len(originalState.GetRawOrdersForNation(France)) != 1 {
		t.Error("Modifying cloned orders affected original")
	}

	if len(cloned.SupplyCenters[France]) != 2 {
		t.Error("Supply centers not cloned correctly")
	}

	cloned.SupplyCenters[France] = append(cloned.SupplyCenters[France], "brest")
	if len(originalState.SupplyCenters[France]) != 2 {
		t.Error("Modifying cloned supply centers affected original")
	}

	if cloned.Board == originalState.Board {
		t.Error("Board should be cloned, not referenced")
	}

	clonedUnit := cloned.Board.GetUnit("paris")
	if clonedUnit == nil {
		t.Error("Unit not found in cloned board")
	}

	if clonedUnit == originalState.Board.GetUnit("paris") {
		t.Error("Unit should be cloned, not referenced")
	}

	cloned.Board.RemoveUnit("paris")
	if originalState.Board.GetUnit("paris") == nil {
		t.Error("Modifying cloned board affected original")
	}
}
