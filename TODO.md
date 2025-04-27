
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

### Units and Movement Refactor
- [ ] Update Unit Data Model
    - Units must have a unit id, not a territory id
- [ ] Create and maintain a territory-to-unit index
    - In-memory dictionary mapping `territory_id` to `unit_id`
- [ ] Introduce Unit Counters
    - Maintain in-memory counters mapping: {owner_id + unit type} → highest number built.
    - Rebuild counters at load time by parsing existing unit_ids.
- [ ] Refactor state mutators
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
    - Create one-off script to migrate unit data model and create `territory_to_unit` mapping

### Resolution Engine (Planned)
- [ ] Load rules and world configuration
- [ ] Ensure bidirectional edge loading
- [ ] Core resolution:
  - [ ] Resolve valid moves
  - [ ] Treat invalid orders as holds
  - [ ] Test basic move/hold cycle

### Order Handling (Planned)
- [ ] Define canonical order JSON structure:
  - [ ] `data/saves/{game_id}/orders/`
  - [ ] One file per player preferred
  - [ ] Orders compressed into history after resolution
- [ ] Implement order validator:
  - [ ] Syntax validation (proper structure)
  - [ ] Semantic validation (move legality, unit type, adjacency)
  - [ ] Unit tests for all validator logic

---

## Phase 3: UX & Input ⏳

- [ ] Raw text order input:
  - [ ] Parse strings like `A PAR - BUR` into structured orders
- [ ] Log invalid orders with human-readable error messages (non-blocking)

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
