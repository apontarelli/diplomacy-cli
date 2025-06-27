# Diplomacy Go Backend

This directory contains the Go source code for the Diplomacy game engine and server.

## Current Status

**Phase 1 Complete**: Basic backend scaffolding with SQLite storage, core domain types, and health endpoint.

## Project Structure

```
backend/
├── cmd/
│   └── server/
│       └── main.go       # Main application entry point
├── internal/
│   ├── game/             # Core game logic, state, and rules
│   │   └── types.go      # Domain types (Game, Player, Unit, etc.)
│   ├── storage/          # Database interaction logic
│   │   └── sqlite.go     # SQLite connection and schema
│   └── api/              # HTTP handlers and routing
│       └── handlers.go   # Health endpoint and server setup
└── go.mod                # Go module definition
```

### Architecture Principles

- **Data-Oriented Design**: Core domain types focus on game logic without serialization concerns
- **Clean Separation**: Storage, domain logic, and API layers are clearly separated
- **Minimal API Surface**: Only essential endpoints implemented; API design driven by actual game mechanics needs

### Directory Purpose

- **`cmd/server`**: Main executable that starts the HTTP server on port 8080 (configurable via PORT env var)
- **`internal/game`**: Core domain types and game logic (no JSON tags - pure data structures)
- **`internal/storage`**: SQLite database layer with automatic schema initialization
- **`internal/api`**: HTTP server with health endpoint (`/health`) for infrastructure verification

## Running the Server

```bash
cd backend
go run ./cmd/server
```

The server will:
- Start on port 8080 (or PORT environment variable)
- Create SQLite database at `~/.diplomacy/diplomacy.db` (or DB_PATH environment variable)
- Serve health check at `http://localhost:8080/health`

## Data Storage: SQLite

Uses **SQLite** for data integrity, ACID transactions, and structured querying without operational complexity.

### Database Schema

Automatically initialized on first run:

**`games`**
- `id` (TEXT, PRIMARY KEY) - e.g., a UUID
- `name` (TEXT)
- `status` (TEXT) - e.g., "forming", "active", "completed"
- `created_at` (TIMESTAMP)

**`turns`**
- `id` (INTEGER, PRIMARY KEY AUTOINCREMENT)
- `game_id` (TEXT, FOREIGN KEY to `games.id`)
- `turn_code` (TEXT) - e.g., "1901-S-M"
- `state_json` (TEXT) - A JSON blob of the full game state for this turn (for history/replay)
- `created_at` (TIMESTAMP)

**`players`**
- `id` (TEXT, PRIMARY KEY) - e.g., a UUID
- `game_id` (TEXT, FOREIGN KEY to `games.id`)
- `nation` (TEXT)
- `status` (TEXT)

**`units`**
- `id` (TEXT, PRIMARY KEY)
- `game_id` (TEXT, FOREIGN KEY to `games.id`)
- `owner_id` (TEXT, FOREIGN KEY to `players.id`)
- `unit_type` (TEXT)
- `territory_id` (TEXT)

**`territory_state`**
- `id` (INTEGER, PRIMARY KEY AUTOINCREMENT)
- `game_id` (TEXT, FOREIGN KEY to `games.id`)
- `turn_id` (INTEGER, FOREIGN KEY to `turns.id`)
- `territory_id` (TEXT)
- `owner_id` (TEXT, FOREIGN KEY to `players.id`)