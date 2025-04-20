# Diplomacy CLI

A turn-based strategy game inspired by Diplomacy, built as a CLI-first prototype using Python and data-oriented design principles.

This project is a work-in-progress, with ongoing tasks tracked in TODO.md

## Folder Structure

```
main.py # Entry point
logic/ # Core game logic
tests/ # Test files
data/ # Game data
data/classic # Immutable data for classic Diplomacy (i.e. starting state)
data/saves # Active game state folders
```

## Getting Started
1. Clone the repo
2. Create a virtual environment

```
python -m venv venv
source venv/bin/activate

```
3. Run the game loop

```
python main.py
```

## Data Model Overview
- Players: id, name, status
- Territories: id, name, type, owner_id
- Edges: territory_a, territory_b, type
- Units: id, owner_id, type, territory_id
- Orders: player_id, type, source_id, target_id

## Running Tests (when completed)

```
pytest tests/
```
