package pipeline

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
)

// BaselineData represents stored baseline performance metrics
type BaselineData struct {
	Timestamp  time.Time                  `json:"timestamp"`
	GoVersion  string                     `json:"go_version"`
	System     string                     `json:"system"`
	Benchmarks map[string]BenchmarkMetric `json:"benchmarks"`
	GitCommit  string                     `json:"git_commit,omitempty"`
	BuildInfo  string                     `json:"build_info,omitempty"`
}

// BenchmarkMetric represents a single benchmark's performance metrics
type BenchmarkMetric struct {
	Name        string  `json:"name"`
	NsPerOp     int64   `json:"ns_per_op"`
	AllocsPerOp int64   `json:"allocs_per_op"`
	BytesPerOp  int64   `json:"bytes_per_op"`
	Iterations  int     `json:"iterations"`
	Variance    float64 `json:"variance,omitempty"`
}

// ComparisonResult represents the result of comparing current vs baseline
type ComparisonResult struct {
	BenchmarkName    string  `json:"benchmark_name"`
	BaselineNsPerOp  int64   `json:"baseline_ns_per_op"`
	CurrentNsPerOp   int64   `json:"current_ns_per_op"`
	PerformanceRatio float64 `json:"performance_ratio"` // current/baseline (lower is better)
	AllocRatio       float64 `json:"alloc_ratio"`
	BytesRatio       float64 `json:"bytes_ratio"`
	Status           string  `json:"status"`       // "improved", "degraded", "stable"
	Significance     string  `json:"significance"` // "major", "minor", "negligible"
}

// BenchmarkComparator handles baseline comparison logic
type BenchmarkComparator struct {
	baselineFile string
	thresholds   ComparisonThresholds
}

// ComparisonThresholds defines what constitutes significant changes
type ComparisonThresholds struct {
	MinorImprovement float64 // e.g., 0.95 (5% improvement)
	MajorImprovement float64 // e.g., 0.80 (20% improvement)
	MinorDegradation float64 // e.g., 1.05 (5% degradation)
	MajorDegradation float64 // e.g., 1.20 (20% degradation)
}

// NewBenchmarkComparator creates a new comparator with default thresholds
func NewBenchmarkComparator(baselineFile string) *BenchmarkComparator {
	return &BenchmarkComparator{
		baselineFile: baselineFile,
		thresholds: ComparisonThresholds{
			MinorImprovement: 0.95,
			MajorImprovement: 0.80,
			MinorDegradation: 1.05,
			MajorDegradation: 1.20,
		},
	}
}

// SaveBaseline saves current benchmark results as the new baseline
func (bc *BenchmarkComparator) SaveBaseline(metrics map[string]BenchmarkMetric) error {
	baseline := BaselineData{
		Timestamp:  time.Now(),
		GoVersion:  getGoVersion(),
		System:     getSystemInfo(),
		Benchmarks: metrics,
		GitCommit:  getGitCommit(),
		BuildInfo:  getBuildInfo(),
	}

	// Ensure directory exists
	dir := filepath.Dir(bc.baselineFile)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create baseline directory: %v", err)
	}

	data, err := json.MarshalIndent(baseline, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal baseline data: %v", err)
	}

	if err := os.WriteFile(bc.baselineFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write baseline file: %v", err)
	}

	return nil
}

// LoadBaseline loads the stored baseline data
func (bc *BenchmarkComparator) LoadBaseline() (*BaselineData, error) {
	data, err := os.ReadFile(bc.baselineFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("baseline file not found: %s", bc.baselineFile)
		}
		return nil, fmt.Errorf("failed to read baseline file: %v", err)
	}

	var baseline BaselineData
	if err := json.Unmarshal(data, &baseline); err != nil {
		return nil, fmt.Errorf("failed to unmarshal baseline data: %v", err)
	}

	return &baseline, nil
}

// Compare compares current metrics against baseline
func (bc *BenchmarkComparator) Compare(currentMetrics map[string]BenchmarkMetric) ([]ComparisonResult, error) {
	baseline, err := bc.LoadBaseline()
	if err != nil {
		return nil, err
	}

	var results []ComparisonResult

	// Compare each current benchmark against baseline
	for name, current := range currentMetrics {
		baselineMetric, exists := baseline.Benchmarks[name]
		if !exists {
			// New benchmark - no comparison possible
			results = append(results, ComparisonResult{
				BenchmarkName:    name,
				BaselineNsPerOp:  0,
				CurrentNsPerOp:   current.NsPerOp,
				PerformanceRatio: 0,
				AllocRatio:       0,
				BytesRatio:       0,
				Status:           "new",
				Significance:     "n/a",
			})
			continue
		}

		result := bc.compareMetrics(name, baselineMetric, current)
		results = append(results, result)
	}

	// Check for removed benchmarks
	for name, baseline := range baseline.Benchmarks {
		if _, exists := currentMetrics[name]; !exists {
			results = append(results, ComparisonResult{
				BenchmarkName:    name,
				BaselineNsPerOp:  baseline.NsPerOp,
				CurrentNsPerOp:   0,
				PerformanceRatio: 0,
				AllocRatio:       0,
				BytesRatio:       0,
				Status:           "removed",
				Significance:     "n/a",
			})
		}
	}

	// Sort results by benchmark name for consistent output
	sort.Slice(results, func(i, j int) bool {
		return results[i].BenchmarkName < results[j].BenchmarkName
	})

	return results, nil
}

// compareMetrics compares individual benchmark metrics
func (bc *BenchmarkComparator) compareMetrics(name string, baseline, current BenchmarkMetric) ComparisonResult {
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

	status, significance := bc.categorizeChange(perfRatio)

	return ComparisonResult{
		BenchmarkName:    name,
		BaselineNsPerOp:  baseline.NsPerOp,
		CurrentNsPerOp:   current.NsPerOp,
		PerformanceRatio: perfRatio,
		AllocRatio:       allocRatio,
		BytesRatio:       bytesRatio,
		Status:           status,
		Significance:     significance,
	}
}

// categorizeChange determines the status and significance of a performance change
func (bc *BenchmarkComparator) categorizeChange(ratio float64) (status, significance string) {
	if ratio == 0 {
		return "unknown", "n/a"
	}

	switch {
	case ratio <= bc.thresholds.MajorImprovement:
		return "improved", "major"
	case ratio <= bc.thresholds.MinorImprovement:
		return "improved", "minor"
	case ratio >= bc.thresholds.MajorDegradation:
		return "degraded", "major"
	case ratio >= bc.thresholds.MinorDegradation:
		return "degraded", "minor"
	default:
		return "stable", "negligible"
	}
}

// GenerateReport generates a human-readable comparison report
func (bc *BenchmarkComparator) GenerateReport(results []ComparisonResult) string {
	var report strings.Builder

	report.WriteString("# Benchmark Comparison Report\n\n")
	report.WriteString(fmt.Sprintf("Generated: %s\n\n", time.Now().Format(time.RFC3339)))

	// Summary statistics
	var improved, degraded, stable, new, removed int
	var majorChanges []ComparisonResult

	for _, result := range results {
		switch result.Status {
		case "improved":
			improved++
		case "degraded":
			degraded++
		case "stable":
			stable++
		case "new":
			new++
		case "removed":
			removed++
		}

		if result.Significance == "major" {
			majorChanges = append(majorChanges, result)
		}
	}

	report.WriteString("## Summary\n\n")
	report.WriteString(fmt.Sprintf("- **Improved**: %d benchmarks\n", improved))
	report.WriteString(fmt.Sprintf("- **Degraded**: %d benchmarks\n", degraded))
	report.WriteString(fmt.Sprintf("- **Stable**: %d benchmarks\n", stable))
	report.WriteString(fmt.Sprintf("- **New**: %d benchmarks\n", new))
	report.WriteString(fmt.Sprintf("- **Removed**: %d benchmarks\n\n", removed))

	// Major changes section
	if len(majorChanges) > 0 {
		report.WriteString("## ⚠️ Major Changes\n\n")
		for _, result := range majorChanges {
			report.WriteString(fmt.Sprintf("### %s\n", result.BenchmarkName))
			report.WriteString(fmt.Sprintf("- **Status**: %s (%s)\n", result.Status, result.Significance))
			report.WriteString(fmt.Sprintf("- **Performance**: %.2fx (%.1f%% change)\n",
				result.PerformanceRatio, (result.PerformanceRatio-1)*100))
			if result.AllocRatio > 0 {
				report.WriteString(fmt.Sprintf("- **Allocations**: %.2fx\n", result.AllocRatio))
			}
			if result.BytesRatio > 0 {
				report.WriteString(fmt.Sprintf("- **Memory**: %.2fx\n", result.BytesRatio))
			}
			report.WriteString("\n")
		}
	}

	// Detailed results table
	report.WriteString("## Detailed Results\n\n")
	report.WriteString("| Benchmark | Status | Performance | Allocations | Memory | Change |\n")
	report.WriteString("|-----------|--------|-------------|-------------|--------|---------|\n")

	for _, result := range results {
		var changeStr string
		if result.PerformanceRatio > 0 {
			changeStr = fmt.Sprintf("%.1f%%", (result.PerformanceRatio-1)*100)
		} else {
			changeStr = "N/A"
		}

		var allocStr, memStr string
		if result.AllocRatio > 0 {
			allocStr = fmt.Sprintf("%.2fx", result.AllocRatio)
		} else {
			allocStr = "N/A"
		}
		if result.BytesRatio > 0 {
			memStr = fmt.Sprintf("%.2fx", result.BytesRatio)
		} else {
			memStr = "N/A"
		}

		statusEmoji := getStatusEmoji(result.Status)
		report.WriteString(fmt.Sprintf("| %s | %s %s | %.2fx | %s | %s | %s |\n",
			result.BenchmarkName, statusEmoji, result.Status, result.PerformanceRatio,
			allocStr, memStr, changeStr))
	}

	return report.String()
}

// getStatusEmoji returns an emoji for the given status
func getStatusEmoji(status string) string {
	switch status {
	case "improved":
		return "✅"
	case "degraded":
		return "❌"
	case "stable":
		return "➖"
	case "new":
		return "🆕"
	case "removed":
		return "🗑️"
	default:
		return "❓"
	}
}

// Helper functions for system information
func getGoVersion() string {
	// This would typically use runtime.Version() but for testing we'll use a placeholder
	return "go1.21.0" // TODO: Use runtime.Version()
}

func getSystemInfo() string {
	// This would typically use runtime.GOOS + runtime.GOARCH
	return "darwin/arm64" // TODO: Use runtime.GOOS + "/" + runtime.GOARCH
}

func getGitCommit() string {
	// TODO: Execute git command to get current commit
	return "unknown"
}

func getBuildInfo() string {
	// TODO: Get build information
	return "development"
}

// ParseBenchmarkOutput parses Go benchmark output into metrics
func ParseBenchmarkOutput(output string) (map[string]BenchmarkMetric, error) {
	metrics := make(map[string]BenchmarkMetric)
	lines := strings.Split(output, "\n")

	for _, line := range lines {
		if !strings.HasPrefix(line, "Benchmark") {
			continue
		}

		parts := strings.Fields(line)
		if len(parts) < 4 {
			continue
		}

		name := parts[0]
		iterations, err := strconv.Atoi(parts[1])
		if err != nil {
			continue
		}

		nsPerOp, err := strconv.ParseInt(strings.TrimSuffix(parts[2], ".0"), 10, 64)
		if err != nil {
			continue
		}

		metric := BenchmarkMetric{
			Name:       name,
			NsPerOp:    nsPerOp,
			Iterations: iterations,
		}

		// Parse additional metrics if present (allocs/op, B/op)
		for i := 3; i < len(parts); i++ {
			part := parts[i]
			if strings.HasSuffix(part, "allocs/op") {
				if allocs, err := strconv.ParseInt(strings.TrimSuffix(part, "allocs/op"), 10, 64); err == nil {
					metric.AllocsPerOp = allocs
				}
			} else if strings.HasSuffix(part, "B/op") {
				if bytes, err := strconv.ParseInt(strings.TrimSuffix(part, "B/op"), 10, 64); err == nil {
					metric.BytesPerOp = bytes
				}
			}
		}

		metrics[name] = metric
	}

	return metrics, nil
}
