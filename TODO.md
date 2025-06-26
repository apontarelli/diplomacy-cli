# Current Focus: Core System Architecture (Go Backend)

## Phase 1: Backend Scaffolding

- [ ] Initialize a Go module within the `backend/` directory.
- [ ] Add SQLite dependency to the Go module.
- [ ] Implement a storage layer in `internal/storage/` that:
  - [ ] Initializes a SQLite database connection.
  - [ ] Creates the database schema if it doesn't exist.
- [ ] Define core game state data structures in `internal/game/`.
- [ ] Set up a basic `net/http` server in `cmd/server/main.go`.
- [ ] Implement a `/health` endpoint that confirms the server is running and can connect to the database.

---

# Future Scope

- [ ] Implement the Go backend, including the resolution engine and API.
- [ ] Implement a TUI for the game.
- [ ] Implement AI agents for the game.
- [ ] Implement a messaging system for players.

---

# Archive

## Completed Python Tasks

- **Phase 1: Core Scaffolding**
  - [x] `main.py` with load → validate → run → save loop
  - [x] Placeholder logic modules: `storage.py`, `validate.py`, `engine.py`, `state.py`
- **Phase 2: Logic & Testing**
  - [x] State Management
  - [x] Unit Management
  - [x] Territory Management
  - [x] Refactor Static Starting State to Dynamic Build
  - [x] Units and Movement Refactor
  - [x] Order Handling
  - [x] Resolution Engine
- **Miscellaneous Tasks**
  - [x] Tooling & Environment
