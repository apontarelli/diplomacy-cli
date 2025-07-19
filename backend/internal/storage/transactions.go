package storage

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"diplomacy-cli/backend/internal/game"
)

// GameOperations provides transactional operations for game management
type GameOperations struct {
	db *Database
}

// NewGameOperations creates a new GameOperations instance
func NewGameOperations(db *Database) *GameOperations {
	return &GameOperations{db: db}
}

// CreateGameWithPlayers creates a new game and adds players in a single transaction
func (ops *GameOperations) CreateGameWithPlayers(game *Game, players []Player) error {
	return ops.db.Transaction(func(tx *sql.Tx) error {
		// Create the game
		result, err := tx.Exec(`
			INSERT INTO games (name, phase, year, status, created_at, updated_at) 
			VALUES (?, ?, ?, ?, ?, ?)`,
			game.Name, game.Phase, game.Year, game.Status, time.Now(), time.Now())
		if err != nil {
			return fmt.Errorf("failed to create game: %w", err)
		}

		gameID, err := result.LastInsertId()
		if err != nil {
			return fmt.Errorf("failed to get game ID: %w", err)
		}
		game.ID = gameID

		// Add players to the game
		for i := range players {
			players[i].GameID = gameID
			result, err := tx.Exec(`
				INSERT INTO players (game_id, user_id, nation, created_at, updated_at)
				VALUES (?, ?, ?, ?, ?)`,
				gameID, players[i].UserID, players[i].Nation, time.Now(), time.Now())
			if err != nil {
				return fmt.Errorf("failed to create player %s: %w", players[i].Nation, err)
			}

			playerID, err := result.LastInsertId()
			if err != nil {
				return fmt.Errorf("failed to get player ID: %w", err)
			}
			players[i].ID = playerID
		}

		return nil
	})
}

// SubmitOrders submits multiple orders for a game in a single transaction
func (ops *GameOperations) SubmitOrders(gameID int64, orders []Order) error {
	return ops.db.Transaction(func(tx *sql.Tx) error {
		// Verify game exists and is in progress
		var game Game
		err := tx.QueryRow("SELECT id, name, phase, year, status FROM games WHERE id = ?", gameID).
			Scan(&game.ID, &game.Name, &game.Phase, &game.Year, &game.Status)
		if err != nil {
			if err == sql.ErrNoRows {
				return fmt.Errorf("game not found")
			}
			return fmt.Errorf("failed to query game: %w", err)
		}

		if game.Status != "in_progress" {
			return fmt.Errorf("cannot submit orders for game in status: %s", game.Status)
		}

		// Clear existing unresolved orders for this game
		_, err = tx.Exec("DELETE FROM orders WHERE game_id = ? AND resolved = FALSE", gameID)
		if err != nil {
			return fmt.Errorf("failed to clear existing orders: %w", err)
		}

		// Create new orders
		for i := range orders {
			orders[i].GameID = gameID
			orders[i].Resolved = false
			result, err := tx.Exec(`
				INSERT INTO orders (game_id, player_id, unit_id, type, target, coast, resolved, 
					support_target, support_destination, convoy_target, created_at, updated_at)
				VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
				gameID, orders[i].PlayerID, orders[i].UnitID, orders[i].Type, orders[i].Target,
				orders[i].Coast, false, orders[i].SupportTarget, orders[i].SupportDestination,
				orders[i].ConvoyTarget, time.Now(), time.Now())
			if err != nil {
				return fmt.Errorf("failed to create order: %w", err)
			}

			orderID, err := result.LastInsertId()
			if err != nil {
				return fmt.Errorf("failed to get order ID: %w", err)
			}
			orders[i].ID = orderID
		}

		return nil
	})
}

// ResolveOrders marks orders as resolved and updates their results
func (ops *GameOperations) ResolveOrders(gameID int64, orderResults map[int64]OrderResult) error {
	return ops.db.Transaction(func(tx *sql.Tx) error {
		for orderID, result := range orderResults {
			_, err := tx.Exec(`
				UPDATE orders 
				SET resolved = TRUE, result = ?, failure_reason = ?, updated_at = ?
				WHERE id = ? AND game_id = ?`,
				result.Result, result.FailureReason, time.Now(), orderID, gameID)
			if err != nil {
				return fmt.Errorf("failed to update order %d: %w", orderID, err)
			}
		}
		return nil
	})
}

// UpdateGameState updates game phase, year, and related state in a transaction
func (ops *GameOperations) UpdateGameState(gameID int64, phase string, year int, units []Unit, supplyCenters []SupplyCenter) error {
	return ops.db.Transaction(func(tx *sql.Tx) error {
		// Update game phase and year
		_, err := tx.Exec(`
			UPDATE games 
			SET phase = ?, year = ?, updated_at = ?
			WHERE id = ?`,
			phase, year, time.Now(), gameID)
		if err != nil {
			return fmt.Errorf("failed to update game state: %w", err)
		}

		// Clear existing units for this game
		_, err = tx.Exec("DELETE FROM units WHERE game_id = ?", gameID)
		if err != nil {
			return fmt.Errorf("failed to clear existing units: %w", err)
		}

		// Create new units
		for i := range units {
			units[i].GameID = gameID
			result, err := tx.Exec(`
				INSERT INTO units (game_id, player_id, type, province, coast, owner, created_at, updated_at)
				VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
				gameID, units[i].PlayerID, units[i].Type, units[i].Province,
				units[i].Coast, units[i].Owner, time.Now(), time.Now())
			if err != nil {
				return fmt.Errorf("failed to create unit: %w", err)
			}

			unitID, err := result.LastInsertId()
			if err != nil {
				return fmt.Errorf("failed to get unit ID: %w", err)
			}
			units[i].ID = unitID
		}

		// Clear existing supply centers for this game
		_, err = tx.Exec("DELETE FROM supply_centers WHERE game_id = ?", gameID)
		if err != nil {
			return fmt.Errorf("failed to clear existing supply centers: %w", err)
		}

		// Create new supply centers
		for i := range supplyCenters {
			supplyCenters[i].GameID = gameID
			result, err := tx.Exec(`
				INSERT INTO supply_centers (game_id, province, owner, created_at, updated_at)
				VALUES (?, ?, ?, ?, ?)`,
				gameID, supplyCenters[i].Province, supplyCenters[i].Owner, time.Now(), time.Now())
			if err != nil {
				return fmt.Errorf("failed to create supply center: %w", err)
			}

			centerID, err := result.LastInsertId()
			if err != nil {
				return fmt.Errorf("failed to get supply center ID: %w", err)
			}
			supplyCenters[i].ID = centerID
		}

		return nil
	})
}

// SaveGameHistory saves a game state snapshot to history
func (ops *GameOperations) SaveGameHistory(history *GameHistory) error {
	return ops.db.Transaction(func(tx *sql.Tx) error {
		result, err := tx.Exec(`
			INSERT INTO game_histories (game_id, phase, year, state_data, timestamp, created_at, updated_at)
			VALUES (?, ?, ?, ?, ?, ?, ?)`,
			history.GameID, history.Phase, history.Year, history.StateData,
			history.Timestamp, time.Now(), time.Now())
		if err != nil {
			return fmt.Errorf("failed to save game history: %w", err)
		}

		historyID, err := result.LastInsertId()
		if err != nil {
			return fmt.Errorf("failed to get history ID: %w", err)
		}
		history.ID = historyID

		return nil
	})
}

// CompleteGame marks a game as completed and performs cleanup
func (ops *GameOperations) CompleteGame(gameID int64, winner string) error {
	return ops.db.Transaction(func(tx *sql.Tx) error {
		// Update game status
		_, err := tx.Exec(`
			UPDATE games 
			SET status = 'completed', updated_at = ?
			WHERE id = ?`,
			time.Now(), gameID)
		if err != nil {
			return fmt.Errorf("failed to complete game: %w", err)
		}

		// Mark all unresolved orders as resolved
		_, err = tx.Exec(`
			UPDATE orders 
			SET resolved = TRUE, updated_at = ?
			WHERE game_id = ? AND resolved = FALSE`,
			time.Now(), gameID)
		if err != nil {
			return fmt.Errorf("failed to resolve remaining orders: %w", err)
		}

		return nil
	})
}

// SaveGameState saves a complete game state to the database using JSON serialization
func (ops *GameOperations) SaveGameState(gameID int64, gameState *game.GameState) error {
	return ops.db.Transaction(func(tx *sql.Tx) error {
		// Serialize the game state to JSON
		stateData, err := SerializeGameState(gameState)
		if err != nil {
			return fmt.Errorf("failed to serialize game state: %w", err)
		}

		// Save to game_histories table with current timestamp
		_, err = tx.Exec(`
			INSERT INTO game_histories (game_id, phase, year, state_data, timestamp, created_at, updated_at)
			VALUES (?, ?, ?, ?, ?, ?, ?)`,
			gameID, gameState.Phase, gameState.Year, string(stateData),
			time.Now(), time.Now(), time.Now())
		if err != nil {
			return fmt.Errorf("failed to save game state: %w", err)
		}

		return nil
	})
}

// LoadGameState loads the most recent game state from the database
func (ops *GameOperations) LoadGameState(gameID int64) (*game.GameState, error) {
	var gameState *game.GameState
	err := ops.db.Transaction(func(tx *sql.Tx) error {
		// Load the most recent game state
		var stateData string
		err := tx.QueryRow(`
			SELECT state_data 
			FROM game_histories 
			WHERE game_id = ? 
			ORDER BY timestamp DESC 
			LIMIT 1`, gameID).Scan(&stateData)
		if err != nil {
			if err == sql.ErrNoRows {
				return fmt.Errorf("no game state found for game %d", gameID)
			}
			return fmt.Errorf("failed to query game state: %w", err)
		}

		// Deserialize the game state
		gameState, err = DeserializeGameState([]byte(stateData))
		if err != nil {
			return fmt.Errorf("failed to deserialize game state: %w", err)
		}

		return nil
	})

	return gameState, err
}

// SaveGameStateSnapshot saves a game state snapshot with a specific timestamp
func (ops *GameOperations) SaveGameStateSnapshot(gameID int64, gameState *game.GameState, timestamp time.Time) error {
	return ops.db.Transaction(func(tx *sql.Tx) error {
		// Serialize the game state to JSON
		stateData, err := SerializeGameState(gameState)
		if err != nil {
			return fmt.Errorf("failed to serialize game state: %w", err)
		}

		// Save to game_histories table with specified timestamp
		_, err = tx.Exec(`
			INSERT INTO game_histories (game_id, phase, year, state_data, timestamp, created_at, updated_at)
			VALUES (?, ?, ?, ?, ?, ?, ?)`,
			gameID, gameState.Phase, gameState.Year, string(stateData),
			timestamp, time.Now(), time.Now())
		if err != nil {
			return fmt.Errorf("failed to save game state snapshot: %w", err)
		}

		return nil
	})
}

// LoadGameStateHistory loads all historical game states for a game with structured error handling
func (ops *GameOperations) LoadGameStateHistory(gameID int64) (*GameStateLoadResult, error) {
	var result *GameStateLoadResult
	err := ops.db.Transaction(func(tx *sql.Tx) error {
		var err error
		result, err = ops.loadGameHistoryWithFailures(tx, gameID)
		return err
	})

	return result, err
}

// LoadGameStateHistoryOrFail provides a convenience method that fails on any corruption
func (ops *GameOperations) LoadGameStateHistoryOrFail(gameID int64) ([]*game.GameState, error) {
	result, err := ops.LoadGameStateHistory(gameID)
	if err != nil {
		return nil, err
	}
	if len(result.Failures) > 0 {
		return nil, fmt.Errorf("failed to load %d of %d states for game %d",
			len(result.Failures), result.TotalCount, gameID)
	}
	return result.States, nil
}

// loadGameHistoryAndCurrent loads historical game states and separates current state
func (ops *GameOperations) loadGameHistoryAndCurrent(tx *sql.Tx, gameID int64) ([]*game.GameState, *game.GameState, error) {
	rows, err := tx.Query(`
		SELECT state_data, timestamp 
		FROM game_histories 
		WHERE game_id = ? 
		ORDER BY timestamp ASC`, gameID)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	var allStates []*game.GameState
	for rows.Next() {
		var stateData string
		var timestamp time.Time
		if err := rows.Scan(&stateData, &timestamp); err != nil {
			return nil, nil, err
		}

		if stateData != "" {
			gameState, err := DeserializeGameState([]byte(stateData))
			if err != nil {
				// Log error but continue - don't fail entire load for one corrupted state
				continue
			}
			allStates = append(allStates, gameState)
		}
	}

	if err := rows.Err(); err != nil {
		return nil, nil, err
	}

	// Separate history from current state
	// The last state is the current state, everything else is history
	var history []*game.GameState
	var currentState *game.GameState

	if len(allStates) > 0 {
		currentState = allStates[len(allStates)-1]
		if len(allStates) > 1 {
			history = allStates[:len(allStates)-1]
		}
	}

	return history, currentState, nil
}

// savePlayers saves all players for a game
func (ops *GameOperations) savePlayers(tx *sql.Tx, gameID int64, players map[game.Nation]string) error {
	// Clear existing players for this game
	_, err := tx.Exec("DELETE FROM players WHERE game_id = ?", gameID)
	if err != nil {
		return fmt.Errorf("failed to clear existing players: %w", err)
	}

	// Insert current players
	for nation, userID := range players {
		_, err := tx.Exec(`
			INSERT INTO players (game_id, user_id, nation, created_at, updated_at)
			VALUES (?, ?, ?, ?, ?)`,
			gameID, userID, string(nation), time.Now(), time.Now())
		if err != nil {
			return fmt.Errorf("failed to insert player %s: %w", nation, err)
		}
	}

	return nil
}

// loadGameHistory loads historical game states
func (ops *GameOperations) loadGameHistory(tx *sql.Tx, gameID int64) ([]*game.GameState, error) {
	rows, err := tx.Query(`
		SELECT state_data 
		FROM game_histories 
		WHERE game_id = ? 
		ORDER BY timestamp ASC`, gameID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var history []*game.GameState
	for rows.Next() {
		var stateData string
		if err := rows.Scan(&stateData); err != nil {
			return nil, err
		}

		if stateData != "" {
			gameState, err := DeserializeGameState([]byte(stateData))
			if err != nil {
				// Log error but continue - don't fail entire load for one corrupted state
				continue
			}
			history = append(history, gameState)
		}
	}

	return history, rows.Err()
}

// loadGameHistoryWithFailures loads historical game states with structured error handling
func (ops *GameOperations) loadGameHistoryWithFailures(tx *sql.Tx, gameID int64) (*GameStateLoadResult, error) {
	rows, err := tx.Query(`
		SELECT state_data, timestamp 
		FROM game_histories 
		WHERE game_id = ? 
		ORDER BY timestamp ASC`, gameID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := &GameStateLoadResult{
		States:   make([]*game.GameState, 0),
		Failures: make([]StateLoadFailure, 0),
	}

	for rows.Next() {
		var stateData string
		var timestamp time.Time
		if err := rows.Scan(&stateData, &timestamp); err != nil {
			return nil, err
		}

		result.TotalCount++

		if stateData != "" {
			gameState, err := DeserializeGameState([]byte(stateData))
			if err != nil {
				// Log the error and collect failure information
				log.Printf("Warning: Failed to deserialize game state for game %d at timestamp %v: %v",
					gameID, timestamp, err)

				failure := StateLoadFailure{
					GameID:    gameID,
					Timestamp: timestamp,
					Error:     err.Error(),
					RawData:   stateData, // Include raw data for debugging
				}
				result.Failures = append(result.Failures, failure)
				continue
			}
			result.States = append(result.States, gameState)
		}
	}

	return result, rows.Err()
}

// OrderResult represents the result of resolving an order
type OrderResult struct {
	Result        string
	FailureReason string
}

// GameStateLoadResult represents the result of loading game states with potential partial failures
type GameStateLoadResult struct {
	States     []*game.GameState  `json:"states"`
	Failures   []StateLoadFailure `json:"failures,omitempty"`
	TotalCount int                `json:"total_count"`
}

// StateLoadFailure represents a failure to load a specific game state
type StateLoadFailure struct {
	GameID    int64     `json:"game_id"`
	Timestamp time.Time `json:"timestamp"`
	Error     string    `json:"error"`
	RawData   string    `json:"raw_data,omitempty"` // For debugging corrupted data
}
