package resolution

import (
	"diplomacy-cli/backend/internal/game"
	"fmt"
)

// ResolutionResultBuilder provides a builder pattern for constructing ResolutionResult
type ResolutionResultBuilder struct {
	result *ResolutionResult
	errors []error
}

// NewResolutionResultBuilder creates a new builder for ResolutionResult
func NewResolutionResultBuilder(turn int, phase game.Phase) *ResolutionResultBuilder {
	return &ResolutionResultBuilder{
		result: NewResolutionResult(turn, phase),
		errors: make([]error, 0),
	}
}

// WithUnitOutcome adds a unit outcome to the resolution result
func (b *ResolutionResultBuilder) WithUnitOutcome(unitID UnitID, outcome UnitOutcome) *ResolutionResultBuilder {
	b.result.AddUnitOutcome(unitID, outcome)
	return b
}

// WithConflict adds a conflict to the resolution result
func (b *ResolutionResultBuilder) WithConflict(conflict Conflict) *ResolutionResultBuilder {
	b.result.AddConflict(conflict)
	return b
}

// WithSupportOutcome adds a support outcome to the resolution result
func (b *ResolutionResultBuilder) WithSupportOutcome(supportID SupportID, outcome SupportOutcome) *ResolutionResultBuilder {
	b.result.AddSupportOutcome(supportID, outcome)
	return b
}

// WithConvoyOutcome adds a convoy outcome to the resolution result
func (b *ResolutionResultBuilder) WithConvoyOutcome(convoyID ConvoyID, outcome ConvoyOutcome) *ResolutionResultBuilder {
	b.result.AddConvoyOutcome(convoyID, outcome)
	return b
}

// WithResolutionStep adds a resolution step to the log
func (b *ResolutionResultBuilder) WithResolutionStep(step ResolutionStep) *ResolutionResultBuilder {
	b.result.AddResolutionStep(step)
	return b
}

// Build finalizes the construction and returns the ResolutionResult
func (b *ResolutionResultBuilder) Build() (*ResolutionResult, error) {
	if len(b.errors) > 0 {
		return nil, fmt.Errorf("validation errors: %v", b.errors)
	}

	if err := b.result.Validate(); err != nil {
		return nil, fmt.Errorf("result validation failed: %w", err)
	}

	return b.result, nil
}
