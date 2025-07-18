package storage

import (
	"database/sql"
	"fmt"
	"time"
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

// OrderResult represents the result of resolving an order
type OrderResult struct {
	Result        string
	FailureReason string
}
