# Performance Regression Testing Framework

This directory contains the comprehensive performance regression testing framework for the Diplomacy CLI resolution engine.

## Overview

The performance regression testing framework provides automated detection of performance degradations, trend analysis, and alerting capabilities to ensure the resolution engine maintains optimal performance over time.

## Components

### 1. Benchmark Runner (`benchmark_runner.go`)
- Orchestrates automated benchmark execution
- Supports multiple benchmark patterns and configurations
- Generates detailed reports with profiling data
- Integrates with CI/CD pipelines

### 2. Regression Detector (`regression_detector.go`)
- Analyzes benchmark results for performance regressions
- Performs statistical analysis and trend detection
- Categorizes regressions by severity (minor, major, critical)
- Maintains historical performance data

### 3. Alert Manager (`alert_manager.go`)
- Sends alerts for performance degradations
- Supports multiple notification channels (GitHub, Slack, webhooks)
- Implements rate limiting and duplicate suppression
- Configurable alert thresholds and templates

### 4. Benchmark Comparator (`benchmark_comparison.go`)
- Compares current results against baseline metrics
- Generates human-readable comparison reports
- Tracks performance improvements and degradations
- Maintains baseline data with metadata

## Directory Structure

```
benchmarks/
├── README.md                    # This file
├── performance-config.json      # Configuration file
├── results/                     # Benchmark results and baselines
│   ├── baseline.json           # Current performance baseline
│   ├── benchmark_*.txt         # Raw benchmark outputs
│   ├── benchmark_*.json        # Parsed benchmark data
│   └── *.prof                  # Profiling data (CPU, memory, trace)
├── reports/                     # Generated reports
│   ├── benchmark_report_*.md   # Markdown reports
│   ├── benchmark_report_*.json # JSON reports
│   └── comparison_*.md         # Comparison reports
├── history/                     # Historical performance data
│   └── Benchmark*.json         # Per-benchmark historical data
└── alerts/                     # Alert system data
    └── history.json            # Alert history for rate limiting
```

## Usage

### Local Development

#### Run Performance Regression Test
```bash
# Basic regression test
./scripts/performance-regression.sh

# With custom configuration
./scripts/performance-regression.sh --bench-time 30s --count 5

# Save new baseline
./scripts/performance-regression.sh --save-baseline

# Skip baseline comparison (for initial setup)
./scripts/performance-regression.sh --skip-baseline
```

#### Manual Benchmark Execution
```bash
# Run all benchmarks
go test -bench=. -benchmem -count=3 ./internal/game/pipeline/

# Run specific benchmark pattern
go test -bench=BenchmarkBaseline -benchmem -count=5 ./internal/game/pipeline/

# With profiling
go test -bench=. -benchmem -cpuprofile=cpu.prof -memprofile=mem.prof ./internal/game/pipeline/
```

### CI/CD Integration

The framework integrates with GitHub Actions through the `.github/workflows/performance-regression.yml` workflow:

- **Automatic Execution**: Runs on pushes to main/develop branches
- **Pull Request Analysis**: Compares PR performance against base branch
- **Scheduled Monitoring**: Daily performance trend monitoring
- **Alert Generation**: Creates GitHub issues for significant regressions

#### Workflow Triggers
- Push to main/develop branches
- Pull requests affecting backend code
- Daily scheduled runs (2 AM UTC)
- Manual workflow dispatch

#### Workflow Outputs
- Performance regression reports in PR comments
- Benchmark artifacts uploaded for analysis
- GitHub issues created for critical regressions
- Performance baselines updated automatically

## Configuration

### Performance Thresholds (`performance-config.json`)

```json
{
  "thresholds": {
    "minor_improvement": 0.95,      // 5% improvement
    "major_improvement": 0.80,      // 20% improvement
    "minor_degradation": 1.05,      // 5% degradation
    "major_degradation": 1.20,      // 20% degradation
    "critical_degradation": 1.50    // 50% degradation
  }
}
```

### Alert Configuration

```json
{
  "alert_config": {
    "enable_alerts": true,
    "alert_on_major_degradation": true,
    "alert_on_critical_degradation": true,
    "notification_channels": ["github_issues", "slack"]
  }
}
```

### Critical Benchmarks

Benchmarks marked as "critical" receive special attention:
- `BenchmarkBaseline`: Core resolution engine performance
- `BenchmarkComplexityScenarios`: Complex game state resolution

## Benchmark Categories

### 1. Baseline Benchmarks
- **BenchmarkBaseline**: Core resolution engine performance
- **Purpose**: Establish performance baseline for simple scenarios
- **Target**: < 1ms per resolution

### 2. DATC Compliance Benchmarks
- **BenchmarkDATCByCategory**: DATC test categories
- **Purpose**: Ensure compliance doesn't impact performance
- **Target**: Maintain current performance levels

### 3. Complexity Benchmarks
- **BenchmarkComplexityScenarios**: Complex game states
- **Purpose**: Test performance under realistic load
- **Target**: Scale linearly with complexity

### 4. Component Benchmarks
- **BenchmarkResolutionEngineComponents**: Individual components
- **Purpose**: Identify performance bottlenecks
- **Target**: Optimize hot paths

## Performance Targets

| Benchmark Category | Target Performance | Memory Usage | Allocations |
|-------------------|-------------------|--------------|-------------|
| Baseline | < 1ms/op | < 1KB/op | < 10 allocs/op |
| DATC Compliance | < 5ms/op | < 5KB/op | < 50 allocs/op |
| Complex Scenarios | < 10ms/op | < 10KB/op | < 100 allocs/op |
| Component Tests | Varies | Minimize | Minimize |

## Regression Detection

### Statistical Analysis
- **Trend Analysis**: Detects performance trends over time
- **Variance Analysis**: Identifies unstable performance
- **Confidence Intervals**: Statistical significance testing
- **Outlier Detection**: Filters anomalous results

### Severity Classification
- **Critical**: > 50% degradation or critical benchmark regression
- **Major**: 20-50% degradation
- **Minor**: 5-20% degradation
- **Negligible**: < 5% change

### Alert Conditions
- Major degradation in critical benchmarks
- Critical degradation in any benchmark
- Consistent degradation trends
- High performance variance

## Troubleshooting

### Common Issues

#### No Baseline Found
```bash
# Create initial baseline
./scripts/performance-regression.sh --save-baseline --skip-baseline
```

#### High Variance Detected
- Check system load during benchmarks
- Increase benchmark sample size (`--count`)
- Run benchmarks in isolated environment

#### False Positive Regressions
- Review benchmark environment consistency
- Check for external factors (system load, thermal throttling)
- Adjust regression thresholds if needed

#### Missing Profiling Data
- Ensure sufficient disk space
- Check file permissions in results directory
- Verify Go toolchain includes profiling tools

### Performance Analysis

#### CPU Profiling
```bash
# Generate CPU profile
go test -bench=BenchmarkBaseline -cpuprofile=cpu.prof ./internal/game/pipeline/

# Analyze profile
go tool pprof cpu.prof
```

#### Memory Profiling
```bash
# Generate memory profile
go test -bench=BenchmarkBaseline -memprofile=mem.prof ./internal/game/pipeline/

# Analyze profile
go tool pprof mem.prof
```

#### Trace Analysis
```bash
# Generate trace
go test -bench=BenchmarkBaseline -trace=trace.out ./internal/game/pipeline/

# Analyze trace
go tool trace trace.out
```

## Best Practices

### Benchmark Development
1. **Consistent Environment**: Run benchmarks in consistent conditions
2. **Sufficient Samples**: Use adequate sample sizes for statistical significance
3. **Realistic Scenarios**: Test with representative game states
4. **Isolation**: Minimize external factors affecting performance

### Baseline Management
1. **Regular Updates**: Update baselines after verified improvements
2. **Version Control**: Track baseline changes with git commits
3. **Documentation**: Document significant baseline changes
4. **Approval Process**: Require review for baseline updates

### Alert Management
1. **Appropriate Thresholds**: Set realistic regression thresholds
2. **Rate Limiting**: Prevent alert spam with rate limiting
3. **Actionable Alerts**: Ensure alerts provide clear next steps
4. **Regular Review**: Periodically review alert effectiveness

## Integration with Development Workflow

### Pre-commit Hooks
Consider adding performance checks to pre-commit hooks for critical changes:

```bash
# Add to .git/hooks/pre-commit
./backend/scripts/performance-regression.sh --bench-time 5s --fail-on-regression true
```

### Code Review Process
1. **Performance Impact Assessment**: Evaluate performance impact of changes
2. **Benchmark Updates**: Update benchmarks for new features
3. **Baseline Approval**: Require approval for baseline changes
4. **Documentation**: Document performance considerations

### Release Process
1. **Performance Validation**: Validate performance before releases
2. **Regression Testing**: Run comprehensive regression tests
3. **Baseline Updates**: Update baselines for new releases
4. **Performance Documentation**: Update performance documentation

## Monitoring and Maintenance

### Regular Tasks
- **Weekly**: Review performance trends and alerts
- **Monthly**: Analyze benchmark coverage and effectiveness
- **Quarterly**: Review and update performance targets
- **Annually**: Comprehensive framework review and updates

### Metrics to Monitor
- **Regression Detection Rate**: Percentage of actual regressions detected
- **False Positive Rate**: Percentage of false regression alerts
- **Alert Response Time**: Time to address performance alerts
- **Benchmark Coverage**: Percentage of code covered by benchmarks

## Future Enhancements

### Planned Features
1. **Machine Learning**: ML-based anomaly detection
2. **Distributed Benchmarking**: Multi-node benchmark execution
3. **Real-time Monitoring**: Live performance monitoring
4. **Advanced Analytics**: Deeper performance insights

### Integration Opportunities
1. **APM Tools**: Integration with application performance monitoring
2. **Cloud Platforms**: Cloud-based benchmark execution
3. **Notification Systems**: Additional notification channels
4. **Reporting Tools**: Enhanced reporting and visualization

---

For questions or issues with the performance regression testing framework, please:
1. Check this documentation
2. Review existing GitHub issues
3. Create a new issue with the `performance` label
4. Contact the development team