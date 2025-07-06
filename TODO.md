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

### Phase 1.5: Multi-Pass Resolution Engine Foundation ✅
- **Hybrid Architecture**: ResolutionEngine with type-safe structs + performance arrays ✅
- **Core Types**: ResolvedOrder, OrderOutcome, ConvoyKey with efficient lookups ✅
- **Multi-Pass Algorithm**: Convoy path stabilization loop with proper termination ✅
- **Order Validation**: Support/convoy relationship validation with detailed errors ✅
- **Basic Resolution**: Move processing, conflict detection, and outcome assignment ✅
- **Comprehensive Testing**: 6 test functions covering basic scenarios and edge cases ✅

## 🎯 Current Focus
### Phase 1.7: DATC Test Suite Implementation ✅ FOUNDATION COMPLETE

**Phase 1.5 COMPLETED** ✅ - Multi-Pass Resolution Engine with all core Diplomacy rules implemented
**Phase 1.6 COMPLETED** ✅ - Integration & Testing with full pipeline implementation
**Phase 1.7 FOUNDATION COMPLETE** ✅ - DATC Test Suite Infrastructure with 28 passing tests across 5 complete categories

## ✅ Recently Completed: DATC Go Test Implementation

### Phase 1.7: DATC (Diplomacy Adjudicator Test Cases) Go Test Suite

Successfully implemented the foundation for DATC Go test generation and execution with optimized infrastructure.

**DATC Source**: https://webdiplomacy.net/doc/DATC_v3_0.html
**Test Categories**: 6.A-6.J (Basic checks, Coastal issues, Circular movement, Supports/Dislodges, Head-to-head battles, Convoys, Adjacent convoys, Retreating, Building, Civil disorder)

#### ✅ Completed Tasks:
1. **1.7.1**: ✅ Create DATC extraction script to download and parse HTML
2. **1.7.2**: ✅ Parse test cases into structured format (JSON)  
3. **1.7.3**: ✅ Generate Go test files from parsed test cases
4. **1.7.4**: ✅ Multi-format order support implementation
5. **1.7.5**: ✅ DATC format parsing support with multi-word provinces
6. **1.7.6**: ✅ DATC Go test infrastructure implementation

#### 🎯 Latest Achievement (January 2025):
**DATC Go Test Infrastructure** - Working test framework with optimized board loading and validation

#### Key Achievement: DATC Go Test Infrastructure Implementation ✅
**Problem Solved**: Need efficient, scalable test infrastructure for 163 DATC test cases including:
- **Unit setup from orders**: Infer unit positions from order strings (e.g., `F North Sea - Picardy` → Fleet at North Sea)
- **Efficient board loading**: Avoid redundant file I/O across test suite
- **Proper test organization**: Logical file structure by test category
- **Validation framework**: Compare engine results against expected DATC outcomes

**Solution**: Comprehensive test infrastructure:
- **Shared board loading**: `sync.Once` pattern loads board exactly once per test suite
- **Helper functions**: `CreateDATCGameState(t)` and `ProcessDATCTest(t, gameState)` for clean test code
- **Unit inference**: Parse unit positions directly from order strings without complex setup
- **Outcome validation**: `ValidateExpectedOutcome()` with detailed logging and error reporting
- **Optimized file structure**: Tests organized by category (basic_checks, convoys, etc.)

**Implementation**: 
- Added `GetDATCTestBoard()` with `sync.Once` for efficient board sharing
- Created `CreateDATCGameState(t)` helper for fresh game state creation
- Implemented unit setup by parsing orders (Fleet/Army type and location inference)
- Built comprehensive validation framework with emoji logging for clear test output
- Organized tests by DATC category with shared helpers in `datc_helpers.go`

**Result**: 
- ✅ **28 working DATC tests** across 5 complete categories with proper validation
- ✅ **Optimized performance** - Board loaded once, reused across all tests
- ✅ **Clean test structure** - Template ready for scaling to remaining 135 tests
- ✅ **Proper validation** - Engine correctly identifies invalid moves with detailed error messages
- ✅ **Complete basic checks** - All 12 fundamental validation tests (6.A.1-6.A.12) passing

#### Implementation Strategy:
- **Extraction Script**: ✅ Python script downloads DATC HTML and extracts test cases
- **Structured Format**: ✅ JSON with test metadata, orders, expected outcomes
- **Test Generation**: ✅ Automated Go test file generation from structured data
- **Multi-Format Support**: ✅ Parser handles DATC, shortcode, and snake_case formats
- **Validation Framework**: 🎯 Compare engine results against DATC expected outcomes
- **Compliance Reporting**: 🎯 Track which test cases pass/fail with detailed reporting

This ensures our engine is fully compliant with official Diplomacy rules before adding persistence layer.

## 🏗️ Hybrid Resolution Architecture Decision

After analyzing the Python PoC's Structure of Arrays (SoA) approach and Go's idioms, we've decided on a **Hybrid Approach** that combines the performance benefits of data-oriented design with Go's type safety and maintainability.

### Why Not Pure SoA in Go?

**Problems with Direct SoA Translation:**
- **Type Safety Loss**: 18+ parallel arrays lose compile-time length consistency guarantees
- **Verbose Code**: Go lacks Python's dynamic typing and dataclass convenience
- **Error Prone**: Manual index synchronization across arrays
- **Maintenance Burden**: Adding fields requires updating multiple arrays and all functions

**Example of Problematic Pure SoA:**
```go
// ❌ Pure SoA - Error prone and verbose
type ResolutionSoA struct {
    UnitID           []string
    OwnerID          []string  
    UnitType         []UnitType
    OrigTerritory    []string
    OrderType        []OrderType
    MoveDestination  []string
    SupportOrigin    []string
    // ... 12 more parallel arrays
}
```

### ✅ Recommended Hybrid Approach

**Core Principle**: Use idiomatic Go structs with performance-conscious design patterns from data-oriented programming.

#### 1. Resolution Engine Structure
```go
type ResolutionEngine struct {
    // Input data - slice of structs for type safety
    orders    []*ResolvedOrder
    
    // Working data - separate arrays for performance-critical operations
    outcomes  []OrderOutcome
    strength  []int
    
    // Lookup maps - efficient access patterns
    conflicts map[string][]int        // territory -> order indices
    supports  map[string][]int        // supported unit -> supporter indices
    convoys   map[ConvoyKey][]string  // convoy path cache
}

type ResolvedOrder struct {
    *Order                    // Embed parsed order
    Index         int         // Position in resolution arrays
    OrigTerritory string      // Normalized territory name
    NewTerritory  string      // Resolved destination
}

type OrderOutcome struct {
    Result       OrderResult
    Dislodged    bool
    SupportCut   bool
    ConvoyPath   []string
    Strength     int
}
```

#### 2. Performance-Critical Hot Paths
For operations that iterate over all orders (strength calculation, conflict detection), we use array-based processing:

```go
// Hot path: Calculate strength for all orders
func (re *ResolutionEngine) calculateStrength() {
    // Reset strength to base value
    for i := range re.strength {
        re.strength[i] = 1
    }
    
    // Add support strength in batch
    for supportedUnit, supporterIndices := range re.supports {
        if supportedIdx := re.findOrderIndex(supportedUnit); supportedIdx != -1 {
            for _, supporterIdx := range supporterIndices {
                if !re.outcomes[supporterIdx].SupportCut {
                    re.strength[supportedIdx]++
                }
            }
        }
    }
}
```

#### 3. Type-Safe Lookup Maps
Instead of error-prone parallel array indexing, use strongly-typed maps:

```go
type ConvoyKey struct {
    Origin      string
    Destination string
}

// Efficient lookups without index synchronization issues
func (re *ResolutionEngine) buildLookupMaps() {
    re.conflicts = make(map[string][]int)
    re.supports = make(map[string][]int)
    
    for i, order := range re.orders {
        // Territory conflicts
        dest := order.NewTerritory
        re.conflicts[dest] = append(re.conflicts[dest], i)
        
        // Support relationships
        if order.Type == Support {
            target := order.SupportTarget
            re.supports[target] = append(re.supports[target], i)
        }
    }
}
```

### Benefits of Hybrid Approach

1. **Type Safety**: Compiler catches mismatched data relationships
2. **Performance**: Cache-friendly arrays for hot paths, efficient maps for lookups
3. **Maintainability**: Clear data relationships, easy to extend
4. **Go Idioms**: Follows Go conventions and patterns
5. **Debugging**: Easier to inspect and debug than parallel arrays
6. **Testing**: Can test individual components in isolation

### Implementation Strategy

#### Phase 1.5.1: Core Types and Engine Structure
```go
// File: backend/internal/game/resolution/types.go
type ResolutionEngine struct { /* ... */ }
type ResolvedOrder struct { /* ... */ }
type OrderOutcome struct { /* ... */ }
type ConvoyKey struct { /* ... */ }

// File: backend/internal/game/resolution/engine.go  
func NewResolutionEngine(orders []*Order) *ResolutionEngine
func (re *ResolutionEngine) Resolve() error
func (re *ResolutionEngine) GetResults() []OrderOutcome
```

#### Phase 1.5.2: Specialized Components
Each component operates on the shared engine state but focuses on specific concerns:

```go
// File: backend/internal/game/resolution/convoy.go
func (re *ResolutionEngine) processConvoys()
func (re *ResolutionEngine) findConvoyPath(origin, dest string) []string

// File: backend/internal/game/resolution/support.go  
func (re *ResolutionEngine) processSupports()
func (re *ResolutionEngine) cutSupports()

// File: backend/internal/game/resolution/conflict.go
func (re *ResolutionEngine) resolveConflicts()
func (re *ResolutionEngine) detectDislodgements()
```

### Migration from Python PoC

The hybrid approach preserves the Python PoC's algorithmic insights while adapting to Go's strengths:

| Python PoC Concept | Go Hybrid Implementation |
|-------------------|-------------------------|
| `ResolutionSoA` parallel arrays | `ResolutionEngine` with typed structs + performance arrays |
| `ResolutionMaps` lookups | Strongly-typed maps (`conflicts`, `supports`, `convoys`) |
| Index-based operations | Method calls on engine with clear data relationships |
| Multi-pass algorithm | Same algorithm, cleaner implementation |
| Convoy path flattening | `ConvoyKey` map with cached paths |

This approach gives us the performance characteristics we need while maintaining Go's type safety and readability advantages.

#### Task 1.5.1: Resolution Data Structures ✅
**File**: `backend/internal/game/resolution/types.go` ✅
- ResolutionEngine with hybrid approach ✅
- ResolvedOrder and OrderOutcome types ✅
- Lookup map types and structures ✅

#### Task 1.5.2: Convoy Path Discovery ✅
**File**: `backend/internal/game/resolution/convoy.go` ✅
- BFS-based convoy path finding with ProvinceCoast value objects ✅
- Convoy chain validation using engine state ✅
- Integration with ResolutionEngine.processConvoys() ✅
- Architectural refactor: eliminated string re-parsing with structured data ✅

#### Task 1.5.3: Support System ✅
**File**: `backend/internal/game/resolution/engine.go` ✅
- Support cutting logic with self-attack rule prevention ✅
- Strength calculation with typed support relationships ✅
- Multi-pass convergence support for convoy-support interaction ✅
- Comprehensive testing including critical Diplomacy rules ✅

#### Task 1.5.4: Dislodgement Detection & Friendly Unit Protection ✅
**File**: `backend/internal/game/resolution/engine.go` ✅
- Implement proper dislodgement detection logic ✅
- **Implement friendly unit protection rule in resolveConflicts()** ✅
  - Friendly units cannot dislodge each other but can bounce enemy attacks ✅
  - Critical Diplomacy rule now correctly implemented ✅
- Handle unit displacement when losing conflicts ✅
- Integration with existing conflict resolution system ✅
- Add comprehensive tests for dislodgement and friendly protection scenarios ✅

#### Task 1.5.5: Main Resolution Engine ✅
**File**: `backend/internal/game/resolution/engine.go` ✅
- Multi-pass iterative algorithm with hybrid data structures ✅
- Convoy path stabilization loop using ConvoyKey maps ✅
- NewResolutionEngine() constructor and Resolve() orchestrator ✅
- GetResults() for final OrderOutcome extraction ✅

#### Task 1.5.6: Resolution Tests ✅
**Files**: `backend/internal/game/resolution/*_test.go` ✅
- Component unit tests ✅
- Integration tests ✅
- DATC compliance tests (basic scenarios covered)

### Phase 1.6: Integration & Testing ✅ COMPLETED

#### Task 1.6.1: Full Pipeline Integration ✅
- ✅ End-to-end tests (raw orders → final state) - `backend/internal/integration/integration_test.go`
- ✅ Multi-turn game progression - `TestMultiTurnProgression` with phase advancement
- ✅ Pipeline orchestrator - `backend/internal/pipeline/orchestrator.go` with complete validation flow
- ✅ Conflict resolution testing - Bounce and support scenarios tested
- ✅ All tests passing - 100% success rate across all packages

#### Task 1.6.2: DATC Implementation (Future Enhancement)
- Diplomacy Adjudicator Test Cases
- Automated compliance testing  
- Regression test suite
- **NOTE**: Re-evaluate DATC v3.0 test extraction and automation for comprehensive test coverage

**Architecture Decision**: Implemented both integration tests (in separate package to avoid import cycles) and pipeline orchestrator for maximum flexibility. The pipeline package provides a clean API for external consumers while integration tests verify end-to-end functionality.

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
**Continue DATC Test Suite Implementation** - Expand to remaining categories and achieve full compliance

**Current Status (January 2025)**: Phase 1 core engine is complete with working DATC test infrastructure. **28/163 tests passing (17% complete)** across 5 complete categories.

### ✅ Completed Categories (28 tests):
- **basic_checks**: 12/12 tests ✅ (6.A.1-6.A.12) - Fundamental validation rules
- **coastal_issues**: 4/4 tests ✅ (6.B.1-6.B.4) - Multi-coast province handling  
- **head_to_head**: 3/3 tests ✅ (6.H.1-6.H.3) - Head-to-head battle mechanics
- **supports_dislodges**: 5/5 tests ✅ (6.D.1-6.D.5) - Support and dislodgement rules
- **convoys**: 4/4 tests ✅ (6.F.1-6.F.3 + Simple) - Basic convoy mechanics

### 🎯 Next Priority Categories (135 remaining tests):
1. **circular_movement** (9 tests) - Unit swaps and circular dependencies
2. **adjacent_convoys** (20 tests) - Adjacent convoy edge cases  
3. **retreating** (16 tests) - Retreat mechanics and validation
4. **building** (7 tests) - Build validation and supply center rules
5. **civil_disorder** (11 tests) - Automatic disbanding and civil disorder
6. **Remaining convoy tests** (21 tests) - Advanced convoy scenarios and paradoxes
7. **Remaining supports_dislodges** (29 tests) - Complex support scenarios
8. **Remaining head_to_head** (12 tests) - Advanced battle mechanics

### After DATC Completion:
**Phase 2**: Persistence Layer Implementation - Database integration and storage layer

**Detailed Plan**: See `DATC_IMPLEMENTATION_PLAN.md` for DATC implementation details

### 🏆 Key Learnings from DATC Infrastructure Implementation:

#### 1. **Efficient Test Organization**
- **File Structure**: Organize by DATC category (10 files for 163 tests)
- **Shared Infrastructure**: `datc_helpers.go` with `sync.Once` board loading
- **Template Pattern**: `CreateDATCGameState(t)` → add units → add orders → process → validate

#### 2. **Unit Setup Strategy**
- **Order-Based Inference**: Parse unit positions from orders (e.g., `F North Sea - Picardy` → Fleet at North Sea)
- **Province Normalization**: Use lowercase province keys (`"liverpool"` not `"Liverpool"`)
- **Minimal Setup**: Only add units that have orders, keep tests focused

#### 3. **Validation Framework**
- **Detailed Logging**: Emoji-based test output for clear pass/fail indication
- **Error Context**: Include test ID and expected outcome in validation messages
- **Flexible Outcomes**: Support various DATC outcome types (order fail, move fail, dislodgement, etc.)

#### 4. **Performance Optimization**
- **Board Sharing**: Load classic board once per test suite using `sync.Once`
- **Fresh Game States**: Each test gets clean GameState while sharing expensive board data
- **Incremental Testing**: Test categories individually to avoid overwhelming output

**What Was Completed in Phase 1**:
- ✅ **Phase 1.1-1.5**: Complete Diplomacy rules engine with all core systems
- ✅ **Phase 1.6**: Full pipeline integration and testing
- ✅ **Friendly unit protection rule** - Critical Diplomacy rule correctly implemented
- ✅ **Proper dislodgement detection** - Units correctly dislodged by successful attacks
- ✅ **Enhanced conflict resolution** - Handles complex scenarios with friendly protection
- ✅ **Comprehensive testing** - Integration tests, pipeline orchestrator, and end-to-end validation
- ✅ **Integration verified** - All tests passing across all packages

**Key Architecture Achievements**:
- Complete validation pipeline: Raw orders → Syntax → Semantic → Resolution
- Multi-pass resolution engine with hybrid data structures
- Convoy system with ProvinceCoast value objects
- Support system with self-attack rule and multi-pass convergence
- Conflict resolution with friendly protection and dislodgement
- Clean separation between integration tests and pipeline orchestrator

**Current Status**: 
- ✅ **Phase 1 COMPLETE** - The Core Domain (The "Engine") with full integration testing
- 🎯 **Ready for Phase 2** - Persistence (Database integration)

---

# Current Actual Go Structure (Updated January 2025)
```
diplomacy-cli/
├── backend/                    # Go backend (current development focus)
│   ├── data/                   # Game variant data
│   │   └── classic/
│   │       ├── start/          # Starting positions
│   │       ├── world/          # Map definition
│   │       └── variant.json
│   ├── internal/
│   │   └── game/               # Core domain package
│   │       ├── loader/         # Map loading
│   │       ├── validation/     # Order parsing & syntax validation
│   │       ├── resolution/     # Rules engine & conflict resolution
│   │       ├── pipeline/       # Pipeline orchestration & DATC tests
│   │       │   ├── orchestrator.go
│   │       │   ├── datc_helpers.go
│   │       │   ├── datc_test.go
│   │       │   └── testdata/
│   │       ├── integration/    # End-to-end integration tests
│   │       ├── board.go        # Core domain types
│   │       └── state.go
│   ├── scripts/                # Temporary build/test scripts (will be removed)
│   ├── go.mod
│   └── go.sum
├── src/                        # Python CLI (legacy/reference)
├── tests/                      # Python tests
├── docs/
├── pyproject.toml
└── README.md
```

## Future Structure (Phase 2+)
```
diplomacy-cli/backend/
├── cmd/
│   └── server/
│       └── main.go
├── data/                       # Game variant data
├── internal/
│   ├── game/                   # Core domain (Phase 1 - COMPLETE)
│   │   ├── loader/
│   │   ├── validation/
│   │   ├── resolution/
│   │   ├── pipeline/
│   │   ├── board.go
│   │   └── state.go
│   ├── model/                  # Database models (Phase 2)
│   ├── storage/                # Database layer (Phase 2)
│   ├── api/                    # HTTP handlers (Phase 3)
│   └── auth/                   # Authentication (Phase 3)
├── go.mod
└── go.sum
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
