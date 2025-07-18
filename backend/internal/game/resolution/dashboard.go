package resolution

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"
)

// PerformanceDashboard provides utilities for displaying performance metrics
type PerformanceDashboard struct {
	monitor *PerformanceMonitor
}

// NewPerformanceDashboard creates a new performance dashboard
func NewPerformanceDashboard(monitor *PerformanceMonitor) *PerformanceDashboard {
	return &PerformanceDashboard{
		monitor: monitor,
	}
}

// GenerateReport creates a comprehensive performance report
func (pd *PerformanceDashboard) GenerateReport() string {
	metrics := pd.monitor.GetMetrics()

	var report strings.Builder

	report.WriteString("=== DATC Engine Performance Report ===\n\n")

	// Resolution Performance
	report.WriteString("📊 Resolution Performance:\n")
	report.WriteString(fmt.Sprintf("  Total Resolutions: %d\n", metrics.TotalResolutions))
	report.WriteString(fmt.Sprintf("  Total Time: %v\n", metrics.TotalResolutionTime))
	report.WriteString(fmt.Sprintf("  Average Time: %v\n", metrics.AverageResolutionTime))
	report.WriteString(fmt.Sprintf("  Min Time: %v\n", metrics.MinResolutionTime))
	report.WriteString(fmt.Sprintf("  Max Time: %v\n", metrics.MaxResolutionTime))
	report.WriteString(fmt.Sprintf("  Resolutions/sec: %.2f\n", pd.monitor.GetResolutionsPerSecond()))
	report.WriteString(fmt.Sprintf("  Avg Orders/Resolution: %.1f\n", pd.monitor.GetAverageOrdersPerResolution()))
	report.WriteString("\n")

	// Memory Performance
	report.WriteString("💾 Memory Performance:\n")
	report.WriteString(fmt.Sprintf("  Total Bytes Allocated: %s\n", formatBytes(metrics.TotalBytesAllocated)))
	report.WriteString(fmt.Sprintf("  Peak Memory Usage: %s\n", formatBytes(metrics.PeakMemoryUsage)))
	report.WriteString(fmt.Sprintf("  GC Count: %d\n", metrics.GCCount))
	report.WriteString(fmt.Sprintf("  GC Pause Time: %v\n", metrics.GCPauseTime))
	report.WriteString("\n")

	// Engine Metrics
	report.WriteString("⚙️ Engine Metrics:\n")
	report.WriteString(fmt.Sprintf("  Orders Processed: %d\n", metrics.OrdersProcessed))
	report.WriteString(fmt.Sprintf("  Conflicts Resolved: %d\n", metrics.ConflictsResolved))
	report.WriteString(fmt.Sprintf("  Strength Calculations: %d\n", metrics.StrengthCalculations))
	report.WriteString(fmt.Sprintf("  Cache Hit Rate: %.1f%%\n", pd.monitor.GetCacheHitRate()))
	report.WriteString(fmt.Sprintf("  Cache Hits: %d\n", metrics.CacheHits))
	report.WriteString(fmt.Sprintf("  Cache Misses: %d\n", metrics.CacheMisses))
	report.WriteString("\n")

	// String Optimization Metrics
	report.WriteString("🔤 String Optimization:\n")
	report.WriteString(fmt.Sprintf("  Strings Interned: %d\n", metrics.StringsInterned))
	report.WriteString(fmt.Sprintf("  String Builder Reuses: %d\n", metrics.StringBuilderReuses))
	report.WriteString(fmt.Sprintf("  String Allocations: %d\n", metrics.StringAllocations))
	report.WriteString("\n")

	// Lazy Evaluation Metrics
	report.WriteString("⏱️ Lazy Evaluation:\n")
	report.WriteString(fmt.Sprintf("  Evaluations Skipped: %d\n", metrics.LazyEvaluationsSkipped))
	report.WriteString(fmt.Sprintf("  Evaluations Forced: %d\n", metrics.LazyEvaluationsForced))
	lazyTotal := metrics.LazyEvaluationsSkipped + metrics.LazyEvaluationsForced
	if lazyTotal > 0 {
		skipRate := float64(metrics.LazyEvaluationsSkipped) / float64(lazyTotal) * 100
		report.WriteString(fmt.Sprintf("  Skip Rate: %.1f%%\n", skipRate))
	}
	report.WriteString("\n")

	// Performance Analysis
	report.WriteString("📈 Performance Analysis:\n")
	report.WriteString(pd.generatePerformanceAnalysis(metrics))

	return report.String()
}

// GenerateJSONReport creates a JSON performance report
func (pd *PerformanceDashboard) GenerateJSONReport() (string, error) {
	metrics := pd.monitor.GetMetrics()

	// Add calculated fields
	report := map[string]interface{}{
		"metrics": metrics,
		"calculated": map[string]interface{}{
			"cache_hit_rate":                pd.monitor.GetCacheHitRate(),
			"resolutions_per_second":        pd.monitor.GetResolutionsPerSecond(),
			"average_orders_per_resolution": pd.monitor.GetAverageOrdersPerResolution(),
		},
		"timestamp": time.Now().UTC(),
	}

	jsonData, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal JSON report: %w", err)
	}

	return string(jsonData), nil
}

// GenerateComparisonReport compares current metrics with baseline
func (pd *PerformanceDashboard) GenerateComparisonReport(baseline PerformanceMetrics) string {
	current := pd.monitor.GetMetrics()

	var report strings.Builder

	report.WriteString("=== Performance Comparison Report ===\n\n")

	// Resolution time comparison
	if baseline.AverageResolutionTime > 0 && current.AverageResolutionTime > 0 {
		improvement := float64(baseline.AverageResolutionTime-current.AverageResolutionTime) / float64(baseline.AverageResolutionTime) * 100
		report.WriteString(fmt.Sprintf("⏱️ Average Resolution Time: %v → %v (%.1f%% %s)\n",
			baseline.AverageResolutionTime, current.AverageResolutionTime,
			abs(improvement), improvementDirection(improvement)))
	}

	// Memory comparison
	if baseline.TotalBytesAllocated > 0 && current.TotalBytesAllocated > 0 {
		improvement := float64(baseline.TotalBytesAllocated-current.TotalBytesAllocated) / float64(baseline.TotalBytesAllocated) * 100
		report.WriteString(fmt.Sprintf("💾 Memory Usage: %s → %s (%.1f%% %s)\n",
			formatBytes(baseline.TotalBytesAllocated), formatBytes(current.TotalBytesAllocated),
			abs(improvement), improvementDirection(improvement)))
	}

	// Cache hit rate comparison
	baselineCacheRate := calculateCacheHitRate(baseline.CacheHits, baseline.CacheMisses)
	currentCacheRate := pd.monitor.GetCacheHitRate()
	if baselineCacheRate > 0 {
		improvement := currentCacheRate - baselineCacheRate
		report.WriteString(fmt.Sprintf("🎯 Cache Hit Rate: %.1f%% → %.1f%% (%+.1f%%)\n",
			baselineCacheRate, currentCacheRate, improvement))
	}

	// String optimization comparison
	if baseline.StringsInterned > 0 && current.StringsInterned > 0 {
		improvement := float64(current.StringsInterned-baseline.StringsInterned) / float64(baseline.StringsInterned) * 100
		report.WriteString(fmt.Sprintf("🔤 Strings Interned: %d → %d (%+.1f%%)\n",
			baseline.StringsInterned, current.StringsInterned, improvement))
	}

	return report.String()
}

// GenerateTopBottlenecks identifies performance bottlenecks
func (pd *PerformanceDashboard) GenerateTopBottlenecks() string {
	metrics := pd.monitor.GetMetrics()

	var bottlenecks []string

	// Check cache hit rate
	cacheHitRate := pd.monitor.GetCacheHitRate()
	if cacheHitRate < 80 {
		bottlenecks = append(bottlenecks, fmt.Sprintf("Low cache hit rate: %.1f%% (target: >80%%)", cacheHitRate))
	}

	// Check lazy evaluation efficiency
	lazyTotal := metrics.LazyEvaluationsSkipped + metrics.LazyEvaluationsForced
	if lazyTotal > 0 {
		skipRate := float64(metrics.LazyEvaluationsSkipped) / float64(lazyTotal) * 100
		if skipRate < 60 {
			bottlenecks = append(bottlenecks, fmt.Sprintf("Low lazy evaluation skip rate: %.1f%% (target: >60%%)", skipRate))
		}
	}

	// Check GC pressure
	if metrics.GCCount > 0 && metrics.TotalResolutions > 0 {
		gcPerResolution := float64(metrics.GCCount) / float64(metrics.TotalResolutions)
		if gcPerResolution > 0.1 {
			bottlenecks = append(bottlenecks, fmt.Sprintf("High GC pressure: %.2f GCs per resolution", gcPerResolution))
		}
	}

	// Check resolution time consistency
	if metrics.MaxResolutionTime > 0 && metrics.AverageResolutionTime > 0 {
		variance := float64(metrics.MaxResolutionTime) / float64(metrics.AverageResolutionTime)
		if variance > 5 {
			bottlenecks = append(bottlenecks, fmt.Sprintf("High resolution time variance: max is %.1fx average", variance))
		}
	}

	var report strings.Builder
	report.WriteString("🚨 Performance Bottlenecks:\n")

	if len(bottlenecks) == 0 {
		report.WriteString("  ✅ No significant bottlenecks detected\n")
	} else {
		for i, bottleneck := range bottlenecks {
			report.WriteString(fmt.Sprintf("  %d. %s\n", i+1, bottleneck))
		}
	}

	return report.String()
}

// generatePerformanceAnalysis provides insights into performance metrics
func (pd *PerformanceDashboard) generatePerformanceAnalysis(metrics PerformanceMetrics) string {
	var analysis strings.Builder

	// Efficiency analysis
	if metrics.TotalResolutions > 0 {
		avgStrengthCalcsPerOrder := float64(metrics.StrengthCalculations) / float64(metrics.OrdersProcessed)

		analysis.WriteString(fmt.Sprintf("  • Average strength calculations per order: %.1f\n", avgStrengthCalcsPerOrder))
		if avgStrengthCalcsPerOrder > 3 {
			analysis.WriteString("    ⚠️ High strength calculation ratio - consider caching improvements\n")
		} else if avgStrengthCalcsPerOrder < 1.5 {
			analysis.WriteString("    ✅ Efficient strength calculation ratio\n")
		}
	}

	// Cache efficiency
	cacheHitRate := pd.monitor.GetCacheHitRate()
	if cacheHitRate > 90 {
		analysis.WriteString("  ✅ Excellent cache performance\n")
	} else if cacheHitRate > 70 {
		analysis.WriteString("  👍 Good cache performance\n")
	} else if cacheHitRate > 50 {
		analysis.WriteString("  ⚠️ Moderate cache performance - room for improvement\n")
	} else {
		analysis.WriteString("  🚨 Poor cache performance - investigate cache strategy\n")
	}

	// String optimization effectiveness
	if metrics.StringsInterned > 0 && metrics.StringBuilderReuses > 0 {
		reuseRatio := float64(metrics.StringBuilderReuses) / float64(metrics.StringsInterned)
		if reuseRatio > 2 {
			analysis.WriteString("  ✅ Excellent string optimization effectiveness\n")
		} else if reuseRatio > 1 {
			analysis.WriteString("  👍 Good string optimization effectiveness\n")
		} else {
			analysis.WriteString("  ⚠️ String optimization could be improved\n")
		}
	}

	return analysis.String()
}

// Helper functions
func formatBytes(bytes int64) string {
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

func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

func improvementDirection(improvement float64) string {
	if improvement > 0 {
		return "improvement"
	}
	return "regression"
}

func calculateCacheHitRate(hits, misses int64) float64 {
	total := hits + misses
	if total == 0 {
		return 0
	}
	return float64(hits) / float64(total) * 100
}

// PerformanceAlert represents a performance alert
type PerformanceAlert struct {
	Level     string    `json:"level"` // "info", "warning", "critical"
	Message   string    `json:"message"`
	Metric    string    `json:"metric"`
	Value     float64   `json:"value"`
	Threshold float64   `json:"threshold"`
	Timestamp time.Time `json:"timestamp"`
}

// AlertManager manages performance alerts
type AlertManager struct {
	alerts     []PerformanceAlert
	thresholds map[string]float64
}

// NewAlertManager creates a new alert manager with default thresholds
func NewAlertManager() *AlertManager {
	return &AlertManager{
		alerts: make([]PerformanceAlert, 0),
		thresholds: map[string]float64{
			"cache_hit_rate_warning":  70.0,
			"cache_hit_rate_critical": 50.0,
			"avg_resolution_time_ms":  100.0,
			"max_resolution_time_ms":  1000.0,
			"gc_per_resolution":       0.1,
			"lazy_skip_rate_warning":  60.0,
		},
	}
}

// CheckAlerts evaluates current metrics against thresholds
func (am *AlertManager) CheckAlerts(metrics PerformanceMetrics, monitor *PerformanceMonitor) []PerformanceAlert {
	var newAlerts []PerformanceAlert
	now := time.Now()

	// Check cache hit rate
	cacheHitRate := monitor.GetCacheHitRate()
	if cacheHitRate < am.thresholds["cache_hit_rate_critical"] {
		newAlerts = append(newAlerts, PerformanceAlert{
			Level:     "critical",
			Message:   "Cache hit rate critically low",
			Metric:    "cache_hit_rate",
			Value:     cacheHitRate,
			Threshold: am.thresholds["cache_hit_rate_critical"],
			Timestamp: now,
		})
	} else if cacheHitRate < am.thresholds["cache_hit_rate_warning"] {
		newAlerts = append(newAlerts, PerformanceAlert{
			Level:     "warning",
			Message:   "Cache hit rate below optimal",
			Metric:    "cache_hit_rate",
			Value:     cacheHitRate,
			Threshold: am.thresholds["cache_hit_rate_warning"],
			Timestamp: now,
		})
	}

	// Check average resolution time
	avgTimeMs := float64(metrics.AverageResolutionTime.Nanoseconds()) / 1e6
	if avgTimeMs > am.thresholds["avg_resolution_time_ms"] {
		newAlerts = append(newAlerts, PerformanceAlert{
			Level:     "warning",
			Message:   "Average resolution time high",
			Metric:    "avg_resolution_time_ms",
			Value:     avgTimeMs,
			Threshold: am.thresholds["avg_resolution_time_ms"],
			Timestamp: now,
		})
	}

	// Check max resolution time
	maxTimeMs := float64(metrics.MaxResolutionTime.Nanoseconds()) / 1e6
	if maxTimeMs > am.thresholds["max_resolution_time_ms"] {
		newAlerts = append(newAlerts, PerformanceAlert{
			Level:     "critical",
			Message:   "Maximum resolution time exceeded",
			Metric:    "max_resolution_time_ms",
			Value:     maxTimeMs,
			Threshold: am.thresholds["max_resolution_time_ms"],
			Timestamp: now,
		})
	}

	am.alerts = append(am.alerts, newAlerts...)
	return newAlerts
}

// GetRecentAlerts returns alerts from the last duration
func (am *AlertManager) GetRecentAlerts(duration time.Duration) []PerformanceAlert {
	cutoff := time.Now().Add(-duration)
	var recent []PerformanceAlert

	for _, alert := range am.alerts {
		if alert.Timestamp.After(cutoff) {
			recent = append(recent, alert)
		}
	}

	// Sort by timestamp, newest first
	sort.Slice(recent, func(i, j int) bool {
		return recent[i].Timestamp.After(recent[j].Timestamp)
	})

	return recent
}

// Global dashboard instance
var globalDashboard = NewPerformanceDashboard(GetGlobalPerformanceMonitor())

// GetGlobalDashboard returns the global performance dashboard
func GetGlobalDashboard() *PerformanceDashboard {
	return globalDashboard
}
