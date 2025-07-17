package resolution

import (
	"sync"
)

// convoy.go - Enhanced convoy support with lazy evaluation for DATC engine

// ConvoyPath represents a path through convoy fleets
type ConvoyPath struct {
	Source      string
	Destination string
	Fleets      []string
	Valid       bool
}

// LazyConvoyPath provides lazy evaluation for convoy path validation
type LazyConvoyPath struct {
	source      string
	destination string
	orders      []*Order
	engine      *DATCEngine

	// Lazy evaluation state
	evaluated bool
	valid     bool
	path      []string
	fleets    []*Order
	mu        sync.RWMutex
}

// ConvoyResolver handles convoy resolution for the DATC engine
type ConvoyResolver struct {
	pathCache map[string]*LazyConvoyPath
	mu        sync.RWMutex
}

// NewConvoyResolver creates a new convoy resolver
func NewConvoyResolver() *ConvoyResolver {
	return &ConvoyResolver{
		pathCache: make(map[string]*LazyConvoyPath),
	}
}

// GetLazyConvoyPath returns a lazy convoy path validator
func (cr *ConvoyResolver) GetLazyConvoyPath(source, destination string, orders []*Order, engine *DATCEngine) *LazyConvoyPath {
	key := source + "->" + destination

	cr.mu.RLock()
	if path, exists := cr.pathCache[key]; exists {
		cr.mu.RUnlock()
		return path
	}
	cr.mu.RUnlock()

	cr.mu.Lock()
	defer cr.mu.Unlock()

	// Double-check after acquiring write lock
	if path, exists := cr.pathCache[key]; exists {
		return path
	}

	path := &LazyConvoyPath{
		source:      source,
		destination: destination,
		orders:      orders,
		engine:      engine,
	}

	cr.pathCache[key] = path
	return path
}

// IsValid lazily evaluates and returns whether the convoy path is valid
func (lcp *LazyConvoyPath) IsValid(optimistic bool) bool {
	lcp.mu.RLock()
	if lcp.evaluated {
		result := lcp.valid
		lcp.mu.RUnlock()
		return result
	}
	lcp.mu.RUnlock()

	lcp.mu.Lock()
	defer lcp.mu.Unlock()

	// Double-check after acquiring write lock
	if lcp.evaluated {
		return lcp.valid
	}

	// Perform the expensive convoy path validation
	lcp.valid = lcp.validateConvoyPath(optimistic)
	lcp.evaluated = true

	return lcp.valid
}

// validateConvoyPath performs the actual convoy path validation
func (lcp *LazyConvoyPath) validateConvoyPath(optimistic bool) bool {
	// Look for convoy orders that could support this move
	var convoyingFleets []*Order

	for _, order := range lcp.orders {
		if order.Type == Convoy && lcp.isConvoyingMove(order) {
			// Check if the convoy succeeds (using optimistic evaluation for convoy success)
			if lcp.engine.resolve(order, optimistic) {
				convoyingFleets = append(convoyingFleets, order)
				lcp.fleets = append(lcp.fleets, order)
			}
		}
	}

	// If we have convoying fleets, check if they form a valid path
	if len(convoyingFleets) > 0 {
		// For now, assume any convoy order makes the path valid
		// TODO: Implement proper convoy path validation using BFS
		lcp.path = []string{lcp.source, lcp.destination} // Simplified path
		return true
	}

	return false
}

// isConvoyingMove checks if a convoy order is convoying this specific move
func (lcp *LazyConvoyPath) isConvoyingMove(convoyOrder *Order) bool {
	if convoyOrder.Type != Convoy {
		return false
	}

	// For convoy orders, auxiliary contains the full move description like "A Norway - Sweden"
	// Check if the auxiliary field matches this move
	expectedFormats := []string{
		"A " + lcp.source + " - " + lcp.destination,
		"F " + lcp.source + " - " + lcp.destination,
		lcp.source + " -> " + lcp.destination,
		lcp.source + " - " + lcp.destination,
	}

	for _, format := range expectedFormats {
		if convoyOrder.Auxiliary == format {
			return true
		}
	}

	return false
}

// GetPath returns the convoy path (only valid after IsValid() returns true)
func (lcp *LazyConvoyPath) GetPath() []string {
	lcp.mu.RLock()
	defer lcp.mu.RUnlock()

	if !lcp.evaluated || !lcp.valid {
		return nil
	}

	result := make([]string, len(lcp.path))
	copy(result, lcp.path)
	return result
}

// GetFleets returns the convoying fleets (only valid after IsValid() returns true)
func (lcp *LazyConvoyPath) GetFleets() []*Order {
	lcp.mu.RLock()
	defer lcp.mu.RUnlock()

	if !lcp.evaluated || !lcp.valid {
		return nil
	}

	result := make([]*Order, len(lcp.fleets))
	copy(result, lcp.fleets)
	return result
}

// Reset clears the lazy evaluation state (used between resolution runs)
func (lcp *LazyConvoyPath) Reset() {
	lcp.mu.Lock()
	defer lcp.mu.Unlock()

	lcp.evaluated = false
	lcp.valid = false
	lcp.path = nil
	lcp.fleets = nil
}

// ClearCache clears the convoy path cache
func (cr *ConvoyResolver) ClearCache() {
	cr.mu.Lock()
	defer cr.mu.Unlock()

	for _, path := range cr.pathCache {
		path.Reset()
	}
	cr.pathCache = make(map[string]*LazyConvoyPath)
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
