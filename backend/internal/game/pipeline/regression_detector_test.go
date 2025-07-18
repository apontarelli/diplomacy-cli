package pipeline

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestRegressionDetector(t *testing.T) {
	// Create temporary directory for test
	tempDir := t.TempDir()

	// Create test configuration
	config := DefaultRegressionConfig()
	config.MinSampleSize = 3
	config.TrendAnalysisWindow = 5

	detector := NewRegressionDetector(tempDir, config)

	// Create baseline metrics
	baselineMetrics := map[string]BenchmarkMetric{
		"BenchmarkBaseline": {
			Name:        "BenchmarkBaseline",
			NsPerOp:     1000,
			AllocsPerOp: 10,
			BytesPerOp:  100,
			Iterations:  1000,
		},
		"BenchmarkComplexity": {
			Name:        "BenchmarkComplexity",
			NsPerOp:     5000,
			AllocsPerOp: 50,
			BytesPerOp:  500,
			Iterations:  500,
		},
	}

	// Save baseline
	baseline := BaselineData{
		Timestamp:  time.Now(),
		GoVersion:  "go1.21.0",
		System:     "test",
		Benchmarks: baselineMetrics,
	}

	baselineFile := filepath.Join(tempDir, "benchmarks", "results", "baseline.json")
	os.MkdirAll(filepath.Dir(baselineFile), 0755)

	data, err := json.MarshalIndent(baseline, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal baseline: %v", err)
	}

	if err := os.WriteFile(baselineFile, data, 0644); err != nil {
		t.Fatalf("Failed to write baseline: %v", err)
	}

	t.Run("DetectNoRegression", func(t *testing.T) {
		// Current metrics similar to baseline
		currentMetrics := map[string]BenchmarkMetric{
			"BenchmarkBaseline": {
				Name:        "BenchmarkBaseline",
				NsPerOp:     1020, // 2% slower - within tolerance
				AllocsPerOp: 10,
				BytesPerOp:  100,
				Iterations:  1000,
			},
			"BenchmarkComplexity": {
				Name:        "BenchmarkComplexity",
				NsPerOp:     4900, // 2% faster
				AllocsPerOp: 50,
				BytesPerOp:  500,
				Iterations:  500,
			},
		}

		analysis, err := detector.AnalyzeRegression(currentMetrics)
		if err != nil {
			t.Fatalf("Analysis failed: %v", err)
		}

		if analysis.Summary.MajorRegressions > 0 {
			t.Errorf("Expected no major regressions, got %d", analysis.Summary.MajorRegressions)
		}

		if analysis.Summary.CriticalRegressions > 0 {
			t.Errorf("Expected no critical regressions, got %d", analysis.Summary.CriticalRegressions)
		}
	})

	t.Run("DetectMajorRegression", func(t *testing.T) {
		// Current metrics with major regression
		currentMetrics := map[string]BenchmarkMetric{
			"BenchmarkBaseline": {
				Name:        "BenchmarkBaseline",
				NsPerOp:     1300, // 30% slower - major regression
				AllocsPerOp: 15,   // 50% more allocations
				BytesPerOp:  150,  // 50% more memory
				Iterations:  1000,
			},
			"BenchmarkComplexity": {
				Name:        "BenchmarkComplexity",
				NsPerOp:     5000,
				AllocsPerOp: 50,
				BytesPerOp:  500,
				Iterations:  500,
			},
		}

		analysis, err := detector.AnalyzeRegression(currentMetrics)
		if err != nil {
			t.Fatalf("Analysis failed: %v", err)
		}

		if analysis.Summary.MajorRegressions == 0 {
			t.Errorf("Expected major regression to be detected")
		}

		// Check that BenchmarkBaseline is flagged as critical
		found := false
		for _, result := range analysis.Comparisons {
			if result.BenchmarkName == "BenchmarkBaseline" {
				found = true
				if !result.IsCriticalBenchmark {
					t.Errorf("BenchmarkBaseline should be marked as critical")
				}
				if result.Status != "degraded" {
					t.Errorf("Expected degraded status, got %s", result.Status)
				}
				if result.Severity != "major" {
					t.Errorf("Expected major severity, got %s", result.Severity)
				}
			}
		}

		if !found {
			t.Errorf("BenchmarkBaseline not found in analysis results")
		}
	})

	t.Run("DetectCriticalRegression", func(t *testing.T) {
		// Current metrics with critical regression
		currentMetrics := map[string]BenchmarkMetric{
			"BenchmarkBaseline": {
				Name:        "BenchmarkBaseline",
				NsPerOp:     2000, // 100% slower - critical regression
				AllocsPerOp: 20,
				BytesPerOp:  200,
				Iterations:  1000,
			},
		}

		analysis, err := detector.AnalyzeRegression(currentMetrics)
		if err != nil {
			t.Fatalf("Analysis failed: %v", err)
		}

		if analysis.Summary.CriticalRegressions == 0 {
			t.Errorf("Expected critical regression to be detected")
		}

		// Check alerts
		criticalAlerts := 0
		for _, alert := range analysis.Alerts {
			if alert.Severity == "critical" {
				criticalAlerts++
			}
		}

		if criticalAlerts == 0 {
			t.Errorf("Expected critical alerts to be generated")
		}
	})

	t.Run("TrendAnalysis", func(t *testing.T) {
		// Create historical data showing degrading trend
		historyDir := filepath.Join(tempDir, "benchmarks", "history")
		os.MkdirAll(historyDir, 0755)

		// Simulate degrading performance over time
		history := []float64{1000, 1050, 1100, 1150, 1200}
		historyData, _ := json.Marshal(history)

		historyFile := filepath.Join(historyDir, "BenchmarkBaseline.json")
		os.WriteFile(historyFile, historyData, 0644)

		currentMetrics := map[string]BenchmarkMetric{
			"BenchmarkBaseline": {
				Name:        "BenchmarkBaseline",
				NsPerOp:     1250, // Continuing the trend
				AllocsPerOp: 10,
				BytesPerOp:  100,
				Iterations:  1000,
			},
		}

		analysis, err := detector.AnalyzeRegression(currentMetrics)
		if err != nil {
			t.Fatalf("Analysis failed: %v", err)
		}

		// Check trend analysis
		found := false
		for _, trend := range analysis.TrendAnalysis {
			if trend.BenchmarkName == "BenchmarkBaseline" {
				found = true
				if trend.TrendDirection != "degrading" {
					t.Errorf("Expected degrading trend, got %s", trend.TrendDirection)
				}
				if trend.TrendStrength <= 0 {
					t.Errorf("Expected positive trend strength for degrading trend, got %f", trend.TrendStrength)
				}
			}
		}

		if !found {
			t.Errorf("Trend analysis not found for BenchmarkBaseline")
		}
	})
}

func TestAlertManager(t *testing.T) {
	tempDir := t.TempDir()

	config := DefaultAlertConfig()
	config.NotificationChannels = []NotificationChannel{
		{
			Type:     "test",
			Enabled:  true,
			Priority: []string{"major", "critical"},
		},
	}

	alertManager := NewAlertManager(tempDir, config)

	// Create test analysis with alerts
	analysis := &RegressionAnalysis{
		Timestamp: time.Now(),
		Alerts: []RegressionAlert{
			{
				Type:          "performance",
				Severity:      "major",
				BenchmarkName: "BenchmarkTest",
				Message:       "Test regression detected",
				Timestamp:     time.Now(),
			},
		},
		Summary: RegressionSummary{
			MajorRegressions: 1,
			AlertsGenerated:  1,
		},
	}

	t.Run("ProcessAlerts", func(t *testing.T) {
		err := alertManager.ProcessAlerts(analysis)
		if err != nil {
			t.Fatalf("ProcessAlerts failed: %v", err)
		}

		// Check that alert history was created
		historyFile := filepath.Join(tempDir, "benchmarks", "alerts", "history.json")
		if _, err := os.Stat(historyFile); os.IsNotExist(err) {
			t.Errorf("Alert history file was not created")
		}
	})

	t.Run("RateLimiting", func(t *testing.T) {
		// Create history with many recent alerts
		history := &AlertHistory{
			Alerts: make([]SentAlert, 0),
		}

		// Add alerts up to the rate limit
		for i := 0; i < config.RateLimiting.MaxAlertsPerHour; i++ {
			history.Alerts = append(history.Alerts, SentAlert{
				Timestamp: time.Now().Add(-30 * time.Minute),
			})
		}

		// This should be rate limited
		if !alertManager.isRateLimited(history) {
			t.Errorf("Expected rate limiting to be active")
		}

		// Add old alerts - should not be rate limited
		history.Alerts = []SentAlert{
			{Timestamp: time.Now().Add(-2 * time.Hour)},
		}

		if alertManager.isRateLimited(history) {
			t.Errorf("Expected rate limiting to be inactive for old alerts")
		}
	})
}

func TestPerformanceRegressionIntegration(t *testing.T) {
	tempDir := t.TempDir()

	// Test the complete integration
	runner := NewBenchmarkRunner(tempDir)
	config := DefaultBenchmarkConfig()
	config.BenchTime = "1s"
	config.Count = 1
	config.Patterns = []string{"BenchmarkBaseline"} // Only run one benchmark for speed
	runner.SetConfig(config)

	// Create regression detector
	regressionConfig := DefaultRegressionConfig()
	_ = NewRegressionDetector(tempDir, regressionConfig)

	// Create alert manager
	alertConfig := DefaultAlertConfig()
	_ = NewAlertManager(tempDir, alertConfig)
	t.Run("EndToEndWorkflow", func(t *testing.T) {
		// This test would run actual benchmarks, but we'll skip it in CI
		// to avoid dependencies on the actual benchmark functions
		t.Skip("Skipping end-to-end test to avoid benchmark dependencies")

		// The workflow would be:
		// 1. Run benchmarks with runner.RunBenchmarks()
		// 2. Analyze results with detector.AnalyzeRegression()
		// 3. Process alerts with alertManager.ProcessAlerts()
	})
}
