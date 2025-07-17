---
title: Research Session
query: "DATC partial information algorithm detailed explanation optimistic pessimistic evaluation circular dependencies resolution order"
date: 7/16/2025
time: 3:36:32 PM
timestamp: 2025-07-16T22:36:32.137Z
exchanges: 1
---

# Research Session

## Initial Query

**Question:** DATC partial information algorithm detailed explanation optimistic pessimistic evaluation circular dependencies resolution order

**Response:**

# DATC Partial Information Algorithm: Detailed Explanation with Optimistic/Pessimistic Evaluation, Circular Dependencies, and Resolution Order

## Introduction to DATC and Its Relevance to the Project

The Diplomacy Adjudication Test Cases (DATC) serve as the de facto standard for validating Diplomacy game adjudicators, encompassing 163 official test cases from datc.zine.net that cover everything from basic moves to complex paradoxes. In the context of your project, which involves building a Go-based Diplomacy engine with tasks like Task 17 (Resolution Engine Interface), Task 18 (Core Move Resolution Algorithm), and Task 21 (DATC Validation Framework), understanding DATC's partial information algorithm is crucial. This algorithm addresses how adjudicators resolve orders under uncertainty, particularly in scenarios with interdependent orders where full information isn't immediately available. It employs multi-pass resolution, dependency tracking, and evaluation strategies to handle ambiguities.

Partial information refers to situations where an order's success depends on outcomes not yet determined, such as in circular movements or convoy paradoxes. DATC specifies a structured approach to resolve these without infinite loops, using concepts like optimistic and pessimistic evaluations to break ties. This directly ties into Task 18's multi-pass resolution architecture and strength calculation, where iterative passes ensure convergence. For Task 21, implementing a parser that validates against DATC expected outcomes will require emulating this algorithm precisely to pass tests involving circular dependencies. By integrating this, your DATCCompliantEngine (from Task 17) can achieve full compliance, enhancing backward compatibility in Task 25's production integration.

From multiple perspectives, the algorithm can be viewed as: (1) a graph-based dependency resolver (modeling orders as nodes with edges for dependencies), (2) a phased evaluator (separating validation, strength calc, and resolution), or (3) a heuristic-based paradox breaker (using optimism/pessimism). In performance terms (relevant to Task 24), inefficient handling of partial info can lead to exponential time complexity in large games, so optimizations like lazy evaluation are key.

## Core Components of the DATC Partial Information Algorithm

The DATC partial information algorithm is designed to adjudicate orders when dependencies create uncertainty. It operates on the principle of iterative refinement: start with partial assumptions, resolve what can be resolved, and propagate information until all orders are determined. This is outlined in DATC section 6 (Advanced Topics), emphasizing that adjudicators must not assume outcomes prematurely but use safe, incremental steps.

Step-by-step breakdown:
1. **Initial Validation Pass**: Validate orders syntactically and semantically using known board state (e.g., unit positions, provinces). This aligns with Task 17's ValidateOrders method, catching invalid moves early.
2. **Dependency Graph Construction**: Model orders as a graph where nodes are orders, and directed edges indicate dependencies (e.g., a support order depends on the supported move's path not being cut).
3. **Partial Resolution Passes**: Resolve independent orders first, then use their outcomes to inform dependents. Repeat until no changes occur.
4. **Handling Uncertainty**: For unresolved dependencies, apply evaluation strategies (detailed below).
5. **Convergence Check**: Ensure the process terminates, typically within a bounded number of passes (DATC recommends detecting non-convergence as an error, though rare).

In your project, this can be implemented in the ConflictResolver struct (from Task 17) as a loop over resolution phases, tracking a "dirty" flag for orders needing re-evaluation. For example, in Go:

```go
type ResolutionState struct {
    Orders       []Order
    Dependencies map[OrderID][]OrderID // Dependency graph
    Resolved     map[OrderID]Outcome
    Dirty        bool
}

func (r *ConflictResolver) ResolvePartial(state *ResolutionState) {
    state.Dirty = true
    for state.Dirty {
        state.Dirty = false
        for _, order := range state.Orders {
            if _, ok := state.Resolved[order.ID]; !ok {
                if allDepsResolved(order, state) {
                    outcome := evaluateOrder(order, state)
                    state.Resolved[order.ID] = outcome
                    state.Dirty = true // Mark for next pass
                }
            }
        }
    }
}
```

This multi-pass approach from Task 18 ensures efficient handling, but watch for performance in large graphs (optimize with topological sorting for Task 24).

## Optimistic and Pessimistic Evaluation Strategies

DATC introduces optimistic and pessimistic evaluations to resolve ambiguities in partial information scenarios, particularly when direct computation fails due to cycles. Optimistic evaluation assumes the best-case scenario (e.g., a convoy succeeds if possible), while pessimistic assumes the worst (e.g., failure if any doubt exists). These are not arbitrary; DATC specifies their use in specific paradoxes, like the "convoy paradox" (test 6.D.1-6.D.4), where a fleet convoys an army that might dislodge it.

From a rules perspective, optimism is used for "positive" resolutions (e.g., assuming a move succeeds to check if it enables another), while pessimism handles "negative" cases (e.g., assuming failure to confirm a dislodgement). Multiple approaches exist: some adjudicators (like DPTG) use pure optimism for speed, but DATC mandates a hybrid for accuracy.

Detailed application:
- **Optimistic Evaluation**: Temporarily assume success for unresolved dependencies and compute. If consistent, accept; else, retry.
- **Pessimistic Evaluation**: Assume failure and compute. Used to break ties in circular supports.

Example in a circular move (A to B, B to A): Optimistically assume both succeed (impossible), then pessimistically evaluate strengths assuming mutual failure, leading to a bounce.

In your engine, integrate this into strength calculation (Task 18):

```go
func calculateStrength(order Order, state *ResolutionState, mode EvaluationMode) int {
    base := 1 // Unit's own strength
    for _, support := range getSupports(order) {
        if mode == Optimistic {
            // Assume support holds unless proven cut
            if !isDefinitelyCut(support, state) {
                base++
            }
        } else { // Pessimistic
            // Assume cut if possibly cut
            if isPossiblyCut(support, state) {
                continue
            }
            base++
        }
    }
    return base
}
```

Edge cases: In convoy paradoxes, optimism can lead to self-dislodgement illusions—DATC clarifies no self-dislodgement, so add checks. Pitfalls: Over-optimism causes infinite loops; mitigate with pass limits.

## Handling Circular Dependencies

Circular dependencies are a hallmark of Diplomacy complexity, occurring in moves (e.g., unit swaps), supports (circular support chains), or convoys (fleet chains depending on each other). DATC's algorithm resolves them by detecting cycles in the dependency graph and applying evaluation strategies to break them.

Approaches:
- **Graph Cycle Detection**: Use DFS to identify cycles, then isolate and resolve subgraphs.
- **SORE (Simultaneous Order Resolution with Evaluation)**: A DATC-recommended method where cycles are evaluated simultaneously using optimism/pessimism.
- **Iterative Weakening**: Gradually reduce assumed strengths in cycles until resolution.

For resolution order in cycles: Process in topological order where possible, falling back to evaluation for cycles. DATC test 6.A.5 (circular movement) expects all moves to fail if strengths are equal.

Step-by-step resolution:
1. Build graph.
2. Topo-sort and resolve acyclic parts.
3. For cycles, apply pessimistic evaluation to assume minimal strengths, resolve, then verify with optimism.
4. Propagate outcomes.

In project terms, this enhances Task 18's dependency tracking. Code snippet:

```go
func resolveCycle(cycle []Order, state *ResolutionState) {
    // Pessimistic pass: assume failures
    tempOutcomes := make(map[OrderID]Outcome)
    for _, ord := range cycle {
        strength := calculateStrength(ord, state, Pessimistic)
        tempOutcomes[ord.ID] = determineOutcome(ord, strength)
    }
    // Verify with optimistic pass for consistency
    consistent := true
    for _, ord := range cycle {
        optStrength := calculateStrength(ord, state.WithTemp(tempOutcomes), Optimistic)
        if determineOutcome(ord, optStrength) != tempOutcomes[ord.ID] {
            consistent = false
            break
        }
    }
    if consistent {
        applyOutcomes(tempOutcomes, state)
    } else {
        // Fallback: all bounce (DATC default for irresolvable)
        for _, ord := range cycle {
            state.Resolved[ord.ID] = OutcomeBounce
        }
    }
}
```

Perspectives: Graph theory view minimizes recomputation; heuristic view prioritizes common cases for speed (Task 24).

## Resolution Order and Multi-Pass Integration

Resolution order in DATC is not arbitrary; it's phased: (1) Validate all, (2) Resolve moves/supports without dependencies, (3) Handle convoys, (4) Resolve conflicts, (5) Determine retreats/builds. For partial info, order is iterative, prioritizing "safe" resolutions.

In circular cases, DATC implies no specific order beyond consistency—adjudicators must produce the same outcome regardless of processing sequence. This is tested in 6.B series.

Apply to project: In Task 18, structure as phases in Resolve method. For Task 4 (Build Phases), extend to adjustment resolutions, ensuring partial info doesn't affect supply calcs.

## Examples and Step-by-Step Guidance

Example: DATC 6.A.1 (Simple Circular Move).
- Orders: Army Paris to Burgundy, Army Burgundy to Paris.
- Step 1: Validate (both valid).
- Step 2: Dependency graph: Mutual dependency.
- Step 3: Pessimistic eval: Each has strength 1 (assume other succeeds, but cycle detected).
- Step 4: Both bounce.
- Outcome: Units hold.

Code integration: Use in DATCCompliantEngine's Resolve.

Another: Convoy Paradox (6.D.1).
- Fleet convoys army that attacks it.
- Optimistic: Assume convoy succeeds, check if attack dislodges—paradox.
- Pessimistic: Assume convoy fails, attack doesn't dislodge.
- DATC resolution: Convoy fails, no dislodgement.

## Edge Cases and Potential Pitfalls

Edge cases:
- Infinite cycles: Rare, but add max-passes (e.g., 100) with error (Task 24 monitoring).
- Ambiguous convoys: DATC 6.E series—pessimistic assumes no path if disrupted.
- Multi-player circles: 7-player games (Task 24 benchmarks)—scale with efficient graphs.
- Backward compatibility: Ensure matches old adjudicator.go (Task 25).

Pitfalls: Assuming global optimism leads to DATC failures; always hybrid. Memory: Large graphs cause allocations—use pools (Task 24).

## Actionable Insights for Project Tasks

For Task 18, extend multi-pass with optimism/pessimism in strength calc. Task 21: Parser must extract expected outcomes for these scenarios, validating ResolutionResult's reasoning field. Task 25: Feature-flag new logic for A/B testing. Task 24: Benchmark cycles with profiling. Overall, this ensures DATC compliance, boosting engine reliability. If implementing, start with graph lib like gonum for dependencies.


---

*Generated by Task Master Research Command*  
*Timestamp: 2025-07-16T22:36:32.137Z*
