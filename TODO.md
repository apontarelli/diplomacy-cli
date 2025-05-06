# Diplomacy CLI Prototype (Python)

## Status
- Phase 1: ✅ Completed
- Phase 2: 🚧 In Progress (State Logic 80% Complete)
- Phase 3: ⏳ Not Started

## Core Principles
- **Data-Oriented Design**: prioritize simple, flat structures for performance and clarity.
- **State-Driven Engine**: all operations mutate or derive from a single game state.
- **CLI-first UX**: fast input, readable output, no graphical dependencies.

---

## Phase 1: Core Scaffolding ✅
- [x] `main.py` with load → validate → run → save loop
- [x] Placeholder logic modules:
  - [x] `storage.py`
  - [x] `validate.py`
  - [x] `engine.py`
  - [x] `state.py`

---

## Phase 2: Logic & Testing 🚧

### State Management
- [x] Create initial world and starting state JSON files
- [x] Implement game storage interface:
  - [x] Start game (prompts for `game_id`, initializes save)
  - [x] View saved games (list + select)
    - [x] Print current state
    - [x] Delete save (with confirmation)
- [x] Load and pretty-print state
- [x] Use turn codes (`1901-S-M`) to unify turn/season/phase data

### Unit Management
- [x] `apply_unit_movements(state, movements)`
- [x] `disband_unit(state, territory_id)`
- [x] `build_unit(state, territory_id, type, owner)`
- [x] `set_territory_owner(state, territory_id, owner)`
- [x] `eliminate_player(player_id)`

### Territory Management
- [x] Merge edges into territories structure
- [x] Normalize territory JSON:
  - [x] Remove redundant `_connections`
  - [x] Add `is_supply_center`, `has_coast` flags
  - [x] Merge home centers
- [x] Create schema validation for territories

### Refactor Static Starting State to Dynamic Build
- [x] Refactor start json files:
    - [x] `starting_units.json` (list of owner_id, type, territory_id)
    - [x] `starting_owners.json` (list of territory_id, owner_id)
    - [x] `starting_players.json` (simple dictionary of nation_id → player status)
- [x] Refactor `start_game`:
    - [x] Create empty dictionaries:
        - `players`
        - `units`
        - `territory_state`
    - [x] Load variant start files:
        - Nations (from immutable `nations.json`)
        - Starting units
        - Starting territory ownerships
        - Starting players
    - [x] For each `starting_unit`, call `build_unit`
    - [x] For each `starting_territory`, call `set_territory_owner`
    - [x] For each `starting_player`, initialize player state
    - [x] Save built game state (`players.json`, `units.json`, `territory_state.json`, etc.) to disk
- [x] Refactor `load_state`:
    - Load state from files
    - Call build_territory_to_unit(units)
    - Call build_counters(units)
    - return state, counters, territory_to_units

### Units and Movement Refactor
- [x] Update Unit Data Model
    - Units must have a unit id, not a territory id
- [x] Create and maintain a territory-to-unit index
    - In-memory dictionary mapping `territory_id` to `unit_id`
- [x] Introduce Unit Counters
    - Maintain in-memory counters mapping: {owner_id + unit type} → highest number built.
    - Rebuild counters at load time by parsing existing unit_ids.
- [x] Refactor state mutators
    - apply_unit_movements(units, territory_to_unit, movements)
	- update units' `teritory_id`
	- update `territory_to_unit`  mapping
    - disband_unit(units, territory_to_unit, unit_id)
	- remove unit from `units`
	- remove corresponding `territory_id` entry from `territory_to_unit`
    - build_unit(units, territory_to_unit, territory_id, type, owner_id)
	- create a new `unit_id`
	    - incrementing on unit counter
	    - FORMAT: `{OWNER}_{UNITTYPE}_{COUNTER}` -- e.g., `ENG_A_1`
	- insert new unit into `units`
	- update `territory_to_unit` mapping


### Order Handling
- [x] Define canonical order JSON structure:
  - [x] Define order format spec: ORDER_FORMAT.md in `docs/ORDER_FORMAT.md`
- [ ] Implement order validator:
  - [x] Load rules and world configuration
    - [x] create Rules data class in `rules_loader.py`
    - [x] returns Rules dataclass
    - [x] Test bidirectional edge loading
  - [x] Define data types in `logic/schema.py`
    - [x] `Order`
    - [x] `SyntaxResult`
    - [x] `SemanticResult`
    - [x] `ValidationResult`
  - [x] Syntax validation (proper structure, required fields) in `validator/syntax.py`
  - [ ] Semantic validation (move legality, unit type, adjacency, phase legality) in `semantic.py
  - [ ] Orchestrate & aggregate results, returning `ValidationReport`
  - [ ] Unit tests for all validator logic
    - [ ] `syntax.py`
    - [ ] `semantic.py`
    - [ ] `orchestrator.py`
  - [ ] Validator should return a `valid` bool and a `reason` string
    - Returns a `ValidationReport` result class type

### Resolution Engine
- [ ] Core resolution:
  - [ ] Resolve valid moves
  - [ ] Treat invalid orders as holds
  - [ ] Test basic move/hold cycle
  - [ ] Create resolution_report.json to explain what happened
    - [ ] applied_orders array
    - [ ] invalid_orders array
- [ ] Implement and test advanced rules resolution:
  - [ ] Add Support Strength
  - [ ] Cut support
  - [ ] Convoys
  - [ ] Retreats
  - [ ] Adjustments (Disband, Builds)

---

## Phase 3: UX & Input ⏳

- [ ] Raw text order input:
  - [ ] Parse strings like `A PAR - BUR` into structured orders
- [ ] Log invalid orders with human-readable error messages (non-blocking)

---

## Future scope

- [ ] Better syntax ParseErrors
  - Currently, syntax parser returns a generic error if all parsers fail
  - Either show all errors for all parsers
  - Rank parser errors by the parser that made it the farthest before failing, representing the most likely parser to be correct



---

## Miscellaneous Tasks

### Tooling & Environment 
- [x] Remove legacy venv / pip workflow
- [x] Adopt uv for env + locking
- [x] Integrate Hatchling
    - [x] Add hatchling
    - [x] Move files to src/diplomacy_cli
    - [x] Update to platformdirs
    - [x] Fix failing tests
    - [ ] Update README.md to reflect new structure and build instructions
- [x] Integrate Ruff (lint / format)
  - [x] Run initial ruff format
  - [x] Run initial ruff check
- [x] Integrate Pyright (type checking)

---

## Core Data Model

### Entities
- **Players**: `id`, `status`
- **Nations**: `id`, `name`
- **Territories**: `id`, `name`, `type`, `is_supply_center`, `has_coast`, `home_center`
- **TerritoryState**: `territory_id`, `owner_id`
- **Territory Edges**: `territory_a`, `territory_b`, `type` (land/sea/both)
- **Units**: `id`, `owner_id`, `type`, `territory_id`
- **Orders**: `player_id`, `type`, `source_id`, `target_id`
- **Messages (future)**: `id`, `sender`, `recipient`, `content`

---

## Basic Game Flow
1. Load game state.
2. Accept player orders.
3. Validate orders.
4. Resolve turn.
5. Update state.
6. Repeat until win or draw.

---

## Project Structure

- `main.py` — entry point, game loop control
- `data/`
  - `{variant}/` — static world definitions (players, territories, edges, units)
  - `saves/{game_id}/` — dynamic game saves (state, orders, units)
- `logic/`
  - `engine.py` — turn processing and resolution
  - `state.py` — game state manipulation
  - `storage.py` — load/save interface
  - `turn_code.py` — turn/season/phase utilities
  - `validate.py` — schema and game rules validation
- `tests/` — unit tests

---

## Future Enhancements

- **TUI/Map Rendering**:  
  Display board via ASCII art or simple grid visualization.

- **Message System**:  
  Allow players to send/receive in-game messages (diplomatic messaging).

- **AI/Testing Sandbox**:  
  Let users input hypothetical orders for any nation to simulate scenarios.

- **Multiplayer Support**:  
  Hotseat first, then basic network support if desired.

---
