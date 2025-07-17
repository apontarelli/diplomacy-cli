package pipeline

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// BenchmarkRunner orchestrates automated benchmark execution and reporting
type BenchmarkRunner struct {
	projectRoot string
	resultsDir  string
	reportsDir  string
	comparator  *BenchmarkComparator
	config      BenchmarkConfig
}

// BenchmarkConfig holds configuration for benchmark execution
type BenchmarkConfig struct {
	BenchTime    string   `json:"bench_time"`    // e.g., "30s"
	Count        int      `json:"count"`         // Number of runs
	CPUProfile   bool     `json:"cpu_profile"`   // Enable CPU profiling
	MemProfile   bool     `json:"mem_profile"`   // Enable memory profiling
	TraceProfile bool     `json:"trace_profile"` // Enable trace profiling
	Timeout      string   `json:"timeout"`       // Overall timeout
	Patterns     []string `json:"patterns"`      // Benchmark patterns to run
	SkipBaseline bool     `json:"skip_baseline"` // Skip baseline comparison
	SaveBaseline bool     `json:"save_baseline"` // Save results as new baseline
	Verbose      bool     `json:"verbose"`       // Verbose output
}

// DefaultBenchmarkConfig returns a sensible default configuration
func DefaultBenchmarkConfig() BenchmarkConfig {
	return BenchmarkConfig{
		BenchTime:    "10s",
		Count:        3,
		CPUProfile:   true,
		MemProfile:   true,
		TraceProfile: false,
		Timeout:      "10m",
		Patterns: []string{
			"BenchmarkBaseline",
			"BenchmarkDATCByCategory",
			"BenchmarkComplexityScenarios",
			"BenchmarkResolutionEngineComponents",
		},
		SkipBaseline: false,
		SaveBaseline: false,
		Verbose:      false,
	}
}

// NewBenchmarkRunner creates a new benchmark runner
func NewBenchmarkRunner(projectRoot string) *BenchmarkRunner {
	resultsDir := filepath.Join(projectRoot, "benchmarks", "results")
	reportsDir := filepath.Join(projectRoot, "benchmarks", "reports")
	baselineFile := filepath.Join(resultsDir, "baseline.json")

	return &BenchmarkRunner{
		projectRoot: projectRoot,
		resultsDir:  resultsDir,
		reportsDir:  reportsDir,
		comparator:  NewBenchmarkComparator(baselineFile),
		config:      DefaultBenchmarkConfig(),
	}
}

// SetConfig updates the benchmark configuration
func (br *BenchmarkRunner) SetConfig(config BenchmarkConfig) {
	br.config = config
}

// RunBenchmarks executes the full benchmark suite
func (br *BenchmarkRunner) RunBenchmarks(ctx context.Context) (*BenchmarkReport, error) {
	// Ensure directories exist
	if err := br.ensureDirectories(); err != nil {
		return nil, fmt.Errorf("failed to create directories: %v", err)
	}

	report := &BenchmarkReport{
		Timestamp: time.Now(),
		Config:    br.config,
		Results:   make(map[string]BenchmarkMetric),
	}

	// Run each benchmark pattern
	for _, pattern := range br.config.Patterns {
		if br.config.Verbose {
			fmt.Printf("Running benchmark pattern: %s\n", pattern)
		}

		result, err := br.runSingleBenchmark(ctx, pattern)
		if err != nil {
			return nil, fmt.Errorf("failed to run benchmark %s: %v", pattern, err)
		}

		// Parse results and add to report
		metrics, err := ParseBenchmarkOutput(result.Output)
		if err != nil {
			return nil, fmt.Errorf("failed to parse benchmark output for %s: %v", pattern, err)
		}

		for name, metric := range metrics {
			report.Results[name] = metric
		}

		report.ExecutionResults = append(report.ExecutionResults, *result)
	}

	// Compare with baseline if not skipped
	if !br.config.SkipBaseline {
		comparison, err := br.comparator.Compare(report.Results)
		if err != nil {
			if br.config.Verbose {
				fmt.Printf("Warning: baseline comparison failed: %v\n", err)
			}
		} else {
			report.Comparison = comparison
		}
	}

	// Save as new baseline if requested
	if br.config.SaveBaseline {
		if err := br.comparator.SaveBaseline(report.Results); err != nil {
			return nil, fmt.Errorf("failed to save baseline: %v", err)
		}
		if br.config.Verbose {
			fmt.Println("Saved new baseline")
		}
	}

	// Generate and save report
	if err := br.saveReport(report); err != nil {
		return nil, fmt.Errorf("failed to save report: %v", err)
	}

	return report, nil
}

// BenchmarkReport contains the complete results of a benchmark run
type BenchmarkReport struct {
	Timestamp        time.Time                  `json:"timestamp"`
	Config           BenchmarkConfig            `json:"config"`
	Results          map[string]BenchmarkMetric `json:"results"`
	Comparison       []ComparisonResult         `json:"comparison,omitempty"`
	ExecutionResults []BenchmarkExecution       `json:"execution_results"`
	Summary          BenchmarkSummary           `json:"summary"`
}

// BenchmarkExecution contains details about a single benchmark execution
type BenchmarkExecution struct {
	Pattern      string        `json:"pattern"`
	Duration     time.Duration `json:"duration"`
	Output       string        `json:"output"`
	Error        string        `json:"error,omitempty"`
	ProfileFiles []string      `json:"profile_files,omitempty"`
}

// BenchmarkSummary provides high-level statistics
type BenchmarkSummary struct {
	TotalBenchmarks int           `json:"total_benchmarks"`
	TotalDuration   time.Duration `json:"total_duration"`
	FastestBench    string        `json:"fastest_bench"`
	SlowestBench    string        `json:"slowest_bench"`
	HighestAllocs   string        `json:"highest_allocs"`
	LowestAllocs    string        `json:"lowest_allocs"`
}

// runSingleBenchmark executes a single benchmark pattern
func (br *BenchmarkRunner) runSingleBenchmark(ctx context.Context, pattern string) (*BenchmarkExecution, error) {
	start := time.Now()

	// Build command
	args := []string{"test", "-bench=" + pattern, "-benchtime=" + br.config.BenchTime,
		"-count=" + fmt.Sprintf("%d", br.config.Count), "-benchmem"}

	// Add profiling flags
	timestamp := time.Now().Format("20060102_150405")
	var profileFiles []string

	if br.config.CPUProfile {
		cpuFile := filepath.Join(br.resultsDir, fmt.Sprintf("%s_%s_cpu.prof", pattern, timestamp))
		args = append(args, "-cpuprofile="+cpuFile)
		profileFiles = append(profileFiles, cpuFile)
	}

	if br.config.MemProfile {
		memFile := filepath.Join(br.resultsDir, fmt.Sprintf("%s_%s_mem.prof", pattern, timestamp))
		args = append(args, "-memprofile="+memFile)
		profileFiles = append(profileFiles, memFile)
	}

	if br.config.TraceProfile {
		traceFile := filepath.Join(br.resultsDir, fmt.Sprintf("%s_%s_trace.out", pattern, timestamp))
		args = append(args, "-trace="+traceFile)
		profileFiles = append(profileFiles, traceFile)
	}

	// Add package path
	args = append(args, "./internal/game/pipeline/")

	// Create command with timeout
	cmdCtx, cancel := context.WithTimeout(ctx, br.parseTimeout())
	defer cancel()

	cmd := exec.CommandContext(cmdCtx, "go", args...)
	cmd.Dir = br.projectRoot

	// Capture output
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// Execute
	err := cmd.Run()
	duration := time.Since(start)

	result := &BenchmarkExecution{
		Pattern:      pattern,
		Duration:     duration,
		Output:       stdout.String(),
		ProfileFiles: profileFiles,
	}

	if err != nil {
		result.Error = fmt.Sprintf("Command failed: %v\nStderr: %s", err, stderr.String())
		return result, err
	}

	return result, nil
}

// parseTimeout parses the timeout string into a duration
func (br *BenchmarkRunner) parseTimeout() time.Duration {
	timeout, err := time.ParseDuration(br.config.Timeout)
	if err != nil {
		return 10 * time.Minute // Default timeout
	}
	return timeout
}

// ensureDirectories creates necessary directories
func (br *BenchmarkRunner) ensureDirectories() error {
	dirs := []string{br.resultsDir, br.reportsDir}
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}
	return nil
}

// saveReport saves the benchmark report to files
func (br *BenchmarkRunner) saveReport(report *BenchmarkReport) error {
	timestamp := report.Timestamp.Format("20060102_150405")

	// Calculate summary
	report.Summary = br.calculateSummary(report.Results)

	// Save JSON report
	jsonFile := filepath.Join(br.reportsDir, fmt.Sprintf("benchmark_report_%s.json", timestamp))
	if err := br.saveJSONReport(report, jsonFile); err != nil {
		return fmt.Errorf("failed to save JSON report: %v", err)
	}

	// Save markdown report
	mdFile := filepath.Join(br.reportsDir, fmt.Sprintf("benchmark_report_%s.md", timestamp))
	if err := br.saveMarkdownReport(report, mdFile); err != nil {
		return fmt.Errorf("failed to save markdown report: %v", err)
	}

	if br.config.Verbose {
		fmt.Printf("Reports saved:\n- JSON: %s\n- Markdown: %s\n", jsonFile, mdFile)
	}

	return nil
}

// calculateSummary generates summary statistics
func (br *BenchmarkRunner) calculateSummary(results map[string]BenchmarkMetric) BenchmarkSummary {
	summary := BenchmarkSummary{
		TotalBenchmarks: len(results),
	}

	var fastestTime, slowestTime int64 = -1, -1
	var highestAllocs, lowestAllocs int64 = -1, -1

	for name, metric := range results {
		// Track fastest/slowest
		if fastestTime == -1 || metric.NsPerOp < fastestTime {
			fastestTime = metric.NsPerOp
			summary.FastestBench = name
		}
		if slowestTime == -1 || metric.NsPerOp > slowestTime {
			slowestTime = metric.NsPerOp
			summary.SlowestBench = name
		}

		// Track highest/lowest allocations
		if highestAllocs == -1 || metric.AllocsPerOp > highestAllocs {
			highestAllocs = metric.AllocsPerOp
			summary.HighestAllocs = name
		}
		if lowestAllocs == -1 || metric.AllocsPerOp < lowestAllocs {
			lowestAllocs = metric.AllocsPerOp
			summary.LowestAllocs = name
		}
	}

	// Calculate total duration from execution results
	// This would be set by the caller based on actual execution time

	return summary
}

// saveJSONReport saves the report as JSON
func (br *BenchmarkRunner) saveJSONReport(report *BenchmarkReport, filename string) error {
	// Implementation would marshal report to JSON and save
	// For now, just create a placeholder
	return os.WriteFile(filename, []byte("{}"), 0644)
}

// saveMarkdownReport saves the report as Markdown
func (br *BenchmarkRunner) saveMarkdownReport(report *BenchmarkReport, filename string) error {
	var content strings.Builder

	content.WriteString("# Benchmark Report\n\n")
	content.WriteString(fmt.Sprintf("**Generated:** %s\n\n", report.Timestamp.Format(time.RFC3339)))

	// Configuration section
	content.WriteString("## Configuration\n\n")
	content.WriteString(fmt.Sprintf("- **Benchmark Time:** %s\n", report.Config.BenchTime))
	content.WriteString(fmt.Sprintf("- **Count:** %d\n", report.Config.Count))
	content.WriteString(fmt.Sprintf("- **CPU Profiling:** %t\n", report.Config.CPUProfile))
	content.WriteString(fmt.Sprintf("- **Memory Profiling:** %t\n", report.Config.MemProfile))
	content.WriteString("\n")

	// Summary section
	content.WriteString("## Summary\n\n")
	content.WriteString(fmt.Sprintf("- **Total Benchmarks:** %d\n", report.Summary.TotalBenchmarks))
	content.WriteString(fmt.Sprintf("- **Fastest:** %s\n", report.Summary.FastestBench))
	content.WriteString(fmt.Sprintf("- **Slowest:** %s\n", report.Summary.SlowestBench))
	content.WriteString(fmt.Sprintf("- **Highest Allocations:** %s\n", report.Summary.HighestAllocs))
	content.WriteString(fmt.Sprintf("- **Lowest Allocations:** %s\n", report.Summary.LowestAllocs))
	content.WriteString("\n")

	// Results table
	content.WriteString("## Results\n\n")
	content.WriteString("| Benchmark | ns/op | allocs/op | bytes/op | iterations |\n")
	content.WriteString("|-----------|-------|-----------|----------|------------|\n")

	for name, metric := range report.Results {
		content.WriteString(fmt.Sprintf("| %s | %d | %d | %d | %d |\n",
			name, metric.NsPerOp, metric.AllocsPerOp, metric.BytesPerOp, metric.Iterations))
	}

	// Comparison section if available
	if len(report.Comparison) > 0 {
		content.WriteString("\n")
		content.WriteString(br.comparator.GenerateReport(report.Comparison))
	}

	return os.WriteFile(filename, []byte(content.String()), 0644)
}
