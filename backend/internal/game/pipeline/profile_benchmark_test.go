package pipeline

import (
	"runtime"
	"testing"

	"diplomacy-cli/backend/internal/game"
	"diplomacy-cli/backend/internal/game/loader"
	"diplomacy-cli/backend/internal/game/resolution"
)

// BenchmarkStrengthCalculationHotPaths focuses on profiling strength calculation performance
func BenchmarkStrengthCalculationHotPaths(b *testing.B) {
	// Load test data
	mapLoader := loader.NewJSONLoader("../../../../backend/data/classic")
	board, err := mapLoader.LoadBoard()
	if err != nil {
		b.Fatalf("Failed to load board: %v", err)
	}

	gameState := game.NewGameState(board, game.SpringMovement, 1901)

	// Create a complex scenario with multiple competing moves and supports
	gameState.Board.Units["london"] = &game.Unit{
		Type:     game.Fleet,
		Owner:    game.England,
		Province: "london",
	}
	gameState.Board.Units["wales"] = &game.Unit{
		Type:     game.Army,
		Owner:    game.England,
		Province: "wales",
	}
	gameState.Board.Units["yorkshire"] = &game.Unit{
		Type:     game.Army,
		Owner:    game.England,
		Province: "yorkshire",
	}
	gameState.Board.Units["brest"] = &game.Unit{
		Type:     game.Fleet,
		Owner:    game.France,
		Province: "brest",
	}
	gameState.Board.Units["paris"] = &game.Unit{
		Type:     game.Army,
		Owner:    game.France,
		Province: "paris",
	}
	gameState.Board.Units["burgundy"] = &game.Unit{
		Type:     game.Army,
		Owner:    game.France,
		Province: "burgundy",
	}

	// Complex orders with multiple supports and competing moves
	gameState.RawOrders[game.England] = []string{
		"f london - english_channel",
		"wales s london - english_channel",
		"yorkshire s london - english_channel",
	}
	gameState.RawOrders[game.France] = []string{
		"f brest - english_channel",
		"paris s brest - english_channel",
		"burgundy s brest - english_channel",
	}

	processor := NewTurnProcessor()

	// Force garbage collection before benchmark
	runtime.GC()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		processor.ProcessTurn(gameState)
	}
}

// BenchmarkConflictResolutionHotPaths focuses on profiling conflict resolution performance
func BenchmarkConflictResolutionHotPaths(b *testing.B) {
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

// BenchmarkCircularDependencyResolution focuses on profiling circular dependency handling
func BenchmarkCircularDependencyResolution(b *testing.B) {
	// Create orders that will trigger circular dependencies
	orders := []resolution.Order{
		{Type: resolution.Move, Source: "london", Destination: "wales", Owner: string(game.England)},
		{Type: resolution.Move, Source: "wales", Destination: "yorkshire", Owner: string(game.England)},
		{Type: resolution.Move, Source: "yorkshire", Destination: "london", Owner: string(game.England)},
		{Type: resolution.Support, Source: "liverpool", Auxiliary: "london -> wales", Owner: string(game.England)},
		{Type: resolution.Support, Source: "edinburgh", Auxiliary: "wales -> yorkshire", Owner: string(game.England)},
		{Type: resolution.Support, Source: "clyde", Auxiliary: "yorkshire -> london", Owner: string(game.England)},
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

// BenchmarkConvoyPathfindingHotPaths focuses on profiling convoy pathfinding performance
func BenchmarkConvoyPathfindingHotPaths(b *testing.B) {
	// Create orders with complex convoy chains
	orders := []resolution.Order{
		{Type: resolution.Move, Source: "london", Destination: "belgium", Owner: string(game.England)},
		{Type: resolution.Convoy, Source: "english_channel", Auxiliary: "A london - belgium", Owner: string(game.England)},
		{Type: resolution.Convoy, Source: "north_sea", Auxiliary: "A london - belgium", Owner: string(game.England)},
		{Type: resolution.Move, Source: "brest", Destination: "london", Owner: string(game.France)},
		{Type: resolution.Convoy, Source: "mid_atlantic", Auxiliary: "A brest - london", Owner: string(game.France)},
		{Type: resolution.Convoy, Source: "irish_sea", Auxiliary: "A brest - london", Owner: string(game.France)},
		{Type: resolution.Support, Source: "wales", Auxiliary: "london -> belgium", Owner: string(game.England)},
		{Type: resolution.Support, Source: "paris", Auxiliary: "brest -> london", Owner: string(game.France)},
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

// BenchmarkMemoryAllocationPatterns focuses on profiling memory allocation patterns
func BenchmarkMemoryAllocationPatterns(b *testing.B) {
	// Create a scenario that will stress memory allocation
	orders := make([]resolution.Order, 0, 100)

	// Generate many orders to stress memory allocation
	for i := 0; i < 50; i++ {
		nations := []game.Nation{game.England, game.France, game.Germany, game.Austria, game.Italy, game.Russia, game.Turkey}
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

// BenchmarkObjectPoolEffectiveness tests the effectiveness of object pooling
func BenchmarkObjectPoolEffectiveness(b *testing.B) {
	orders := []resolution.Order{
		{Type: resolution.Move, Source: "london", Destination: "english_channel", Owner: string(game.England)},
		{Type: resolution.Move, Source: "brest", Destination: "english_channel", Owner: string(game.France)},
		{Type: resolution.Support, Source: "wales", Auxiliary: "london -> english_channel", Owner: string(game.England)},
		{Type: resolution.Support, Source: "paris", Auxiliary: "brest -> english_channel", Owner: string(game.France)},
	}

	// Get a shared pool for reuse
	pool := resolution.GetGlobalPool()

	// Force garbage collection before benchmark
	runtime.GC()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		engine := resolution.NewDATCEngineWithPool(orders, pool)
		engine.ResolveAll()
		engine.Cleanup()
	}
}
