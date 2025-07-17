package resolution

import (
	"diplomacy-cli/backend/internal/game"
	"sync"
)

// ObjectPool provides memory pooling for frequently allocated objects
// to reduce garbage collection pressure during resolution
type ObjectPool struct {
	orders          sync.Pool
	unitOutcomes    sync.Pool
	conflicts       sync.Pool
	supportOutcomes sync.Pool
	convoyOutcomes  sync.Pool
	resolutionSteps sync.Pool
	unitIDSlices    sync.Pool
	stringSlices    sync.Pool
	orderSlices     sync.Pool
}

// NewObjectPool creates a new object pool with pre-configured pools
func NewObjectPool() *ObjectPool {
	return &ObjectPool{
		orders: sync.Pool{
			New: func() interface{} {
				return &Order{}
			},
		},
		unitOutcomes: sync.Pool{
			New: func() interface{} {
				return &UnitOutcome{}
			},
		},
		conflicts: sync.Pool{
			New: func() interface{} {
				return &Conflict{
					competitors: make([]UnitID, 0, 4), // Pre-allocate for typical conflicts
					strengths:   make(map[UnitID]int, 4),
				}
			},
		},
		supportOutcomes: sync.Pool{
			New: func() interface{} {
				return &SupportOutcome{
					cutBy: make([]UnitID, 0, 2), // Pre-allocate for typical support cuts
				}
			},
		},
		convoyOutcomes: sync.Pool{
			New: func() interface{} {
				return &ConvoyOutcome{
					disruptedBy: make([]UnitID, 0, 2), // Pre-allocate for typical disruptions
					path:        make([]string, 0, 8), // Pre-allocate for typical convoy paths
				}
			},
		},
		resolutionSteps: sync.Pool{
			New: func() interface{} {
				return &ResolutionStep{
					dependencies: make([]UnitID, 0, 3), // Pre-allocate for typical dependencies
				}
			},
		},
		unitIDSlices: sync.Pool{
			New: func() interface{} {
				return make([]UnitID, 0, 8) // Pre-allocate for typical slice usage
			},
		},
		stringSlices: sync.Pool{
			New: func() interface{} {
				return make([]string, 0, 8) // Pre-allocate for typical slice usage
			},
		},
		orderSlices: sync.Pool{
			New: func() interface{} {
				return make([]*Order, 0, 8) // Pre-allocate for typical slice usage
			},
		},
	}
}

// GetOrder retrieves a clean Order from the pool
func (p *ObjectPool) GetOrder() *Order {
	order := p.orders.Get().(*Order)
	order.reset()
	return order
}

// PutOrder returns an Order to the pool after clearing sensitive data
func (p *ObjectPool) PutOrder(order *Order) {
	if order == nil {
		return
	}

	// Clear all fields to prevent memory leaks and ensure clean state
	*order = Order{}

	p.orders.Put(order)
}

// GetUnitOutcome retrieves a clean UnitOutcome from the pool
func (p *ObjectPool) GetUnitOutcome() *UnitOutcome {
	outcome := p.unitOutcomes.Get().(*UnitOutcome)
	*outcome = UnitOutcome{} // Zero out the struct
	return outcome
}

// PutUnitOutcome returns a UnitOutcome to the pool
func (p *ObjectPool) PutUnitOutcome(outcome *UnitOutcome) {
	if outcome == nil {
		return
	}

	// Clear pointer fields to prevent memory leaks
	outcome.toProvince = nil
	outcome.dislodgedBy = nil

	p.unitOutcomes.Put(outcome)
}

// GetConflict retrieves a clean Conflict from the pool
func (p *ObjectPool) GetConflict() *Conflict {
	conflict := p.conflicts.Get().(*Conflict)

	// Reset slices but keep capacity
	conflict.competitors = conflict.competitors[:0]

	// Clear map but keep capacity
	for k := range conflict.strengths {
		delete(conflict.strengths, k)
	}

	// Clear other fields
	conflict.province = ""
	conflict.winner = nil
	conflict.resolution = 0
	conflict.explanation = ""

	return conflict
}

// PutConflict returns a Conflict to the pool
func (p *ObjectPool) PutConflict(conflict *Conflict) {
	if conflict == nil {
		return
	}

	// Clear pointer fields
	conflict.winner = nil

	p.conflicts.Put(conflict)
}

// GetSupportOutcome retrieves a clean SupportOutcome from the pool
func (p *ObjectPool) GetSupportOutcome() *SupportOutcome {
	outcome := p.supportOutcomes.Get().(*SupportOutcome)

	// Reset slice but keep capacity
	outcome.cutBy = outcome.cutBy[:0]

	// Clear other fields
	*outcome = SupportOutcome{
		cutBy: outcome.cutBy, // Preserve the slice with capacity
	}

	return outcome
}

// PutSupportOutcome returns a SupportOutcome to the pool
func (p *ObjectPool) PutSupportOutcome(outcome *SupportOutcome) {
	if outcome == nil {
		return
	}
	p.supportOutcomes.Put(outcome)
}

// GetConvoyOutcome retrieves a clean ConvoyOutcome from the pool
func (p *ObjectPool) GetConvoyOutcome() *ConvoyOutcome {
	outcome := p.convoyOutcomes.Get().(*ConvoyOutcome)

	// Reset slices but keep capacity
	outcome.disruptedBy = outcome.disruptedBy[:0]
	outcome.path = outcome.path[:0]

	// Clear other fields
	*outcome = ConvoyOutcome{
		disruptedBy: outcome.disruptedBy, // Preserve slices with capacity
		path:        outcome.path,
	}

	return outcome
}

// PutConvoyOutcome returns a ConvoyOutcome to the pool
func (p *ObjectPool) PutConvoyOutcome(outcome *ConvoyOutcome) {
	if outcome == nil {
		return
	}
	p.convoyOutcomes.Put(outcome)
}

// GetResolutionStep retrieves a clean ResolutionStep from the pool
func (p *ObjectPool) GetResolutionStep() *ResolutionStep {
	step := p.resolutionSteps.Get().(*ResolutionStep)

	// Reset slice but keep capacity
	step.dependencies = step.dependencies[:0]

	// Clear other fields
	*step = ResolutionStep{
		dependencies: step.dependencies, // Preserve slice with capacity
	}

	return step
}

// PutResolutionStep returns a ResolutionStep to the pool
func (p *ObjectPool) PutResolutionStep(step *ResolutionStep) {
	if step == nil {
		return
	}
	p.resolutionSteps.Put(step)
}

// GetUnitIDSlice retrieves a clean UnitID slice from the pool
func (p *ObjectPool) GetUnitIDSlice() []UnitID {
	slice := p.unitIDSlices.Get().([]UnitID)
	return slice[:0] // Reset length but keep capacity
}

// PutUnitIDSlice returns a UnitID slice to the pool
func (p *ObjectPool) PutUnitIDSlice(slice []UnitID) {
	if slice == nil {
		return
	}

	// Only pool slices that aren't too large to avoid memory waste
	if cap(slice) <= 32 {
		p.unitIDSlices.Put(slice)
	}
}

// GetStringSlice retrieves a clean string slice from the pool
func (p *ObjectPool) GetStringSlice() []string {
	slice := p.stringSlices.Get().([]string)
	return slice[:0] // Reset length but keep capacity
}

// PutStringSlice returns a string slice to the pool
func (p *ObjectPool) PutStringSlice(slice []string) {
	if slice == nil {
		return
	}

	// Only pool slices that aren't too large to avoid memory waste
	if cap(slice) <= 32 {
		p.stringSlices.Put(slice)
	}
}

// GetOrderSlice retrieves a clean Order pointer slice from the pool
func (p *ObjectPool) GetOrderSlice() []*Order {
	slice := p.orderSlices.Get().([]*Order)
	return slice[:0] // Reset length but keep capacity
}

// PutOrderSlice returns an Order pointer slice to the pool
func (p *ObjectPool) PutOrderSlice(slice []*Order) {
	if slice == nil {
		return
	}

	// Clear all pointers to prevent memory leaks
	for i := range slice {
		slice[i] = nil
	}

	// Only pool slices that aren't too large to avoid memory waste
	if cap(slice) <= 32 {
		p.orderSlices.Put(slice)
	}
}

// PooledResolutionResult wraps ResolutionResult with pool management
type PooledResolutionResult struct {
	*ResolutionResult
	pool *ObjectPool
}

// NewPooledResolutionResult creates a new ResolutionResult using pooled objects
func (p *ObjectPool) NewPooledResolutionResult(turn int, phase game.Phase) *PooledResolutionResult {
	result := NewResolutionResult(turn, phase)
	return &PooledResolutionResult{
		ResolutionResult: result,
		pool:             p,
	}
}

// AddPooledUnitOutcome adds a unit outcome using a pooled object
func (pr *PooledResolutionResult) AddPooledUnitOutcome(unitID UnitID, unit game.Unit, order game.Order, status UnitStatus, fromProvince string, toProvince *string, dislodgedBy *UnitID, reason string, strength, supportCount int) {
	outcome := pr.pool.GetUnitOutcome()

	outcome.unit = unit
	outcome.orderGiven = order
	outcome.finalStatus = status
	outcome.fromProvince = fromProvince
	outcome.toProvince = toProvince
	outcome.dislodgedBy = dislodgedBy
	outcome.reason = reason
	outcome.strength = strength
	outcome.supportCount = supportCount

	pr.AddUnitOutcome(unitID, *outcome)

	// Return to pool after copying
	pr.pool.PutUnitOutcome(outcome)
}

// AddPooledConflict adds a conflict using a pooled object
func (pr *PooledResolutionResult) AddPooledConflict(province string, competitors []UnitID, winner *UnitID, resolution ConflictType, explanation string, strengths map[UnitID]int) {
	conflict := pr.pool.GetConflict()

	conflict.province = province
	conflict.competitors = append(conflict.competitors, competitors...)
	conflict.winner = winner
	conflict.resolution = resolution
	conflict.explanation = explanation

	// Copy strengths map
	for k, v := range strengths {
		conflict.strengths[k] = v
	}

	pr.AddConflict(*conflict)

	// Return to pool after copying
	pr.pool.PutConflict(conflict)
}

// Release returns all pooled objects back to their respective pools
// This should be called when the ResolutionResult is no longer needed
func (pr *PooledResolutionResult) Release() {
	// Note: The actual objects in ResolutionResult are copies, so we don't need
	// to return them to pools. The pooling happens during construction.
	// This method is here for future expansion if we need cleanup logic.
}

// Global pool instance for the resolution package
var globalPool = NewObjectPool()

// GetGlobalPool returns the global object pool instance
func GetGlobalPool() *ObjectPool {
	return globalPool
}
