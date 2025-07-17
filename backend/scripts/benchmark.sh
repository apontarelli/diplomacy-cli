#!/bin/bash

# Enhanced Benchmarking Script for Diplomacy CLI
# This script runs comprehensive benchmarks and generates detailed reports

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
BENCHMARK_DIR="$PROJECT_ROOT/benchmarks"
RESULTS_DIR="$BENCHMARK_DIR/results"
REPORTS_DIR="$BENCHMARK_DIR/reports"

# Create directories if they don't exist
mkdir -p "$RESULTS_DIR"
mkdir -p "$REPORTS_DIR"

# Configuration
BENCHMARK_TIME="30s"
BENCHMARK_COUNT=5
CPU_PROFILE=true
MEM_PROFILE=true
TRACE_PROFILE=false

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

log() {
    echo -e "${BLUE}[$(date +'%Y-%m-%d %H:%M:%S')]${NC} $1"
}

error() {
    echo -e "${RED}[ERROR]${NC} $1" >&2
}

success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

warn() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

# Function to run benchmarks with profiling
run_benchmark() {
    local test_name="$1"
    local output_prefix="$2"
    local benchmark_pattern="$3"
    
    log "Running benchmark: $test_name"
    
    cd "$PROJECT_ROOT"
    
    # Base benchmark command
    local cmd="go test -bench=$benchmark_pattern -benchtime=$BENCHMARK_TIME -count=$BENCHMARK_COUNT -benchmem"
    
    # Add profiling flags if enabled
    if [ "$CPU_PROFILE" = true ]; then
        cmd="$cmd -cpuprofile=$RESULTS_DIR/${output_prefix}_cpu.prof"
    fi
    
    if [ "$MEM_PROFILE" = true ]; then
        cmd="$cmd -memprofile=$RESULTS_DIR/${output_prefix}_mem.prof"
    fi
    
    if [ "$TRACE_PROFILE" = true ]; then
        cmd="$cmd -trace=$RESULTS_DIR/${output_prefix}_trace.out"
    fi
    
    # Run the benchmark and save results
    $cmd ./internal/game/pipeline/ > "$RESULTS_DIR/${output_prefix}_results.txt" 2>&1
    
    if [ $? -eq 0 ]; then
        success "Completed benchmark: $test_name"
    else
        error "Failed benchmark: $test_name"
        return 1
    fi
}

# Function to generate benchmark report
generate_report() {
    local timestamp=$(date +'%Y%m%d_%H%M%S')
    local report_file="$REPORTS_DIR/benchmark_report_$timestamp.md"
    
    log "Generating benchmark report: $report_file"
    
    cat > "$report_file" << EOF
# Diplomacy CLI Benchmark Report

**Generated:** $(date)
**Go Version:** $(go version)
**System:** $(uname -a)

## Configuration
- Benchmark Time: $BENCHMARK_TIME
- Benchmark Count: $BENCHMARK_COUNT
- CPU Profiling: $CPU_PROFILE
- Memory Profiling: $MEM_PROFILE
- Trace Profiling: $TRACE_PROFILE

## Results Summary

EOF

    # Process each result file
    for result_file in "$RESULTS_DIR"/*_results.txt; do
        if [ -f "$result_file" ]; then
            local test_name=$(basename "$result_file" _results.txt)
            echo "### $test_name" >> "$report_file"
            echo '```' >> "$report_file"
            cat "$result_file" >> "$report_file"
            echo '```' >> "$report_file"
            echo "" >> "$report_file"
        fi
    done
    
    # Add profiling information
    cat >> "$report_file" << EOF

## Profiling Data

The following profiling files were generated:

EOF

    for prof_file in "$RESULTS_DIR"/*.prof; do
        if [ -f "$prof_file" ]; then
            local prof_name=$(basename "$prof_file")
            echo "- \`$prof_name\`" >> "$report_file"
        fi
    done
    
    cat >> "$report_file" << EOF

### Analyzing Profiles

To analyze CPU profiles:
\`\`\`bash
go tool pprof $RESULTS_DIR/baseline_cpu.prof
\`\`\`

To analyze memory profiles:
\`\`\`bash
go tool pprof $RESULTS_DIR/baseline_mem.prof
\`\`\`

### Performance Analysis Commands

\`\`\`bash
# View top functions by CPU usage
go tool pprof -top $RESULTS_DIR/baseline_cpu.prof

# Generate flame graph (requires graphviz)
go tool pprof -web $RESULTS_DIR/baseline_cpu.prof

# Memory allocation analysis
go tool pprof -alloc_space $RESULTS_DIR/baseline_mem.prof
\`\`\`

EOF

    success "Generated report: $report_file"
}

# Function to compare with baseline
compare_with_baseline() {
    local baseline_file="$RESULTS_DIR/baseline_results.txt"
    local current_file="$1"
    
    if [ ! -f "$baseline_file" ]; then
        warn "No baseline file found. Current results will be used as baseline."
        cp "$current_file" "$baseline_file"
        return
    fi
    
    log "Comparing with baseline..."
    
    # Simple comparison - in practice, you'd want more sophisticated analysis
    echo "=== BASELINE vs CURRENT ===" >> "$current_file"
    echo "Baseline:" >> "$current_file"
    grep "Benchmark" "$baseline_file" >> "$current_file" || true
    echo "" >> "$current_file"
    echo "Current:" >> "$current_file"
    grep "Benchmark" "$current_file" >> "$current_file" || true
}

# Function to clean old results
cleanup_old_results() {
    log "Cleaning up old results..."
    
    # Keep only the last 10 result sets
    find "$RESULTS_DIR" -name "*_results.txt" -type f | sort -r | tail -n +11 | xargs rm -f
    find "$RESULTS_DIR" -name "*.prof" -type f -mtime +7 -delete
    find "$REPORTS_DIR" -name "*.md" -type f -mtime +30 -delete
    
    success "Cleanup completed"
}

# Main execution
main() {
    log "Starting enhanced benchmark suite..."
    
    # Clean up old results
    cleanup_old_results
    
    # Run baseline benchmark
    run_benchmark "Baseline Performance" "baseline" "BenchmarkBaseline"
    
    # Run DATC category benchmarks
    run_benchmark "DATC Categories" "datc_categories" "BenchmarkDATCByCategory"
    
    # Run complexity scenarios
    run_benchmark "Complexity Scenarios" "complexity" "BenchmarkComplexityScenarios"
    
    # Run component benchmarks
    run_benchmark "Resolution Components" "components" "BenchmarkResolutionEngineComponents"
    
    # Run memory pressure tests
    run_benchmark "Memory Pressure" "memory_pressure" "BenchmarkMemoryPressure"
    
    # Compare with baseline
    compare_with_baseline "$RESULTS_DIR/baseline_results.txt"
    
    # Generate comprehensive report
    generate_report
    
    success "Benchmark suite completed successfully!"
    
    # Display quick summary
    log "Quick Summary:"
    echo "Results directory: $RESULTS_DIR"
    echo "Reports directory: $REPORTS_DIR"
    echo "Latest report: $(ls -t "$REPORTS_DIR"/*.md | head -1)"
}

# Handle command line arguments
case "${1:-}" in
    "clean")
        log "Cleaning all benchmark data..."
        rm -rf "$BENCHMARK_DIR"
        success "Benchmark data cleaned"
        ;;
    "report")
        generate_report
        ;;
    "baseline")
        run_benchmark "Baseline Only" "baseline" "BenchmarkBaseline"
        ;;
    "help"|"-h"|"--help")
        echo "Usage: $0 [command]"
        echo ""
        echo "Commands:"
        echo "  (no args)  Run full benchmark suite"
        echo "  clean      Clean all benchmark data"
        echo "  report     Generate report from existing results"
        echo "  baseline   Run only baseline benchmark"
        echo "  help       Show this help message"
        ;;
    *)
        main
        ;;
esac