# Go Refactor Plan



### Phase 1: The Core Domain (The "Engine")

This phase is done in complete isolation. You don't need a database or a web server. The goal is a fully testable Go library that understands the rules of Diplomacy.

*   **Packages to Build**: `internal/game/`
*   **Order of Operations**:
    1.  **Define Structs**: Start in `board.go` and `state.go`. Define your core structs: `Province`, `Unit`, `Board`, `Game`, `Order`, etc.
    2.  **Implement the Loader**: Create the `loader/` sub-package. Implement the `MapLoader` interface and the `StaticLoader` that returns the hardcoded classic map. This allows you to create a `Board` to test against.
    3.  **Implement the Adjudicator**: This is the most complex part. In `resolution/adjudicator.go`, write the logic that resolves moves.
    4.  **Write Tests (`_test.go`)**: This is the most critical step of Phase 1. Your `adjudicator_test.go` should be extensive. A great goal is to implement the [Diplomacy Adjudicator Test Cases (DATC)](http://web.inter.nl.net/users/L.B.Kruijswijk/), which are a standard suite for verifying a correct implementation.

*   **Milestone**: You have a Go package that can be given a game state and a set of orders, and it can correctly calculate the next game state.

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
- internal/game/loader/static_loader.go: An implementation of MapLoader that returns the hardcoded classic map, allowing for fast startup.
- internal/game/validation/: A sub-package for validating that player orders are legal.
- internal/game/validation/semantic.go: Checks if an order makes sense according to the game rules and current state.
- internal/game/resolution/: A sub-package for the complex logic of adjudicating a turn's orders.
- internal/game/resolution/adjudicator.go: The rules engine that resolves all orders according to the official rules.
- internal/game/resolution/adjudicator_test.go: Holds the critical unit tests (like the DATC) for the rules engine.
- internal/model/: Defines the Go structs that map directly to your database tables or documents.
- internal/storage/: The data persistence layer, abstracting all database interactions.
- internal/storage/store.go: Defines the database interfaces (e.g., GameStore, UserStore) for dependency injection.
- internal/storage/sqlite/: A sub-package containing the concrete database implementation for SQLite.
- internal/storage/sqlite/store.go: The actual implementation of the storage interfaces using the SQLite database driver.
- internal/storage/sqlite/store_test.go: Contains integration tests that run against a real (or in-memory) SQLite database.
