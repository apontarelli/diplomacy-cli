package storage

import (
	"database/sql"
	"os"
	"testing"
	"time"
)

func TestDatabaseInitialization(t *testing.T) {
	// Use in-memory SQLite for testing
	config := Config{
		DatabasePath: ":memory:",
		LogQueries:   false,
	}

	db, err := NewDatabase(config)
	if err != nil {
		t.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	// Test database health
	if err := db.Health(); err != nil {
		t.Errorf("Database health check failed: %v", err)
	}
}

func TestGameOperations(t *testing.T) {
	// Use in-memory SQLite for testing
	config := Config{
		DatabasePath: ":memory:",
		LogQueries:   false,
	}

	db, err := NewDatabase(config)
	if err != nil {
		t.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	ops := NewGameOperations(db)

	// Test creating a game with players
	game := &Game{
		Name:   "Test Game",
		Phase:  "spring_movement",
		Year:   1901,
		Status: "waiting_for_players",
	}

	players := []Player{
		{
			UserID: "user1",
			Nation: "england",
		},
		{
			UserID: "user2",
			Nation: "france",
		},
	}

	err = ops.CreateGameWithPlayers(game, players)
	if err != nil {
		t.Fatalf("Failed to create game with players: %v", err)
	}

	// Verify game was created
	var retrievedGame Game
	err = db.QueryRow("SELECT id, name, phase, year, status FROM games WHERE id = ?", game.ID).
		Scan(&retrievedGame.ID, &retrievedGame.Name, &retrievedGame.Phase, &retrievedGame.Year, &retrievedGame.Status)
	if err != nil {
		t.Errorf("Failed to retrieve created game: %v", err)
	}

	if retrievedGame.Name != "Test Game" {
		t.Errorf("Expected game name 'Test Game', got '%s'", retrievedGame.Name)
	}

	// Verify players were created
	rows, err := db.Query("SELECT id, game_id, user_id, nation FROM players WHERE game_id = ?", game.ID)
	if err != nil {
		t.Errorf("Failed to retrieve players: %v", err)
	}
	defer rows.Close()

	var retrievedPlayers []Player
	for rows.Next() {
		var player Player
		err := rows.Scan(&player.ID, &player.GameID, &player.UserID, &player.Nation)
		if err != nil {
			t.Errorf("Failed to scan player: %v", err)
		}
		retrievedPlayers = append(retrievedPlayers, player)
	}

	if len(retrievedPlayers) != 2 {
		t.Errorf("Expected 2 players, got %d", len(retrievedPlayers))
	}
}

func TestOrderSubmission(t *testing.T) {
	config := Config{
		DatabasePath: ":memory:",
		LogQueries:   false,
	}

	db, err := NewDatabase(config)
	if err != nil {
		t.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	ops := NewGameOperations(db)

	// Create a game first
	game := &Game{
		Name:   "Test Game",
		Phase:  "spring_movement",
		Year:   1901,
		Status: "in_progress",
	}

	result, err := db.Exec(`
		INSERT INTO games (name, phase, year, status, created_at, updated_at) 
		VALUES (?, ?, ?, ?, ?, ?)`,
		game.Name, game.Phase, game.Year, game.Status, time.Now(), time.Now())
	if err != nil {
		t.Fatalf("Failed to create test game: %v", err)
	}

	gameID, err := result.LastInsertId()
	if err != nil {
		t.Fatalf("Failed to get game ID: %v", err)
	}
	game.ID = gameID

	// Create a player
	player := &Player{
		GameID: game.ID,
		UserID: "user1",
		Nation: "england",
	}

	result, err = db.Exec(`
		INSERT INTO players (game_id, user_id, nation, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?)`,
		player.GameID, player.UserID, player.Nation, time.Now(), time.Now())
	if err != nil {
		t.Fatalf("Failed to create test player: %v", err)
	}

	playerID, err := result.LastInsertId()
	if err != nil {
		t.Fatalf("Failed to get player ID: %v", err)
	}
	player.ID = playerID

	// Submit orders
	orders := []Order{
		{
			PlayerID: player.ID,
			Type:     "move",
			Target:   "london",
		},
		{
			PlayerID: player.ID,
			Type:     "hold",
		},
	}

	err = ops.SubmitOrders(game.ID, orders)
	if err != nil {
		t.Fatalf("Failed to submit orders: %v", err)
	}

	// Verify orders were created
	rows, err := db.Query("SELECT id, game_id, player_id, type, resolved FROM orders WHERE game_id = ?", game.ID)
	if err != nil {
		t.Errorf("Failed to retrieve orders: %v", err)
	}
	defer rows.Close()

	var retrievedOrders []Order
	for rows.Next() {
		var order Order
		err := rows.Scan(&order.ID, &order.GameID, &order.PlayerID, &order.Type, &order.Resolved)
		if err != nil {
			t.Errorf("Failed to scan order: %v", err)
		}
		retrievedOrders = append(retrievedOrders, order)
	}

	if len(retrievedOrders) != 2 {
		t.Errorf("Expected 2 orders, got %d", len(retrievedOrders))
	}

	for _, order := range retrievedOrders {
		if order.Resolved {
			t.Errorf("Expected order to be unresolved, but it was resolved")
		}
	}
}

func TestGameStateUpdate(t *testing.T) {
	config := Config{
		DatabasePath: ":memory:",
		LogQueries:   false,
	}

	db, err := NewDatabase(config)
	if err != nil {
		t.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	ops := NewGameOperations(db)

	// Create a game first
	game := &Game{
		Name:   "Test Game",
		Phase:  "spring_movement",
		Year:   1901,
		Status: "in_progress",
	}

	result, err := db.Exec(`
		INSERT INTO games (name, phase, year, status, created_at, updated_at) 
		VALUES (?, ?, ?, ?, ?, ?)`,
		game.Name, game.Phase, game.Year, game.Status, time.Now(), time.Now())
	if err != nil {
		t.Fatalf("Failed to create test game: %v", err)
	}

	gameID, err := result.LastInsertId()
	if err != nil {
		t.Fatalf("Failed to get game ID: %v", err)
	}
	game.ID = gameID

	// Update game state
	units := []Unit{
		{
			Type:     "army",
			Province: "london",
			Owner:    "england",
		},
		{
			Type:     "fleet",
			Province: "english_channel",
			Owner:    "england",
		},
	}

	supplyCenters := []SupplyCenter{
		{
			Province: "london",
			Owner:    "england",
		},
		{
			Province: "liverpool",
			Owner:    "england",
		},
	}

	err = ops.UpdateGameState(game.ID, "fall_movement", 1901, units, supplyCenters)
	if err != nil {
		t.Fatalf("Failed to update game state: %v", err)
	}

	// Verify game was updated
	var updatedGame Game
	err = db.QueryRow("SELECT id, name, phase, year, status FROM games WHERE id = ?", game.ID).
		Scan(&updatedGame.ID, &updatedGame.Name, &updatedGame.Phase, &updatedGame.Year, &updatedGame.Status)
	if err != nil {
		t.Errorf("Failed to retrieve updated game: %v", err)
	}

	if updatedGame.Phase != "fall_movement" {
		t.Errorf("Expected phase 'fall_movement', got '%s'", updatedGame.Phase)
	}

	// Verify units were created
	rows, err := db.Query("SELECT id, game_id, type, province, owner FROM units WHERE game_id = ?", game.ID)
	if err != nil {
		t.Errorf("Failed to retrieve units: %v", err)
	}
	defer rows.Close()

	var retrievedUnits []Unit
	for rows.Next() {
		var unit Unit
		err := rows.Scan(&unit.ID, &unit.GameID, &unit.Type, &unit.Province, &unit.Owner)
		if err != nil {
			t.Errorf("Failed to scan unit: %v", err)
		}
		retrievedUnits = append(retrievedUnits, unit)
	}

	if len(retrievedUnits) != 2 {
		t.Errorf("Expected 2 units, got %d", len(retrievedUnits))
	}

	// Verify supply centers were created
	rows, err = db.Query("SELECT id, game_id, province, owner FROM supply_centers WHERE game_id = ?", game.ID)
	if err != nil {
		t.Errorf("Failed to retrieve supply centers: %v", err)
	}
	defer rows.Close()

	var retrievedCenters []SupplyCenter
	for rows.Next() {
		var center SupplyCenter
		err := rows.Scan(&center.ID, &center.GameID, &center.Province, &center.Owner)
		if err != nil {
			t.Errorf("Failed to scan supply center: %v", err)
		}
		retrievedCenters = append(retrievedCenters, center)
	}

	if len(retrievedCenters) != 2 {
		t.Errorf("Expected 2 supply centers, got %d", len(retrievedCenters))
	}
}

func TestGameHistory(t *testing.T) {
	config := Config{
		DatabasePath: ":memory:",
		LogQueries:   false,
	}

	db, err := NewDatabase(config)
	if err != nil {
		t.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	ops := NewGameOperations(db)

	// Create a game first
	game := &Game{
		Name:   "Test Game",
		Phase:  "spring_movement",
		Year:   1901,
		Status: "in_progress",
	}

	result, err := db.Exec(`
		INSERT INTO games (name, phase, year, status, created_at, updated_at) 
		VALUES (?, ?, ?, ?, ?, ?)`,
		game.Name, game.Phase, game.Year, game.Status, time.Now(), time.Now())
	if err != nil {
		t.Fatalf("Failed to create test game: %v", err)
	}

	gameID, err := result.LastInsertId()
	if err != nil {
		t.Fatalf("Failed to get game ID: %v", err)
	}
	game.ID = gameID

	// Save game history
	history := &GameHistory{
		GameID:    game.ID,
		Phase:     "spring_movement",
		Year:      1901,
		StateData: `{"units": [], "orders": []}`,
		Timestamp: time.Now(),
	}

	err = ops.SaveGameHistory(history)
	if err != nil {
		t.Fatalf("Failed to save game history: %v", err)
	}

	// Verify history was saved
	rows, err := db.Query("SELECT id, game_id, phase, year FROM game_histories WHERE game_id = ?", game.ID)
	if err != nil {
		t.Errorf("Failed to retrieve game history: %v", err)
	}
	defer rows.Close()

	var retrievedHistory []GameHistory
	for rows.Next() {
		var hist GameHistory
		err := rows.Scan(&hist.ID, &hist.GameID, &hist.Phase, &hist.Year)
		if err != nil {
			t.Errorf("Failed to scan game history: %v", err)
		}
		retrievedHistory = append(retrievedHistory, hist)
	}

	if len(retrievedHistory) != 1 {
		t.Errorf("Expected 1 history entry, got %d", len(retrievedHistory))
	}

	if retrievedHistory[0].Phase != "spring_movement" {
		t.Errorf("Expected phase 'spring_movement', got '%s'", retrievedHistory[0].Phase)
	}
}

func TestTransactionRollback(t *testing.T) {
	config := Config{
		DatabasePath: ":memory:",
		LogQueries:   false,
	}

	db, err := NewDatabase(config)
	if err != nil {
		t.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	// Test transaction rollback on error
	err = db.Transaction(func(tx *sql.Tx) error {
		// Create a game
		_, err := tx.Exec(`
			INSERT INTO games (name, phase, year, status, created_at, updated_at) 
			VALUES (?, ?, ?, ?, ?, ?)`,
			"Test Game", "spring_movement", 1901, "in_progress", time.Now(), time.Now())
		if err != nil {
			return err
		}

		// Force an error to trigger rollback
		return tx.QueryRow("INVALID SQL").Err()
	})

	if err == nil {
		t.Error("Expected transaction to fail and rollback")
	}

	// Verify no game was created due to rollback
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM games").Scan(&count)
	if err != nil {
		t.Errorf("Failed to count games: %v", err)
	}
	if count != 0 {
		t.Errorf("Expected 0 games after rollback, got %d", count)
	}
}

func TestMain(m *testing.M) {
	// Run tests
	code := m.Run()

	// Cleanup any test database files
	os.Remove("test.db")

	os.Exit(code)
}
