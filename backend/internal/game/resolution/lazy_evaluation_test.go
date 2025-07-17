package resolution

import (
	"testing"
	"time"
)

// TestLazyConvoyPathEvaluation tests that convoy path evaluation is lazy
func TestLazyConvoyPathEvaluation(t *testing.T) {
	// Create a simple scenario with convoy
	orders := []Order{
		{Unit: "A London", Type: Move, Source: "london", Destination: "belgium", Owner: "england"},
		{Unit: "F North Sea", Type: Convoy, Source: "north_sea", Auxiliary: "A london - belgium", Owner: "england"},
	}

	engine := NewDATCEngine(orders)
	defer engine.Cleanup()

	// Get lazy convoy path - should not evaluate yet
	lazyPath := engine.convoyResolver.GetLazyConvoyPath("london", "belgium", engine.orderedOrders, engine)

	// Verify it hasn't been evaluated yet
	lazyPath.mu.RLock()
	evaluated := lazyPath.evaluated
	lazyPath.mu.RUnlock()

	if evaluated {
		t.Error("Convoy path should not be evaluated immediately")
	}

	// Now trigger evaluation
	isValid := lazyPath.IsValid(true)

	// Verify it's now evaluated
	lazyPath.mu.RLock()
	evaluatedAfter := lazyPath.evaluated
	lazyPath.mu.RUnlock()

	if !evaluatedAfter {
		t.Error("Convoy path should be evaluated after IsValid() call")
	}

	if !isValid {
		t.Error("Convoy path should be valid")
	}

	// Second call should use cached result
	startTime := time.Now()
	isValid2 := lazyPath.IsValid(true)
	duration := time.Since(startTime)

	if isValid != isValid2 {
		t.Error("Cached result should be the same")
	}

	// Should be very fast (cached)
	if duration > time.Microsecond*100 {
		t.Errorf("Cached convoy path evaluation took too long: %v", duration)
	}
}

// TestLazyStrengthCalculation tests that strength calculations are lazy
func TestLazyStrengthCalculation(t *testing.T) {
	orders := []Order{
		{Unit: "A Berlin", Type: Move, Source: "berlin", Destination: "munich", Owner: "germany"},
		{Unit: "A Vienna", Type: Support, Source: "vienna", Auxiliary: "berlin -> munich", Owner: "austria"},
	}

	engine := NewDATCEngine(orders)
	defer engine.Cleanup()

	// Create a lazy strength calculator
	calculatorCalled := false
	calculator := func() int {
		calculatorCalled = true
		return 2 // Base strength + 1 support
	}

	lazyCalc := &LazyStrengthCalculator{
		key:        "test",
		calculator: calculator,
	}

	// Verify calculator hasn't been called yet
	if calculatorCalled {
		t.Error("Calculator should not be called immediately")
	}

	// Now trigger calculation
	result := lazyCalc.Calculate()

	if !calculatorCalled {
		t.Error("Calculator should be called after Calculate()")
	}

	if result != 2 {
		t.Errorf("Expected strength 2, got %d", result)
	}

	// Reset flag and call again - should use cached result
	calculatorCalled = false
	result2 := lazyCalc.Calculate()

	if calculatorCalled {
		t.Error("Calculator should not be called again for cached result")
	}

	if result != result2 {
		t.Error("Cached result should be the same")
	}
}

// TestLazyReasonGeneration tests that resolution reasoning is lazy
func TestLazyReasonGeneration(t *testing.T) {
	orders := []Order{
		{Unit: "A Berlin", Type: Move, Source: "berlin", Destination: "munich", Owner: "germany"},
	}

	engine := NewDATCEngine(orders)
	defer engine.Cleanup()

	// Resolve the order first
	results := engine.ResolveAll()
	if len(results) != 1 {
		t.Fatalf("Expected 1 result, got %d", len(results))
	}

	order := &results[0].Order

	// Create lazy reason generator
	reasonGen := &LazyReasonGenerator{
		order:  order,
		engine: engine,
	}

	// Verify it hasn't been evaluated yet
	reasonGen.mu.RLock()
	evaluated := reasonGen.evaluated
	reasonGen.mu.RUnlock()

	if evaluated {
		t.Error("Reason should not be generated immediately")
	}

	// Now trigger reason generation
	startTime := time.Now()
	reason := reasonGen.GetReason()
	firstCallDuration := time.Since(startTime)

	// Verify it's now evaluated
	reasonGen.mu.RLock()
	evaluatedAfter := reasonGen.evaluated
	reasonGen.mu.RUnlock()

	if !evaluatedAfter {
		t.Error("Reason should be generated after GetReason() call")
	}

	if reason == "" {
		t.Error("Reason should not be empty")
	}

	// Second call should use cached result and be faster
	startTime = time.Now()
	reason2 := reasonGen.GetReason()
	secondCallDuration := time.Since(startTime)

	if reason != reason2 {
		t.Error("Cached reason should be the same")
	}

	// Second call should be significantly faster
	if secondCallDuration >= firstCallDuration {
		t.Errorf("Cached reason generation should be faster: first=%v, second=%v", firstCallDuration, secondCallDuration)
	}
}

// BenchmarkLazyVsEagerEvaluation compares lazy vs eager evaluation performance
func BenchmarkLazyVsEagerEvaluation(b *testing.B) {
	orders := []Order{
		{Unit: "A Berlin", Type: Move, Source: "berlin", Destination: "munich", Owner: "germany"},
		{Unit: "A Vienna", Type: Support, Source: "vienna", Auxiliary: "berlin -> munich", Owner: "austria"},
		{Unit: "A Paris", Type: Move, Source: "paris", Destination: "burgundy", Owner: "france"},
		{Unit: "F North Sea", Type: Convoy, Source: "north_sea", Auxiliary: "A london - belgium", Owner: "england"},
	}

	b.Run("LazyEvaluation", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			engine := NewDATCEngine(orders)

			// Only resolve orders, don't generate detailed reasons unless needed
			results := engine.ResolveAll()

			// Only access reason for first order (lazy evaluation)
			if len(results) > 0 {
				_ = results[0].Reason
			}

			engine.Cleanup()
		}
	})

	b.Run("EagerEvaluation", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			engine := NewDATCEngine(orders)

			results := engine.ResolveAll()

			// Access all reasons (eager evaluation)
			for _, result := range results {
				_ = result.Reason
			}

			engine.Cleanup()
		}
	})
}

// BenchmarkConvoyPathCaching tests convoy path caching effectiveness
func BenchmarkConvoyPathCaching(b *testing.B) {
	orders := []Order{
		{Unit: "A London", Type: Move, Source: "london", Destination: "belgium", Owner: "england"},
		{Unit: "F North Sea", Type: Convoy, Source: "north_sea", Auxiliary: "A london - belgium", Owner: "england"},
		{Unit: "F English Channel", Type: Convoy, Source: "english_channel", Auxiliary: "A london - belgium", Owner: "england"},
	}

	engine := NewDATCEngine(orders)
	defer engine.Cleanup()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		// Multiple calls to same convoy path should benefit from caching
		lazyPath := engine.convoyResolver.GetLazyConvoyPath("london", "belgium", engine.orderedOrders, engine)
		_ = lazyPath.IsValid(true)
		_ = lazyPath.IsValid(false) // Different optimism should still use cache for path structure
	}
}
