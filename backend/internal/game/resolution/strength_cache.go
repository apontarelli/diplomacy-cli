package resolution

import (
	"sync"
)

// LazyStrengthCalculator represents a deferred strength calculation
type LazyStrengthCalculator struct {
	key        string
	calculator func() int
	evaluated  bool
	result     int
	mu         sync.RWMutex
}

// Calculate performs the lazy evaluation and returns the result
func (lsc *LazyStrengthCalculator) Calculate() int {
	lsc.mu.RLock()
	if lsc.evaluated {
		result := lsc.result
		lsc.mu.RUnlock()
		return result
	}
	lsc.mu.RUnlock()

	lsc.mu.Lock()
	defer lsc.mu.Unlock()

	// Double-check after acquiring write lock
	if lsc.evaluated {
		return lsc.result
	}

	lsc.result = lsc.calculator()
	lsc.evaluated = true
	return lsc.result
}

// LazySupportCalculator represents a deferred support calculation
type LazySupportCalculator struct {
	key        string
	calculator func() bool
	evaluated  bool
	result     bool
	mu         sync.RWMutex
}

// Calculate performs the lazy evaluation and returns the result
func (lsc *LazySupportCalculator) Calculate() bool {
	lsc.mu.RLock()
	if lsc.evaluated {
		result := lsc.result
		lsc.mu.RUnlock()
		return result
	}
	lsc.mu.RUnlock()

	lsc.mu.Lock()
	defer lsc.mu.Unlock()

	// Double-check after acquiring write lock
	if lsc.evaluated {
		return lsc.result
	}

	lsc.result = lsc.calculator()
	lsc.evaluated = true
	return lsc.result
}

// StrengthCache provides caching for expensive strength calculations with lazy evaluation
type StrengthCache struct {
	attackStrengthCache  map[string]*LazyStrengthCalculator
	holdStrengthCache    map[string]*LazyStrengthCalculator
	preventStrengthCache map[string]*LazyStrengthCalculator
	supportCache         map[string]*LazySupportCalculator
	competingMovesCache  map[string][]*Order
	mu                   sync.RWMutex
}

// NewStrengthCache creates a new strength calculation cache
func NewStrengthCache() *StrengthCache {
	return &StrengthCache{
		attackStrengthCache:  make(map[string]*LazyStrengthCalculator),
		holdStrengthCache:    make(map[string]*LazyStrengthCalculator),
		preventStrengthCache: make(map[string]*LazyStrengthCalculator),
		supportCache:         make(map[string]*LazySupportCalculator),
		competingMovesCache:  make(map[string][]*Order),
	}
}

// GetAttackStrength retrieves cached attack strength or creates lazy calculator if not cached
func (sc *StrengthCache) GetAttackStrength(key string, calculator func() int) int {
	sc.mu.RLock()
	if lazyCalc, exists := sc.attackStrengthCache[key]; exists {
		sc.mu.RUnlock()
		GetGlobalPerformanceMonitor().RecordCacheHit()
		return lazyCalc.Calculate()
	}
	sc.mu.RUnlock()

	GetGlobalPerformanceMonitor().RecordCacheMiss()

	sc.mu.Lock()
	// Double-check after acquiring write lock
	if lazyCalc, exists := sc.attackStrengthCache[key]; exists {
		sc.mu.Unlock()
		return lazyCalc.Calculate()
	}

	lazyCalc := &LazyStrengthCalculator{
		key:        key,
		calculator: calculator,
	}
	sc.attackStrengthCache[key] = lazyCalc
	sc.mu.Unlock()

	return lazyCalc.Calculate()
}

// GetHoldStrength retrieves cached hold strength or creates lazy calculator if not cached
func (sc *StrengthCache) GetHoldStrength(key string, calculator func() int) int {
	sc.mu.RLock()
	if lazyCalc, exists := sc.holdStrengthCache[key]; exists {
		sc.mu.RUnlock()
		return lazyCalc.Calculate()
	}
	sc.mu.RUnlock()

	sc.mu.Lock()
	// Double-check after acquiring write lock
	if lazyCalc, exists := sc.holdStrengthCache[key]; exists {
		sc.mu.Unlock()
		return lazyCalc.Calculate()
	}

	lazyCalc := &LazyStrengthCalculator{
		key:        key,
		calculator: calculator,
	}
	sc.holdStrengthCache[key] = lazyCalc
	sc.mu.Unlock()

	return lazyCalc.Calculate()
}

// GetPreventStrength retrieves cached prevent strength or creates lazy calculator if not cached
func (sc *StrengthCache) GetPreventStrength(key string, calculator func() int) int {
	sc.mu.RLock()
	if lazyCalc, exists := sc.preventStrengthCache[key]; exists {
		sc.mu.RUnlock()
		return lazyCalc.Calculate()
	}
	sc.mu.RUnlock()

	sc.mu.Lock()
	// Double-check after acquiring write lock
	if lazyCalc, exists := sc.preventStrengthCache[key]; exists {
		sc.mu.Unlock()
		return lazyCalc.Calculate()
	}

	lazyCalc := &LazyStrengthCalculator{
		key:        key,
		calculator: calculator,
	}
	sc.preventStrengthCache[key] = lazyCalc
	sc.mu.Unlock()

	return lazyCalc.Calculate()
}

// GetSupportResult retrieves cached support result or creates lazy calculator if not cached
func (sc *StrengthCache) GetSupportResult(key string, calculator func() bool) bool {
	sc.mu.RLock()
	if lazyCalc, exists := sc.supportCache[key]; exists {
		sc.mu.RUnlock()
		return lazyCalc.Calculate()
	}
	sc.mu.RUnlock()

	sc.mu.Lock()
	// Double-check after acquiring write lock
	if lazyCalc, exists := sc.supportCache[key]; exists {
		sc.mu.Unlock()
		return lazyCalc.Calculate()
	}

	lazyCalc := &LazySupportCalculator{
		key:        key,
		calculator: calculator,
	}
	sc.supportCache[key] = lazyCalc
	sc.mu.Unlock()

	return lazyCalc.Calculate()
}

// GetCompetingMoves retrieves cached competing moves or calculates if not cached
func (sc *StrengthCache) GetCompetingMoves(destination string, calculator func() []*Order) []*Order {
	sc.mu.RLock()
	if moves, exists := sc.competingMovesCache[destination]; exists {
		sc.mu.RUnlock()
		return moves
	}
	sc.mu.RUnlock()

	moves := calculator()

	sc.mu.Lock()
	sc.competingMovesCache[destination] = moves
	sc.mu.Unlock()

	return moves
}

// Clear resets all caches
func (sc *StrengthCache) Clear() {
	sc.mu.Lock()
	defer sc.mu.Unlock()

	sc.attackStrengthCache = make(map[string]*LazyStrengthCalculator)
	sc.holdStrengthCache = make(map[string]*LazyStrengthCalculator)
	sc.preventStrengthCache = make(map[string]*LazyStrengthCalculator)
	sc.supportCache = make(map[string]*LazySupportCalculator)
	sc.competingMovesCache = make(map[string][]*Order)
}

// generateStrengthKey creates a cache key for strength calculations
func generateStrengthKey(orderSource, orderDest string, optimistic bool) string {
	if optimistic {
		return orderSource + "->" + orderDest + ":opt"
	}
	return orderSource + "->" + orderDest + ":pes"
}

// generateSupportKey creates a cache key for support calculations
func generateSupportKey(supportSource, moveSource, moveDest string) string {
	return supportSource + ":supports:" + moveSource + "->" + moveDest
}
