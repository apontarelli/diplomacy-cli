package storage

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"diplomacy-backend/internal/game"
	_ "github.com/mattn/go-sqlite3"
)

type DB struct {
	conn *sql.DB
}

func NewDB(dbPath string) (*DB, error) {
	if err := os.MkdirAll(filepath.Dir(dbPath), 0755); err != nil {
		return nil, fmt.Errorf("failed to create database directory: %w", err)
	}

	conn, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := conn.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	db := &DB{conn: conn}
	if err := db.initSchema(); err != nil {
		return nil, fmt.Errorf("failed to initialize schema: %w", err)
	}

	return db, nil
}

func (db *DB) Close() error {
	return db.conn.Close()
}

func (db *DB) Ping() error {
	return db.conn.Ping()
}

func (db *DB) initSchema() error {
	schema := `
	CREATE TABLE IF NOT EXISTS games (
		id TEXT PRIMARY KEY,
		name TEXT NOT NULL,
		status TEXT NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS turns (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		game_id TEXT NOT NULL,
		year INTEGER NOT NULL,
		season TEXT NOT NULL,
		phase TEXT NOT NULL,
		status TEXT NOT NULL,
		deadline TIMESTAMP,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (game_id) REFERENCES games(id)
	);

	CREATE TABLE IF NOT EXISTS players (
		id TEXT PRIMARY KEY,
		game_id TEXT NOT NULL,
		nation TEXT NOT NULL,
		status TEXT NOT NULL,
		FOREIGN KEY (game_id) REFERENCES games(id)
	);

	CREATE TABLE IF NOT EXISTS units (
		id TEXT PRIMARY KEY,
		game_id TEXT NOT NULL,
		owner_id TEXT NOT NULL,
		unit_type TEXT NOT NULL,
		territory_id TEXT NOT NULL,
		FOREIGN KEY (game_id) REFERENCES games(id),
		FOREIGN KEY (owner_id) REFERENCES players(id)
	);

	CREATE TABLE IF NOT EXISTS territory_state (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		game_id TEXT NOT NULL,
		turn_id INTEGER NOT NULL,
		territory_id TEXT NOT NULL,
		owner_id TEXT,
		FOREIGN KEY (game_id) REFERENCES games(id),
		FOREIGN KEY (turn_id) REFERENCES turns(id),
		FOREIGN KEY (owner_id) REFERENCES players(id)
	);

	CREATE TABLE IF NOT EXISTS orders (
		id TEXT PRIMARY KEY,
		game_id TEXT NOT NULL,
		turn_id INTEGER NOT NULL,
		player_id TEXT NOT NULL,
		unit_id TEXT NOT NULL,
		order_type TEXT NOT NULL,
		from_territory TEXT NOT NULL,
		to_territory TEXT,
		support_unit TEXT,
		status TEXT NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (game_id) REFERENCES games(id),
		FOREIGN KEY (turn_id) REFERENCES turns(id),
		FOREIGN KEY (player_id) REFERENCES players(id),
		FOREIGN KEY (unit_id) REFERENCES units(id)
	);
	`

	_, err := db.conn.Exec(schema)
	return err
}

func (db *DB) CreateGame(g game.Game) error {
	query := `INSERT INTO games (id, name, status, created_at) VALUES (?, ?, ?, ?)`
	_, err := db.conn.Exec(query, g.ID, g.Name, string(g.Status), g.CreatedAt)
	return err
}

func (db *DB) GetGame(gameID string) (*game.Game, error) {
	query := `SELECT id, name, status, created_at FROM games WHERE id = ?`
	row := db.conn.QueryRow(query, gameID)

	var g game.Game
	var status string
	err := row.Scan(&g.ID, &g.Name, &status, &g.CreatedAt)
	if err != nil {
		return nil, err
	}

	g.Status = game.GameStatus(status)
	return &g, nil
}

func (db *DB) CreatePlayer(p game.Player) error {
	query := `INSERT INTO players (id, game_id, nation, status) VALUES (?, ?, ?, ?)`
	_, err := db.conn.Exec(query, p.ID, p.GameID, p.Nation, string(p.Status))
	return err
}

func (db *DB) GetPlayersByGame(gameID string) ([]game.Player, error) {
	query := `SELECT id, game_id, nation, status FROM players WHERE game_id = ?`
	rows, err := db.conn.Query(query, gameID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var players []game.Player
	for rows.Next() {
		var p game.Player
		var status string
		err := rows.Scan(&p.ID, &p.GameID, &p.Nation, &status)
		if err != nil {
			return nil, err
		}
		p.Status = game.PlayerStatus(status)
		players = append(players, p)
	}

	return players, rows.Err()
}

func (db *DB) GetAssignedNations(gameID string) ([]string, error) {
	query := `SELECT nation FROM players WHERE game_id = ?`
	rows, err := db.conn.Query(query, gameID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var nations []string
	for rows.Next() {
		var nation string
		err := rows.Scan(&nation)
		if err != nil {
			return nil, err
		}
		nations = append(nations, nation)
	}

	return nations, rows.Err()
}

func (db *DB) GetGameState(gameID string) (*game.GameState, error) {
	gameInfo, err := db.GetGame(gameID)
	if err != nil {
		return nil, err
	}

	players, err := db.GetPlayersByGame(gameID)
	if err != nil {
		return nil, err
	}

	units, err := db.GetUnitsByGame(gameID)
	if err != nil {
		return nil, err
	}

	territories, err := db.GetTerritoriesByGame(gameID)
	if err != nil {
		return nil, err
	}

	var orders []game.Order
	currentTurn, err := db.GetCurrentTurn(gameID)
	if err == nil {
		orders, _ = db.GetOrdersByTurn(gameID, currentTurn.ID)
	}

	return &game.GameState{
		Game:        *gameInfo,
		Players:     players,
		Units:       units,
		Territories: territories,
		Orders:      orders,
	}, nil
}

func (db *DB) GetUnitsByGame(gameID string) ([]game.Unit, error) {
	query := `SELECT id, game_id, owner_id, unit_type, territory_id FROM units WHERE game_id = ?`
	rows, err := db.conn.Query(query, gameID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var units []game.Unit
	for rows.Next() {
		var u game.Unit
		var unitType string
		err := rows.Scan(&u.ID, &u.GameID, &u.OwnerID, &unitType, &u.TerritoryID)
		if err != nil {
			return nil, err
		}
		u.UnitType = game.UnitType(unitType)
		units = append(units, u)
	}

	return units, rows.Err()
}

func (db *DB) GetTerritoriesByGame(gameID string) ([]game.TerritoryState, error) {
	query := `SELECT id, game_id, turn_id, territory_id, owner_id FROM territory_state WHERE game_id = ?`
	rows, err := db.conn.Query(query, gameID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var territories []game.TerritoryState
	for rows.Next() {
		var t game.TerritoryState
		var ownerID sql.NullString
		err := rows.Scan(&t.ID, &t.GameID, &t.TurnID, &t.TerritoryID, &ownerID)
		if err != nil {
			return nil, err
		}
		if ownerID.Valid {
			t.OwnerID = ownerID.String
		}
		territories = append(territories, t)
	}

	return territories, rows.Err()
}

func (db *DB) CreateTurn(t game.Turn) error {
	query := `INSERT INTO turns (game_id, year, season, phase, status, deadline, created_at) VALUES (?, ?, ?, ?, ?, ?, ?)`
	_, err := db.conn.Exec(query, t.GameID, t.Year, string(t.Season), string(t.Phase), string(t.Status), t.Deadline, t.CreatedAt)
	return err
}

func (db *DB) GetCurrentTurn(gameID string) (*game.Turn, error) {
	query := `SELECT id, game_id, year, season, phase, status, deadline, created_at FROM turns WHERE game_id = ? AND status = ? ORDER BY id DESC LIMIT 1`
	row := db.conn.QueryRow(query, gameID, string(game.TurnStatusActive))

	var t game.Turn
	var season, phase, status string
	var deadline sql.NullTime
	err := row.Scan(&t.ID, &t.GameID, &t.Year, &season, &phase, &status, &deadline, &t.CreatedAt)
	if err != nil {
		return nil, err
	}

	t.Season = game.Season(season)
	t.Phase = game.Phase(phase)
	t.Status = game.TurnStatus(status)
	if deadline.Valid {
		t.Deadline = &deadline.Time
	}

	return &t, nil
}

func (db *DB) GetTurnsByGame(gameID string) ([]game.Turn, error) {
	query := `SELECT id, game_id, year, season, phase, status, deadline, created_at FROM turns WHERE game_id = ? ORDER BY id`
	rows, err := db.conn.Query(query, gameID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var turns []game.Turn
	for rows.Next() {
		var t game.Turn
		var season, phase, status string
		var deadline sql.NullTime
		err := rows.Scan(&t.ID, &t.GameID, &t.Year, &season, &phase, &status, &deadline, &t.CreatedAt)
		if err != nil {
			return nil, err
		}

		t.Season = game.Season(season)
		t.Phase = game.Phase(phase)
		t.Status = game.TurnStatus(status)
		if deadline.Valid {
			t.Deadline = &deadline.Time
		}
		turns = append(turns, t)
	}

	return turns, rows.Err()
}

func (db *DB) CreateOrder(o game.Order) error {
	query := `INSERT INTO orders (id, game_id, turn_id, player_id, unit_id, order_type, from_territory, to_territory, support_unit, status, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	_, err := db.conn.Exec(query, o.ID, o.GameID, o.TurnID, o.PlayerID, o.UnitID, string(o.OrderType), o.FromTerritory, o.ToTerritory, o.SupportUnit, string(o.Status), o.CreatedAt, o.UpdatedAt)
	return err
}

func (db *DB) GetOrdersByTurn(gameID string, turnID int) ([]game.Order, error) {
	query := `SELECT id, game_id, turn_id, player_id, unit_id, order_type, from_territory, to_territory, support_unit, status, created_at, updated_at FROM orders WHERE game_id = ? AND turn_id = ?`
	rows, err := db.conn.Query(query, gameID, turnID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []game.Order
	for rows.Next() {
		var o game.Order
		var orderType, status string
		var toTerritory, supportUnit sql.NullString
		err := rows.Scan(&o.ID, &o.GameID, &o.TurnID, &o.PlayerID, &o.UnitID, &orderType, &o.FromTerritory, &toTerritory, &supportUnit, &status, &o.CreatedAt, &o.UpdatedAt)
		if err != nil {
			return nil, err
		}

		o.OrderType = game.OrderType(orderType)
		o.Status = game.OrderStatus(status)
		if toTerritory.Valid {
			o.ToTerritory = toTerritory.String
		}
		if supportUnit.Valid {
			o.SupportUnit = supportUnit.String
		}
		orders = append(orders, o)
	}

	return orders, rows.Err()
}

func (db *DB) GetOrdersByPlayer(gameID string, turnID int, playerID string) ([]game.Order, error) {
	query := `SELECT id, game_id, turn_id, player_id, unit_id, order_type, from_territory, to_territory, support_unit, status, created_at, updated_at FROM orders WHERE game_id = ? AND turn_id = ? AND player_id = ?`
	rows, err := db.conn.Query(query, gameID, turnID, playerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []game.Order
	for rows.Next() {
		var o game.Order
		var orderType, status string
		var toTerritory, supportUnit sql.NullString
		err := rows.Scan(&o.ID, &o.GameID, &o.TurnID, &o.PlayerID, &o.UnitID, &orderType, &o.FromTerritory, &toTerritory, &supportUnit, &status, &o.CreatedAt, &o.UpdatedAt)
		if err != nil {
			return nil, err
		}

		o.OrderType = game.OrderType(orderType)
		o.Status = game.OrderStatus(status)
		if toTerritory.Valid {
			o.ToTerritory = toTerritory.String
		}
		if supportUnit.Valid {
			o.SupportUnit = supportUnit.String
		}
		orders = append(orders, o)
	}

	return orders, rows.Err()
}

func (db *DB) UpdateOrderStatus(orderID string, status game.OrderStatus) error {
	query := `UPDATE orders SET status = ?, updated_at = ? WHERE id = ?`
	_, err := db.conn.Exec(query, string(status), time.Now(), orderID)
	return err
}

func (db *DB) DeleteOrder(orderID string) error {
	query := `DELETE FROM orders WHERE id = ?`
	_, err := db.conn.Exec(query, orderID)
	return err
}

func (db *DB) UpdateUnitPosition(unitID, newTerritoryID string) error {
	query := `UPDATE units SET territory_id = ? WHERE id = ?`
	_, err := db.conn.Exec(query, newTerritoryID, unitID)
	return err
}

func (db *DB) GetOrderByID(orderID string) (*game.Order, error) {
	query := `SELECT id, game_id, turn_id, player_id, unit_id, order_type, from_territory, to_territory, support_unit, status, created_at, updated_at FROM orders WHERE id = ?`
	row := db.conn.QueryRow(query, orderID)

	var o game.Order
	var orderType, status string
	var toTerritory, supportUnit sql.NullString
	err := row.Scan(&o.ID, &o.GameID, &o.TurnID, &o.PlayerID, &o.UnitID, &orderType, &o.FromTerritory, &toTerritory, &supportUnit, &status, &o.CreatedAt, &o.UpdatedAt)
	if err != nil {
		return nil, err
	}

	o.OrderType = game.OrderType(orderType)
	o.Status = game.OrderStatus(status)
	if toTerritory.Valid {
		o.ToTerritory = toTerritory.String
	}
	if supportUnit.Valid {
		o.SupportUnit = supportUnit.String
	}

	return &o, nil
}
func (db *DB) UpdateTurnStatus(turnID int, status game.TurnStatus) error {
	query := `UPDATE turns SET status = ? WHERE id = ?`
	_, err := db.conn.Exec(query, string(status), turnID)
	return err
}

func (db *DB) CreateUnit(u game.Unit) error {
	query := `INSERT INTO units (id, game_id, owner_id, unit_type, territory_id) VALUES (?, ?, ?, ?, ?)`
	_, err := db.conn.Exec(query, u.ID, u.GameID, u.OwnerID, string(u.UnitType), u.TerritoryID)
	return err
}

func (db *DB) CreateTerritoryState(ts game.TerritoryState) error {
	query := `INSERT INTO territory_state (game_id, turn_id, territory_id, owner_id) VALUES (?, ?, ?, ?)`
	_, err := db.conn.Exec(query, ts.GameID, ts.TurnID, ts.TerritoryID, ts.OwnerID)
	return err
}

func (db *DB) UpdateGameStatus(gameID string, status game.GameStatus) error {
	query := `UPDATE games SET status = ? WHERE id = ?`
	_, err := db.conn.Exec(query, string(status), gameID)
	return err
}

func (db *DB) GetPlayerByNation(gameID, nation string) (*game.Player, error) {
	query := `SELECT id, game_id, nation, status FROM players WHERE game_id = ? AND nation = ?`
	row := db.conn.QueryRow(query, gameID, nation)

	var p game.Player
	var status string
	err := row.Scan(&p.ID, &p.GameID, &p.Nation, &status)
	if err != nil {
		return nil, err
	}

	p.Status = game.PlayerStatus(status)
	return &p, nil
}

func (db *DB) GetExpiredTurns() ([]game.Turn, error) {
	query := `SELECT id, game_id, year, season, phase, status, deadline, created_at FROM turns WHERE status = ? AND deadline < ?`
	rows, err := db.conn.Query(query, string(game.TurnStatusActive), time.Now())
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var turns []game.Turn
	for rows.Next() {
		var t game.Turn
		var season, phase, status string
		var deadline sql.NullTime
		err := rows.Scan(&t.ID, &t.GameID, &t.Year, &season, &phase, &status, &deadline, &t.CreatedAt)
		if err != nil {
			return nil, err
		}

		t.Season = game.Season(season)
		t.Phase = game.Phase(phase)
		t.Status = game.TurnStatus(status)
		if deadline.Valid {
			t.Deadline = &deadline.Time
		}
		turns = append(turns, t)
	}

	return turns, rows.Err()
}