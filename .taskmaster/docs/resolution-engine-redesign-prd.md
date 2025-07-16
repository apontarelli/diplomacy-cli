# Resolution Engine Redesign - Product Requirements Document

## Overview

The current resolution engine has architectural limitations that make DATC compliance validation difficult and the codebase harder to maintain. This PRD outlines a redesign focused on clean architecture, rich domain models, and comprehensive observability to achieve full DATC compliance and long-term maintainability.

## Problem Statement

### Current Issues
1. **Information Loss**: Resolution details are lost after state transition, making DATC validation impossible
2. **Complex Reconstruction**: Pipeline code does complex work to figure out what the resolution engine decided
3. **Poor Testability**: Cannot validate intermediate resolution steps or reasoning
4. **Debugging Difficulty**: Hard to understand why specific resolutions occurred
5. **Architectural Debt**: Multi-phase application logic suggests insufficient resolution detail

### Impact
- 163 official DATC tests (171 total test functions) cannot be properly validated (only checking "no crash" vs actual outcomes)
- Game engine correctness cannot be verified against official Diplomacy rules
- Maintenance burden from complex state reconstruction logic
- Risk of subtle bugs in complex scenarios (convoys, paradoxes, circular movement)

## Goals

### Primary Goals
1. **DATC Compliance**: Enable validation of all 163 DATC test cases with specific outcome checking
2. **Clean Architecture**: Single responsibility components with clear interfaces
3. **Rich Observability**: Complete information about resolution reasoning and outcomes
4. **Maintainability**: Self-documenting code that's easy to understand and extend

### Secondary Goals
1. **Performance**: Maintain or improve current resolution performance
2. **Backward Compatibility**: Preserve existing game state and order interfaces
3. **Extensibility**: Easy to add new order types and resolution rules

## Solution Architecture

### Core Components

#### 1. Rich Domain Models
```go
type ResolutionResult struct {
    TurnNumber     int
    Phase          Phase
    UnitOutcomes   map[UnitID]UnitOutcome
    Conflicts      []Conflict
    SupportResults map[SupportID]SupportOutcome
    ConvoyResults  map[ConvoyID]ConvoyOutcome
    ResolutionLog  []ResolutionStep  // For debugging/DATC validation
}

type UnitOutcome struct {
    Unit         Unit
    OrderGiven   Order
    FinalStatus  UnitStatus  // Moved, Held, Dislodged, Bounced
    FromProvince Province
    ToProvince   *Province   // nil if didn't move
    DislodgedBy  *UnitID     // who dislodged this unit
    Reason       string      // human-readable explanation
    Strength     int         // final calculated strength
}

type Conflict struct {
    Province     Province
    Competitors  []UnitID
    Winner       *UnitID     // nil if bounce
    Resolution   ConflictType // HeadToHead, Support, Convoy, etc.
    Explanation  string
}
```

#### 2. Clean Resolution Engine Interface
```go
type ResolutionEngine interface {
    // Resolve all orders and return comprehensive results
    Resolve(orders []Order, board *Board) (*ResolutionResult, error)
    
    // Validate orders before resolution (syntax + semantic)
    ValidateOrders(orders []Order, board *Board) ([]ValidationError, error)
}

type DATCCompliantEngine struct {
    // Implementation that produces rich ResolutionResult
}
```

#### 3. DATC Validation Framework
```go
type DATCValidator struct {
    testCases map[string]DATCTestCase
}

func (v *DATCValidator) ValidateTest(
    testID string,
    initialState *GameState,
    orders []Order,
    result *ResolutionResult,
) ValidationResult {
    // Parse expected outcomes from DATC specification
    // Validate against rich ResolutionResult data
    // Return detailed pass/fail with explanations
}
```

#### 4. State Transition Simplification
```go
func ApplyResolution(
    gameState *GameState, 
    result *ResolutionResult,
) *GameState {
    // Simple application of explicit outcomes
    // No complex reconstruction logic needed
    // All information is in ResolutionResult
}
```

### Data Flow
```
Orders → Validation → Resolution → ResolutionResult → GameState
                                        ↓
                                  DATC Validation
                                  Debug Logging
                                  AI Analysis
```

## Technical Requirements

### Functional Requirements

#### FR1: Rich Resolution Results
- **Requirement**: Resolution engine must produce comprehensive outcome data
- **Acceptance Criteria**: 
  - Every unit's fate is explicitly recorded (moved/held/dislodged/bounced)
  - All conflicts are documented with reasoning
  - Support and convoy outcomes are tracked
  - Human-readable explanations for all decisions

#### FR2: DATC Test Validation
- **Requirement**: All 163 official DATC tests (171 total test functions) must validate specific expected outcomes
- **Acceptance Criteria**:
  - Tests check actual unit positions, not just "no crash"
  - Failed tests show expected vs actual outcomes clearly
  - Complex scenarios (convoys, paradoxes) are properly validated

#### FR3: Immutable Results
- **Requirement**: ResolutionResult must be immutable after creation
- **Acceptance Criteria**:
  - No methods that modify ResolutionResult after construction
  - Thread-safe for concurrent access
  - Prevents accidental state corruption

#### FR4: Performance Parity
- **Requirement**: New engine must match or exceed current performance
- **Acceptance Criteria**:
  - Resolution time ≤ current implementation for standard games
  - Memory usage within 20% of current implementation
  - Benchmarks for complex scenarios (7-player, many units)

### Non-Functional Requirements

#### NFR1: Code Quality
- All public APIs have comprehensive documentation
- 90%+ test coverage for resolution engine
- No cyclic dependencies between packages
- Clear separation of concerns

#### NFR2: Maintainability
- Self-documenting domain models
- Minimal cognitive complexity in resolution logic
- Easy to add new order types or resolution rules
- Clear error messages for debugging

#### NFR3: Backward Compatibility
- Existing Order and GameState interfaces preserved
- Current game serialization format maintained
- Existing tests continue to pass (with enhanced validation)

## Implementation Plan

### Phase 1: Domain Model Design (Week 1)
- Design ResolutionResult and related structures
- Create comprehensive test cases for domain models
- Validate design with complex DATC scenarios
- **Deliverable**: Complete domain model with tests

### Phase 2: Resolution Engine Core (Week 2)
- Implement new resolution algorithm with rich output
- Focus on move resolution, conflicts, and basic support
- Ensure all information needed for DATC validation is captured
- **Deliverable**: Core resolution engine with basic functionality

### Phase 3: Advanced Features (Week 3)
- Implement convoy resolution with paradox handling
- Add retreat and build phase support
- Handle civil disorder and edge cases
- **Deliverable**: Complete resolution engine

### Phase 4: DATC Integration (Week 4)
- Build DATC validation framework
- Convert existing tests to use rich validation
- Achieve 100% DATC compliance
- **Deliverable**: Fully DATC-compliant engine

### Phase 5: Integration & Performance (Week 5)
- Integrate with existing pipeline
- Performance optimization and benchmarking
- Documentation and cleanup
- **Deliverable**: Production-ready resolution engine

## Success Metrics

### Primary Metrics
1. **DATC Compliance**: 163/163 official DATC tests (171/171 total test functions) passing with specific outcome validation
2. **Code Quality**: 90%+ test coverage, zero cyclic dependencies
3. **Performance**: Resolution time ≤ current implementation

### Secondary Metrics
1. **Maintainability**: Reduced complexity in state transition logic
2. **Debuggability**: Rich logging and explanation capabilities
3. **Extensibility**: Easy addition of new features (measured by LOC for new order types)

## Risk Assessment

### High Risk
- **Performance Regression**: Rich data structures may impact performance
  - *Mitigation*: Benchmark early, optimize hot paths, consider lazy evaluation
- **Scope Creep**: Complex DATC scenarios may require architectural changes
  - *Mitigation*: Design for most complex scenarios upfront, validate with hardest tests

### Medium Risk
- **Integration Complexity**: Existing code may have hidden dependencies on current behavior
  - *Mitigation*: Comprehensive integration testing, gradual rollout
- **Timeline**: 5 weeks may be optimistic for full DATC compliance
  - *Mitigation*: Prioritize core functionality, defer edge cases if needed

### Low Risk
- **Backward Compatibility**: Well-defined interfaces should prevent breaking changes
- **Team Knowledge**: Current team understands both existing code and DATC requirements

## Alternatives Considered

### Alternative 1: Incremental Enhancement
- **Approach**: Add event sourcing to existing pipeline
- **Rejected Because**: Maintains architectural complexity, doesn't address root causes

### Alternative 2: Property-Based Testing
- **Approach**: Use DATC as oracle for generated test cases
- **Rejected Because**: Doesn't solve observability issues, adds complexity

### Alternative 3: External Validation Service
- **Approach**: Separate service for DATC validation
- **Rejected Because**: Doesn't improve core architecture, adds deployment complexity

## Conclusion

The resolution engine redesign addresses fundamental architectural issues while enabling full DATC compliance. The investment in clean architecture will pay dividends in maintainability, debuggability, and correctness. The 5-week timeline is aggressive but achievable with focused execution and clear priorities.

The rich domain models and comprehensive observability will not only solve the immediate DATC validation problem but also enable future features like AI analysis, replay systems, and advanced debugging tools.