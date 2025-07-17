package resolution

import (
	"sync"
)

// StrengthCache provides caching for expensive strength calculations
type StrengthCache struct {
	attackStrengthCache  map[string]int
	holdStrengthCache    map[string]int
	preventStrengthCache map[string]int
	supportCache         map[string]bool
	competingMovesCache  map[string][]*Order
	mu                   sync.RWMutex
}

// NewStrengthCache creates a new strength calculation cache
func NewStrengthCache() *StrengthCache {
	return &StrengthCache{
		attackStrengthCache:  make(map[string]int),
		holdStrengthCache:    make(map[string]int),
		preventStrengthCache: make(map[string]int),
		supportCache:         make(map[string]bool),
		competingMovesCache:  make(map[string][]*Order),
	}
}

// GetAttackStrength retrieves cached attack strength or calculates if not cached
func (sc *StrengthCache) GetAttackStrength(key string, calculator func() int) int {
	sc.mu.RLock()
	if strength, exists := sc.attackStrengthCache[key]; exists {
		sc.mu.RUnlock()
		return strength
	}
	sc.mu.RUnlock()

	strength := calculator()

	sc.mu.Lock()
	sc.attackStrengthCache[key] = strength
	sc.mu.Unlock()

	return strength
}

// GetHoldStrength retrieves cached hold strength or calculates if not cached
func (sc *StrengthCache) GetHoldStrength(key string, calculator func() int) int {
	sc.mu.RLock()
	if strength, exists := sc.holdStrengthCache[key]; exists {
		sc.mu.RUnlock()
		return strength
	}
	sc.mu.RUnlock()

	strength := calculator()

	sc.mu.Lock()
	sc.holdStrengthCache[key] = strength
	sc.mu.Unlock()

	return strength
}

// GetPreventStrength retrieves cached prevent strength or calculates if not cached
func (sc *StrengthCache) GetPreventStrength(key string, calculator func() int) int {
	sc.mu.RLock()
	if strength, exists := sc.preventStrengthCache[key]; exists {
		sc.mu.RUnlock()
		return strength
	}
	sc.mu.RUnlock()

	strength := calculator()

	sc.mu.Lock()
	sc.preventStrengthCache[key] = strength
	sc.mu.Unlock()

	return strength
}

// GetSupportResult retrieves cached support result or calculates if not cached
func (sc *StrengthCache) GetSupportResult(key string, calculator func() bool) bool {
	sc.mu.RLock()
	if result, exists := sc.supportCache[key]; exists {
		sc.mu.RUnlock()
		return result
	}
	sc.mu.RUnlock()

	result := calculator()

	sc.mu.Lock()
	sc.supportCache[key] = result
	sc.mu.Unlock()

	return result
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

	sc.attackStrengthCache = make(map[string]int)
	sc.holdStrengthCache = make(map[string]int)
	sc.preventStrengthCache = make(map[string]int)
	sc.supportCache = make(map[string]bool)
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
