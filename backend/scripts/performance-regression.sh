#!/bin/bash

# Performance Regression Detection Script
# This script runs benchmarks and detects performance regressions
# Can be used locally or in CI/CD pipelines

set -euo pipefail

# Configuration
PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
BENCHMARKS_DIR="${PROJECT_ROOT}/benchmarks"
RESULTS_DIR="${BENCHMARKS_DIR}/results"
REPORTS_DIR="${BENCHMARKS_DIR}/reports"
BASELINE_FILE="${RESULTS_DIR}/baseline.json"

# Default values
BENCH_TIME="10s"
BENCH_COUNT=3
SAVE_BASELINE=false
SKIP_BASELINE=false
VERBOSE=false
FAIL_ON_REGRESSION=true
REGRESSION_THRESHOLD=1.20  # 20% degradation threshold

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Usage function
usage() {
    cat << EOF
Usage: $0 [OPTIONS]

Performance regression detection script for the Diplomacy CLI resolution engine.

OPTIONS:
    -t, --bench-time TIME       Benchmark duration (default: ${BENCH_TIME})
    -c, --count COUNT          Number of benchmark runs (default: ${BENCH_COUNT})
    -s, --save-baseline        Save current results as new baseline
    -k, --skip-baseline        Skip baseline comparison
    -v, --verbose              Enable verbose output
    -f, --fail-on-regression   Fail script on performance regression (default: true)
    -r, --threshold RATIO      Regression threshold ratio (default: ${REGRESSION_THRESHOLD})
    -h, --help                 Show this help message

EXAMPLES:
    # Run standard performance regression test
    $0

    # Run with longer benchmark time and save as baseline
    $0 --bench-time 30s --save-baseline

    # Run without failing on regressions (for analysis)
    $0 --fail-on-regression false

    # Run with custom regression threshold (10% degradation)
    $0 --threshold 1.10

EOF
}

# Logging functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Parse command line arguments
parse_args() {
    while [[ $# -gt 0 ]]; do
        case $1 in
            -t|--bench-time)
                BENCH_TIME="$2"
                shift 2
                ;;
            -c|--count)
                BENCH_COUNT="$2"
                shift 2
                ;;
            -s|--save-baseline)
                SAVE_BASELINE=true
                shift
                ;;
            -k|--skip-baseline)
                SKIP_BASELINE=true
                shift
                ;;
            -v|--verbose)
                VERBOSE=true
                shift
                ;;
            -f|--fail-on-regression)
                FAIL_ON_REGRESSION="$2"
                shift 2
                ;;
            -r|--threshold)
                REGRESSION_THRESHOLD="$2"
                shift 2
                ;;
            -h|--help)
                usage
                exit 0
                ;;
            *)
                log_error "Unknown option: $1"
                usage
                exit 1
                ;;
        esac
    done
}

# Setup directories
setup_directories() {
    log_info "Setting up benchmark directories..."
    mkdir -p "${RESULTS_DIR}" "${REPORTS_DIR}"
}

# Check Go installation
check_go() {
    if ! command -v go &> /dev/null; then
        log_error "Go is not installed or not in PATH"
        exit 1
    fi
    
    local go_version
    go_version=$(go version | awk '{print $3}')
    log_info "Using Go version: ${go_version}"
}

# Run benchmarks
run_benchmarks() {
    log_info "Running performance benchmarks..."
    log_info "Configuration: time=${BENCH_TIME}, count=${BENCH_COUNT}"
    
    cd "${PROJECT_ROOT}"
    
    local timestamp
    timestamp=$(date +"%Y%m%d_%H%M%S")
    local results_file="${RESULTS_DIR}/benchmark_${timestamp}.txt"
    local json_file="${RESULTS_DIR}/benchmark_${timestamp}.json"
    
    # Run Go benchmarks
    local bench_cmd="go test -bench=. -benchmem -count=${BENCH_COUNT} -benchtime=${BENCH_TIME}"
    
    if [[ "${VERBOSE}" == "true" ]]; then
        bench_cmd="${bench_cmd} -v"
    fi
    
    # Add profiling for detailed analysis
    local cpu_profile="${RESULTS_DIR}/cpu_${timestamp}.prof"
    local mem_profile="${RESULTS_DIR}/mem_${timestamp}.prof"
    
    bench_cmd="${bench_cmd} -cpuprofile=${cpu_profile} -memprofile=${mem_profile}"
    bench_cmd="${bench_cmd} ./internal/game/pipeline/"
    
    log_info "Executing: ${bench_cmd}"
    
    if ! eval "${bench_cmd}" > "${results_file}" 2>&1; then
        log_error "Benchmark execution failed"
        if [[ "${VERBOSE}" == "true" ]]; then
            cat "${results_file}"
        fi
        exit 1
    fi
    
    log_success "Benchmarks completed successfully"
    
    # Parse results to JSON format
    parse_benchmark_results "${results_file}" "${json_file}"
    
    echo "${json_file}"
}

# Parse benchmark results to JSON
parse_benchmark_results() {
    local results_file="$1"
    local json_file="$2"
    
    log_info "Parsing benchmark results..."
    
    # Create a simple Go program to parse benchmark output
    cat > /tmp/parse_benchmarks.go << 'EOF'
package main

import (
    "bufio"
    "encoding/json"
    "fmt"
    "os"
    "regexp"
    "strconv"
    "strings"
    "time"
)

type BenchmarkResult struct {
    Name        string  `json:"name"`
    NsPerOp     int64   `json:"ns_per_op"`
    AllocsPerOp int64   `json:"allocs_per_op"`
    BytesPerOp  int64   `json:"bytes_per_op"`
    Iterations  int     `json:"iterations"`
}

type BenchmarkData struct {
    Timestamp  time.Time         `json:"timestamp"`
    GoVersion  string           `json:"go_version"`
    Benchmarks []BenchmarkResult `json:"benchmarks"`
}

func main() {
    if len(os.Args) != 3 {
        fmt.Fprintf(os.Stderr, "Usage: %s <input_file> <output_file>\n", os.Args[0])
        os.Exit(1)
    }
    
    inputFile := os.Args[1]
    outputFile := os.Args[2]
    
    file, err := os.Open(inputFile)
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error opening file: %v\n", err)
        os.Exit(1)
    }
    defer file.Close()
    
    var benchmarks []BenchmarkResult
    benchmarkRegex := regexp.MustCompile(`^Benchmark(\w+)\s+(\d+)\s+(\d+(?:\.\d+)?)\s+ns/op(?:\s+(\d+)\s+allocs/op)?(?:\s+(\d+)\s+B/op)?`)
    
    scanner := bufio.NewScanner(file)
    for scanner.Scan() {
        line := scanner.Text()
        matches := benchmarkRegex.FindStringSubmatch(line)
        
        if len(matches) >= 4 {
            name := "Benchmark" + matches[1]
            iterations, _ := strconv.Atoi(matches[2])
            nsPerOp, _ := strconv.ParseFloat(matches[3], 64)
            
            result := BenchmarkResult{
                Name:       name,
                NsPerOp:    int64(nsPerOp),
                Iterations: iterations,
            }
            
            if len(matches) > 4 && matches[4] != "" {
                result.AllocsPerOp, _ = strconv.ParseInt(matches[4], 10, 64)
            }
            
            if len(matches) > 5 && matches[5] != "" {
                result.BytesPerOp, _ = strconv.ParseInt(matches[5], 10, 64)
            }
            
            benchmarks = append(benchmarks, result)
        }
    }
    
    data := BenchmarkData{
        Timestamp:  time.Now(),
        GoVersion:  "unknown", // Could be extracted from go version
        Benchmarks: benchmarks,
    }
    
    jsonData, err := json.MarshalIndent(data, "", "  ")
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error marshaling JSON: %v\n", err)
        os.Exit(1)
    }
    
    if err := os.WriteFile(outputFile, jsonData, 0644); err != nil {
        fmt.Fprintf(os.Stderr, "Error writing output file: %v\n", err)
        os.Exit(1)
    }
    
    fmt.Printf("Parsed %d benchmarks\n", len(benchmarks))
}
EOF
    
    go run /tmp/parse_benchmarks.go "${results_file}" "${json_file}"
    rm /tmp/parse_benchmarks.go
}

# Compare with baseline
compare_with_baseline() {
    local current_results="$1"
    
    if [[ "${SKIP_BASELINE}" == "true" ]]; then
        log_info "Skipping baseline comparison"
        return 0
    fi
    
    if [[ ! -f "${BASELINE_FILE}" ]]; then
        log_warning "No baseline file found at ${BASELINE_FILE}"
        log_info "Run with --save-baseline to create initial baseline"
        return 0
    fi
    
    log_info "Comparing with baseline..."
    
    # Create comparison script
    cat > /tmp/compare_benchmarks.go << 'EOF'
package main

import (
    "encoding/json"
    "fmt"
    "os"
    "strconv"
)

type BenchmarkResult struct {
    Name        string  `json:"name"`
    NsPerOp     int64   `json:"ns_per_op"`
    AllocsPerOp int64   `json:"allocs_per_op"`
    BytesPerOp  int64   `json:"bytes_per_op"`
    Iterations  int     `json:"iterations"`
}

type BenchmarkData struct {
    Benchmarks []BenchmarkResult `json:"benchmarks"`
}

type ComparisonResult struct {
    Name             string  `json:"name"`
    BaselineNsPerOp  int64   `json:"baseline_ns_per_op"`
    CurrentNsPerOp   int64   `json:"current_ns_per_op"`
    Ratio            float64 `json:"ratio"`
    Status           string  `json:"status"`
}

func main() {
    if len(os.Args) != 4 {
        fmt.Fprintf(os.Stderr, "Usage: %s <baseline_file> <current_file> <threshold>\n", os.Args[0])
        os.Exit(1)
    }
    
    baselineFile := os.Args[1]
    currentFile := os.Args[2]
    threshold, _ := strconv.ParseFloat(os.Args[3], 64)
    
    // Load baseline
    baselineData, err := os.ReadFile(baselineFile)
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error reading baseline: %v\n", err)
        os.Exit(1)
    }
    
    var baseline BenchmarkData
    if err := json.Unmarshal(baselineData, &baseline); err != nil {
        fmt.Fprintf(os.Stderr, "Error parsing baseline: %v\n", err)
        os.Exit(1)
    }
    
    // Load current results
    currentData, err := os.ReadFile(currentFile)
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error reading current results: %v\n", err)
        os.Exit(1)
    }
    
    var current BenchmarkData
    if err := json.Unmarshal(currentData, &current); err != nil {
        fmt.Fprintf(os.Stderr, "Error parsing current results: %v\n", err)
        os.Exit(1)
    }
    
    // Create lookup map for baseline
    baselineMap := make(map[string]BenchmarkResult)
    for _, b := range baseline.Benchmarks {
        baselineMap[b.Name] = b
    }
    
    var comparisons []ComparisonResult
    var regressions int
    
    fmt.Printf("%-40s %15s %15s %10s %s\n", "Benchmark", "Baseline (ns)", "Current (ns)", "Ratio", "Status")
    fmt.Printf("%s\n", strings.Repeat("-", 90))
    
    for _, curr := range current.Benchmarks {
        base, exists := baselineMap[curr.Name]
        if !exists {
            fmt.Printf("%-40s %15s %15d %10s %s\n", curr.Name, "N/A", curr.NsPerOp, "N/A", "NEW")
            continue
        }
        
        ratio := float64(curr.NsPerOp) / float64(base.NsPerOp)
        var status string
        
        if ratio >= threshold {
            status = "REGRESSION"
            regressions++
        } else if ratio <= 0.95 {
            status = "IMPROVEMENT"
        } else {
            status = "STABLE"
        }
        
        fmt.Printf("%-40s %15d %15d %10.2f %s\n", curr.Name, base.NsPerOp, curr.NsPerOp, ratio, status)
        
        comparisons = append(comparisons, ComparisonResult{
            Name:            curr.Name,
            BaselineNsPerOp: base.NsPerOp,
            CurrentNsPerOp:  curr.NsPerOp,
            Ratio:           ratio,
            Status:          status,
        })
    }
    
    fmt.Printf("\nSummary: %d regressions detected\n", regressions)
    
    if regressions > 0 {
        os.Exit(1)
    }
}
EOF
    
    if go run /tmp/compare_benchmarks.go "${BASELINE_FILE}" "${current_results}" "${REGRESSION_THRESHOLD}"; then
        log_success "No performance regressions detected"
        rm /tmp/compare_benchmarks.go
        return 0
    else
        log_error "Performance regressions detected!"
        rm /tmp/compare_benchmarks.go
        return 1
    fi
}

# Save baseline
save_baseline() {
    local results_file="$1"
    
    if [[ "${SAVE_BASELINE}" == "true" ]]; then
        log_info "Saving new baseline..."
        cp "${results_file}" "${BASELINE_FILE}"
        log_success "Baseline saved to ${BASELINE_FILE}"
    fi
}

# Generate report
generate_report() {
    local results_file="$1"
    local timestamp
    timestamp=$(date +"%Y%m%d_%H%M%S")
    local report_file="${REPORTS_DIR}/performance_report_${timestamp}.md"
    
    log_info "Generating performance report..."
    
    cat > "${report_file}" << EOF
# Performance Regression Report

**Generated:** $(date -Iseconds)
**Configuration:** time=${BENCH_TIME}, count=${BENCH_COUNT}
**Threshold:** ${REGRESSION_THRESHOLD}x ($(echo "scale=1; (${REGRESSION_THRESHOLD} - 1) * 100" | bc)% degradation)

## Results

EOF
    
    if [[ -f "${results_file}" ]]; then
        echo '```' >> "${report_file}"
        cat "${results_file}" >> "${report_file}"
        echo '```' >> "${report_file}"
    fi
    
    log_success "Report generated: ${report_file}"
}

# Main execution
main() {
    log_info "Starting performance regression detection..."
    
    parse_args "$@"
    setup_directories
    check_go
    
    local results_file
    results_file=$(run_benchmarks)
    
    local regression_detected=false
    if ! compare_with_baseline "${results_file}"; then
        regression_detected=true
    fi
    
    save_baseline "${results_file}"
    generate_report "${results_file}"
    
    if [[ "${regression_detected}" == "true" && "${FAIL_ON_REGRESSION}" == "true" ]]; then
        log_error "Performance regression detected - failing script"
        exit 1
    fi
    
    log_success "Performance regression detection completed successfully"
}

# Run main function with all arguments
main "$@"