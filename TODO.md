# Go Refactor Plan

## Implementation Decisions & Follow-up Tasks

### Phase 1 Decisions Made (Updated after Python PoC Analysis):
- **Domain Types**: No JSON tags on core domain structs (following architecture decision)
- **Order Types**: Implement all order types from start (Move, Hold, Support, Convoy) based on Python PoC learnings
- **Game Phases**: Begin with movement phase, add retreat/build phases in later iterations
- **Validation Pipeline**: Implement full Syntax → Semantic → Resolution pipeline from Python PoC
- **Resolution Engine**: Use multi-pass algorithm with SoA pattern for performance
- **Testing**: Comprehensive unit tests including DATC compliance

### Key Architectural Insights from Python PoC:
- **Validation Pipeline**: Raw orders → Syntax validation → Semantic validation → Resolution
- **Data-Oriented Design**: Structure of Arrays (SoA) for resolution performance
- **Multi-Pass Resolution**: Iterative algorithm with convoy path discovery and conflict resolution
- **Orchestrator Pattern**: Coordinates entire validation and resolution pipeline
- **Immutable State**: Game states with explicit transitions and history tracking

### Phase 1: The Core Domain (The "Engine") - Enhanced Approach

This phase builds a complete Diplomacy rules engine in isolation, implementing the full validation and resolution pipeline learned from the Python PoC.

*   **Packages to Build**: `internal/game/` with comprehensive sub-packages
*   **Order of Operations**:
    1.  **Define Core Domain** (`board.go`, `state.go`): Complete structs with all order types, outcomes, and validation result types
    2.  **Implement Map Loader** (`loader/`): Interface and static loader for classic map 
    3.  **Build Validation Pipeline** (`validation/`):
        - **Syntax Validation**: Parse raw order strings into structured orders
        - **Semantic Validation**: Validate orders against game rules and current state  
        - **Orchestrator**: Coordinate the validation pipeline
    4.  **Implement Resolution Engine** (`resolution/`):
        - **SoA Data Structures**: Structure of Arrays for performance
        - **Multi-Pass Algorithm**: Iterative resolution with convoy discovery
        - **Support & Strength**: Support cutting and strength calculation
        - **Conflict Resolution**: Bouncing, dislodgement, and final outcomes
    5.  **Comprehensive Testing**: Unit tests, integration tests, and DATC compliance

*   **Milestone**: A complete Diplomacy rules engine that can process raw order strings through syntax validation, semantic validation, and multi-pass resolution to produce the next game state with full outcome reporting.

### Phase 2: Persistence (Making the Engine's State "Real")

Now you connect your abstract game engine to a concrete database.

*   **Packages to Build**: `internal/model/` and `internal/storage/`
*   **Order of Operations**:
    1.  **Define DB Models**: In `model/`, create the `User` and `Game` structs that will map to your database tables. You'll need to decide how to store the complex `game.State` (a common solution is to serialize it to a JSON string and store it in a single text column in your `games` table).
    2.  **Define Storage Interfaces**: In `storage/store.go`, define the `UserStore` and `GameStore` interfaces (e.g., `CreateGame(g *model.Game)`, `GetGameByID(id string)`).
    3.  **Implement Concrete Store**: In `storage/sqlite/store.go`, write the actual SQL code to implement the interfaces.
    4.  **Write Integration Tests**: Your `storage/sqlite/store_test.go` will be an integration test. It will need to connect to a real (or in-memory) SQLite database to verify that your SQL queries work as expected.

*   **Milestone**: You have a data layer that can create, read, and update users and games in a database.

### Phase 3: The API & Auth (Exposing the Engine to the World)

This is where you build the web-facing part and tie everything together. It's often best to build `auth` and `api` in parallel, as most API endpoints will require authentication.

*   **Packages to Build**: `internal/api/`, `internal/auth/`, and `cmd/server/main.go`
*   **Order of Operations**:
    1.  **Basic Server Wiring**: Create `cmd/server/main.go` and `internal/api/router.go`. Get a basic HTTP server running that can respond to a simple `/health` check.
    2.  **Implement Auth**: Build the `internal/auth/` package for password hashing and JWT management.
    3.  **Build User Endpoints**: In `handlers.go`, create the `POST /register` and `POST /login` endpoints. These will use your `auth` package and the `UserStore` from Phase 2.
    4.  **Build Protected Game Endpoints**: Now, use your auth middleware. Implement the core game endpoints like `POST /games` (to create a new game) and `GET /games/{id}` (to view game state). These handlers will orchestrate calls to both your `storage` and `game` packages.
    5.  **Document with OpenAPI**: As you build endpoints, keep your `api/openapi.yaml` file updated.

*   **Milestone**: A running API server. You can use a tool like Postman or `curl` to register a user, log in, create a game, and view its state.

### Phase 4: The Frontend (Giving the Game a "Face")

This is technically outside the scope of the Go project directory, but it's the final piece of the puzzle.

*   **Order of Operations**:
    1.  **Choose a Framework**: Select your frontend technology (React, Vue, Svelte, etc.).
    2.  **Build UI Components**: Create components for the map, order submission forms, game list, etc.
    3.  **Connect to the API**: Use the `openapi.yaml` specification as the definitive guide for how to interact with your backend. This contract ensures the frontend and backend can be developed independently.

*   **Milestone**: A playable web-based Diplomacy game!

This phased approach builds from the inside out, ensuring each layer rests on a solid, well-tested foundation. You're essentially building the engine (`game`), then the chassis and transmission (`storage`), then the dashboard and controls (`api`), and finally the car's body (`frontend`). It's a fantastic and rewarding project. Good luck

---

# Current Implementation Status & Next Steps

## ✅ Completed
### Phase 1.1: Core Domain Foundation
- **Core Domain Structs**: `board.go`, `state.go` with comprehensive types ✅
- **Board Management**: Province creation, unit management, neighbor relationships ✅
- **Game State**: Turn management, phase tracking, order handling ✅
- **Comprehensive Testing**: Full test coverage for board and state logic ✅

### Phase 1.2: Map Loader Implementation  
- **MapLoader Interface**: Clean abstraction for loading game maps ✅
- **JSONLoader**: Complete implementation for loading classic Diplomacy map ✅
- **Coast Handling**: Support for complex provinces with multiple coasts ✅
- **Comprehensive Testing**: Full test coverage including edge cases ✅
- **Classic Map Data**: Complete JSON data files for provinces, edges, nations, starting positions ✅

### Phase 1.3: Syntax Parser Foundation
- **Core Type System**: Token types, error handling, parser interfaces ✅
- **Lexer Implementation**: Complete tokenization with normalization and error detection ✅
- **Province Resolver**: Handle abbreviations and space-to-underscore conversion ✅
- **Comprehensive Testing**: Full test suite with 100% pass rate ✅
- **Coast Support**: Handle multi-coast provinces like `stp/sc` ✅
- **Abbreviation Support**: Keywords (`s` → SUPPORT) and unit types (`a` → UNIT_TYPE) ✅

### Phase 1.4: Order-Specific Parsers Implementation ✅
- **Single File Architecture**: All parsers in `parser.go` following Python PoC pattern ✅
- **Move Parser**: Handles `A par - bur`, `par - bur`, `F stp/sc - bot` with adjacency validation ✅
- **Hold Parser**: Handles `A par hold`, `par h` with coast support ✅
- **Support Parsers**: Both support hold (`A bur s par`) and support move (`A bur s par - pic`) ✅
- **Convoy Parser**: Handles `F eng c lon - bel` with fleet-only validation ✅
- **Build/Disband Parsers**: Handles `build army paris` and `disband army paris` ✅
- **Parser Registry**: Phase-aware parsing with fallback mechanism ✅
- **Comprehensive Testing**: 100% test coverage with integration tests ✅
- **Province Resolution**: Full abbreviation and coast support ✅
- **Error Handling**: Detailed error messages with context ✅

## 🎯 Current Focus
### Phase 1.5: Multi-Pass Resolution Engine

#### Task 1.5.1: Resolution Data Structures
**File**: `backend/internal/game/resolution/types.go` (new)
- Structure of Arrays (SoA) for performance
- ResolutionSoA with parallel arrays
- ResolutionMaps for efficient lookups

#### Task 1.5.2: Convoy Path Discovery
**File**: `backend/internal/game/resolution/convoy.go` (new)
- BFS-based convoy path finding
- Convoy chain validation
- Integration with move resolution

#### Task 1.5.3: Support System
**File**: `backend/internal/game/resolution/support.go` (new)
- Support cutting logic
- Strength calculation
- Support effectiveness determination

#### Task 1.5.4: Conflict Resolution
**File**: `backend/internal/game/resolution/conflict.go` (new)
- Move conflict detection
- Bouncing and dislodgement logic
- Final territory assignment

#### Task 1.5.5: Main Resolution Engine
**File**: `backend/internal/game/resolution/engine.go` (new)
- Multi-pass iterative algorithm
- Convoy path stabilization loop
- Integration of all components
- Final outcome assignment

#### Task 1.5.6: Resolution Tests
**Files**: `backend/internal/game/resolution/*_test.go`
- Component unit tests
- Integration tests
- DATC compliance tests

### Phase 1.6: Integration & Testing

#### Task 1.6.1: Full Pipeline Integration
- End-to-end tests (raw orders → final state)
- Multi-turn game progression
- Performance testing

#### Task 1.6.2: DATC Implementation
- Diplomacy Adjudicator Test Cases
- Automated compliance testing
- Regression test suite

---

## Pipeline Architecture Analysis & Decision

### Current Domain Model Reality Check
After reviewing the Go implementation, the existing `semantic.go` has **architectural mismatches**, not just missing types:

**What We Actually Have:**
- `GameState.RawOrders` - `map[Nation][]string` (raw order strings)
- `Province.CoastNeighbors` - `map[string][]string` (coast name -> neighbors)  
- `OrderResult` - already defined result types
- `Unit.Coast` - string field for coast specification

**What semantic.go Incorrectly Assumes:**
- `sv.state.Orders` ❌ - we have `RawOrders` instead
- `game.Coast` type ❌ - we use strings for coast names
- `toProvince.Coasts` field ❌ - we have `CoastNeighbors` map

### Correct Pipeline Flow
```
RawOrders (strings) → Syntax Parser → []*Order → Semantic Validator → SemanticResult
```

The semantic validator should work with **parsed orders**, not raw strings.

### Strategic Decision: Syntax First vs Semantic First

**Option A: Fix Semantic First**
- ❌ Semantic validator expects parsed `[]*Order` objects
- ❌ Without syntax parser, we can't test semantic validation properly
- ❌ Would require creating mock parsed orders for testing

**Option B: Build Syntax Parser First** ⭐ **RECOMMENDED**
- ✅ Syntax parser converts `RawOrders` → `[]*Order` 
- ✅ Enables proper testing of semantic validation with real parsed orders
- ✅ Follows natural data flow: raw → parsed → validated
- ✅ Semantic validator can be designed knowing exact `Order` structure

## Next Immediate Task
**Task 1.5.1**: Implement Resolution Data Structures in `backend/internal/game/resolution/types.go`

**Rationale**: With parsing complete, we need the resolution engine foundation. The Structure of Arrays (SoA) pattern from the Python PoC enables:
1. High-performance resolution with cache-friendly data access
2. Efficient batch operations on orders
3. Clean separation between resolution data and domain objects
4. Foundation for the multi-pass iterative algorithm

**Implementation Approach**: 
- Create ResolutionSoA with parallel arrays for orders, outcomes, strengths
- Implement ResolutionMaps for efficient lookups by province/unit
- Design for the multi-pass convoy discovery algorithm
- Ensure compatibility with existing Order and GameState types

**After Resolution Types**: Implement convoy path discovery, support system, and conflict resolution components.

---

# Current Proposed Go Structure
```
diplomacy-game/
├── api/
│   └── openapi.yaml
├── cmd/
│   └── server/
│       └── main.go
├── configs/
│   ├── config.yaml
│   └── maps/
│       └── classic.json
├── internal/
│   ├── api/
│   │   ├── handlers.go
│   │   ├── handlers_test.go
│   │   ├── middleware.go
│   │   └── router.go
│   ├── auth/
│   │   ├── jwt.go
│   │   └── password.go
│   ├── game/
│   │   ├── board.go
│   │   ├── loader/
│   │   │   ├── interface.go
│   │   │   └── static_loader.go
│   │   ├── resolution/
│   │   │   ├── adjudicator.go
│   │   │   └── adjudicator_test.go
│   │   ├── state.go
│   │   └── validation/
│   │       ├── syntax.go
│   │       └── semantic.go
│   ├── model/
│   │   ├── game.go
│   │   └── user.go
│   └── storage/
│       ├── store.go
│       └── sqlite/
│           ├── store.go
│           └── store_test.go
├── pkg/
├── web/
│   ├── static/
│   └── templates/
├── go.mod
├── go.sum
└── Makefile
```
File & Package Purposes

Root Level

- api/: Contains API specification files, not Go source code.
- api/openapi.yaml: Defines your REST API using the OpenAPI standard for documentation and tooling.
- cmd/: Contains the main application entry points for compilation.
- cmd/server/main.go: Initializes and starts all services (database, router, etc.) and runs the HTTP server.
- configs/: Holds static configuration files for the application.
- configs/config.yaml: Contains runtime configuration like database connection strings and server ports.
- configs/maps/: Stores game map definition files (e.g., JSON) for when you support multiple variants.
- internal/: Contains all the private application logic not meant for external import.
- pkg/: Contains any utility code that is safe and intended to be shared with external projects (often empty initially).
- web/: Contains frontend assets like HTML templates, CSS, and JavaScript.
- go.mod & go.sum: The standard Go files for managing project dependencies.
- Makefile: Provides simple command shortcuts for common development tasks like building, running, and testing.

internal/ Packages

- internal/api/: Manages all HTTP-related concerns, including routing and request handling.
- internal/api/router.go: Sets up the HTTP router, mapping URL paths to their corresponding handlers.
- internal/api/handlers.go: Defines the functions that handle specific API endpoints (e.g., creating a game, submitting an order).
- internal/api/handlers_test.go: Contains unit and integration tests for your HTTP handlers.
- internal/api/middleware.go: Holds HTTP middleware for tasks like logging, authentication checks, and CORS.
- internal/auth/: Contains logic for user authentication (login) and authorization (permissions).
- internal/auth/password.go: Handles hashing and verifying user passwords securely.
- internal/auth/jwt.go: Manages the creation and validation of JSON Web Tokens for API sessions.
- internal/game/: The core domain package containing all the rules and logic of Diplomacy, independent of the API or database.
- internal/game/board.go: Defines the Go structs for the game board, provinces, and units.
- internal/game/state.go: Manages the overall state of a single game, including phase, year, and unit positions.
- internal/game/loader/: A sub-package responsible for loading map definitions.
- internal/game/loader/interface.go: Defines the MapLoader interface, abstracting how maps are loaded.
- internal/game/loader/json_loader.go: An implementation of MapLoader that loads the classic map from JSON files.
- internal/game/validation/: A sub-package implementing the complete validation pipeline.
- internal/game/validation/lexer.go: Tokenizes raw order strings with normalization and error detection.
- internal/game/validation/parser.go: Comprehensive order parsing system with all order types and phase-aware registry.
- internal/game/validation/semantic.go: (Future) Validates parsed orders against game rules and current state.
- internal/game/validation/types.go: Token types, validation result types, and error handling.
- internal/game/validation/*_test.go: Comprehensive unit tests for all validation components.
- internal/game/resolution/: A sub-package implementing the multi-pass resolution engine.
- internal/game/resolution/types.go: SoA data structures and resolution maps for performance.
- internal/game/resolution/engine.go: Main resolution engine with multi-pass algorithm.
- internal/game/resolution/convoy.go: Convoy path discovery and validation logic.
- internal/game/resolution/support.go: Support cutting and strength calculation.
- internal/game/resolution/conflict.go: Conflict resolution, bouncing, and dislodgement detection.
- internal/game/resolution/*_test.go: Comprehensive tests including DATC compliance.
- internal/model/: Defines the Go structs that map directly to your database tables or documents.
- internal/storage/: The data persistence layer, abstracting all database interactions.
- internal/storage/store.go: Defines the database interfaces (e.g., GameStore, UserStore) for dependency injection.
- internal/storage/sqlite/: A sub-package containing the concrete database implementation for SQLite.
- internal/storage/sqlite/store.go: The actual implementation of the storage interfaces using the SQLite database driver.
- internal/storage/sqlite/store_test.go: Contains integration tests that run against a real (or in-memory) SQLite database.
