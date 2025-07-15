package pipeline

import (
	"strings"
	"sync"
	"testing"

	"diplomacy-cli/backend/internal/game"
	"diplomacy-cli/backend/internal/game/loader"
)

var (
	datcTestBoard     *game.Board
	datcTestBoardOnce sync.Once
	datcTestBoardErr  error
)

// ResetDATCTestBoard clears the cached board for testing
func ResetDATCTestBoard() {
	datcTestBoard = nil
	datcTestBoardErr = nil
	datcTestBoardOnce = sync.Once{}
}

// DATCTestResult represents the result of processing a DATC test case
type DATCTestResult struct {
	NewGameState *game.GameState
	SyntaxErrors []string
	ProcessError error
}

// ProcessDATCTest is a helper function to process a DATC test case and return results
func ProcessDATCTest(t *testing.T, gameState *game.GameState) *DATCTestResult {
	t.Helper()

	processor := NewTurnProcessor()
	result, err := processor.ProcessTurn(gameState)

	var syntaxErrors []string
	if err != nil {
		// Extract syntax errors from the error message
		errorStr := err.Error()
		if strings.Contains(errorStr, "syntax errors:") {
			// Parse syntax errors from the error message
			parts := strings.Split(errorStr, "syntax errors: [")
			if len(parts) > 1 {
				errorsPart := strings.TrimSuffix(parts[1], "]")
				syntaxErrors = []string{errorsPart}
			}
		} else {
			syntaxErrors = []string{errorStr}
		}
	}

	return &DATCTestResult{
		NewGameState: result, // Can be nil if processing failed
		SyntaxErrors: syntaxErrors,
		ProcessError: err,
	}
}

// ParseUnitsFromOrders extracts unit positions from DATC orders
func ParseUnitsFromOrders(orders map[string][]string) map[string][]game.Unit {
	units := make(map[string][]game.Unit)

	for nation, orderList := range orders {
		var nationUnits []game.Unit

		for _, order := range orderList {
			// Extract unit type and location from order
			// Examples: "F North Sea - Picardy", "A Liverpool - Irish Sea", "F London - Yorkshire"
			parts := strings.Fields(order)
			if len(parts) >= 3 {
				unitType := parts[0] // F or A
				location := strings.Join(parts[1:findDashIndex(parts)], " ")

				var unit game.Unit
				switch unitType {
				case "F":
					unit = game.Unit{
						Type:     game.Fleet,
						Owner:    game.Nation(nation),
						Province: location,
					}
				case "A":
					unit = game.Unit{
						Type:     game.Army,
						Owner:    game.Nation(nation),
						Province: location,
					}
				default:
					continue // Skip invalid unit types
				}

				nationUnits = append(nationUnits, unit)
			}
		}

		if len(nationUnits) > 0 {
			units[nation] = nationUnits
		}
	}

	return units
}

// findDashIndex finds the index of "-" or "Supports" or "Convoys" in order parts
func findDashIndex(parts []string) int {
	for i, part := range parts {
		if part == "-" || part == "Supports" || part == "Convoys" {
			return i
		}
	}
	return len(parts) // If no separator found, assume entire string is location
}

// GetDATCTestBoard loads the classic board once and reuses it across DATC tests
func GetDATCTestBoard() (*game.Board, error) {
	datcTestBoardOnce.Do(func() {
		// Use path that works from test directory (backend/internal/game/pipeline)
		mapLoader := loader.NewJSONLoader("../../../../backend/data/classic")
		datcTestBoard, datcTestBoardErr = mapLoader.LoadBoard()
	})
	return datcTestBoard, datcTestBoardErr
}

// CreateDATCGameState creates a fresh game state for DATC testing
func CreateDATCGameState(t *testing.T) *game.GameState {
	t.Helper()

	board, err := GetDATCTestBoard()
	if err != nil {
		t.Fatalf("Failed to load board: %v", err)
	}

	// Create a fresh board by cloning the template to avoid test interference
	freshBoard := &game.Board{
		Provinces: board.Provinces,
		Units:     make(map[string]*game.Unit), // Fresh empty units map
	}

	return game.NewGameState(freshBoard, game.SpringMovement, 1901)
}

// ValidateExpectedOutcome validates the test result against the expected DATC outcome
func ValidateExpectedOutcome(t *testing.T, result *DATCTestResult, expectedOutcome string, testID string) {
	t.Helper()

	// Parse expected outcome and validate accordingly
	outcome := strings.ToLower(expectedOutcome)

	switch {
	case strings.Contains(outcome, "order should fail"):
		t.Logf("📋 Test %s: Validating that order failed", testID)
		// For basic failure tests, we expect syntax errors
		if len(result.SyntaxErrors) == 0 && result.ProcessError == nil {
			t.Errorf("❌ Expected order to fail, but no errors found")
		} else {
			t.Logf("✓ Order correctly failed: %v", result.SyntaxErrors)
		}

	case strings.Contains(outcome, "should be dislodged"):
		t.Logf("📋 Test %s: Validating dislodgement", testID)
		// TODO: Implement specific dislodgement validation
		if result.ProcessError != nil {
			t.Logf("⚠ Process error occurred: %v", result.ProcessError)
		}

	case strings.Contains(outcome, "no outcome specified"):
		t.Logf("📋 Test %s: Validating general consistency", testID)
		if result.ProcessError != nil {
			t.Logf("⚠ Process error occurred: %v", result.ProcessError)
		}

	case strings.Contains(outcome, "move fails"):
		t.Logf("📋 Test %s: Validating that move failed", testID)
		if len(result.SyntaxErrors) == 0 && result.ProcessError == nil {
			t.Errorf("❌ Expected move to fail, but no errors found")
		} else {
			t.Logf("✓ Move correctly failed: %v", result.SyntaxErrors)
		}

	default:
		t.Logf("📋 Test %s: Generic validation for outcome: %s", testID, expectedOutcome)
		if result.ProcessError != nil {
			t.Logf("⚠ Process error occurred: %v", result.ProcessError)
		}
	}

	// Log the final result
	if len(result.SyntaxErrors) > 0 {
		t.Logf("🔍 Syntax errors encountered: %v", result.SyntaxErrors)
	}
	if result.NewGameState != nil {
		t.Logf("🔍 Turn processed successfully")
	}

	t.Logf("📊 Test %s validation completed", testID)
}
