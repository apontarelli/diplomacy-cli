package resolution

import (
	"context"
	"runtime"
	"sync"
	"sync/atomic"
	"time"
)

// PerformanceMetrics tracks key performance indicators for the resolution engine
type PerformanceMetrics struct {
	// Resolution timing metrics
	TotalResolutions      int64         `json:"total_resolutions"`
	TotalResolutionTime   time.Duration `json:"total_resolution_time"`
	AverageResolutionTime time.Duration `json:"average_resolution_time"`
	MaxResolutionTime     time.Duration `json:"max_resolution_time"`
	MinResolutionTime     time.Duration `json:"min_resolution_time"`

	// Memory metrics
	TotalAllocations    int64         `json:"total_allocations"`
	TotalBytesAllocated int64         `json:"total_bytes_allocated"`
	PeakMemoryUsage     int64         `json:"peak_memory_usage"`
	GCCount             int64         `json:"gc_count"`
	GCPauseTime         time.Duration `json:"gc_pause_time"`

	// Engine-specific metrics
	OrdersProcessed      int64 `json:"orders_processed"`
	ConflictsResolved    int64 `json:"conflicts_resolved"`
	StrengthCalculations int64 `json:"strength_calculations"`
	CacheHits            int64 `json:"cache_hits"`
	CacheMisses          int64 `json:"cache_misses"`

	// String optimization metrics
	StringsInterned     int64 `json:"strings_interned"`
	StringBuilderReuses int64 `json:"string_builder_reuses"`
	StringAllocations   int64 `json:"string_allocations"`

	// Lazy evaluation metrics
	LazyEvaluationsSkipped int64 `json:"lazy_evaluations_skipped"`
	LazyEvaluationsForced  int64 `json:"lazy_evaluations_forced"`

	mu sync.RWMutex
}

// PerformanceMonitor provides centralized performance monitoring for the resolution engine
type PerformanceMonitor struct {
	metrics   *PerformanceMetrics
	enabled   bool
	startTime time.Time
	mu        sync.RWMutex
}

// NewPerformanceMonitor creates a new performance monitor
func NewPerformanceMonitor() *PerformanceMonitor {
	return &PerformanceMonitor{
		metrics: &PerformanceMetrics{
			MinResolutionTime: time.Duration(^uint64(0) >> 1), // Max duration as initial min
		},
		enabled:   true,
		startTime: time.Now(),
	}
}

// Enable turns on performance monitoring
func (pm *PerformanceMonitor) Enable() {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	pm.enabled = true
}

// Disable turns off performance monitoring
func (pm *PerformanceMonitor) Disable() {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	pm.enabled = false
}

// IsEnabled returns whether monitoring is enabled
func (pm *PerformanceMonitor) IsEnabled() bool {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	return pm.enabled
}

// StartResolution begins timing a resolution operation
func (pm *PerformanceMonitor) StartResolution(ctx context.Context, orderCount int) *ResolutionTimer {
	if !pm.IsEnabled() {
		return &ResolutionTimer{enabled: false}
	}

	return &ResolutionTimer{
		monitor:    pm,
		startTime:  time.Now(),
		orderCount: orderCount,
		enabled:    true,
		ctx:        ctx,
	}
}

// RecordCacheHit records a cache hit
func (pm *PerformanceMonitor) RecordCacheHit() {
	if !pm.IsEnabled() {
		return
	}
	atomic.AddInt64(&pm.metrics.CacheHits, 1)
}

// RecordCacheMiss records a cache miss
func (pm *PerformanceMonitor) RecordCacheMiss() {
	if !pm.IsEnabled() {
		return
	}
	atomic.AddInt64(&pm.metrics.CacheMisses, 1)
}

// RecordStrengthCalculation records a strength calculation
func (pm *PerformanceMonitor) RecordStrengthCalculation() {
	if !pm.IsEnabled() {
		return
	}
	atomic.AddInt64(&pm.metrics.StrengthCalculations, 1)
}

// RecordStringInterned records a string interning operation
func (pm *PerformanceMonitor) RecordStringInterned() {
	if !pm.IsEnabled() {
		return
	}
	atomic.AddInt64(&pm.metrics.StringsInterned, 1)
}

// RecordStringBuilderReuse records a string builder reuse
func (pm *PerformanceMonitor) RecordStringBuilderReuse() {
	if !pm.IsEnabled() {
		return
	}
	atomic.AddInt64(&pm.metrics.StringBuilderReuses, 1)
}

// RecordLazyEvaluationSkipped records a skipped lazy evaluation
func (pm *PerformanceMonitor) RecordLazyEvaluationSkipped() {
	if !pm.IsEnabled() {
		return
	}
	atomic.AddInt64(&pm.metrics.LazyEvaluationsSkipped, 1)
}

// RecordLazyEvaluationForced records a forced lazy evaluation
func (pm *PerformanceMonitor) RecordLazyEvaluationForced() {
	if !pm.IsEnabled() {
		return
	}
	atomic.AddInt64(&pm.metrics.LazyEvaluationsForced, 1)
}

// GetMetrics returns a copy of current metrics
func (pm *PerformanceMonitor) GetMetrics() PerformanceMetrics {
	pm.metrics.mu.RLock()
	defer pm.metrics.mu.RUnlock()

	// Update memory stats
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	// Create a copy without the mutex
	metrics := PerformanceMetrics{
		TotalResolutions:       pm.metrics.TotalResolutions,
		TotalResolutionTime:    pm.metrics.TotalResolutionTime,
		AverageResolutionTime:  pm.metrics.AverageResolutionTime,
		MaxResolutionTime:      pm.metrics.MaxResolutionTime,
		MinResolutionTime:      pm.metrics.MinResolutionTime,
		TotalAllocations:       pm.metrics.TotalAllocations,
		TotalBytesAllocated:    int64(memStats.TotalAlloc),
		PeakMemoryUsage:        int64(memStats.Sys),
		GCCount:                int64(memStats.NumGC),
		GCPauseTime:            time.Duration(memStats.PauseTotalNs),
		OrdersProcessed:        pm.metrics.OrdersProcessed,
		ConflictsResolved:      pm.metrics.ConflictsResolved,
		StrengthCalculations:   pm.metrics.StrengthCalculations,
		CacheHits:              pm.metrics.CacheHits,
		CacheMisses:            pm.metrics.CacheMisses,
		StringsInterned:        pm.metrics.StringsInterned,
		StringBuilderReuses:    pm.metrics.StringBuilderReuses,
		StringAllocations:      pm.metrics.StringAllocations,
		LazyEvaluationsSkipped: pm.metrics.LazyEvaluationsSkipped,
		LazyEvaluationsForced:  pm.metrics.LazyEvaluationsForced,
	}

	// Calculate average resolution time
	if metrics.TotalResolutions > 0 {
		metrics.AverageResolutionTime = time.Duration(int64(metrics.TotalResolutionTime) / metrics.TotalResolutions)
	}

	return metrics
}

// Reset clears all metrics
func (pm *PerformanceMonitor) Reset() {
	pm.metrics.mu.Lock()
	defer pm.metrics.mu.Unlock()

	pm.metrics = &PerformanceMetrics{
		MinResolutionTime: time.Duration(^uint64(0) >> 1), // Max duration as initial min
	}
	pm.startTime = time.Now()
}

// ResolutionTimer tracks the timing of a single resolution operation
type ResolutionTimer struct {
	monitor    *PerformanceMonitor
	startTime  time.Time
	orderCount int
	enabled    bool
	ctx        context.Context
}

// Finish completes the timing and records metrics
func (rt *ResolutionTimer) Finish() {
	if !rt.enabled {
		return
	}

	duration := time.Since(rt.startTime)
	rt.monitor.recordResolution(duration, rt.orderCount)
}

// recordResolution records a completed resolution with timing and order count
func (pm *PerformanceMonitor) recordResolution(duration time.Duration, orderCount int) {
	pm.metrics.mu.Lock()
	defer pm.metrics.mu.Unlock()

	pm.metrics.TotalResolutions++
	pm.metrics.TotalResolutionTime += duration
	pm.metrics.OrdersProcessed += int64(orderCount)

	if duration > pm.metrics.MaxResolutionTime {
		pm.metrics.MaxResolutionTime = duration
	}

	if duration < pm.metrics.MinResolutionTime {
		pm.metrics.MinResolutionTime = duration
	}
}

// GetCacheHitRate returns the cache hit rate as a percentage
func (pm *PerformanceMonitor) GetCacheHitRate() float64 {
	metrics := pm.GetMetrics()
	total := metrics.CacheHits + metrics.CacheMisses
	if total == 0 {
		return 0
	}
	return float64(metrics.CacheHits) / float64(total) * 100
}

// GetAverageOrdersPerResolution returns the average number of orders per resolution
func (pm *PerformanceMonitor) GetAverageOrdersPerResolution() float64 {
	metrics := pm.GetMetrics()
	if metrics.TotalResolutions == 0 {
		return 0
	}
	return float64(metrics.OrdersProcessed) / float64(metrics.TotalResolutions)
}

// GetResolutionsPerSecond returns the average resolutions per second since monitoring started
func (pm *PerformanceMonitor) GetResolutionsPerSecond() float64 {
	metrics := pm.GetMetrics()
	elapsed := time.Since(pm.startTime)
	if elapsed == 0 {
		return 0
	}
	return float64(metrics.TotalResolutions) / elapsed.Seconds()
}

// Global performance monitor instance
var globalPerformanceMonitor = NewPerformanceMonitor()

// GetGlobalPerformanceMonitor returns the global performance monitor
func GetGlobalPerformanceMonitor() *PerformanceMonitor {
	return globalPerformanceMonitor
}

// EnableGlobalMonitoring enables global performance monitoring
func EnableGlobalMonitoring() {
	globalPerformanceMonitor.Enable()
}

// DisableGlobalMonitoring disables global performance monitoring
func DisableGlobalMonitoring() {
	globalPerformanceMonitor.Disable()
}
