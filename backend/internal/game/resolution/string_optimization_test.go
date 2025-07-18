package resolution

import (
	"fmt"
	"runtime"
	"testing"
)

// TestStringInterning tests that string interning works correctly
func TestStringInterning(t *testing.T) {
	interner := NewStringInterner()

	// Test basic interning
	s1 := interner.Intern("test")
	s2 := interner.Intern("test")

	// Should be the same string value
	if s1 != s2 {
		t.Error("Interned strings should be equal")
	}

	// Test different strings
	s3 := interner.Intern("different")
	if s1 == s3 {
		t.Error("Different strings should not be equal")
	}

	// Test size tracking
	if interner.Size() != 2 {
		t.Errorf("Expected 2 interned strings, got %d", interner.Size())
	}
}

// TestStringBuilderPool tests the string builder pool functionality
func TestStringBuilderPool(t *testing.T) {
	pool := NewStringBuilderPool()

	// Get a builder
	sb1 := pool.Get()
	sb1.WriteString("test")

	// Return it
	pool.Put(sb1)

	// Get another - should be the same instance (reset)
	sb2 := pool.Get()
	if sb2.Len() != 0 {
		t.Error("String builder should be reset when returned to pool")
	}

	pool.Put(sb2)
}

// TestOrderStringInterning tests that orders properly intern their strings
func TestOrderStringInterning(t *testing.T) {
	// Create two orders with same territories
	order1 := NewOrder("A Berlin", "berlin", "munich", "", "germany", Move)
	order2 := NewOrder("F Berlin", "berlin", "kiel", "", "germany", Move)

	// Source territories should be equal (interned)
	if order1.Source != order2.Source {
		t.Error("Same territory names should be interned to same value")
	}

	// Owners should be equal (interned)
	if order1.Owner != order2.Owner {
		t.Error("Same owner names should be interned to same value")
	}
}

// BenchmarkStringInterning benchmarks string interning performance
func BenchmarkStringInterning(b *testing.B) {
	territories := []string{"berlin", "munich", "vienna", "paris", "london", "moscow", "warsaw", "rome"}

	b.Run("WithInterning", func(b *testing.B) {
		interner := NewStringInterner()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			territory := territories[i%len(territories)]
			_ = interner.Intern(territory)
		}
	})

	b.Run("WithoutInterning", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			territory := territories[i%len(territories)]
			// Just copy the string (no interning)
			_ = string([]byte(territory))
		}
	})
}

// BenchmarkStringBuilderVsConcat benchmarks string builder vs concatenation
func BenchmarkStringBuilderVsConcat(b *testing.B) {
	source := "berlin"
	destination := "munich"

	b.Run("StringBuilder", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			sb := GetStringBuilder()
			sb.WriteString(source)
			sb.WriteString(" -> ")
			sb.WriteString(destination)
			result := sb.String()
			PutStringBuilder(sb)
			_ = result
		}
	})

	b.Run("Concatenation", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			result := source + " -> " + destination
			_ = result
		}
	})

	b.Run("Sprintf", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			result := fmt.Sprintf("%s -> %s", source, destination)
			_ = result
		}
	})
}

// BenchmarkCacheKeyBuilding benchmarks optimized cache key building
func BenchmarkCacheKeyBuilding(b *testing.B) {
	source := "berlin"
	destination := "munich"

	b.Run("OptimizedCacheKey", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			key := BuildArrowKey(source, destination)
			_ = key
		}
	})

	b.Run("SimpleConcatenation", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			key := source + " -> " + destination
			_ = key
		}
	})

	b.Run("SprintfCacheKey", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			key := fmt.Sprintf("%s -> %s", source, destination)
			_ = key
		}
	})
}

// BenchmarkResolutionEngineStringOptimizations benchmarks the full resolution engine with string optimizations
func BenchmarkResolutionEngineStringOptimizations(b *testing.B) {
	// Create a complex scenario with many string operations
	orders := []Order{
		{Unit: "A Berlin", Type: Move, Source: "berlin", Destination: "munich", Owner: "germany"},
		{Unit: "A Vienna", Type: Support, Source: "vienna", Auxiliary: "berlin -> munich", Owner: "austria"},
		{Unit: "A Paris", Type: Move, Source: "paris", Destination: "burgundy", Owner: "france"},
		{Unit: "F London", Type: Move, Source: "london", Destination: "belgium", Owner: "england"},
		{Unit: "F North Sea", Type: Convoy, Source: "north_sea", Auxiliary: "A london - belgium", Owner: "england"},
		{Unit: "A Moscow", Type: Hold, Source: "moscow", Destination: "moscow", Owner: "russia"},
		{Unit: "A Warsaw", Type: Move, Source: "warsaw", Destination: "berlin", Owner: "russia"},
		{Unit: "F Rome", Type: Support, Source: "rome", Auxiliary: "vienna -> munich", Owner: "italy"},
	}

	b.Run("WithStringOptimizations", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			engine := NewDATCEngine(orders)
			results := engine.ResolveAll()

			// Access reasons to trigger lazy evaluation and string building
			for _, result := range results {
				_ = result.Reason
			}

			engine.Cleanup()
		}
	})
}

// BenchmarkMemoryUsage benchmarks memory usage with string optimizations
func BenchmarkMemoryUsage(b *testing.B) {
	orders := make([]Order, 100)
	territories := []string{"berlin", "munich", "vienna", "paris", "london", "moscow", "warsaw", "rome"}
	owners := []string{"germany", "austria", "france", "england", "russia", "italy", "turkey"}

	// Create many orders with repeated territory and owner names
	for i := range orders {
		orders[i] = Order{
			Unit:        fmt.Sprintf("A %s", territories[i%len(territories)]),
			Type:        Move,
			Source:      territories[i%len(territories)],
			Destination: territories[(i+1)%len(territories)],
			Owner:       owners[i%len(owners)],
		}
	}

	b.Run("WithInterning", func(b *testing.B) {
		var m1, m2 runtime.MemStats
		runtime.GC()
		runtime.ReadMemStats(&m1)

		for i := 0; i < b.N; i++ {
			engine := NewDATCEngine(orders) // This will intern strings
			_ = engine.ResolveAll()
			engine.Cleanup()
		}

		runtime.GC()
		runtime.ReadMemStats(&m2)

		b.ReportMetric(float64(m2.TotalAlloc-m1.TotalAlloc)/float64(b.N), "bytes/op")
		b.ReportMetric(float64(m2.Mallocs-m1.Mallocs)/float64(b.N), "allocs/op")
	})
}

// BenchmarkStringInternerConcurrency tests string interner under concurrent load
func BenchmarkStringInternerConcurrency(b *testing.B) {
	interner := NewStringInterner()
	territories := []string{"berlin", "munich", "vienna", "paris", "london", "moscow", "warsaw", "rome"}

	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			territory := territories[i%len(territories)]
			_ = interner.Intern(territory)
			i++
		}
	})
}

// BenchmarkStringBuilderPoolConcurrency tests string builder pool under concurrent load
func BenchmarkStringBuilderPoolConcurrency(b *testing.B) {
	pool := NewStringBuilderPool()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			sb := pool.Get()
			sb.WriteString("test")
			sb.WriteString(" -> ")
			sb.WriteString("destination")
			_ = sb.String()
			pool.Put(sb)
		}
	})
}

// TestStringOptimizationCorrectness verifies that string optimizations don't affect correctness
func TestStringOptimizationCorrectness(t *testing.T) {
	orders := []Order{
		{Unit: "A Berlin", Type: Move, Source: "berlin", Destination: "munich", Owner: "germany"},
		{Unit: "A Vienna", Type: Support, Source: "vienna", Auxiliary: "berlin -> munich", Owner: "austria"},
		{Unit: "A Paris", Type: Move, Source: "paris", Destination: "munich", Owner: "france"},
	}

	// Run resolution multiple times and ensure consistent results
	var firstResults []AdjudicationResult

	for run := 0; run < 5; run++ {
		engine := NewDATCEngine(orders)
		results := engine.ResolveAll()

		if run == 0 {
			firstResults = results
		} else {
			// Compare with first run
			if len(results) != len(firstResults) {
				t.Fatalf("Run %d: Different number of results: %d vs %d", run, len(results), len(firstResults))
			}

			for i, result := range results {
				if result.Success != firstResults[i].Success {
					t.Errorf("Run %d: Order %d success differs: %v vs %v", run, i, result.Success, firstResults[i].Success)
				}
			}
		}

		engine.Cleanup()
	}
}

// BenchmarkLazyReasonGeneration benchmarks lazy vs eager reason generation
func BenchmarkLazyReasonGeneration(b *testing.B) {
	orders := []Order{
		{Unit: "A Berlin", Type: Move, Source: "berlin", Destination: "munich", Owner: "germany"},
		{Unit: "A Vienna", Type: Support, Source: "vienna", Auxiliary: "berlin -> munich", Owner: "austria"},
		{Unit: "A Paris", Type: Move, Source: "paris", Destination: "munich", Owner: "france"},
	}

	b.Run("LazyReasonGeneration", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			engine := NewDATCEngine(orders)
			results := engine.ResolveAll()

			// Only access first reason (lazy)
			if len(results) > 0 {
				_ = results[0].Reason
			}

			engine.Cleanup()
		}
	})

	b.Run("EagerReasonGeneration", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			engine := NewDATCEngine(orders)
			results := engine.ResolveAll()

			// Access all reasons (eager)
			for _, result := range results {
				_ = result.Reason
			}

			engine.Cleanup()
		}
	})
}

// TestGlobalInternerStats tests global interner statistics
func TestGlobalInternerStats(t *testing.T) {
	// Clear global interner
	globalInterner.Clear()

	// Create some orders
	orders := []Order{
		{Unit: "A Berlin", Type: Move, Source: "berlin", Destination: "munich", Owner: "germany"},
		{Unit: "A Vienna", Type: Support, Source: "vienna", Auxiliary: "berlin -> munich", Owner: "austria"},
	}

	engine := NewDATCEngine(orders)
	_ = engine.ResolveAll()

	// Check that strings were interned
	size := globalInterner.Size()
	if size < 4 { // At least berlin, munich, vienna, germany, austria should be interned
		t.Errorf("Expected at least 4 interned strings, got %d", size)
	}

	t.Logf("Global interner contains %d unique strings", size)
	engine.Cleanup()
}
