package resolution

// convoy.go - Minimal convoy support for DATC engine
// This file contains only the convoy functionality needed by the DATC engine

// ConvoyPath represents a path through convoy fleets
type ConvoyPath struct {
	Source      string
	Destination string
	Fleets      []string
	Valid       bool
}

// ConvoyResolver handles convoy resolution for the DATC engine
type ConvoyResolver struct {
	// Minimal implementation for now
}

// NewConvoyResolver creates a new convoy resolver
func NewConvoyResolver() *ConvoyResolver {
	return &ConvoyResolver{}
}

// ResolveConvoy resolves a convoy operation
func (cr *ConvoyResolver) ResolveConvoy(convoyID ConvoyID, orders []Order) ConvoyOutcome {
	// Simple implementation: convoy succeeds if not disrupted
	// TODO: Implement proper convoy path validation
	return NewConvoyOutcome(
		UnitID{},               // convoyingFleet - TODO: extract from convoyID
		UnitID{},               // convoyTarget - TODO: extract from convoyID
		true,                   // successful
		false,                  // disrupted
		"Convoy not disrupted", // reason
		[]string{},             // path - TODO: calculate actual path
		[]UnitID{},             // disruptedBy
	)
}
