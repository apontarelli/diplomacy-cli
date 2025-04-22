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

[x] Create world and starting state json files
[x] Create game storage interface
[x] Create basic start game, list game, and delete game storage interface
    [x] menu to choose between starting a new game or view saved games
	[x] new game prompts for a game_id and calls start_game
	[x] view saved game lists the current saved games
	    [x] selecting a game from the list gives two options
		[x] print state
		[x] delete game (with a confirmation dialog)  
[] Implement game state logic
    [x] start_game()
    [x] load_state(game_id)
	[x] load from JSON
	[x] return in-memory dict of players, territories, units, game (turn, season)
    [x] format_state(state)
	[x] pretty print the current state
    [] move_unit(state, unit_id, territory_id)
	[x] convert units to a territory keyed dictionary
    [] advance_turn(state)
	[] implement turn history
	    [] compress turn files into single state file
	    [] create year-phase identifier
    [] remove_unit(state, unit_id)
	[] optionally, disband_unit()
    [] change_territory(state, territory_id, new_owner_id)
	[] only applies to supply centers
	[] consider having a method that loops through units in supply centers to call this method to change ownership
    [] update_player_status(state):
	[] if player owns no supply centers, mark as eliminated
	[] also allows for forfeit
    [] end_game()
    [] build_unit(state, player_id, territory_id, type)    
[] Define canonical order JSON structure  
[] Implement order validator
    [] Syntactic validation (is the structure correct?)
    [] Semantic validation (is the move valid for unit, edge, etc.)
    [] Test all validation logic

[] Implement core resolution engine
    [] Run turn using valid/invalid orders
    [] Treat invalid orders as holds
    [] Test basic move + hold resolution

---

##  Phase 3: UX & Input


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
