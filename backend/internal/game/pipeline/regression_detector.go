package pipeline

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"sort"
	"time"
)

// RegressionDetector analyzes benchmark results for performance regressions
type RegressionDetector struct {
	config       RegressionConfig
	baselineFile string
	historyDir   string
}

// RegressionConfig defines thresholds and detection parameters
type RegressionConfig struct {
	MinorImprovementThreshold    float64 `json:"minor_improvement_threshold"`    // 0.95 = 5% improvement
	MajorImprovementThreshold    float64 `json:"major_improvement_threshold"`    // 0.80 = 20% improvement
	MinorDegradationThreshold    float64 `json:"minor_degradation_threshold"`    // 1.05 = 5% degradation
	MajorDegradationThreshold    float64 `json:"major_degradation_threshold"`    // 1.20 = 20% degradation
	CriticalDegradationThreshold float64 `json:"critical_degradation_threshold"` // 1.50 = 50% degradation

	// Statistical analysis parameters
	MinSampleSize           int     `json:"min_sample_size"`          // Minimum samples for trend analysis
	VarianceThreshold       float64 `json:"variance_threshold"`       // Maximum acceptable variance
	TrendAnalysisWindow     int     `json:"trend_analysis_window"`    // Number of recent runs to analyze
	StatisticalSignificance float64 `json:"statistical_significance"` // P-value threshold

	// Alert configuration
	AlertOnMinorDegradation    bool `json:"alert_on_minor_degradation"`
	AlertOnMajorDegradation    bool `json:"alert_on_major_degradation"`
	AlertOnCriticalDegradation bool `json:"alert_on_critical_degradation"`
	AlertOnHighVariance        bool `json:"alert_on_high_variance"`

	// Critical benchmarks that should never regress significantly
	CriticalBenchmarks []string `json:"critical_benchmarks"`
}

// DefaultRegressionConfig returns sensible defaults
func DefaultRegressionConfig() RegressionConfig {
	return RegressionConfig{
		MinorImprovementThreshold:    0.95,
		MajorImprovementThreshold:    0.80,
		MinorDegradationThreshold:    1.05,
		MajorDegradationThreshold:    1.20,
		CriticalDegradationThreshold: 1.50,

		MinSampleSize:           5,
		VarianceThreshold:       0.15, // 15% variance
		TrendAnalysisWindow:     10,
		StatisticalSignificance: 0.05, // 5% p-value

		AlertOnMinorDegradation:    false,
		AlertOnMajorDegradation:    true,
		AlertOnCriticalDegradation: true,
		AlertOnHighVariance:        true,

		CriticalBenchmarks: []string{
			"BenchmarkBaseline",
			"BenchmarkComplexityScenarios",
		},
	}
}

// RegressionAnalysis contains the complete analysis results
type RegressionAnalysis struct {
	Timestamp       time.Time                  `json:"timestamp"`
	Config          RegressionConfig           `json:"config"`
	BaselineMetrics map[string]BenchmarkMetric `json:"baseline_metrics"`
	CurrentMetrics  map[string]BenchmarkMetric `json:"current_metrics"`
	Comparisons     []RegressionResult         `json:"comparisons"`
	TrendAnalysis   []TrendAnalysis            `json:"trend_analysis"`
	Alerts          []RegressionAlert          `json:"alerts"`
	Summary         RegressionSummary          `json:"summary"`
}

// RegressionResult represents the analysis of a single benchmark
type RegressionResult struct {
	BenchmarkName           string     `json:"benchmark_name"`
	BaselineNsPerOp         int64      `json:"baseline_ns_per_op"`
	CurrentNsPerOp          int64      `json:"current_ns_per_op"`
	PerformanceRatio        float64    `json:"performance_ratio"`
	AllocRatio              float64    `json:"alloc_ratio"`
	BytesRatio              float64    `json:"bytes_ratio"`
	Status                  string     `json:"status"`
	Severity                string     `json:"severity"`
	IsCriticalBenchmark     bool       `json:"is_critical_benchmark"`
	StatisticalSignificance bool       `json:"statistical_significance"`
	ConfidenceInterval      [2]float64 `json:"confidence_interval"`
}

// TrendAnalysis analyzes performance trends over time
type TrendAnalysis struct {
	BenchmarkName     string    `json:"benchmark_name"`
	DataPoints        []float64 `json:"data_points"`
	TrendDirection    string    `json:"trend_direction"` // "improving", "degrading", "stable"
	TrendStrength     float64   `json:"trend_strength"`  // Correlation coefficient
	Variance          float64   `json:"variance"`
	Mean              float64   `json:"mean"`
	StandardDeviation float64   `json:"standard_deviation"`
	IsVolatile        bool      `json:"is_volatile"`
}

// RegressionAlert represents an alert condition
type RegressionAlert struct {
	Type           string    `json:"type"`     // "performance", "variance", "trend"
	Severity       string    `json:"severity"` // "minor", "major", "critical"
	BenchmarkName  string    `json:"benchmark_name"`
	Message        string    `json:"message"`
	Timestamp      time.Time `json:"timestamp"`
	ActionRequired bool      `json:"action_required"`
}

// RegressionSummary provides high-level statistics
type RegressionSummary struct {
	TotalBenchmarks        int `json:"total_benchmarks"`
	ImprovedBenchmarks     int `json:"improved_benchmarks"`
	DegradedBenchmarks     int `json:"degraded_benchmarks"`
	StableBenchmarks       int `json:"stable_benchmarks"`
	CriticalRegressions    int `json:"critical_regressions"`
	MajorRegressions       int `json:"major_regressions"`
	MinorRegressions       int `json:"minor_regressions"`
	HighVarianceBenchmarks int `json:"high_variance_benchmarks"`
	AlertsGenerated        int `json:"alerts_generated"`
}

// NewRegressionDetector creates a new regression detector
func NewRegressionDetector(projectRoot string, config RegressionConfig) *RegressionDetector {
	baselineFile := filepath.Join(projectRoot, "benchmarks", "results", "baseline.json")
	historyDir := filepath.Join(projectRoot, "benchmarks", "history")

	// Ensure history directory exists
	os.MkdirAll(historyDir, 0755)

	return &RegressionDetector{
		config:       config,
		baselineFile: baselineFile,
		historyDir:   historyDir,
	}
}

// AnalyzeRegression performs comprehensive regression analysis
func (rd *RegressionDetector) AnalyzeRegression(currentMetrics map[string]BenchmarkMetric) (*RegressionAnalysis, error) {
	analysis := &RegressionAnalysis{
		Timestamp:      time.Now(),
		Config:         rd.config,
		CurrentMetrics: currentMetrics,
		Comparisons:    make([]RegressionResult, 0),
		TrendAnalysis:  make([]TrendAnalysis, 0),
		Alerts:         make([]RegressionAlert, 0),
	}

	// Load baseline metrics
	baselineMetrics, err := rd.loadBaseline()
	if err != nil {
		return nil, fmt.Errorf("failed to load baseline: %v", err)
	}
	analysis.BaselineMetrics = baselineMetrics

	// Perform comparison analysis
	rd.performComparisonAnalysis(analysis)

	// Perform trend analysis if sufficient history exists
	rd.performTrendAnalysis(analysis)

	// Generate alerts based on analysis
	rd.generateAlerts(analysis)

	// Calculate summary statistics
	rd.calculateSummary(analysis)

	// Save current metrics to history
	rd.saveToHistory(currentMetrics)

	return analysis, nil
}

// performComparisonAnalysis compares current metrics against baseline
func (rd *RegressionDetector) performComparisonAnalysis(analysis *RegressionAnalysis) {
	for name, current := range analysis.CurrentMetrics {
		baseline, exists := analysis.BaselineMetrics[name]
		if !exists {
			// New benchmark - no comparison possible
			continue
		}

		result := rd.analyzeIndividualBenchmark(name, baseline, current)
		analysis.Comparisons = append(analysis.Comparisons, result)
	}

	// Sort by severity and performance impact
	sort.Slice(analysis.Comparisons, func(i, j int) bool {
		a, b := analysis.Comparisons[i], analysis.Comparisons[j]

		// Critical benchmarks first
		if a.IsCriticalBenchmark != b.IsCriticalBenchmark {
			return a.IsCriticalBenchmark
		}

		// Then by severity
		severityOrder := map[string]int{"critical": 0, "major": 1, "minor": 2, "negligible": 3}
		if severityOrder[a.Severity] != severityOrder[b.Severity] {
			return severityOrder[a.Severity] < severityOrder[b.Severity]
		}

		// Finally by performance impact
		return math.Abs(a.PerformanceRatio-1) > math.Abs(b.PerformanceRatio-1)
	})
}

// analyzeIndividualBenchmark performs detailed analysis of a single benchmark
func (rd *RegressionDetector) analyzeIndividualBenchmark(name string, baseline, current BenchmarkMetric) RegressionResult {
	var perfRatio, allocRatio, bytesRatio float64

	if baseline.NsPerOp > 0 {
		perfRatio = float64(current.NsPerOp) / float64(baseline.NsPerOp)
	}
	if baseline.AllocsPerOp > 0 {
		allocRatio = float64(current.AllocsPerOp) / float64(baseline.AllocsPerOp)
	}
	if baseline.BytesPerOp > 0 {
		bytesRatio = float64(current.BytesPerOp) / float64(baseline.BytesPerOp)
	}

	status, severity := rd.categorizePerformanceChange(perfRatio)
	isCritical := rd.isCriticalBenchmark(name)

	// Calculate confidence interval (simplified)
	confidenceInterval := rd.calculateConfidenceInterval(baseline, current)

	return RegressionResult{
		BenchmarkName:           name,
		BaselineNsPerOp:         baseline.NsPerOp,
		CurrentNsPerOp:          current.NsPerOp,
		PerformanceRatio:        perfRatio,
		AllocRatio:              allocRatio,
		BytesRatio:              bytesRatio,
		Status:                  status,
		Severity:                severity,
		IsCriticalBenchmark:     isCritical,
		StatisticalSignificance: rd.isStatisticallySignificant(baseline, current),
		ConfidenceInterval:      confidenceInterval,
	}
}

// categorizePerformanceChange determines status and severity
func (rd *RegressionDetector) categorizePerformanceChange(ratio float64) (status, severity string) {
	if ratio == 0 {
		return "unknown", "negligible"
	}

	switch {
	case ratio <= rd.config.MajorImprovementThreshold:
		return "improved", "major"
	case ratio <= rd.config.MinorImprovementThreshold:
		return "improved", "minor"
	case ratio >= rd.config.CriticalDegradationThreshold:
		return "degraded", "critical"
	case ratio >= rd.config.MajorDegradationThreshold:
		return "degraded", "major"
	case ratio >= rd.config.MinorDegradationThreshold:
		return "degraded", "minor"
	default:
		return "stable", "negligible"
	}
}

// isCriticalBenchmark checks if a benchmark is marked as critical
func (rd *RegressionDetector) isCriticalBenchmark(name string) bool {
	for _, critical := range rd.config.CriticalBenchmarks {
		if name == critical {
			return true
		}
	}
	return false
}

// calculateConfidenceInterval calculates a simplified confidence interval
func (rd *RegressionDetector) calculateConfidenceInterval(baseline, current BenchmarkMetric) [2]float64 {
	// Simplified calculation - in practice, would use proper statistical methods
	ratio := float64(current.NsPerOp) / float64(baseline.NsPerOp)
	variance := rd.estimateVariance(baseline, current)
	margin := 1.96 * math.Sqrt(variance) // 95% confidence interval

	return [2]float64{ratio - margin, ratio + margin}
}

// estimateVariance estimates variance from limited data
func (rd *RegressionDetector) estimateVariance(baseline, current BenchmarkMetric) float64 {
	// Simplified variance estimation
	// In practice, would use historical data or multiple runs
	return 0.01 // 1% variance assumption
}

// isStatisticallySignificant determines if the difference is statistically significant
func (rd *RegressionDetector) isStatisticallySignificant(baseline, current BenchmarkMetric) bool {
	// Simplified significance test
	// In practice, would use proper statistical tests (t-test, etc.)
	ratio := float64(current.NsPerOp) / float64(baseline.NsPerOp)
	return math.Abs(ratio-1) > 0.05 // 5% threshold
}

// performTrendAnalysis analyzes performance trends over time
func (rd *RegressionDetector) performTrendAnalysis(analysis *RegressionAnalysis) {
	for name := range analysis.CurrentMetrics {
		history, err := rd.loadBenchmarkHistory(name)
		if err != nil || len(history) < rd.config.MinSampleSize {
			continue
		}

		trend := rd.analyzeTrend(name, history)
		analysis.TrendAnalysis = append(analysis.TrendAnalysis, trend)
	}
}

// analyzeTrend performs trend analysis on historical data
func (rd *RegressionDetector) analyzeTrend(name string, dataPoints []float64) TrendAnalysis {
	n := len(dataPoints)
	if n < 2 {
		return TrendAnalysis{BenchmarkName: name}
	}

	// Calculate basic statistics
	mean := rd.calculateMean(dataPoints)
	variance := rd.calculateVariance(dataPoints, mean)
	stdDev := math.Sqrt(variance)

	// Calculate trend using linear regression
	trendStrength := rd.calculateTrendStrength(dataPoints)

	var trendDirection string
	if trendStrength > 0.1 {
		trendDirection = "degrading"
	} else if trendStrength < -0.1 {
		trendDirection = "improving"
	} else {
		trendDirection = "stable"
	}

	isVolatile := variance > rd.config.VarianceThreshold

	return TrendAnalysis{
		BenchmarkName:     name,
		DataPoints:        dataPoints,
		TrendDirection:    trendDirection,
		TrendStrength:     trendStrength,
		Variance:          variance,
		Mean:              mean,
		StandardDeviation: stdDev,
		IsVolatile:        isVolatile,
	}
}

// calculateMean calculates the arithmetic mean
func (rd *RegressionDetector) calculateMean(data []float64) float64 {
	sum := 0.0
	for _, v := range data {
		sum += v
	}
	return sum / float64(len(data))
}

// calculateVariance calculates the variance
func (rd *RegressionDetector) calculateVariance(data []float64, mean float64) float64 {
	sum := 0.0
	for _, v := range data {
		diff := v - mean
		sum += diff * diff
	}
	return sum / float64(len(data)-1)
}

// calculateTrendStrength calculates trend strength using correlation
func (rd *RegressionDetector) calculateTrendStrength(data []float64) float64 {
	n := len(data)
	if n < 2 {
		return 0
	}

	// Simple linear correlation with time
	var sumX, sumY, sumXY, sumX2, sumY2 float64

	for i, y := range data {
		x := float64(i)
		sumX += x
		sumY += y
		sumXY += x * y
		sumX2 += x * x
		sumY2 += y * y
	}

	nf := float64(n)
	numerator := nf*sumXY - sumX*sumY
	denominator := math.Sqrt((nf*sumX2 - sumX*sumX) * (nf*sumY2 - sumY*sumY))

	if denominator == 0 {
		return 0
	}

	return numerator / denominator
}

// generateAlerts creates alerts based on analysis results
func (rd *RegressionDetector) generateAlerts(analysis *RegressionAnalysis) {
	for _, result := range analysis.Comparisons {
		rd.checkPerformanceAlerts(analysis, result)
	}

	for _, trend := range analysis.TrendAnalysis {
		rd.checkTrendAlerts(analysis, trend)
	}
}

// checkPerformanceAlerts checks for performance-related alerts
func (rd *RegressionDetector) checkPerformanceAlerts(analysis *RegressionAnalysis, result RegressionResult) {
	shouldAlert := false

	switch result.Severity {
	case "critical":
		shouldAlert = rd.config.AlertOnCriticalDegradation
	case "major":
		shouldAlert = rd.config.AlertOnMajorDegradation
	case "minor":
		shouldAlert = rd.config.AlertOnMinorDegradation
	}

	if shouldAlert && result.Status == "degraded" {
		alert := RegressionAlert{
			Type:          "performance",
			Severity:      result.Severity,
			BenchmarkName: result.BenchmarkName,
			Message: fmt.Sprintf("Performance regression detected: %s is %.1f%% slower than baseline",
				result.BenchmarkName, (result.PerformanceRatio-1)*100),
			Timestamp:      analysis.Timestamp,
			ActionRequired: result.IsCriticalBenchmark || result.Severity == "critical",
		}
		analysis.Alerts = append(analysis.Alerts, alert)
	}
}

// checkTrendAlerts checks for trend-related alerts
func (rd *RegressionDetector) checkTrendAlerts(analysis *RegressionAnalysis, trend TrendAnalysis) {
	if rd.config.AlertOnHighVariance && trend.IsVolatile {
		alert := RegressionAlert{
			Type:          "variance",
			Severity:      "minor",
			BenchmarkName: trend.BenchmarkName,
			Message: fmt.Sprintf("High variance detected in %s: %.1f%% variance",
				trend.BenchmarkName, trend.Variance*100),
			Timestamp:      analysis.Timestamp,
			ActionRequired: false,
		}
		analysis.Alerts = append(analysis.Alerts, alert)
	}

	if trend.TrendDirection == "degrading" && math.Abs(trend.TrendStrength) > 0.5 {
		alert := RegressionAlert{
			Type:          "trend",
			Severity:      "major",
			BenchmarkName: trend.BenchmarkName,
			Message: fmt.Sprintf("Degrading performance trend detected in %s (strength: %.2f)",
				trend.BenchmarkName, trend.TrendStrength),
			Timestamp:      analysis.Timestamp,
			ActionRequired: rd.isCriticalBenchmark(trend.BenchmarkName),
		}
		analysis.Alerts = append(analysis.Alerts, alert)
	}
}

// calculateSummary generates summary statistics
func (rd *RegressionDetector) calculateSummary(analysis *RegressionAnalysis) {
	summary := RegressionSummary{
		TotalBenchmarks: len(analysis.Comparisons),
		AlertsGenerated: len(analysis.Alerts),
	}

	for _, result := range analysis.Comparisons {
		switch result.Status {
		case "improved":
			summary.ImprovedBenchmarks++
		case "degraded":
			summary.DegradedBenchmarks++
			switch result.Severity {
			case "critical":
				summary.CriticalRegressions++
			case "major":
				summary.MajorRegressions++
			case "minor":
				summary.MinorRegressions++
			}
		case "stable":
			summary.StableBenchmarks++
		}
	}

	for _, trend := range analysis.TrendAnalysis {
		if trend.IsVolatile {
			summary.HighVarianceBenchmarks++
		}
	}

	analysis.Summary = summary
}

// loadBaseline loads baseline metrics from file
func (rd *RegressionDetector) loadBaseline() (map[string]BenchmarkMetric, error) {
	data, err := os.ReadFile(rd.baselineFile)
	if err != nil {
		return nil, err
	}

	var baseline BaselineData
	if err := json.Unmarshal(data, &baseline); err != nil {
		return nil, err
	}

	return baseline.Benchmarks, nil
}

// loadBenchmarkHistory loads historical data for a specific benchmark
func (rd *RegressionDetector) loadBenchmarkHistory(benchmarkName string) ([]float64, error) {
	historyFile := filepath.Join(rd.historyDir, benchmarkName+".json")

	data, err := os.ReadFile(historyFile)
	if err != nil {
		return nil, err
	}

	var history []float64
	if err := json.Unmarshal(data, &history); err != nil {
		return nil, err
	}

	// Return only the most recent data points within the analysis window
	if len(history) > rd.config.TrendAnalysisWindow {
		return history[len(history)-rd.config.TrendAnalysisWindow:], nil
	}

	return history, nil
}

// saveToHistory saves current metrics to historical data
func (rd *RegressionDetector) saveToHistory(metrics map[string]BenchmarkMetric) error {
	for name, metric := range metrics {
		historyFile := filepath.Join(rd.historyDir, name+".json")

		// Load existing history
		var history []float64
		if data, err := os.ReadFile(historyFile); err == nil {
			json.Unmarshal(data, &history)
		}

		// Append new data point
		history = append(history, float64(metric.NsPerOp))

		// Keep only recent data points to prevent unbounded growth
		maxHistory := rd.config.TrendAnalysisWindow * 2
		if len(history) > maxHistory {
			history = history[len(history)-maxHistory:]
		}

		// Save updated history
		data, err := json.MarshalIndent(history, "", "  ")
		if err != nil {
			continue // Skip this benchmark on error
		}

		os.WriteFile(historyFile, data, 0644)
	}

	return nil
}
