package resolution

import (
	"fmt"
	"runtime"
	"testing"
	"time"
)

// BenchmarkDATCEngineWithMonitoring benchmarks the DATC engine with performance monitoring enabled
func BenchmarkDATCEngineWithMonitoring(b *testing.B) {
	// Create test orders for benchmarking
	orders := []Order{
		{Unit: "A Berlin", Type: Move, Source: "berlin", Destination: "munich", Owner: "germany"},
		{Unit: "A Vienna", Type: Support, Source: "vienna", Auxiliary: "berlin -> munich", Owner: "austria"},
		{Unit: "A Paris", Type: Move, Source: "paris", Destination: "munich", Owner: "france"},
		{Unit: "F London", Type: Move, Source: "london", Destination: "belgium", Owner: "england"},
		{Unit: "F North Sea", Type: Convoy, Source: "north_sea", Auxiliary: "A london - belgium", Owner: "england"},
		{Unit: "A Moscow", Type: Hold, Source: "moscow", Destination: "moscow", Owner: "russia"},
		{Unit: "A Warsaw", Type: Move, Source: "warsaw", Destination: "berlin", Owner: "russia"},
		{Unit: "F Rome", Type: Support, Source: "rome", Auxiliary: "vienna -> munich", Owner: "italy"},
	}

	b.Run("WithMonitoring", func(b *testing.B) {
		// Enable monitoring
		EnableGlobalMonitoring()
		defer DisableGlobalMonitoring()

		// Reset metrics before benchmark
		GetGlobalPerformanceMonitor().Reset()

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			engine := NewDATCEngine(orders)
			results := engine.ResolveAll()

			// Access reasons to trigger lazy evaluation
			for _, result := range results {
				_ = result.Reason
			}

			engine.Cleanup()
		}
		b.StopTimer()

		// Report performance metrics
		metrics := GetGlobalPerformanceMonitor().GetMetrics()
		dashboard := GetGlobalDashboard()

		b.Logf("Performance Metrics:\n%s", dashboard.GenerateReport())

		// Report custom metrics
		b.ReportMetric(float64(metrics.TotalResolutions), "resolutions")
		b.ReportMetric(float64(metrics.StrengthCalculations), "strength_calcs")
		b.ReportMetric(GetGlobalPerformanceMonitor().GetCacheHitRate(), "cache_hit_rate_%")
		b.ReportMetric(float64(metrics.StringsInterned), "strings_interned")
	})

	b.Run("WithoutMonitoring", func(b *testing.B) {
		// Disable monitoring for comparison
		DisableGlobalMonitoring()

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			engine := NewDATCEngine(orders)
			results := engine.ResolveAll()

			// Access reasons to trigger lazy evaluation
			for _, result := range results {
				_ = result.Reason
			}

			engine.Cleanup()
		}
	})
}

// BenchmarkComplexScenarioWithMetrics benchmarks a complex scenario with detailed metrics
func BenchmarkComplexScenarioWithMetrics(b *testing.B) {
	// Create a more complex scenario with 14 orders (full 7-player game)
	orders := []Order{
		// Germany
		{Unit: "A Berlin", Type: Move, Source: "berlin", Destination: "munich", Owner: "germany"},
		{Unit: "F Kiel", Type: Support, Source: "kiel", Auxiliary: "berlin -> munich", Owner: "germany"},

		// Austria
		{Unit: "A Vienna", Type: Move, Source: "vienna", Destination: "munich", Owner: "austria"},
		{Unit: "A Budapest", Type: Support, Source: "budapest", Auxiliary: "vienna -> munich", Owner: "austria"},

		// France
		{Unit: "A Paris", Type: Move, Source: "paris", Destination: "burgundy", Owner: "france"},
		{Unit: "A Marseilles", Type: Support, Source: "marseilles", Auxiliary: "paris -> burgundy", Owner: "france"},

		// England
		{Unit: "F London", Type: Move, Source: "london", Destination: "belgium", Owner: "england"},
		{Unit: "F North Sea", Type: Convoy, Source: "north_sea", Auxiliary: "A london - belgium", Owner: "england"},

		// Russia
		{Unit: "A Moscow", Type: Move, Source: "moscow", Destination: "warsaw", Owner: "russia"},
		{Unit: "A St Petersburg", Type: Support, Source: "st_petersburg", Auxiliary: "moscow -> warsaw", Owner: "russia"},

		// Italy
		{Unit: "A Rome", Type: Move, Source: "rome", Destination: "venice", Owner: "italy"},
		{Unit: "F Naples", Type: Support, Source: "naples", Auxiliary: "rome -> venice", Owner: "italy"},

		// Turkey
		{Unit: "A Constantinople", Type: Hold, Source: "constantinople", Destination: "constantinople", Owner: "turkey"},
		{Unit: "F Ankara", Type: Hold, Source: "ankara", Destination: "ankara", Owner: "turkey"},
	}

	EnableGlobalMonitoring()
	defer DisableGlobalMonitoring()

	// Reset metrics
	GetGlobalPerformanceMonitor().Reset()

	var memStatsBefore, memStatsAfter runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&memStatsBefore)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		engine := NewDATCEngine(orders)
		results := engine.ResolveAll()

		// Force all lazy evaluations
		for _, result := range results {
			_ = result.Reason
		}

		engine.Cleanup()
	}
	b.StopTimer()

	runtime.GC()
	runtime.ReadMemStats(&memStatsAfter)

	// Generate comprehensive report
	metrics := GetGlobalPerformanceMonitor().GetMetrics()
	dashboard := GetGlobalDashboard()

	b.Logf("Complex Scenario Performance Report:\n%s", dashboard.GenerateReport())
	b.Logf("Bottleneck Analysis:\n%s", dashboard.GenerateTopBottlenecks())

	// Memory analysis
	memoryUsed := memStatsAfter.TotalAlloc - memStatsBefore.TotalAlloc
	b.Logf("Memory used during benchmark: %s", formatBytesLocal(int64(memoryUsed)))

	// Report detailed metrics
	b.ReportMetric(float64(metrics.TotalResolutions), "resolutions")
	b.ReportMetric(float64(metrics.OrdersProcessed), "orders_processed")
	b.ReportMetric(float64(metrics.StrengthCalculations), "strength_calculations")
	b.ReportMetric(GetGlobalPerformanceMonitor().GetCacheHitRate(), "cache_hit_rate_%")
	b.ReportMetric(float64(metrics.StringsInterned), "strings_interned")
	b.ReportMetric(float64(metrics.LazyEvaluationsSkipped), "lazy_skipped")
	b.ReportMetric(float64(metrics.LazyEvaluationsForced), "lazy_forced")
	b.ReportMetric(float64(memoryUsed)/float64(b.N), "bytes_per_op")
}

// BenchmarkStringOptimizationImpact measures the impact of string optimizations
func BenchmarkStringOptimizationImpact(b *testing.B) {
	orders := []Order{
		{Unit: "A Berlin", Type: Move, Source: "berlin", Destination: "munich", Owner: "germany"},
		{Unit: "A Vienna", Type: Support, Source: "vienna", Auxiliary: "berlin -> munich", Owner: "austria"},
		{Unit: "A Paris", Type: Move, Source: "paris", Destination: "munich", Owner: "france"},
	}

	b.Run("WithStringOptimizations", func(b *testing.B) {
		EnableGlobalMonitoring()
		defer DisableGlobalMonitoring()
		GetGlobalPerformanceMonitor().Reset()

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			engine := NewDATCEngine(orders)
			results := engine.ResolveAll()

			// Force reason generation to test string building
			for _, result := range results {
				_ = result.Reason
			}

			engine.Cleanup()
		}
		b.StopTimer()

		metrics := GetGlobalPerformanceMonitor().GetMetrics()
		b.ReportMetric(float64(metrics.StringsInterned), "strings_interned")
		b.ReportMetric(float64(metrics.StringBuilderReuses), "builder_reuses")
	})
}

// BenchmarkCacheEffectiveness measures cache effectiveness
func BenchmarkCacheEffectiveness(b *testing.B) {
	// Create orders that will result in repeated strength calculations
	orders := []Order{
		{Unit: "A Berlin", Type: Move, Source: "berlin", Destination: "munich", Owner: "germany"},
		{Unit: "A Vienna", Type: Move, Source: "vienna", Destination: "munich", Owner: "austria"},
		{Unit: "A Paris", Type: Move, Source: "paris", Destination: "munich", Owner: "france"},
		{Unit: "A Warsaw", Type: Support, Source: "warsaw", Auxiliary: "berlin -> munich", Owner: "russia"},
		{Unit: "A Rome", Type: Support, Source: "rome", Auxiliary: "vienna -> munich", Owner: "italy"},
	}

	EnableGlobalMonitoring()
	defer DisableGlobalMonitoring()

	b.Run("CacheEffectiveness", func(b *testing.B) {
		GetGlobalPerformanceMonitor().Reset()

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			engine := NewDATCEngine(orders)
			_ = engine.ResolveAll()
			engine.Cleanup()
		}
		b.StopTimer()

		metrics := GetGlobalPerformanceMonitor().GetMetrics()
		cacheHitRate := GetGlobalPerformanceMonitor().GetCacheHitRate()

		b.Logf("Cache Performance:")
		b.Logf("  Hit Rate: %.1f%%", cacheHitRate)
		b.Logf("  Hits: %d", metrics.CacheHits)
		b.Logf("  Misses: %d", metrics.CacheMisses)
		b.Logf("  Strength Calculations: %d", metrics.StrengthCalculations)

		b.ReportMetric(cacheHitRate, "cache_hit_rate_%")
		b.ReportMetric(float64(metrics.CacheHits), "cache_hits")
		b.ReportMetric(float64(metrics.CacheMisses), "cache_misses")
	})
}

// BenchmarkLazyEvaluationEfficiency measures lazy evaluation efficiency
func BenchmarkLazyEvaluationEfficiency(b *testing.B) {
	orders := []Order{
		{Unit: "A Berlin", Type: Move, Source: "berlin", Destination: "munich", Owner: "germany"},
		{Unit: "A Vienna", Type: Support, Source: "vienna", Auxiliary: "berlin -> munich", Owner: "austria"},
		{Unit: "A Paris", Type: Move, Source: "paris", Destination: "burgundy", Owner: "france"},
	}

	EnableGlobalMonitoring()
	defer DisableGlobalMonitoring()

	b.Run("OnlyResolution", func(b *testing.B) {
		GetGlobalPerformanceMonitor().Reset()

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			engine := NewDATCEngine(orders)
			_ = engine.ResolveAll() // Don't access reasons
			engine.Cleanup()
		}
		b.StopTimer()

		metrics := GetGlobalPerformanceMonitor().GetMetrics()
		b.Logf("Lazy Evaluation (Resolution Only):")
		b.Logf("  Skipped: %d", metrics.LazyEvaluationsSkipped)
		b.Logf("  Forced: %d", metrics.LazyEvaluationsForced)

		b.ReportMetric(float64(metrics.LazyEvaluationsSkipped), "lazy_skipped")
		b.ReportMetric(float64(metrics.LazyEvaluationsForced), "lazy_forced")
	})

	b.Run("WithReasonGeneration", func(b *testing.B) {
		GetGlobalPerformanceMonitor().Reset()

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			engine := NewDATCEngine(orders)
			results := engine.ResolveAll()

			// Force all reason generation
			for _, result := range results {
				_ = result.Reason
			}

			engine.Cleanup()
		}
		b.StopTimer()

		metrics := GetGlobalPerformanceMonitor().GetMetrics()
		b.Logf("Lazy Evaluation (With Reasons):")
		b.Logf("  Skipped: %d", metrics.LazyEvaluationsSkipped)
		b.Logf("  Forced: %d", metrics.LazyEvaluationsForced)

		total := metrics.LazyEvaluationsSkipped + metrics.LazyEvaluationsForced
		if total > 0 {
			skipRate := float64(metrics.LazyEvaluationsSkipped) / float64(total) * 100
			b.Logf("  Skip Rate: %.1f%%", skipRate)
			b.ReportMetric(skipRate, "lazy_skip_rate_%")
		}
	})
}

// BenchmarkMemoryEfficiency measures memory efficiency improvements
func BenchmarkMemoryEfficiency(b *testing.B) {
	// Create many orders with repeated strings to test interning effectiveness
	orders := make([]Order, 50)
	territories := []string{"berlin", "munich", "vienna", "paris", "london", "moscow", "warsaw", "rome"}
	owners := []string{"germany", "austria", "france", "england", "russia", "italy", "turkey"}

	for i := range orders {
		orders[i] = Order{
			Unit:        fmt.Sprintf("A %s", territories[i%len(territories)]),
			Type:        Move,
			Source:      territories[i%len(territories)],
			Destination: territories[(i+1)%len(territories)],
			Owner:       owners[i%len(owners)],
		}
	}

	EnableGlobalMonitoring()
	defer DisableGlobalMonitoring()

	var m1, m2 runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&m1)

	GetGlobalPerformanceMonitor().Reset()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		engine := NewDATCEngine(orders)
		results := engine.ResolveAll()

		// Access some reasons to test string building
		for j, result := range results {
			if j%3 == 0 { // Access every third reason
				_ = result.Reason
			}
		}

		engine.Cleanup()
	}
	b.StopTimer()

	runtime.GC()
	runtime.ReadMemStats(&m2)

	metrics := GetGlobalPerformanceMonitor().GetMetrics()

	b.Logf("Memory Efficiency Report:")
	b.Logf("  Strings Interned: %d", metrics.StringsInterned)
	b.Logf("  String Builder Reuses: %d", metrics.StringBuilderReuses)
	b.Logf("  Total Allocations: %d", m2.TotalAlloc-m1.TotalAlloc)
	b.Logf("  Mallocs: %d", m2.Mallocs-m1.Mallocs)

	b.ReportMetric(float64(m2.TotalAlloc-m1.TotalAlloc)/float64(b.N), "bytes/op")
	b.ReportMetric(float64(m2.Mallocs-m1.Mallocs)/float64(b.N), "allocs/op")
	b.ReportMetric(float64(metrics.StringsInterned), "strings_interned")
	b.ReportMetric(float64(metrics.StringBuilderReuses), "builder_reuses")
}

// TestPerformanceRegression tests for performance regressions
func TestPerformanceRegression(t *testing.T) {
	orders := []Order{
		{Unit: "A Berlin", Type: Move, Source: "berlin", Destination: "munich", Owner: "germany"},
		{Unit: "A Vienna", Type: Support, Source: "vienna", Auxiliary: "berlin -> munich", Owner: "austria"},
		{Unit: "A Paris", Type: Move, Source: "paris", Destination: "munich", Owner: "france"},
	}

	EnableGlobalMonitoring()
	defer DisableGlobalMonitoring()

	// Performance thresholds (adjusted for simple test scenarios)
	const (
		maxAvgResolutionTimeMs = 10.0 // 10ms max average resolution time
		minCacheHitRate        = 15.0 // 15% minimum cache hit rate (simple scenarios have low cache reuse)
		maxMemoryPerOpMB       = 1.0  // 1MB max memory per operation
	)

	GetGlobalPerformanceMonitor().Reset()

	// Run multiple iterations to get stable metrics
	iterations := 100
	var memBefore, memAfter runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&memBefore)

	start := time.Now()
	for i := 0; i < iterations; i++ {
		engine := NewDATCEngine(orders)
		results := engine.ResolveAll()

		// Access reasons to trigger lazy evaluation
		for _, result := range results {
			_ = result.Reason
		}

		engine.Cleanup()
	}
	elapsed := time.Since(start)

	runtime.GC()
	runtime.ReadMemStats(&memAfter)

	metrics := GetGlobalPerformanceMonitor().GetMetrics()
	avgResolutionTimeMs := float64(elapsed.Nanoseconds()) / float64(iterations) / 1e6
	cacheHitRate := GetGlobalPerformanceMonitor().GetCacheHitRate()
	memoryPerOpMB := float64(memAfter.TotalAlloc-memBefore.TotalAlloc) / float64(iterations) / 1024 / 1024

	t.Logf("Performance Regression Test Results:")
	t.Logf("  Average Resolution Time: %.2f ms (threshold: %.2f ms)", avgResolutionTimeMs, maxAvgResolutionTimeMs)
	t.Logf("  Cache Hit Rate: %.1f%% (threshold: %.1f%%)", cacheHitRate, minCacheHitRate)
	t.Logf("  Memory Per Operation: %.2f MB (threshold: %.2f MB)", memoryPerOpMB, maxMemoryPerOpMB)
	t.Logf("  Total Resolutions: %d", metrics.TotalResolutions)
	t.Logf("  Strength Calculations: %d", metrics.StrengthCalculations)

	// Check thresholds
	if avgResolutionTimeMs > maxAvgResolutionTimeMs {
		t.Errorf("Performance regression: average resolution time %.2f ms exceeds threshold %.2f ms",
			avgResolutionTimeMs, maxAvgResolutionTimeMs)
	}

	if cacheHitRate < minCacheHitRate {
		t.Errorf("Performance regression: cache hit rate %.1f%% below threshold %.1f%%",
			cacheHitRate, minCacheHitRate)
	}

	if memoryPerOpMB > maxMemoryPerOpMB {
		t.Errorf("Performance regression: memory usage %.2f MB per operation exceeds threshold %.2f MB",
			memoryPerOpMB, maxMemoryPerOpMB)
	}
}

// formatBytesLocal formats bytes for local use (avoiding import cycle with dashboard.go)
func formatBytesLocal(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}
