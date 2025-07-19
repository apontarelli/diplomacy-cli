package storage

import (
	"os"
	"testing"
	"time"

	"diplomacy-cli/backend/internal/game"
)

func TestGameStatePersistence_RoundTrip(t *testing.T) {
	// Create test database
	db, err := NewDatabase(Config{DatabasePath: ":memory:"})
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	defer db.Close()

	ops := NewGameOperations(db)

	// Create a game record first
	testGame := &Game{
		Name:   "Test Game",
		Phase:  "spring_movement",
		Year:   1901,
		Status: "in_progress",
	}
	players := []Player{}
	err = ops.CreateGameWithPlayers(testGame, players)
	if err != nil {
		t.Fatalf("Failed to create game: %v", err)
	}

	// Create a test game state
	testGameState := createTestGameState()

	// Save the game state
	err = ops.SaveGameState(testGame.ID, testGameState)
	if err != nil {
		t.Fatalf("Failed to save game state: %v", err)
	}

	// Load the game state
	loadedGameState, err := ops.LoadGameState(testGame.ID)
	if err != nil {
		t.Fatalf("Failed to load game state: %v", err)
	}

	// Verify round-trip integrity
	assertGameStateEqualBasic(t, testGameState, loadedGameState)
}

func TestGameStateSnapshot(t *testing.T) {
	// Create test database
	db, err := NewDatabase(Config{DatabasePath: ":memory:"})
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	defer db.Close()

	ops := NewGameOperations(db)

	// Create a game record first
	testGame := &Game{
		Name:   "Test Game",
		Phase:  "spring_movement",
		Year:   1901,
		Status: "in_progress",
	}
	players := []Player{}
	err = ops.CreateGameWithPlayers(testGame, players)
	if err != nil {
		t.Fatalf("Failed to create game: %v", err)
	}

	// Create test game state and save initial state
	testGameState := createTestGameState()
	err = ops.SaveGameState(testGame.ID, testGameState)
	if err != nil {
		t.Fatalf("Failed to save initial game state: %v", err)
	}

	// Create a modified game state
	modifiedState := createTestGameState()
	modifiedState.Phase = game.FallMovement
	modifiedState.Year = 1902

	// Save snapshot with specific timestamp
	snapshotTime := time.Now().Add(-1 * time.Hour)
	err = ops.SaveGameStateSnapshot(testGame.ID, modifiedState, snapshotTime)
	if err != nil {
		t.Fatalf("Failed to save game state snapshot: %v", err)
	}

	// Load history
	result, err := ops.LoadGameStateHistory(testGame.ID)
	if err != nil {
		t.Fatalf("Failed to load game state history: %v", err)
	}

	// Verify no failures occurred
	if len(result.Failures) > 0 {
		t.Errorf("Expected no failures, got %d failures: %v", len(result.Failures), result.Failures)
	}

	// Verify history contains both states
	if len(result.States) < 2 {
		t.Errorf("Expected at least 2 historical states, got %d", len(result.States))
	}

	// Verify the states are in chronological order
	if len(result.States) >= 2 {
		// The snapshot should be first (older timestamp)
		if result.States[0].Phase != game.FallMovement || result.States[0].Year != 1902 {
			t.Errorf("Expected first state to be Fall 1902, got %v %d", result.States[0].Phase, result.States[0].Year)
		}
		// The current state should be second (newer timestamp)
		if result.States[1].Phase != game.SpringMovement || result.States[1].Year != 1901 {
			t.Errorf("Expected second state to be Spring 1901, got %v %d", result.States[1].Phase, result.States[1].Year)
		}
	}
}

func TestGameStatePersistence_ErrorHandling(t *testing.T) {
	// Create test database
	db, err := NewDatabase(Config{DatabasePath: ":memory:"})
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	defer db.Close()

	ops := NewGameOperations(db)

	tests := []struct {
		name string
		test func(t *testing.T)
	}{
		{
			name: "load non-existent game",
			test: func(t *testing.T) {
				_, err := ops.LoadGameState(999)
				if err == nil {
					t.Error("Expected error for non-existent game")
				}
			},
		},
		{
			name: "save nil game state",
			test: func(t *testing.T) {
				err := ops.SaveGameState(1, nil)
				if err == nil {
					t.Error("Expected error for nil game state")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, tt.test)
	}
}

func TestGameStatePersistence_ConcurrentAccess(t *testing.T) {
	// Create test database file (not in-memory for concurrent access)
	tempFile, err := os.CreateTemp("", "test_concurrent_*.db")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tempFile.Name())
	tempFile.Close()

	db, err := NewDatabase(Config{DatabasePath: tempFile.Name()})
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	defer db.Close()

	ops := NewGameOperations(db)

	// Create a game record first
	testGame := &Game{
		Name:   "Test Game",
		Phase:  "spring_movement",
		Year:   1901,
		Status: "in_progress",
	}
	players := []Player{}
	err = ops.CreateGameWithPlayers(testGame, players)
	if err != nil {
		t.Fatalf("Failed to create game: %v", err)
	}

	// Create test game state
	testGameState := createTestGameState()

	// Save an initial state so loads don't fail
	err = ops.SaveGameState(testGame.ID, testGameState)
	if err != nil {
		t.Fatalf("Failed to save initial game state: %v", err)
	}

	// Test concurrent saves and loads
	done := make(chan bool, 2)

	// Goroutine 1: Save game state
	go func() {
		defer func() { done <- true }()
		for i := 0; i < 5; i++ {
			modifiedState := createTestGameState()
			modifiedState.Year = 1901 + i
			err := ops.SaveGameState(testGame.ID, modifiedState)
			if err != nil {
				t.Errorf("Concurrent save failed: %v", err)
			}
			time.Sleep(10 * time.Millisecond)
		}
	}()

	// Goroutine 2: Load game state
	go func() {
		defer func() { done <- true }()
		for i := 0; i < 5; i++ {
			_, err := ops.LoadGameState(testGame.ID)
			if err != nil {
				t.Errorf("Concurrent load failed: %v", err)
			}
			time.Sleep(10 * time.Millisecond)
		}
	}()

	// Wait for both goroutines to complete
	<-done
	<-done
}

func TestGameStateHistory_StructuredErrorHandling(t *testing.T) {
	// Create test database
	db, err := NewDatabase(Config{DatabasePath: ":memory:"})
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	defer db.Close()

	ops := NewGameOperations(db)

	// Create a game record first
	testGame := &Game{
		Name:   "Test Game",
		Phase:  "spring_movement",
		Year:   1901,
		Status: "in_progress",
	}
	players := []Player{}
	err = ops.CreateGameWithPlayers(testGame, players)
	if err != nil {
		t.Fatalf("Failed to create game: %v", err)
	}

	// Insert a valid game state
	validState := createTestGameState()
	err = ops.SaveGameState(testGame.ID, validState)
	if err != nil {
		t.Fatalf("Failed to save valid game state: %v", err)
	}

	// Manually insert corrupted JSON data directly into the database
	_, err = db.Exec(`
		INSERT INTO game_histories (game_id, phase, year, state_data, timestamp, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)`,
		testGame.ID, "spring_movement", 1902, "invalid json data",
		time.Now().Add(1*time.Hour), time.Now(), time.Now())
	if err != nil {
		t.Fatalf("Failed to insert corrupted data: %v", err)
	}

	// Load history using structured error handling
	result, err := ops.LoadGameStateHistory(testGame.ID)
	if err != nil {
		t.Fatalf("Failed to load game state history: %v", err)
	}

	// Verify we got partial results
	if len(result.States) != 1 {
		t.Errorf("Expected 1 valid state, got %d", len(result.States))
	}

	// Verify we captured the failure
	if len(result.Failures) != 1 {
		t.Errorf("Expected 1 failure, got %d", len(result.Failures))
	}

	// Verify failure details
	if len(result.Failures) > 0 {
		failure := result.Failures[0]
		if failure.GameID != testGame.ID {
			t.Errorf("Expected failure GameID %d, got %d", testGame.ID, failure.GameID)
		}
		if failure.Error == "" {
			t.Error("Expected failure to have error message")
		}
		if failure.RawData != "invalid json data" {
			t.Errorf("Expected failure RawData to be 'invalid json data', got '%s'", failure.RawData)
		}
	}

	// Verify total count
	if result.TotalCount != 2 {
		t.Errorf("Expected TotalCount 2, got %d", result.TotalCount)
	}

	// Test the convenience method
	_, err = ops.LoadGameStateHistoryOrFail(testGame.ID)
	if err == nil {
		t.Error("Expected LoadGameStateHistoryOrFail to fail when there are corrupted states")
	}
}

// Helper functions for creating test data

func createTestGameState() *game.GameState {
	// Create a simple board
	board := game.NewBoard()

	// Add a few test provinces
	provinces := []*game.Province{
		{
			Name:           "vienna",
			ShortCode:      "vie",
			DisplayName:    "Vienna",
			Type:           game.Land,
			SupplyCenter:   true,
			CoastNeighbors: make(map[string][]string),
			ArmyNeighbors:  []string{"bud", "gal"},
			FleetNeighbors: []string{},
		},
		{
			Name:           "london",
			ShortCode:      "lon",
			DisplayName:    "London",
			Type:           game.Land,
			SupplyCenter:   true,
			CoastNeighbors: make(map[string][]string),
			ArmyNeighbors:  []string{"yor"},
			FleetNeighbors: []string{"nth"},
		},
	}

	for _, province := range provinces {
		board.AddProvince(province)
	}

	// Add some test units
	units := []*game.Unit{
		{Type: game.Army, Owner: game.Austria, Province: "vienna", Coast: ""},
		{Type: game.Fleet, Owner: game.England, Province: "london", Coast: ""},
	}

	for _, unit := range units {
		board.PlaceUnit(unit)
	}

	// Create game state
	gameState := game.NewGameState(board, game.SpringMovement, 1901)

	// Add some test data
	gameState.RawOrders[game.Austria] = []string{"A vie-bud"}
	gameState.RawOrders[game.England] = []string{"F lon-nth"}

	gameState.SupplyCenters[game.Austria] = []string{"vienna"}
	gameState.SupplyCenters[game.England] = []string{"london"}

	return gameState
}

// Assertion helper functions

func assertGameStateEqualBasic(t *testing.T, expected, actual *game.GameState) {
	if expected.Phase != actual.Phase {
		t.Errorf("Phase mismatch: expected %v, got %v", expected.Phase, actual.Phase)
	}
	if expected.Year != actual.Year {
		t.Errorf("Year mismatch: expected %v, got %v", expected.Year, actual.Year)
	}

	// Compare board units count
	if len(expected.Board.Units) != len(actual.Board.Units) {
		t.Errorf("Units count mismatch: expected %d, got %d", len(expected.Board.Units), len(actual.Board.Units))
	}

	// Compare raw orders count
	if len(expected.RawOrders) != len(actual.RawOrders) {
		t.Errorf("Raw orders count mismatch: expected %d, got %d", len(expected.RawOrders), len(actual.RawOrders))
	}

	// Compare supply centers count
	if len(expected.SupplyCenters) != len(actual.SupplyCenters) {
		t.Errorf("Supply centers count mismatch: expected %d, got %d", len(expected.SupplyCenters), len(actual.SupplyCenters))
	}

	// Compare dislodged units count
	if len(expected.DislodgedUnits) != len(actual.DislodgedUnits) {
		t.Errorf("Dislodged units count mismatch: expected %d, got %d", len(expected.DislodgedUnits), len(actual.DislodgedUnits))
	}
}
