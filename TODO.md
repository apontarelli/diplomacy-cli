# Current Focus: Core System Architecture (Go Backend)

## Phase 1: Backend Scaffolding ✅

- [x] Initialize a Go module within the `backend/` directory.
- [x] Add SQLite dependency to the Go module (latest stable: v1.14.28).
- [x] Implement a storage layer in `internal/storage/` that:
  - [x] Initializes a SQLite database connection.
  - [x] Creates the database schema if it doesn't exist.
- [x] Define core game state data structures in `internal/game/` (clean domain types, no JSON tags).
- [x] Set up a basic `net/http` server in `cmd/server/main.go`.
- [x] Implement a `/health` endpoint that confirms the server is running and can connect to the database.



## Phase 2: Core Game Logic

**Sub-phases:**
- Phase 2A: Game Foundation ✅
- Phase 2B: Turn Management ✅
- Phase 2C: Order System (current)
- Phase 2D: Basic Resolution

### Phase 2A: Game Foundation ✅

- [x] Load classic Diplomacy game data (nations, territories, adjacencies)
- [x] Implement game creation endpoints with proper initialization
- [x] Add player registration and nation assignment
- [x] Basic game state query endpoints

### Phase 2B: Turn Management ✅

- [x] Implement turn creation and progression system
- [x] Add phase management (Spring/Fall seasons, Movement/Retreat/Build phases)
- [x] Create turn deadline and timing system
- [x] Add endpoints for turn operations (start turn, advance phase, etc.)
- [x] Initialize starting game state with units and territory ownership

### Phase 2C: Order System ✅

- [x] Implement order types (Move, Hold, Support, Convoy)
- [x] Add order validation logic (legal moves, unit ownership, adjacency)
- [x] Create order submission and retrieval endpoints
- [x] Implement order preview and modification system
- [x] Add order conflict detection and NMR handling

---

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
