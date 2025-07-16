# DATC Engine Core Invariants and Assumptions

## Architecture Principles

### 1. DATC Partial Information Algorithm Compliance
- **Invariant**: All resolution follows the exact DATC partial information algorithm
- **Implementation**: `resolve()` method with optimistic/pessimistic evaluation
- **Guarantee**: Deterministic resolution that matches DATC specification

### 2. Order-by-Order Resolution
- **Invariant**: Orders are resolved individually using recursive dependency resolution
- **Implementation**: Each order calls `resolve()` which may recursively resolve dependencies
- **Guarantee**: No order is resolved until all its dependencies are known or cycles are detected

### 3. Cycle Detection and Backup Rules
- **Invariant**: Dependency cycles are detected and resolved using DATC backup rules
- **Implementation**: `cycle` tracking array and `recursionHits` counter
- **Guarantee**: All cycles result in failed moves (backup rule application)

## Strength Calculation Invariants

### 4. Four Types of Strength (DATC 5.B)
- **Attack Strength**: 1 + supports (0 if PATH fails)
- **Hold Strength**: 1 (if occupied) + hold supports (0 if unit moves away successfully)
- **Defend Strength**: Same as attack strength (for head-to-head battles)
- **Prevent Strength**: Same as attack strength (for competing moves)

### 5. PATH Validation (DATC 5.B.6, 5.B.8)
- **Invariant**: Attack and prevent strength are 0 if PATH fails
- **Implementation**: `hasValidPath()` check before strength calculation
- **Guarantee**: Invalid moves (no path) cannot succeed

### 6. Optimistic/Pessimistic Evaluation
- **Invariant**: Uncertain information is evaluated both optimistically and pessimistically
- **Implementation**: All strength calculations take `optimistic` parameter
- **Guarantee**: Orders with definitive outcomes are resolved immediately

## Resolution Logic Invariants

### 7. Move Success Conditions
- **Single Move**: Attack strength > Hold strength
- **Multiple Moves**: Highest prevent strength > Hold strength AND unique winner
- **Equal Strength**: Results in standoff (all moves fail)

### 8. Support Cutting Rules
- **Invariant**: Support is cut only if the attacking move actually succeeds
- **Implementation**: `adjudicateSupport()` checks if attackers succeed
- **Guarantee**: Failed attacks do not cut support

### 9. Hold Strength for Non-Moving Units
- **Invariant**: Units not moving (Hold, Support, Convoy) provide hold strength of 1
- **Implementation**: `calculateHoldStrength()` checks order type
- **Guarantee**: Supporting/convoying units can still defend themselves

## Data Structure Invariants

### 10. Order State Management
- **Invariant**: Each order has exactly one resolution state (resolved/unresolved)
- **Fields**: `isResolved`, `resolution`, `isVisited`
- **Guarantee**: No order can be in an inconsistent state

### 11. Immutable Order Processing
- **Invariant**: Original order data is never modified during resolution
- **Implementation**: Resolution state is separate from order data
- **Guarantee**: Orders can be re-resolved with different parameters

### 12. Cycle State Isolation
- **Invariant**: Cycle detection state is reset between top-level resolutions
- **Implementation**: `cycle`, `recursionHits`, `uncertain` reset per order
- **Guarantee**: No cross-contamination between order resolutions

## Performance and Scalability Assumptions

### 13. Bounded Recursion
- **Assumption**: Maximum recursion depth prevents infinite loops
- **Implementation**: `maxRecursionDepth = 100` safety limit
- **Guarantee**: Algorithm terminates even with complex cycles

### 14. Linear Order Processing
- **Assumption**: Orders are processed in input order for deterministic results
- **Implementation**: `orderedOrders` slice maintains original order
- **Guarantee**: Consistent resolution regardless of internal data structures

### 15. Efficient Dependency Resolution
- **Assumption**: Most orders resolve without deep recursion
- **Implementation**: Early termination when optimistic == pessimistic
- **Guarantee**: Performance scales well with typical game scenarios

## Error Handling Invariants

### 16. Graceful Degradation
- **Invariant**: Invalid or malformed orders fail gracefully
- **Implementation**: Type checks and validation in adjudication methods
- **Guarantee**: Engine never crashes on bad input

### 17. Deterministic Failure Modes
- **Invariant**: Same input always produces same output
- **Implementation**: No random elements or undefined behavior
- **Guarantee**: Reproducible results for debugging and testing

## Testing and Validation Assumptions

### 18. DATC Test Case Compliance
- **Assumption**: Engine passes all relevant DATC test cases
- **Implementation**: Comprehensive test suite covering edge cases
- **Guarantee**: Behavior matches official Diplomacy rules

### 19. Regression Prevention
- **Assumption**: Changes maintain backward compatibility
- **Implementation**: Full test suite runs on every change
- **Guarantee**: No unintended behavior changes

## Future Extension Points

### 20. Variant Rule Support
- **Design**: Core algorithm can be extended for rule variants
- **Implementation**: Configurable behavior through engine parameters
- **Guarantee**: Base DATC compliance is never compromised

### 21. Performance Optimization
- **Design**: Algorithm structure allows for caching and memoization
- **Implementation**: Stateless strength calculations enable optimization
- **Guarantee**: Optimizations preserve correctness