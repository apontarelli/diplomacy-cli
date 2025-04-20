# Idea:

Python CLI Diplomacy prototype
Data-oriented design

# Plan:

## Phase 1: Core Scaffolding

[x] Create `main.py` with load → validate → run → save loop  
[x] Create placeholder logic files:
    [x] `storage.py`
    [x] `validate.py`
    [x] `engine.py`
    [x] `state.py`

---

##  Phase 2: Logic + Testing (in parallel)

[] Define canonical order JSON structure  
[] Implement order validator
    [] Syntactic validation (is the structure correct?)
    [] Semantic validation (is the move valid for unit, edge, etc.)
    [] Test all validation logic

[] Implement core resolution engine
    [] Run turn using valid/invalid orders
    [] Treat invalid orders as holds
    [] Test basic move + hold resolution

[] Implement game state helpers
    [] Get/set unit at territory
    [] Transfer ownership of territories
    [] Advance turn & phase
    [] Test state mutation logic

[] Implement load/save functions
    [] Load from `data/reference/`
    [] Save to `data/saves/`
    [] Support multiple game folders
    [] Test load/save round-trip

---

##  Phase 3: Reference Game Data

[] Define reference data (stored in `data/reference/classic/`)
    [] `territories.json` – ID, name, type, supply center
    [] `edges.json` – bidirectional connections, typed
    [] `nations.json` – list of major powers
    [] `starting_units.json` – unit type, nation_id, territory_id

---

##  Phase 4: UX & Input

[] Update `main.py` / loader:
    [] List available saves
    [] Create new game from reference

[] Text-based order input:
    [] Accept raw input (e.g. `A PAR - BUR`)
    [] Parse into order struct
    [] Store raw + parsed orders for review

[] Print/log invalid orders with errors (but don’t block)

---

# Core Concepts:

1. Entities
- Players
	- id
	- name
	- status
- Territories
	- id
	- name
	- owner_id
	- type
        - coast
        - land
        - sea
- Territory_edges
    - territory_a
    - territory_b
    - type (land, sea, both)
- Units
	- id
	- owner_id
	- type
    - territory_id
- Orders
	- player_id
	- type
	- source_id
	- target_id
- Messages (eventually)
	- id
	- sender
	- recipient
	- content

# 2. Basic Game Flow
- Game loads state (players, phase, territories, units)
- Each player submits an order
- Turn resolves when all orders are in (or time limit is reached)
- Update territories, units, and player status
- repeat until win or draw condition is met

# 3. Project structure
- Initial structure for validating logic

- main.py
- data/
    - reference/
        - players
        - territories
        - edges
        - units
    - saves/
        [game_#]/
            - units
            - territories
            - players
            - orders
            - game_state
- logic/
    - load.py
    - engine.py
    - validate.py
    - state.py
- tests/

# 4. Future plans
- UX (CLI/TUI)
    - Rendering game map
- Messages
- Multiplayer
- Order testing (allowing users to input orders for their enemies to test)
