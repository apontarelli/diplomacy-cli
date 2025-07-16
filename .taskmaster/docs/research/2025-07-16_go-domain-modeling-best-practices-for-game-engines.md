---
title: Research Session
query: "Go domain modeling best practices for game engines immutable data structures rich domain models"
date: 7/15/2025
time: 6:16:23 PM
timestamp: 2025-07-16T01:16:23.300Z
exchanges: 1
---

# Research Session

## Initial Query

**Question:** Go domain modeling best practices for game engines immutable data structures rich domain models

**Response:**

# Go Domain Modeling Best Practices for Game Engines: Focus on Immutable Data Structures and Rich Domain Models

## Introduction

Domain modeling in Go, particularly for game engines like the Diplomacy resolution engine in this project, involves creating structured representations of core concepts that encapsulate business logic, ensure data integrity, and promote maintainability. Best practices emphasize clarity, type safety, and performance, given Go's strengths in concurrency and simplicity. For game engines, models must handle complex states (e.g., board configurations, unit movements, conflicts) efficiently, often under real-time constraints. Immutable data structures are crucial to prevent unintended mutations in concurrent environments, while rich domain models (inspired by Domain-Driven Design or DDD) embed behavior and validation within the models themselves, reducing boilerplate and improving expressiveness.

This response draws directly from the project context, such as Task 16 (Design Rich Domain Models) which requires modeling structures like `ResolutionResult`, `UnitOutcome`, and `Conflict` for DATC compliance. It also ties into Task 17 (Implement Resolution Engine Interface) for validation and resolution, and broader tasks like Task 2 (Convoy System) and Task 24 (Performance Optimization), where immutability aids in safe parallelism and benchmarking. I'll provide actionable insights, including code examples tailored to these tasks, multiple approaches (e.g., structs vs. interfaces), step-by-step guidance, edge cases, and pitfalls to ensure the models are robust, testable, and performant.

## Core Best Practices for Domain Modeling in Go

In Go, domain models are typically implemented as structs with embedded types for composition, leveraging Go's lack of classes to favor simple, composable types. A key practice is to use value types (structs) over pointers for immutability and to avoid nil-related errors, which is especially relevant for game engines where state consistency is paramount. Follow the "data-oriented design" principle, common in game development, by aligning models with CPU cache lines for performance—e.g., using arrays or slices instead of maps for frequently accessed data like unit positions.

From a DDD perspective, models should be "rich," meaning they include methods for behavior (e.g., validation, computation) rather than being anemic data holders. This contrasts with procedural styles where logic is external. For this project, apply this to `ResolutionResult` by embedding methods like `IsResolved()` or `GetDislodgedUnits()`. Use custom types for domain-specific concepts (e.g., `type Province string`) to add type safety and prevent invalid states at compile time.

Multiple perspectives: A functional approach treats models as immutable values, enabling pure functions for transformations (useful for resolution engines). An object-oriented lens might use interfaces for polymorphism (e.g., different `UnitOutcome` implementations). For game engines, hybrid approaches work best—structs for data, interfaces for extensibility. Reference Go's standard library (e.g., `time.Time` as an immutable struct with methods) as a model.

## Implementing Immutable Data Structures in Go

Immutability in Go is achieved by designing structs without exported fields that can be mutated post-construction, using constructor functions (e.g., `NewResolutionResult()`) to enforce invariants. All fields should be private (lowercase), with read-only access via methods or getters. This prevents accidental mutations, crucial for concurrent game resolution where multiple goroutines might access shared state (e.g., in Task 24's parallel processing optimizations).

Step-by-step guidance:
1. Define the struct with private fields: `type ResolutionResult struct { turnNumber int; phase Phase; ... }`.
2. Create a constructor: `func NewResolutionResult(turn int, phase Phase, ...) *ResolutionResult { return &ResolutionResult{turnNumber: turn, ...} }` (use pointers sparingly; prefer values for immutability).
3. Provide getters: `func (r ResolutionResult) TurnNumber() int { return r.turnNumber }`.
4. For collections, use read-only views: Return copies of slices/maps to prevent external mutation.

Examples tailored to Task 16:
```go
type ResolutionResult struct {
    turnNumber    int
    phase         Phase
    unitOutcomes  map[UnitID]UnitOutcome // Private for immutability
    // Other fields as per task
}

func NewResolutionResult(turn int, phase Phase, outcomes map[UnitID]UnitOutcome, ...) ResolutionResult {
    // Validation logic here, e.g., check for nil maps
    if outcomes == nil {
        outcomes = make(map[UnitID]UnitOutcome)
    }
    return ResolutionResult{turnNumber: turn, phase: phase, unitOutcomes: outcomes}
}

func (r ResolutionResult) UnitOutcomes() map[UnitID]UnitOutcome {
    // Return a copy to enforce immutability
    copy := make(map[UnitID]UnitOutcome, len(r.unitOutcomes))
    for k, v := range r.unitOutcomes {
        copy[k] = v // Assuming UnitOutcome is also immutable
    }
    return copy
}
```
This ensures `ResolutionResult` can't be altered after creation, aligning with Task 16's immutability guarantees in tests.

Perspectives: For performance-critical game engines, immutability can increase allocations (e.g., copying maps), so use arenas or pools (Go 1.23's experimental features) as suggested in Task 24. Edge case: Handle empty collections by initializing them in constructors to avoid nil panics.

## Building Rich Domain Models

Rich domain models encapsulate logic within the model, making them self-contained. In Go, attach methods to structs for behaviors like validation or computation. For game engines, this means models can simulate resolutions or check invariants without external services, improving testability (e.g., Task 16's unit tests for construction validation).

Approaches: Use embedded structs for inheritance-like composition (e.g., embed `Unit` in `UnitOutcome`). For polymorphism, define interfaces (e.g., `Outcome` interface with `Status()` method). Enrich models with domain events (e.g., a `ResolutionLog` slice that records steps immutably).

Step-by-step for Task 16's `UnitOutcome`:
1. Define the struct: `type UnitOutcome struct { unit Unit; order Order; status UnitStatus; ... }`.
2. Add methods: `func (uo UnitOutcome) IsDislodged() bool { return uo.status == Dislodged }`.
3. Embed validation: `func NewUnitOutcome(unit Unit, order Order, status UnitStatus) (UnitOutcome, error) { if unit == nil { return UnitOutcome{}, errors.New("unit cannot be nil") } return UnitOutcome{...}, nil }`.
4. Ensure richness: Add computation like `func (uo UnitOutcome) EffectivePower() int { /* logic based on supports */ }`.

Code snippet for richness:
```go
type UnitOutcome struct {
    unit       Unit
    orderGiven Order
    finalStatus UnitStatus
    // ...
}

func (uo UnitOutcome) Validate() error {
    if uo.unit.ID == "" {
        return errors.New("invalid unit ID")
    }
    // More checks, tying into Task 17's validation
    return nil
}

func (uo UnitOutcome) ApplyToBoard(board *Board) {
    // Immutable update: Create new board state
    // Useful for simulation in game engines
}
```
This makes models "rich" by including behavior, directly applicable to DATC compliance in Task 16.

## Application to Game Engines and Project-Specific Tasks

In game engines, immutable rich models facilitate undo/redo stacks, AI simulations, and multiplayer synchronization without locks, reducing bugs in concurrent scenarios (e.g., Task 9's API for order submission). For this Diplomacy project, apply to `Conflict` models by making them immutable graphs for resolution, using BFS as in Task 2's convoy system.

Actionable for Task 16: Extend `ResolutionResult` with methods like `AggregateConflicts()` to compute derived data immutably. For Task 17, integrate validation into models: `func (r ResolutionResult) ValidateOrders(orders []Order) ([]ValidationError, error) { /* use internal state */ }`. This ties into Task 1's validation pipeline by making models the source of truth.

For Task 6's database schema, ensure models map to GORM via immutable structs with serialization methods (e.g., `MarshalJSON()`), preserving immutability during persistence. In Task 25's integration, immutable models ensure backward compatibility by providing adapter methods.

## Detailed Examples and Code Snippets

Comprehensive example for `Conflict` in Task 16:
```go
type Conflict struct {
    involvedUnits []UnitID // Immutable slice
    winner        UnitID
}

func NewConflict(units []UnitID, winner UnitID) Conflict {
    sort.Slice(units, func(i, j int) bool { return units[i] < units[j] }) // Normalize for equality checks
    return Conflict{involvedUnits: append([]UnitID(nil), units...), winner: winner} // Defensive copy
}

func (c Conflict) InvolvedUnits() []UnitID {
    return append([]UnitID(nil), c.involvedUnits...) // Return copy
}

func (c Conflict) IsResolved() bool {
    return c.winner != ""
}
```
This is immutable and rich, with methods for querying.

Another for `SupportResults` as a map: Use a custom type with methods to avoid direct map access.

## Edge Cases and Potential Pitfalls

Edge cases: Nil values—always check in constructors (e.g., nil `Unit` in `UnitOutcome` should error, as per Task 16 tests). Empty collections—initialize to empty slices/maps to avoid nil dereferences. Concurrency—immutable structs are safe for sharing, but if using pointers, ensure no shared mutable state.

Pitfalls: Overusing immutability can lead to high allocation overhead; mitigate with Task 24's memory pools. Anemic models (data-only) lead to scattered logic—avoid by always adding methods. Serialization issues: Immutable structs with private fields need custom marshaling for JSON (e.g., in Task 9's API). Performance in large games: Maps in `ResolutionResult` can be slow; consider sorted slices for O(1) access if keys are dense.

Multiple approaches to pitfalls: For serialization, use struct tags or custom Marshalers. Test edge cases exhaustively, as in Task 16's strategy: Cover nil, empty, and DATC extremes.

## Conclusion and Actionable Recommendations

Adopting these best practices will make the project's domain models robust, aligning with DATC requirements and enabling scalable game engine features. Start by refactoring Task 16's models to be immutable and rich, then integrate into Task 17's engine. Benchmark changes per Task 24 to ensure no regressions. If speculation is needed: In future expansions, consider generics (Go 1.18+) for reusable immutable collections, but stick to project-spec Go 1.23 features. This approach not only satisfies immediate tasks but positions the engine for production deployment in Task 25.


---

*Generated by Task Master Research Command*  
*Timestamp: 2025-07-16T01:16:23.300Z*
