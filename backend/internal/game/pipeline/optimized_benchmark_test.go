package pipeline

import (
	"runtime"
	"testing"

	"diplomacy-cli/backend/internal/game"
	"diplomacy-cli/backend/internal/game/resolution"
)

// BenchmarkOptimizedConflictResolution tests the optimized version with caching
func BenchmarkOptimizedConflictResolution(b *testing.B) {
	// Create orders that will trigger complex conflict resolution
	orders := []resolution.Order{
		{Type: resolution.Move, Source: "london", Destination: "english_channel", Owner: string(game.England)},
		{Type: resolution.Move, Source: "brest", Destination: "english_channel", Owner: string(game.France)},
		{Type: resolution.Move, Source: "belgium", Destination: "english_channel", Owner: string(game.Germany)},
		{Type: resolution.Support, Source: "wales", Auxiliary: "london -> english_channel", Owner: string(game.England)},
		{Type: resolution.Support, Source: "paris", Auxiliary: "brest -> english_channel", Owner: string(game.France)},
		{Type: resolution.Support, Source: "holland", Auxiliary: "belgium -> english_channel", Owner: string(game.Germany)},
		{Type: resolution.Support, Source: "yorkshire", Auxiliary: "london -> english_channel", Owner: string(game.England)},
		{Type: resolution.Support, Source: "burgundy", Auxiliary: "brest -> english_channel", Owner: string(game.France)},
	}

	// Force garbage collection before benchmark
	runtime.GC()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		engine := resolution.NewDATCEngine(orders)
		engine.ResolveAll()
		engine.Cleanup()
	}
}

// BenchmarkOptimizedStrengthCalculation tests optimized strength calculations
func BenchmarkOptimizedStrengthCalculation(b *testing.B) {
	// Create a scenario with many competing moves and supports
	orders := []resolution.Order{
		{Type: resolution.Move, Source: "london", Destination: "english_channel", Owner: string(game.England)},
		{Type: resolution.Move, Source: "brest", Destination: "english_channel", Owner: string(game.France)},
		{Type: resolution.Move, Source: "belgium", Destination: "english_channel", Owner: string(game.Germany)},
		{Type: resolution.Move, Source: "holland", Destination: "english_channel", Owner: string(game.Germany)},
		{Type: resolution.Support, Source: "wales", Auxiliary: "london -> english_channel", Owner: string(game.England)},
		{Type: resolution.Support, Source: "yorkshire", Auxiliary: "london -> english_channel", Owner: string(game.England)},
		{Type: resolution.Support, Source: "liverpool", Auxiliary: "london -> english_channel", Owner: string(game.England)},
		{Type: resolution.Support, Source: "paris", Auxiliary: "brest -> english_channel", Owner: string(game.France)},
		{Type: resolution.Support, Source: "burgundy", Auxiliary: "brest -> english_channel", Owner: string(game.France)},
		{Type: resolution.Support, Source: "marseilles", Auxiliary: "brest -> english_channel", Owner: string(game.France)},
		{Type: resolution.Support, Source: "ruhr", Auxiliary: "belgium -> english_channel", Owner: string(game.Germany)},
		{Type: resolution.Support, Source: "kiel", Auxiliary: "holland -> english_channel", Owner: string(game.Germany)},
	}

	// Force garbage collection before benchmark
	runtime.GC()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		engine := resolution.NewDATCEngine(orders)
		engine.ResolveAll()
		engine.Cleanup()
	}
}

// BenchmarkOptimizedMemoryAllocation tests memory allocation patterns with optimizations
func BenchmarkOptimizedMemoryAllocation(b *testing.B) {
	// Create a scenario that will stress memory allocation
	orders := make([]resolution.Order, 0, 100)

	nations := []game.Nation{game.England, game.France, game.Germany, game.Austria, game.Italy, game.Russia, game.Turkey}

	// Generate many orders to stress memory allocation
	for i := 0; i < 50; i++ {
		owner := string(nations[i%len(nations)])

		orders = append(orders, resolution.Order{
			Type:        resolution.Move,
			Source:      "province_" + string(rune('a'+i%26)),
			Destination: "province_" + string(rune('z'-i%26)),
			Owner:       owner,
		})
		orders = append(orders, resolution.Order{
			Type:      resolution.Support,
			Source:    "support_" + string(rune('a'+i%26)),
			Auxiliary: "province_" + string(rune('a'+i%26)) + " -> " + "province_" + string(rune('z'-i%26)),
			Owner:     owner,
		})
	}

	// Force garbage collection before benchmark
	runtime.GC()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		engine := resolution.NewDATCEngine(orders)
		results := engine.ResolveAll()
		_ = results // Prevent optimization
		engine.Cleanup()
	}
}

// BenchmarkCacheEffectiveness compares performance with and without caching
func BenchmarkCacheEffectiveness(b *testing.B) {
	// Create orders that will benefit from caching (repeated calculations)
	orders := []resolution.Order{
		{Type: resolution.Move, Source: "london", Destination: "english_channel", Owner: string(game.England)},
		{Type: resolution.Move, Source: "brest", Destination: "english_channel", Owner: string(game.France)},
		{Type: resolution.Move, Source: "belgium", Destination: "english_channel", Owner: string(game.Germany)},
		{Type: resolution.Support, Source: "wales", Auxiliary: "london -> english_channel", Owner: string(game.England)},
		{Type: resolution.Support, Source: "paris", Auxiliary: "brest -> english_channel", Owner: string(game.France)},
		{Type: resolution.Support, Source: "holland", Auxiliary: "belgium -> english_channel", Owner: string(game.Germany)},
		// Add more orders that will cause repeated strength calculations
		{Type: resolution.Move, Source: "yorkshire", Destination: "north_sea", Owner: string(game.England)},
		{Type: resolution.Move, Source: "burgundy", Destination: "north_sea", Owner: string(game.France)},
		{Type: resolution.Support, Source: "liverpool", Auxiliary: "yorkshire -> north_sea", Owner: string(game.England)},
		{Type: resolution.Support, Source: "marseilles", Auxiliary: "burgundy -> north_sea", Owner: string(game.France)},
	}

	// Force garbage collection before benchmark
	runtime.GC()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		engine := resolution.NewDATCEngine(orders)
		engine.ResolveAll()
		engine.Cleanup()
	}
}
