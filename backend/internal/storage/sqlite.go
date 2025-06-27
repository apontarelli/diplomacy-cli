package storage

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

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
		turn_code TEXT NOT NULL,
		state_json TEXT NOT NULL,
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