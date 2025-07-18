package storage

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

// Database wraps the SQL database connection
type Database struct {
	*sql.DB
}

// Config holds database configuration
type Config struct {
	DatabasePath string
	LogQueries   bool
}

// NewDatabase creates a new database connection with schema initialization
func NewDatabase(config Config) (*Database, error) {
	if config.DatabasePath == "" {
		config.DatabasePath = "diplomacy.db"
	}

	db, err := sql.Open("sqlite3", config.DatabasePath+"?_foreign_keys=on")
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	database := &Database{DB: db}

	// Test connection
	if err := database.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// Create schema
	if err := database.CreateSchema(); err != nil {
		return nil, fmt.Errorf("failed to create schema: %w", err)
	}

	// Create indexes
	if err := database.CreateIndexes(); err != nil {
		return nil, fmt.Errorf("failed to create indexes: %w", err)
	}

	log.Println("Database initialized successfully")
	return database, nil
}

// CreateSchema creates all database tables
func (db *Database) CreateSchema() error {
	schema := `
	-- Users table
	CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		username TEXT UNIQUE NOT NULL,
		email TEXT UNIQUE NOT NULL,
		password_hash TEXT NOT NULL,
		display_name TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	-- Games table
	CREATE TABLE IF NOT EXISTS games (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		phase TEXT NOT NULL DEFAULT 'spring_movement',
		year INTEGER NOT NULL DEFAULT 1901,
		status TEXT NOT NULL DEFAULT 'waiting_for_players',
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	-- Players table
	CREATE TABLE IF NOT EXISTS players (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		game_id INTEGER NOT NULL,
		user_id TEXT NOT NULL,
		nation TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (game_id) REFERENCES games(id) ON DELETE CASCADE,
		UNIQUE(game_id, nation)
	);

	-- Units table
	CREATE TABLE IF NOT EXISTS units (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		game_id INTEGER NOT NULL,
		player_id INTEGER,
		type TEXT NOT NULL,
		province TEXT NOT NULL,
		coast TEXT,
		owner TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (game_id) REFERENCES games(id) ON DELETE CASCADE,
		FOREIGN KEY (player_id) REFERENCES players(id) ON DELETE SET NULL,
		UNIQUE(game_id, province)
	);

	-- Orders table
	CREATE TABLE IF NOT EXISTS orders (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		game_id INTEGER NOT NULL,
		player_id INTEGER NOT NULL,
		unit_id INTEGER,
		type TEXT NOT NULL,
		target TEXT,
		coast TEXT,
		resolved BOOLEAN NOT NULL DEFAULT FALSE,
		result TEXT,
		support_target TEXT,
		support_destination TEXT,
		convoy_target TEXT,
		failure_reason TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (game_id) REFERENCES games(id) ON DELETE CASCADE,
		FOREIGN KEY (player_id) REFERENCES players(id) ON DELETE CASCADE,
		FOREIGN KEY (unit_id) REFERENCES units(id) ON DELETE CASCADE
	);

	-- Supply centers table
	CREATE TABLE IF NOT EXISTS supply_centers (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		game_id INTEGER NOT NULL,
		province TEXT NOT NULL,
		owner TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (game_id) REFERENCES games(id) ON DELETE CASCADE,
		UNIQUE(game_id, province)
	);

	-- Game history table
	CREATE TABLE IF NOT EXISTS game_histories (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		game_id INTEGER NOT NULL,
		phase TEXT NOT NULL,
		year INTEGER NOT NULL,
		state_data TEXT NOT NULL,
		timestamp DATETIME NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (game_id) REFERENCES games(id) ON DELETE CASCADE
	);

	-- Triggers to update updated_at timestamps
	CREATE TRIGGER IF NOT EXISTS update_users_updated_at 
		AFTER UPDATE ON users 
		BEGIN 
			UPDATE users SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id; 
		END;

	CREATE TRIGGER IF NOT EXISTS update_games_updated_at 
		AFTER UPDATE ON games 
		BEGIN 
			UPDATE games SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id; 
		END;

	CREATE TRIGGER IF NOT EXISTS update_players_updated_at 
		AFTER UPDATE ON players 
		BEGIN 
			UPDATE players SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id; 
		END;

	CREATE TRIGGER IF NOT EXISTS update_units_updated_at 
		AFTER UPDATE ON units 
		BEGIN 
			UPDATE units SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id; 
		END;

	CREATE TRIGGER IF NOT EXISTS update_orders_updated_at 
		AFTER UPDATE ON orders 
		BEGIN 
			UPDATE orders SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id; 
		END;

	CREATE TRIGGER IF NOT EXISTS update_supply_centers_updated_at 
		AFTER UPDATE ON supply_centers 
		BEGIN 
			UPDATE supply_centers SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id; 
		END;

	CREATE TRIGGER IF NOT EXISTS update_game_histories_updated_at 
		AFTER UPDATE ON game_histories 
		BEGIN 
			UPDATE game_histories SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id; 
		END;
	`

	_, err := db.Exec(schema)
	return err
}

// CreateIndexes creates performance indexes
func (db *Database) CreateIndexes() error {
	indexes := []string{
		"CREATE INDEX IF NOT EXISTS idx_games_status_phase ON games(status, phase)",
		"CREATE INDEX IF NOT EXISTS idx_players_game_nation ON players(game_id, nation)",
		"CREATE INDEX IF NOT EXISTS idx_players_user_id ON players(user_id)",
		"CREATE INDEX IF NOT EXISTS idx_units_game_owner ON units(game_id, owner)",
		"CREATE INDEX IF NOT EXISTS idx_units_province ON units(province)",
		"CREATE INDEX IF NOT EXISTS idx_orders_game_resolved ON orders(game_id, resolved)",
		"CREATE INDEX IF NOT EXISTS idx_orders_player_id ON orders(player_id)",
		"CREATE INDEX IF NOT EXISTS idx_supply_centers_game_owner ON supply_centers(game_id, owner)",
		"CREATE INDEX IF NOT EXISTS idx_game_histories_game_phase_year ON game_histories(game_id, phase, year)",
	}

	for _, index := range indexes {
		if _, err := db.Exec(index); err != nil {
			return fmt.Errorf("failed to create index: %w", err)
		}
	}

	return nil
}

// Transaction executes a function within a database transaction
func (db *Database) Transaction(fn func(*sql.Tx) error) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}

	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p)
		} else if err != nil {
			tx.Rollback()
		} else {
			err = tx.Commit()
		}
	}()

	err = fn(tx)
	return err
}

// Health checks database connectivity
func (db *Database) Health() error {
	return db.Ping()
}
