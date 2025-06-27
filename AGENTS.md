# Project: diplomacy-cli

## Overview

This project is a turn-based strategy game inspired by Diplomacy, built as a command-line interface (CLI) application. It serves as a proof of concept for exploring data-oriented design principles in Python.

- **Technology:** Python 3.12, Go
- **Key Dependencies:** `platformdirs`
- **Dev Dependencies:** `ruff`, `pyright`, `pytest`, `jsonschema`, `build`

## Project Structure

- `src/diplomacy_cli`: Python CLI (legacy/reference implementation)
  - `cli/ux`: User interface and user-facing interactions.
  - `core/logic`: Core game logic, state management, and rules processing.
  - `data`: Game data definitions (e.g., classic map variant).
- `backend/`: Go backend (current development focus)
  - `cmd/server/`: HTTP server entry point
  - `internal/game/`: Core domain types and game logic
  - `internal/storage/`: SQLite database layer
  - `internal/api/`: HTTP handlers and routing
- `tests/`: Automated tests.
- `pyproject.toml`: Python project configuration.
- `flake.nix`: Nix development environment.

## Development Environment

This project uses [Nix](https://nixos.org/) to manage the development environment, ensuring that all contributors have a consistent set of tools.

To activate the development environment, `cd` into the project directory and allow `direnv` when prompted. It will automatically configure a shell with all the necessary dependencies.

## Development Workflow

1.  **Activate the Nix Shell:** `nix develop`
2.  **Linting & Formatting:** `ruff` is used for code quality. Configuration is in `pyproject.toml`.
3.  **Type Checking:** `pyright` is used for static type analysis. Configuration is in `pyproject.toml`.
4.  **Testing:** `pytest` is used for running tests.

## Best Practices & Conventions

- **Dependency Management:** When adding new dependencies, always search for the latest stable version and pin to it. For Nix packages, prefer stable channels (e.g., `nixos-25.05`) over `nixos-unstable`.
- **Task Management:** The `TODO.md` file is used to track the current development task. Please keep it up to date with the current focus.
- **Architecture:**
  - **Data-Oriented Design:** The project prefers a data-oriented approach, keeping data and logic separate. However, this is a guiding principle, not a dogmatic rule.
  - **Function Structure:** Lean towards larger, self-contained functions. Only break logic into smaller functions if it is clearly reusable or needs to be composed in different ways.
- **Code Style:**
  - **Python**: Follow PEP 8, use type hints for all function signatures and complex types.
  - **Go**: Follow standard Go conventions (gofmt, standard project layout). No JSON tags on domain types.
- **General:**
  - Follow existing patterns in the codebase for consistency.

## Architecture Decisions

- **Data-Oriented Design**: Core domain types focus on game logic without serialization concerns
- **Clean Layer Separation**: Storage, domain logic, and API layers are distinct
- **SQLite over JSON**: Structured data with ACID properties for game state integrity
- **Go Backend + Python AI**: Performance/concurrency for engine, flexibility for AI strategy
- **API Design Philosophy**: Endpoints driven by actual game mechanics needs, not REST conventions
- **Domain-First Approach**: Core game logic drives API design

## High-Level Architecture Vision

- **Core System (Go Backend)**: Authoritative game state with SQLite storage, resolution engine, HTTP API, turn lifecycle management
- **Game Mechanics**: Player registration, order submission/validation, turn management, resolution engine
- **AI Agents**: Tactical AI (Go) + LLM strategist (Python/DSpy) for autonomous gameplay
- **Frontends**: TUI (Bubble Tea) for human players, web interface (optional)
- **Messaging**: Private negotiations between players
- **Advanced Features**: AI vs AI simulation, tournament management, custom variants

## Detailed Roadmap

### Game State & Mechanics
- Players: registration, assignment to powers
- Orders: legal generation, submission, preview, overwriting
- Phase timing: deadline enforcement, NMR handling
- Territory, unit, and adjacency modeling (supports coastal provinces, fleets, etc.)
- Resolution engine using SoA or optimized data layout
- Game history + replay capability

### AI Agents
- **Tactical AI (Go):**
  - Legal move enumeration
  - Evaluation functions (supply center gain, unit safety, alliance value)
  - Multi-move simulation (concurrent rollout of future orders)
  - Tool interface for LLM use
- **LLM Strategist (Python + DSpy):**
  - Propose strategic moves based on game state + memory
  - Generate and interpret private messages
  - Track trust, alliances, betrayals
  - Adapt strategy over time
  - Use tactical engine as a tool during planning
  - (Eventually) represent distinct personalities per power

### Messaging Layer
- Private message threads (per player-pair, per turn)
- Read/unread tracking
- Message deadlines (cut off negotiation at turn end)
- LLM-generated messages for AI players
- Optional moderation / replay log

### Frontends
- **TUI (Bubble Tea, Go):**
  - Lightweight client for humans or AI debugging
  - Phase viewer, unit/map display
  - Order submission and messaging
  - History view (scrollback for previous turns/messages)
- **Web (React, optional):**
  - Drag-and-drop map orders
  - Real-time updates via WebSockets
  - Private message UI per player
  - Phase history and game log

### Simulation / Sandbox Mode
- AI-vs-AI mode for testing strategies
- Run full games autonomously
- Evaluate emergent behavior, bluffing, betrayal, etc.
- Log and replay simulated games
- Run scenario trees for LLM training/debugging

### Advanced Features (Later Scope)
- Fog of war / custom variants
- Anonymized or masked identities
- Alternative rule sets (no-press, gunboat, etc.)
- Tournament bracket management
- Discord bot for turn reminders and game notifications
- Observer/spectator mode
- Game replayer / visualizer

### Dev Infrastructure
- Containerized DSpy AI service (FastAPI)
- Test suite for game logic (unit + integration tests)
- Parallel simulation runner for AI evaluation
- DB migrations and schema evolution tools
