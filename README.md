# Diplomacy CLI

A turn-based strategy game inspired by Diplomacy, built as a CLI-first prototype using Python and data-oriented design principles.

This project is a work-in-progress, with ongoing tasks tracked in TODO.md

## Folder Structure

```
src/diplomacy_cli/  # Main source code
├── cli/            # User interface
├── core/           # Core game logic
└── data/           # Game data
backend/            # Go backend for the game engine
tests/              # Automated tests
pyproject.toml      # Project definition and dependencies
flake.nix           # Nix flake for development environment
```

## Getting Started

This project uses [Nix](https://nixos.org/) with [direnv](https://direnv.net/) to automatically manage the development environment.

### Prerequisites

- [Nix](https://nixos.org/download.html) (with flakes enabled)
- [direnv](https://direnv.net/docs/installation.html)

### Setup

1. **Clone the repository**
2. **Allow direnv:** The first time you `cd` into the project directory, `direnv` will prompt you for permission. Type `direnv allow` to activate the environment.

The development environment will automatically:
- Set up Python 3.12 with uv package manager
- Create a virtual environment in `.venv/`
- Install all development dependencies (pytest, ruff, pyright, etc.)
- Configure your PATH to use the virtual environment

### Manual Setup (Alternative)

If you prefer not to use Nix/direnv:

```sh
# Create virtual environment
python3.12 -m venv .venv
source .venv/bin/activate

# Install dependencies
pip install -e ".[dev]"
```

## How to Run

### Python CLI (Legacy)

Once the development environment is active, run the Python CLI:

```sh
python -m diplomacy_cli
```

### Go Backend (Current Development)

The new Go backend server can be started with:

```sh
cd backend
go run ./cmd/server
```

This starts the game engine server on `http://localhost:8080` with a health check endpoint at `/health`.

## Running Tests

To run the test suite, use `pytest`:

```sh
pytest
```