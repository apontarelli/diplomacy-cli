package storage

import (
	"encoding/json"
	"testing"
	"time"

	"diplomacy-cli/backend/internal/game"
)

func TestGameStateSerialization_RoundTrip(t *testing.T) {
	// Create a complex game state for testing
	board := createTestBoard()
	gameState := &game.GameState{
		Board:         board,
		Phase:         game.SpringMovement,
		Year:          1901,
		RawOrders:     make(map[game.Nation][]string),
		SupplyCenters: make(map[game.Nation][]string),
		CreatedAt:     time.Now(),
	}

	// Add some raw orders
	gameState.RawOrders[game.Austria] = []string{"A vie-bud", "F tri-alb"}
	gameState.RawOrders[game.England] = []string{"F lon-nth", "A lvp-yor"}

	// Add supply centers
	gameState.SupplyCenters[game.Austria] = []string{"vienna"}
	gameState.SupplyCenters[game.England] = []string{"london"}

	// Add dislodged units
	dislodgedUnit := &game.Unit{
		Type:      game.Army,
		Owner:     game.France,
		Province:  "vienna", // Use existing province
		Coast:     "",
		Dislodged: true,
	}
	gameState.DislodgedUnits = []*game.Unit{dislodgedUnit}

	// Test serialization
	data, err := SerializeGameState(gameState)
	if err != nil {
		t.Fatalf("Failed to serialize game state: %v", err)
	}

	// Test deserialization
	deserializedState, err := DeserializeGameState(data)
	if err != nil {
		t.Fatalf("Failed to deserialize game state: %v", err)
	}

	// Verify round-trip integrity
	assertGameStateEqual(t, gameState, deserializedState)
}

func TestGameSerialization_RoundTrip(t *testing.T) {
	// Create a test game
	game := createTestGame()

	// Test serialization
	data, err := SerializeGame(game)
	if err != nil {
		t.Fatalf("Failed to serialize game: %v", err)
	}

	// Test deserialization
	deserializedGame, err := DeserializeGame(data)
	if err != nil {
		t.Fatalf("Failed to deserialize game: %v", err)
	}

	// Verify round-trip integrity
	assertGameEqual(t, game, deserializedGame)
}

func TestUnitSerialization_RoundTrip(t *testing.T) {
	unit := &game.Unit{
		Type:      game.Fleet,
		Owner:     game.Russia,
		Province:  "stp",
		Coast:     "nc",
		Dislodged: false,
	}

	dto, err := ToUnitDTO(unit)
	if err != nil {
		t.Fatalf("Failed to convert unit to DTO: %v", err)
	}

	convertedUnit, err := FromUnitDTO(dto)
	if err != nil {
		t.Fatalf("Failed to convert unit from DTO: %v", err)
	}

	assertUnitEqual(t, unit, convertedUnit)
}

func TestProvinceSerialization_RoundTrip(t *testing.T) {
	province := &game.Province{
		Name:         "st_petersburg",
		ShortCode:    "stp",
		DisplayName:  "St. Petersburg",
		Type:         game.Land,
		SupplyCenter: true,
		CoastNeighbors: map[string][]string{
			"nc": {"bar", "nwy"},
			"sc": {"lvn", "bot"},
		},
		ArmyNeighbors:  []string{"lvn", "fin", "mos"},
		FleetNeighbors: []string{"bar", "nwy", "bot"},
	}

	dto, err := ToProvinceDTO(province)
	if err != nil {
		t.Fatalf("Failed to convert province to DTO: %v", err)
	}

	convertedProvince, err := FromProvinceDTO(dto)
	if err != nil {
		t.Fatalf("Failed to convert province from DTO: %v", err)
	}

	assertProvinceEqual(t, province, convertedProvince)
}

func TestOrderSerialization_RoundTrip(t *testing.T) {
	order := &game.Order{
		ID:                 "order-1",
		UnitType:           game.Army,
		From:               "vie",
		FromCoast:          "",
		Type:               game.Move,
		To:                 "bud",
		ToCoast:            "",
		Owner:              game.Austria,
		Result:             game.Success,
		FailureReason:      "",
		SupportTarget:      "",
		SupportDestination: "",
		ConvoyTarget:       "",
	}

	dto, err := ToOrderDTO(order)
	if err != nil {
		t.Fatalf("Failed to convert order to DTO: %v", err)
	}

	convertedOrder, err := FromOrderDTO(dto)
	if err != nil {
		t.Fatalf("Failed to convert order from DTO: %v", err)
	}

	assertOrderEqual(t, order, convertedOrder)
}

func TestSerialization_ErrorHandling(t *testing.T) {
	tests := []struct {
		name string
		test func(t *testing.T)
	}{
		{
			name: "nil game state",
			test: func(t *testing.T) {
				_, err := SerializeGameState(nil)
				if err == nil {
					t.Error("Expected error for nil game state")
				}
			},
		},
		{
			name: "empty JSON data",
			test: func(t *testing.T) {
				_, err := DeserializeGameState([]byte{})
				if err == nil {
					t.Error("Expected error for empty JSON data")
				}
			},
		},
		{
			name: "invalid JSON data",
			test: func(t *testing.T) {
				_, err := DeserializeGameState([]byte("invalid json"))
				if err == nil {
					t.Error("Expected error for invalid JSON data")
				}
			},
		},
		{
			name: "nil unit DTO",
			test: func(t *testing.T) {
				_, err := FromUnitDTO(nil)
				if err == nil {
					t.Error("Expected error for nil unit DTO")
				}
			},
		},
		{
			name: "nil province DTO",
			test: func(t *testing.T) {
				_, err := FromProvinceDTO(nil)
				if err == nil {
					t.Error("Expected error for nil province DTO")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, tt.test)
	}
}

func TestGameStateValidation(t *testing.T) {
	tests := []struct {
		name          string
		setupState    func() *game.GameState
		expectError   bool
		errorContains string
	}{
		{
			name: "valid game state",
			setupState: func() *game.GameState {
				return createTestGameStateForValidation()
			},
			expectError: false,
		},
		{
			name: "nil game state",
			setupState: func() *game.GameState {
				return nil
			},
			expectError:   true,
			errorContains: "game state cannot be nil",
		},
		{
			name: "nil board",
			setupState: func() *game.GameState {
				gs := createTestGameStateForValidation()
				gs.Board = nil
				return gs
			},
			expectError:   true,
			errorContains: "board cannot be nil",
		},
		{
			name: "invalid year",
			setupState: func() *game.GameState {
				gs := createTestGameStateForValidation()
				gs.Year = 1800
				return gs
			},
			expectError:   true,
			errorContains: "invalid year 1800",
		},
		{
			name: "empty phase",
			setupState: func() *game.GameState {
				gs := createTestGameStateForValidation()
				gs.Phase = ""
				return gs
			},
			expectError:   true,
			errorContains: "phase cannot be empty",
		},
		{
			name: "unit in non-existent province",
			setupState: func() *game.GameState {
				gs := createTestGameStateForValidation()
				// Add unit to non-existent province
				gs.Board.Units["nonexistent"] = &game.Unit{
					Type:     game.Army,
					Owner:    game.Austria,
					Province: "nonexistent",
				}
				return gs
			},
			expectError:   true,
			errorContains: "unit in non-existent province",
		},
		{
			name: "army in sea province",
			setupState: func() *game.GameState {
				gs := createTestGameStateForValidation()
				// Add sea province
				seaProvince := &game.Province{
					Name:        "test_sea",
					ShortCode:   "ts",
					DisplayName: "Test Sea",
					Type:        game.Sea,
				}
				gs.Board.AddProvince(seaProvince)
				// Try to place army in sea
				gs.Board.Units["test_sea"] = &game.Unit{
					Type:     game.Army,
					Owner:    game.Austria,
					Province: "test_sea",
				}
				return gs
			},
			expectError:   true,
			errorContains: "army cannot be placed in sea province",
		},
		{
			name: "supply center assigned to multiple nations",
			setupState: func() *game.GameState {
				gs := createTestGameStateForValidation()
				// Assign same supply center to multiple nations
				gs.SupplyCenters[game.Austria] = []string{"vienna"}
				gs.SupplyCenters[game.England] = []string{"vienna"}
				return gs
			},
			expectError:   true,
			errorContains: "supply center vienna assigned to multiple nations",
		},
		{
			name: "supply center references non-existent province",
			setupState: func() *game.GameState {
				gs := createTestGameStateForValidation()
				gs.SupplyCenters[game.Austria] = []string{"nonexistent"}
				return gs
			},
			expectError:   true,
			errorContains: "supply center references non-existent province",
		},
		{
			name: "dislodged unit without dislodged flag",
			setupState: func() *game.GameState {
				gs := createTestGameStateForValidation()
				// Add unit to dislodged list without Dislodged=true
				gs.DislodgedUnits = []*game.Unit{
					{
						Type:      game.Army,
						Owner:     game.Austria,
						Province:  "vienna",
						Dislodged: false, // Should be true
					},
				}
				return gs
			},
			expectError:   true,
			errorContains: "should have Dislodged=true",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gameState := tt.setupState()
			err := ValidateGameState(gameState)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				} else if tt.errorContains != "" && !containsString(err.Error(), tt.errorContains) {
					t.Errorf("Expected error to contain '%s', got: %v", tt.errorContains, err)
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
			}
		})
	}
}

func TestDeserializationWithValidation(t *testing.T) {
	// Test that validation is actually called during deserialization
	validState := createTestGameStateForValidation()

	// Serialize valid state
	data, err := SerializeGameState(validState)
	if err != nil {
		t.Fatalf("Failed to serialize valid state: %v", err)
	}

	// Should deserialize successfully
	_, err = DeserializeGameState(data)
	if err != nil {
		t.Errorf("Valid state should deserialize successfully: %v", err)
	}

	// Now test with invalid data that passes JSON unmarshaling but fails validation
	// We'll manually create invalid JSON that has correct structure but invalid data
	invalidJSON := `{
		"phase": "",
		"year": 1800,
		"board": {
			"provinces": {},
			"units": {}
		},
		"raw_orders": {},
		"supply_centers": {},
		"dislodged_units": [],
		"created_at": "2023-01-01T00:00:00Z"
	}`

	_, err = DeserializeGameState([]byte(invalidJSON))
	if err == nil {
		t.Error("Expected deserialization to fail due to validation")
	}
	if !containsString(err.Error(), "validation failed") {
		t.Errorf("Expected validation error, got: %v", err)
	}
}

// Helper functions

func createTestGameStateForValidation() *game.GameState {
	board := game.NewBoard()

	// Add valid provinces
	vienna := &game.Province{
		Name:           "vienna",
		ShortCode:      "vie",
		DisplayName:    "Vienna",
		Type:           game.Land,
		SupplyCenter:   true,
		CoastNeighbors: make(map[string][]string),
		ArmyNeighbors:  []string{"budapest"},
		FleetNeighbors: []string{},
	}
	board.AddProvince(vienna)

	budapest := &game.Province{
		Name:           "budapest",
		ShortCode:      "bud",
		DisplayName:    "Budapest",
		Type:           game.Land,
		SupplyCenter:   true,
		CoastNeighbors: make(map[string][]string),
		ArmyNeighbors:  []string{"vienna"},
		FleetNeighbors: []string{},
	}
	board.AddProvince(budapest)

	// Add valid unit
	unit := &game.Unit{
		Type:     game.Army,
		Owner:    game.Austria,
		Province: "vienna",
		Coast:    "",
	}
	board.PlaceUnit(unit)

	// Create game state
	gameState := game.NewGameState(board, game.SpringMovement, 1901)
	gameState.SupplyCenters[game.Austria] = []string{"vienna"}
	gameState.RawOrders[game.Austria] = []string{"A vie-bud"}

	return gameState
}

func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr ||
		(len(s) > len(substr) &&
			(s[:len(substr)] == substr ||
				s[len(s)-len(substr):] == substr ||
				containsSubstring(s, substr))))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func TestJSONMarshaling_ComplexStructures(t *testing.T) {
	// Test that complex nested structures can be marshaled to valid JSON
	gameState := createComplexGameState()

	dto, err := ToGameStateDTO(gameState)
	if err != nil {
		t.Fatalf("Failed to convert to DTO: %v", err)
	}

	// Marshal to JSON
	jsonData, err := json.Marshal(dto)
	if err != nil {
		t.Fatalf("Failed to marshal to JSON: %v", err)
	}

	// Verify it's valid JSON by unmarshaling
	var unmarshaled GameStateDTO
	if err := json.Unmarshal(jsonData, &unmarshaled); err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	// Convert back to domain object
	reconstructed, err := FromGameStateDTO(&unmarshaled)
	if err != nil {
		t.Fatalf("Failed to convert from DTO: %v", err)
	}

	// Verify integrity
	assertGameStateEqual(t, gameState, reconstructed)
}

// Helper functions for creating test data

func createTestBoard() *game.Board {
	board := game.NewBoard()

	// Add some test provinces
	vienna := &game.Province{
		Name:           "vienna",
		ShortCode:      "vie",
		DisplayName:    "Vienna",
		Type:           game.Land,
		SupplyCenter:   true,
		CoastNeighbors: make(map[string][]string),
		ArmyNeighbors:  []string{"bud", "gal", "boh", "tyr", "tri"},
		FleetNeighbors: []string{},
	}
	board.AddProvince(vienna)

	london := &game.Province{
		Name:           "london",
		ShortCode:      "lon",
		DisplayName:    "London",
		Type:           game.Land,
		SupplyCenter:   true,
		CoastNeighbors: make(map[string][]string),
		ArmyNeighbors:  []string{"yor", "wal"},
		FleetNeighbors: []string{"nth", "eng"},
	}
	board.AddProvince(london)

	// Add some test units
	viennaUnit := &game.Unit{
		Type:     game.Army,
		Owner:    game.Austria,
		Province: "vienna",
		Coast:    "",
	}
	board.PlaceUnit(viennaUnit)

	londonUnit := &game.Unit{
		Type:     game.Fleet,
		Owner:    game.England,
		Province: "london",
		Coast:    "",
	}
	board.PlaceUnit(londonUnit)

	return board
}

func createTestGame() *game.Game {
	testGame := game.NewGame("test-game-1", "Test Game")
	testGame.AddPlayer(game.Austria, "player-1")
	testGame.AddPlayer(game.England, "player-2")

	// Set current state
	board := createTestBoard()
	gameState := game.NewGameState(board, game.SpringMovement, 1901)
	testGame.CurrentState = gameState

	// Add some history
	historyState := gameState.Clone()
	historyState.Phase = game.FallMovement
	testGame.History = append(testGame.History, historyState)

	return testGame
}

func createComplexGameState() *game.GameState {
	board := createTestBoard()
	gameState := &game.GameState{
		Board:         board,
		Phase:         game.SpringMovement,
		Year:          1901,
		RawOrders:     make(map[game.Nation][]string),
		SupplyCenters: make(map[game.Nation][]string),
		CreatedAt:     time.Now(),
	}

	// Add complex data for all nations
	nations := []game.Nation{game.Austria, game.England, game.France, game.Germany, game.Italy, game.Russia, game.Turkey}
	for _, nation := range nations {
		gameState.RawOrders[nation] = []string{
			"A vie-bud",
			"F tri S A vie-bud",
			"A boh-mun",
		}
		gameState.SupplyCenters[nation] = []string{"vienna"}
	}

	// Add multiple dislodged units
	for i, nation := range nations[:3] {
		unit := &game.Unit{
			Type:      game.Army,
			Owner:     nation,
			Province:  "test-province-" + string(rune(i)),
			Coast:     "",
			Dislodged: true,
		}
		gameState.DislodgedUnits = append(gameState.DislodgedUnits, unit)
	}

	return gameState
}

// Assertion helper functions

func assertGameStateEqual(t *testing.T, expected, actual *game.GameState) {
	if expected.Phase != actual.Phase {
		t.Errorf("Phase mismatch: expected %v, got %v", expected.Phase, actual.Phase)
	}
	if expected.Year != actual.Year {
		t.Errorf("Year mismatch: expected %v, got %v", expected.Year, actual.Year)
	}

	// Compare raw orders
	if len(expected.RawOrders) != len(actual.RawOrders) {
		t.Errorf("RawOrders length mismatch: expected %d, got %d", len(expected.RawOrders), len(actual.RawOrders))
	}
	for nation, orders := range expected.RawOrders {
		actualOrders, exists := actual.RawOrders[nation]
		if !exists {
			t.Errorf("Missing raw orders for nation %v", nation)
			continue
		}
		if len(orders) != len(actualOrders) {
			t.Errorf("Raw orders length mismatch for %v: expected %d, got %d", nation, len(orders), len(actualOrders))
		}
		for i, order := range orders {
			if i < len(actualOrders) && order != actualOrders[i] {
				t.Errorf("Raw order mismatch for %v[%d]: expected %v, got %v", nation, i, order, actualOrders[i])
			}
		}
	}

	// Compare supply centers
	if len(expected.SupplyCenters) != len(actual.SupplyCenters) {
		t.Errorf("SupplyCenters length mismatch: expected %d, got %d", len(expected.SupplyCenters), len(actual.SupplyCenters))
	}

	// Compare dislodged units
	if len(expected.DislodgedUnits) != len(actual.DislodgedUnits) {
		t.Errorf("DislodgedUnits length mismatch: expected %d, got %d", len(expected.DislodgedUnits), len(actual.DislodgedUnits))
	}
}

func assertGameEqual(t *testing.T, expected, actual *game.Game) {
	if expected.ID != actual.ID {
		t.Errorf("ID mismatch: expected %v, got %v", expected.ID, actual.ID)
	}
	if expected.Name != actual.Name {
		t.Errorf("Name mismatch: expected %v, got %v", expected.Name, actual.Name)
	}
	if expected.Status != actual.Status {
		t.Errorf("Status mismatch: expected %v, got %v", expected.Status, actual.Status)
	}

	// Compare players
	if len(expected.Players) != len(actual.Players) {
		t.Errorf("Players length mismatch: expected %d, got %d", len(expected.Players), len(actual.Players))
	}

	// Compare history length
	if len(expected.History) != len(actual.History) {
		t.Errorf("History length mismatch: expected %d, got %d", len(expected.History), len(actual.History))
	}
}

func assertUnitEqual(t *testing.T, expected, actual *game.Unit) {
	if expected.Type != actual.Type {
		t.Errorf("Unit type mismatch: expected %v, got %v", expected.Type, actual.Type)
	}
	if expected.Owner != actual.Owner {
		t.Errorf("Unit owner mismatch: expected %v, got %v", expected.Owner, actual.Owner)
	}
	if expected.Province != actual.Province {
		t.Errorf("Unit province mismatch: expected %v, got %v", expected.Province, actual.Province)
	}
	if expected.Coast != actual.Coast {
		t.Errorf("Unit coast mismatch: expected %v, got %v", expected.Coast, actual.Coast)
	}
	if expected.Dislodged != actual.Dislodged {
		t.Errorf("Unit dislodged mismatch: expected %v, got %v", expected.Dislodged, actual.Dislodged)
	}
}

func assertProvinceEqual(t *testing.T, expected, actual *game.Province) {
	if expected.Name != actual.Name {
		t.Errorf("Province name mismatch: expected %v, got %v", expected.Name, actual.Name)
	}
	if expected.Type != actual.Type {
		t.Errorf("Province type mismatch: expected %v, got %v", expected.Type, actual.Type)
	}
	if expected.SupplyCenter != actual.SupplyCenter {
		t.Errorf("Province supply center mismatch: expected %v, got %v", expected.SupplyCenter, actual.SupplyCenter)
	}
}

func assertOrderEqual(t *testing.T, expected, actual *game.Order) {
	if expected.ID != actual.ID {
		t.Errorf("Order ID mismatch: expected %v, got %v", expected.ID, actual.ID)
	}
	if expected.Type != actual.Type {
		t.Errorf("Order type mismatch: expected %v, got %v", expected.Type, actual.Type)
	}
	if expected.Owner != actual.Owner {
		t.Errorf("Order owner mismatch: expected %v, got %v", expected.Owner, actual.Owner)
	}
	if expected.From != actual.From {
		t.Errorf("Order from mismatch: expected %v, got %v", expected.From, actual.From)
	}
	if expected.To != actual.To {
		t.Errorf("Order to mismatch: expected %v, got %v", expected.To, actual.To)
	}
}
