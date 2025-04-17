Idea:

Python CLI Diplomacy prototype
Data-oriented design

Plan:
[] Create main.py with load, validate, run, save loop - hardcode save game folder
[] Create placeholder load, validate, run, and save files
[] Write each logic file with tests simultaneously
[] Create reference data for starting game state
[] Update main.py and/or loader to show list of save games and ability to create new game

Core Concepts:

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

2. Basic Game Flow
- Game loads state (players, phase, territories, units)
- Each player submits an order
- Turn resolves when all orders are in (or time limit is reached)
- Update territories, units, and player status
- repeat until win or draw condition is met

3. Project structure
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
