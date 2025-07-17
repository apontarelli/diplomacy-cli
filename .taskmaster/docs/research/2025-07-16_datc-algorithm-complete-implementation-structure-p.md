---
title: Research Session
query: "DATC algorithm complete implementation structure phases dependency graph resolution order backup rules"
date: 7/16/2025
time: 3:45:16 PM
timestamp: 2025-07-16T22:45:16.339Z
exchanges: 1
---

# Research Session

## Initial Query

**Question:** DATC algorithm complete implementation structure phases dependency graph resolution order backup rules

**Response:**

# Comprehensive Research on DATC Algorithm: Complete Implementation Structure, Phases, Dependency Graph, Resolution Order, and Backup Rules

## Introduction
The Diplomacy Adjudication Test Cases (DATC) provide a standardized suite of 163 test cases for validating the correctness of a Diplomacy game's order resolution algorithm. Based on the official DATC documentation (version 2.4, available at diplomacy-archive.com), the algorithm adjudicates movement phase orders in a turn-based strategy game involving units, supports, convoys, and conflicts. Your query focuses on a complete implementation structure, including phases, dependency graphs, resolution order, and backup rules. This is highly relevant to your project, particularly **Task 5: Achieve Full DATC Compliance and Edge Case Handling**, which aims to pass the remaining 46 tests (focusing on categories like COASTAL ISSUES, CONVOYS, ADJACENT CONVOYS, and CIVIL DISORDER) using the 4-phase roadmap in `datc-implementation-roadmap.md`. It also ties into **Task 18: Implement Core Move Resolution Algorithm**, which outlines multi-pass resolution with phases like validation, strength calculation, conflict resolution, and outcome determination, and **Task 17: Implement Resolution Engine Interface**, which defines the `ResolutionEngine` interface.

This response draws from DATC specifications, open-source Diplomacy adjudicators (e.g., implementations in Python's `diplomacy` library and Go-based projects like `dippy`), and your existing project files (e.g., `adjudicator.go`). I'll provide a detailed, actionable structure for implementation in Go, aligned with your project's use of Gin, JWT, WebSockets, and database integration. The goal is to achieve 100% DATC compliance, as per Task 5, by iteratively implementing and testing against the categorized failures in `datc-test-categorization.md`. I'll include code snippets, step-by-step guidance, multiple approaches (e.g., graph-based vs. iterative), edge cases, and pitfalls, ensuring applicability to tasks like **Task 22: Implement State Transition Simplification** for applying resolution results immutably.

## Overview of DATC Algorithm
The DATC algorithm resolves orders in Diplomacy's movement phase, handling moves, holds, supports, and convoys while resolving conflicts like head-to-head battles, dislodgements, and paradoxes (e.g., circular convoys). It emphasizes determinism, where outcomes are predictable based on rules without randomness. Key principles include:
- **Unit Strength**: A unit has base strength 1, plus supports.
- **Conflict Resolution**: Moves to occupied spaces cause bounces unless one dislodges the other.
- **Convoy Paradoxes**: Handled via backup rules to avoid infinite loops.
- **Backup Rules**: Fallback mechanisms for unresolved paradoxes, such as the "Szykman rule" or "all-or-nothing" convoy resolution.

From a project perspective, this aligns with Task 18's multi-pass architecture, where resolution iterates until stability. An alternative perspective is a graph-based approach (modeling orders as nodes with dependencies), which could integrate with your `ConvoyEngine` in Task 17. Performance is crucial, as Task 5 mentions using Go pprof for profiling—aim for O(n log n) complexity where n is the number of orders, avoiding exponential blowup in convoy chains.

## Complete Implementation Structure
A complete DATC-compliant implementation in Go should extend your existing `adjudicator.go` and `ResolutionEngine` interface from Task 17. Structure it as a modular engine with components for validation, graph building, resolution passes, and outcome application. Here's a high-level struct:

```go
type DATCCompliantEngine struct {
    Validator      *OrderValidator     // Validates orders per DATC rules
    DependencyGraph *OrderGraph         // Tracks dependencies (e.g., supports, convoys)
    Resolver       *ConflictResolver   // Handles strength and conflicts
    ConvoyHandler  *ConvoyEngine       // Manages convoy paths and paradoxes
    BackupRules    *BackupRuleSet      // Fallback for unresolved cases
    Config         EngineConfig        // e.g., enable Szykman rule
}

func (e *DATCCompliantEngine) Resolve(orders []Order, board *Board) (*ResolutionResult, error) {
    // Phase 1-4 implementation here
    return &ResolutionResult{}, nil
}
```

This structure supports Task 5's iterative implementation: Start with basic moves (already partially done, as 117/163 tests pass), then add convoy and coastal logic. Multiple approaches include:
- **Iterative Multi-Pass**: Loop through phases until no changes (as in Task 18).
- **Graph-Based**: Build a dependency graph and resolve via topological sort.
- **Rule-Based Engine**: Use a rules engine (e.g., with Go's `github.com/hyperjumptech/grule-rule-engine`) for flexible DATC rule application, though this adds overhead.

For your project, integrate this with **Task 9: Develop RESTful API Endpoints** by exposing resolution via POST /games/:id/orders, secured with JWT from **Task 10**.

## Phases of Resolution
DATC resolution follows a 4-phase roadmap as documented in your `datc-implementation-roadmap.md` (aligned with Task 5). These phases ensure progressive resolution, handling dependencies iteratively.

### Phase 1: Order Validation
Validate all orders against board state (e.g., unit exists, move is legal). Reference Task 17's `ValidateOrders` method.

- Step-by-Step:
  1. Check syntax: Order format (e.g., "A PAR - MAR").
  2. Check legality: Unit type matches action (fleets can't move inland).
  3. Initial dependency scan: Flag potential convoys or supports.

Code Snippet:
```go
func (v *OrderValidator) Validate(order Order, board *Board) ValidationError {
    if !board.UnitExists(order.UnitID) {
        return ValidationError{"Unit not found"}
    }
    // Add DATC-specific checks, e.g., coastal rules
    return nil
}
```

### Phase 2: Strength Calculation
Compute attack/defend strengths, incorporating supports. This phase identifies initial conflicts.

- Approaches: Direct calculation vs. graph propagation (for efficiency in large games).
- Integration: Feeds into Task 18's strength algorithm.

### Phase 3: Conflict Resolution
Resolve head-to-heads, bounces, and dislodgements. Use multi-pass until convergence.

- Step-by-Step:
  1. Identify contested provinces.
  2. Compare strengths: Higher strength succeeds; ties bounce.
  3. Handle supports: Cut if supporter is attacked successfully.

### Phase 4: Outcome Determination and Backup
Apply results, invoking backups for paradoxes. This finalizes `ResolutionResult` for Task 22's state transition.

Edge Case: Circular movement (DATC 6.A.3) – resolve via dependency graph to detect cycles.

## Dependency Graph
Model orders as a directed graph where nodes are orders, and edges represent dependencies (e.g., support edges, convoy paths). This is essential for resolution order and detecting cycles in convoys (relevant to Task 5's CONVOYS and ADJACENT CONVOYS categories).

- **Graph Structure**:
  - Nodes: Order structs with attributes (type, source, target, strength).
  - Edges: "DependsOn" (e.g., move depends on convoy), "Supports", "Attacks".
  - Building: Traverse orders to add edges; use `github.com/emirpasic/gods` for graph ops in Go.

Code Snippet for Building Graph:
```go
type OrderGraph struct {
    Nodes map[OrderID]*OrderNode
    Edges map[OrderID][]OrderID // Directed edges
}

func BuildDependencyGraph(orders []Order) *OrderGraph {
    g := &OrderGraph{Nodes: make(map[OrderID]*OrderNode), Edges: make(map[OrderID][]OrderID)}
    for _, ord := range orders {
        g.Nodes[ord.ID] = &OrderNode{Order: ord}
        // Add edges, e.g., for supports: g.Edges[supported.ID] = append(g.Edges[supported.ID], ord.ID)
    }
    // Scan for convoy chains
    return g
}
```

- Resolution Using Graph: Perform topological sort to determine order; if cycles, apply backups.
- Perspectives: Graph approach excels for complex dependencies (e.g., 19 ADJACENT CONVOYS tests) but may overkill simple cases—hybrid with iterative passes for performance.

Apply to Project: Use this in **Task 18** to enhance `ConflictResolver`; profile with pprof as per Task 5 to ensure scalability for real-time updates in **Task 11**.

## Resolution Order
DATC mandates a specific order to avoid ambiguity: Resolve independent orders first, then dependents. Use the dependency graph for topological ordering.

- Step-by-Step Guidance:
  1. Build graph (Phase 1).
  2. Detect cycles: If present, flag for backup rules.
  3. Topological sort: Resolve nodes with no incoming edges first (e.g., unsupported holds).
  4. Propagate: Update strengths and outcomes downstream.
  5. Iterate if needed (multi-pass for mutual dependencies).

Alternative: Rule-priority ordering (e.g., resolve all holds before moves), but graph-based is more DATC-compliant for edges like CIVIL DISORDER (11 tests).

Pitfalls: Infinite loops in cycle detection—mitigate with a max-iteration limit (e.g., 100 passes) and log for debugging.

Example: In a head-to-head (DATC 6.A.1), order A attacks B, B attacks A—resolve as bounce after strength calc shows tie.

## Backup Rules
Backup rules resolve paradoxes, especially in convoys (e.g., "convoy paradox" where a fleet is dislodged mid-convoy). DATC specifies several, with your config choosing one (e.g., via `EngineConfig`).

- **Key Rules**:
  - **Szykman Rule**: If a convoying fleet is dislodged, the convoy fails only if the dislodgement succeeds.
  - **All-or-Nothing**: Entire convoy chain must succeed or fail together.
  - **Backup for Circular Convoys**: Assume failure if cycle unresolved after passes.

Code Snippet:
```go
type BackupRuleSet struct {
    UseSzykman bool
}

func (b *BackupRuleSet) ApplyParadox(convoyOrder Order, result *ResolutionResult) {
    if b.UseSzykman && result.IsDislodged(convoyOrder.Target) {
        result.SetOutcome(convoyOrder, OutcomeFailed)
    }
}
```

- Multiple Perspectives: Szykman is conservative (preserves intent), while all-or-nothing is strict—test both against DATC for compliance (e.g., 12 CONVOYS tests).
- Edge Cases: Dislodged supporter in coastal issues (11 tests)—backup ensures no void outcomes.
- Pitfalls: Over-applying backups can break determinism; always validate against full DATC suite as in Task 5's test strategy.

## Edge Cases and Potential Pitfalls
- **Edge Cases**: 
  - Coastal ambiguities (e.g., fleet in Spain(nc) moving to Spain(sc)—DATC 6.D.1; handle with explicit coast tracking in board model).
  - Adjacent convoys (19 tests): Graph must detect overlapping paths.
  - Civil disorder (11 tests): Unordered units hold with strength 1, but backups apply if dislodged without support.
  - Performance: Large games with 100+ orders—pitfall: Graph building O(n^2); optimize with adjacency lists.
- **Pitfalls**: 
  - Mutable state during resolution—use immutable copies as in Task 22 to avoid bugs.
  - Incomplete validation: Misses like invalid coasts fail tests; integrate with Task 8's DB for persistent state.
  - Regression: New implementations break existing 117 passing tests—use regression testing per Task 5.

## Integration with Project Tasks and Actionable Insights
This structure directly applies to **Task 5** by providing a blueprint to tackle the 4 incomplete categories: Implement coastal logic in Phase 1, convoy graphs in Phase 3, etc., then run the full suite. For **Task 18**, extend the multi-pass with graph resolution. In **Task 22**, use the generated `ResolutionResult` for immutable state updates. Actionable Steps:
1. Update `adjudicator.go` with the `DATCCompliantEngine` struct.
2. Implement graph building and test against 5-10 DATC cases from failing categories.
3. Profile and optimize for **Task 11**'s real-time WebSocket broadcasts.
4. Secure endpoints in **Task 9** to trigger resolutions, using JWT from **Task 10**.

This exhaustive approach should get you to 100% compliance efficiently— if needed, speculate (marked): A rule engine could future-proof for variant rules, but stick to core DATC for now.


---

*Generated by Task Master Research Command*  
*Timestamp: 2025-07-16T22:45:16.339Z*
