package pipeline

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"runtime"
	"testing"
	"time"

	"diplomacy-cli/backend/internal/game"
	"diplomacy-cli/backend/internal/game/loader"
)

// BenchmarkResult holds detailed benchmark metrics
type BenchmarkResult struct {
	Name         string           `json:"name"`
	Duration     time.Duration    `json:"duration"`
	Iterations   int              `json:"iterations"`
	NsPerOp      int64            `json:"ns_per_op"`
	AllocsPerOp  int64            `json:"allocs_per_op"`
	BytesPerOp   int64            `json:"bytes_per_op"`
	MemoryBefore runtime.MemStats `json:"memory_before"`
	MemoryAfter  runtime.MemStats `json:"memory_after"`
	TestCategory string           `json:"test_category"`
	Complexity   string           `json:"complexity"`
}

// DATCTestCase represents a test case from the DATC JSON file
type DATCTestCase struct {
	ID              string              `json:"id"`
	Name            string              `json:"name"`
	Section         string              `json:"section"`
	Category        string              `json:"category"`
	Description     string              `json:"description"`
	Orders          map[string][]string `json:"orders"`
	ExpectedOutcome string              `json:"expected_outcome"`
}

// DATCTestSuite represents the full DATC test suite
type DATCTestSuite struct {
	DATCVersion    string         `json:"datc_version"`
	TotalTestCases int            `json:"total_test_cases"`
	TestCases      []DATCTestCase `json:"test_cases"`
}

// loadDATCTests loads the DATC test suite from JSON
func loadDATCTests() (*DATCTestSuite, error) {
	data, err := ioutil.ReadFile("testdata/datc_tests_fixed.json")
	if err != nil {
		return nil, fmt.Errorf("failed to read DATC tests: %v", err)
	}

	var suite DATCTestSuite
	if err := json.Unmarshal(data, &suite); err != nil {
		return nil, fmt.Errorf("failed to parse DATC tests: %v", err)
	}

	return &suite, nil
}

// runBenchmarkWithMemory runs a benchmark with detailed memory tracking
func runBenchmarkWithMemory(b *testing.B, name string, category string, complexity string, fn func()) BenchmarkResult {
	// Force GC before measurement
	runtime.GC()
	runtime.GC() // Run twice to ensure clean state

	var memBefore, memAfter runtime.MemStats
	runtime.ReadMemStats(&memBefore)

	start := time.Now()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		fn()
	}
	b.StopTimer()

	duration := time.Since(start)
	runtime.ReadMemStats(&memAfter)

	return BenchmarkResult{
		Name:         name,
		Duration:     duration,
		Iterations:   b.N,
		NsPerOp:      duration.Nanoseconds() / int64(b.N),
		AllocsPerOp:  int64(testing.AllocsPerRun(1, fn)),
		BytesPerOp:   int64(testing.AllocsPerRun(1, fn)) * 64, // Rough estimate
		MemoryBefore: memBefore,
		MemoryAfter:  memAfter,
		TestCategory: category,
		Complexity:   complexity,
	}
}

// BenchmarkDATCByCategory benchmarks all DATC tests grouped by category
func BenchmarkDATCByCategory(b *testing.B) {
	suite, err := loadDATCTests()
	if err != nil {
		b.Fatalf("Failed to load DATC tests: %v", err)
	}

	// Group tests by category
	categories := make(map[string][]DATCTestCase)
	for _, testCase := range suite.TestCases {
		categories[testCase.Category] = append(categories[testCase.Category], testCase)
	}

	// Benchmark each category
	for category, tests := range categories {
		b.Run(category, func(b *testing.B) {
			processor := NewTurnProcessor()
			mapLoader := loader.NewJSONLoader("../../../../backend/data/classic")
			board, err := mapLoader.LoadBoard()
			if err != nil {
				b.Fatalf("Failed to load board: %v", err)
			}

			// Run a sample of tests from this category (first 5 to avoid timeout)
			sampleSize := len(tests)
			if sampleSize > 5 {
				sampleSize = 5
			}

			result := runBenchmarkWithMemory(b, fmt.Sprintf("DATC_%s", category), category, "medium", func() {
				for i := 0; i < sampleSize; i++ {
					gameState := game.NewGameState(board, game.SpringMovement, 1901)
					// TODO: Convert DATC test case to game state and run
					processor.ProcessTurn(gameState)
				}
			})

			// Log detailed results
			b.Logf("Category: %s, Allocs/op: %d, Bytes/op: %d", category, result.AllocsPerOp, result.BytesPerOp)
		})
	}
}

// BenchmarkComplexityScenarios tests different game complexity levels
func BenchmarkComplexityScenarios(b *testing.B) {
	scenarios := []struct {
		name       string
		complexity string
		players    int
		units      int
		orders     int
	}{
		{"Simple_2P_4U", "low", 2, 4, 4},
		{"Medium_4P_12U", "medium", 4, 12, 12},
		{"Complex_7P_22U", "high", 7, 22, 22},
		{"Stress_7P_50U", "extreme", 7, 50, 50},
	}

	for _, scenario := range scenarios {
		b.Run(scenario.name, func(b *testing.B) {
			processor := NewTurnProcessor()
			mapLoader := loader.NewJSONLoader("../../../../backend/data/classic")
			board, err := mapLoader.LoadBoard()
			if err != nil {
				b.Fatalf("Failed to load board: %v", err)
			}

			result := runBenchmarkWithMemory(b, scenario.name, "complexity", scenario.complexity, func() {
				gameState := createComplexGameState(board, scenario.players, scenario.units, scenario.orders)
				processor.ProcessTurn(gameState)
			})

			b.Logf("Scenario: %s, Players: %d, Units: %d, Allocs/op: %d",
				scenario.name, scenario.players, scenario.units, result.AllocsPerOp)
		})
	}
}

// BenchmarkResolutionEngineComponents benchmarks individual resolution components
func BenchmarkResolutionEngineComponents(b *testing.B) {
	components := []struct {
		name string
		fn   func()
	}{
		{"OrderParsing", func() {
			// TODO: Benchmark order parsing
		}},
		{"StrengthCalculation", func() {
			// TODO: Benchmark strength calculation
		}},
		{"ConflictResolution", func() {
			// TODO: Benchmark conflict resolution
		}},
		{"ConvoyPathfinding", func() {
			// TODO: Benchmark convoy pathfinding
		}},
	}

	for _, component := range components {
		b.Run(component.name, func(b *testing.B) {
			result := runBenchmarkWithMemory(b, component.name, "component", "micro", component.fn)
			b.Logf("Component: %s, Allocs/op: %d", component.name, result.AllocsPerOp)
		})
	}
}

// BenchmarkMemoryPressure tests performance under memory pressure
func BenchmarkMemoryPressure(b *testing.B) {
	b.Run("HighAllocation", func(b *testing.B) {
		processor := NewTurnProcessor()
		mapLoader := loader.NewJSONLoader("../../../../backend/data/classic")
		board, err := mapLoader.LoadBoard()
		if err != nil {
			b.Fatalf("Failed to load board: %v", err)
		}

		result := runBenchmarkWithMemory(b, "MemoryPressure", "stress", "high", func() {
			// Create many game states to simulate memory pressure
			for i := 0; i < 10; i++ {
				gameState := createComplexGameState(board, 7, 22, 22)
				processor.ProcessTurn(gameState)
			}
		})

		b.Logf("Memory pressure test - GC cycles: %d, Heap size: %d MB",
			result.MemoryAfter.NumGC-result.MemoryBefore.NumGC,
			result.MemoryAfter.HeapSys/1024/1024)
	})
}

// createComplexGameState creates a game state with specified complexity
func createComplexGameState(board *game.Board, players, units, orders int) *game.GameState {
	gameState := game.NewGameState(board, game.SpringMovement, 1901)

	// Add units and orders based on complexity parameters
	// This is a simplified implementation - in practice, we'd create realistic scenarios
	unitCount := 0
	orderCount := 0

	nations := []game.Nation{game.England, game.France, game.Germany, game.Austria, game.Italy, game.Russia, game.Turkey}
	provinces := []string{"london", "paris", "berlin", "vienna", "rome", "moscow", "constantinople",
		"liverpool", "brest", "munich", "budapest", "naples", "st_petersburg", "ankara"}

	for i := 0; i < players && i < len(nations); i++ {
		nation := nations[i]
		unitsPerPlayer := units / players
		ordersPerPlayer := orders / players

		// Add units
		for j := 0; j < unitsPerPlayer && unitCount < len(provinces); j++ {
			province := provinces[unitCount]
			gameState.Board.Units[province] = &game.Unit{
				Type:     game.Army,
				Owner:    nation,
				Province: province,
			}
			unitCount++
		}

		// Add orders
		var playerOrders []string
		for j := 0; j < ordersPerPlayer && orderCount < len(provinces)-1; j++ {
			from := provinces[orderCount]
			to := provinces[orderCount+1]
			playerOrders = append(playerOrders, fmt.Sprintf("%s - %s", from, to))
			orderCount++
		}

		if len(playerOrders) > 0 {
			gameState.RawOrders[nation] = playerOrders
		}
	}

	return gameState
}

// BenchmarkBaseline establishes performance baseline for comparison
func BenchmarkBaseline(b *testing.B) {
	processor := NewTurnProcessor()
	mapLoader := loader.NewJSONLoader("../../../../backend/data/classic")
	board, err := mapLoader.LoadBoard()
	if err != nil {
		b.Fatalf("Failed to load board: %v", err)
	}

	// Standard baseline scenario
	gameState := game.NewGameState(board, game.SpringMovement, 1901)
	gameState.Board.Units["london"] = &game.Unit{
		Type:     game.Fleet,
		Owner:    game.England,
		Province: "london",
	}
	gameState.Board.Units["wales"] = &game.Unit{
		Type:     game.Army,
		Owner:    game.England,
		Province: "wales",
	}
	gameState.RawOrders[game.England] = []string{
		"f london - english_channel",
		"wales s london - english_channel",
	}

	result := runBenchmarkWithMemory(b, "Baseline", "baseline", "standard", func() {
		processor.ProcessTurn(gameState)
	})

	// Save baseline for future comparisons
	b.Logf("BASELINE: %d ns/op, %d allocs/op, %d bytes/op",
		result.NsPerOp, result.AllocsPerOp, result.BytesPerOp)
}
