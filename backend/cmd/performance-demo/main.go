package main

import (
	"fmt"
	"log"
	"time"

	"diplomacy-cli/backend/internal/game/resolution"
)

func main() {
	fmt.Println("🚀 DATC Engine Performance Monitoring Demo")
	fmt.Println("==========================================")

	// Enable performance monitoring
	resolution.EnableGlobalMonitoring()

	// Create test scenarios
	scenarios := []struct {
		name   string
		orders []resolution.Order
	}{
		{
			name: "Simple Conflict",
			orders: []resolution.Order{
				{Unit: "A Berlin", Type: resolution.Move, Source: "berlin", Destination: "munich", Owner: "germany"},
				{Unit: "A Vienna", Type: resolution.Support, Source: "vienna", Auxiliary: "berlin -> munich", Owner: "austria"},
				{Unit: "A Paris", Type: resolution.Move, Source: "paris", Destination: "munich", Owner: "france"},
			},
		},
		{
			name: "Complex Multi-Nation",
			orders: []resolution.Order{
				{Unit: "A Berlin", Type: resolution.Move, Source: "berlin", Destination: "munich", Owner: "germany"},
				{Unit: "F Kiel", Type: resolution.Support, Source: "kiel", Auxiliary: "berlin -> munich", Owner: "germany"},
				{Unit: "A Vienna", Type: resolution.Move, Source: "vienna", Destination: "munich", Owner: "austria"},
				{Unit: "A Budapest", Type: resolution.Support, Source: "budapest", Auxiliary: "vienna -> munich", Owner: "austria"},
				{Unit: "A Paris", Type: resolution.Move, Source: "paris", Destination: "burgundy", Owner: "france"},
				{Unit: "A Marseilles", Type: resolution.Support, Source: "marseilles", Auxiliary: "paris -> burgundy", Owner: "france"},
				{Unit: "F London", Type: resolution.Move, Source: "london", Destination: "belgium", Owner: "england"},
				{Unit: "F North Sea", Type: resolution.Convoy, Source: "north_sea", Auxiliary: "A london - belgium", Owner: "england"},
				{Unit: "A Moscow", Type: resolution.Move, Source: "moscow", Destination: "warsaw", Owner: "russia"},
				{Unit: "A St Petersburg", Type: resolution.Support, Source: "st_petersburg", Auxiliary: "moscow -> warsaw", Owner: "russia"},
			},
		},
	}

	dashboard := resolution.GetGlobalDashboard()
	monitor := resolution.GetGlobalPerformanceMonitor()

	for i, scenario := range scenarios {
		fmt.Printf("\n📊 Running Scenario %d: %s\n", i+1, scenario.name)
		fmt.Printf("Orders: %d\n", len(scenario.orders))

		// Reset metrics for this scenario
		monitor.Reset()

		// Run multiple iterations to get meaningful metrics
		iterations := 100
		start := time.Now()

		for j := 0; j < iterations; j++ {
			engine := resolution.NewDATCEngine(scenario.orders)
			results := engine.ResolveAll()

			// Access some reasons to trigger lazy evaluation
			for k, result := range results {
				if k%2 == 0 { // Access every other reason
					_ = result.Reason
				}
			}

			engine.Cleanup()
		}

		elapsed := time.Since(start)
		fmt.Printf("Completed %d iterations in %v\n", iterations, elapsed)

		// Generate performance report
		fmt.Println("\n" + dashboard.GenerateReport())
		fmt.Println(dashboard.GenerateTopBottlenecks())
	}

	// Demonstrate profiling capabilities
	fmt.Println("\n🔍 Profiling Demo")
	fmt.Println("=================")

	profileManager := resolution.GetGlobalProfileManager()

	// Start pprof server (in background)
	go func() {
		if err := profileManager.StartPProfServer("localhost:6060"); err != nil {
			log.Printf("Failed to start pprof server: %v", err)
		}
	}()

	fmt.Println("pprof server started at http://localhost:6060/debug/pprof/")
	fmt.Println("You can now use:")
	fmt.Println("  go tool pprof http://localhost:6060/debug/pprof/profile")
	fmt.Println("  go tool pprof http://localhost:6060/debug/pprof/heap")

	// Demonstrate profiled resolution
	fmt.Println("\n📈 Running profiled resolution...")
	engine := resolution.NewDATCEngine(scenarios[1].orders)

	results, err := profileManager.ProfiledResolution(nil, engine, "demo_profile")
	if err != nil {
		log.Printf("Profiling error: %v", err)
	} else {
		fmt.Printf("Profiled resolution completed with %d results\n", len(results))
		fmt.Println("Profile files generated:")
		fmt.Println("  - demo_profile_cpu.prof")
		fmt.Println("  - demo_profile_mem.prof")
		fmt.Println("  - demo_profile_trace.out")
	}

	// Demonstrate JSON export
	fmt.Println("\n📄 JSON Report Export")
	fmt.Println("======================")

	jsonReport, err := dashboard.GenerateJSONReport()
	if err != nil {
		log.Printf("Failed to generate JSON report: %v", err)
	} else {
		fmt.Println("JSON report generated (first 500 chars):")
		if len(jsonReport) > 500 {
			fmt.Println(jsonReport[:500] + "...")
		} else {
			fmt.Println(jsonReport)
		}
	}

	// Demonstrate alert system
	fmt.Println("\n🚨 Alert System Demo")
	fmt.Println("====================")

	alertManager := resolution.NewAlertManager()
	metrics := monitor.GetMetrics()
	alerts := alertManager.CheckAlerts(metrics, monitor)

	if len(alerts) == 0 {
		fmt.Println("✅ No performance alerts detected")
	} else {
		fmt.Printf("⚠️ %d performance alerts detected:\n", len(alerts))
		for i, alert := range alerts {
			fmt.Printf("  %d. [%s] %s: %.2f (threshold: %.2f)\n",
				i+1, alert.Level, alert.Message, alert.Value, alert.Threshold)
		}
	}

	fmt.Println("\n🎉 Performance monitoring demo completed!")
	fmt.Println("The DATC engine now has comprehensive performance monitoring capabilities.")
}
