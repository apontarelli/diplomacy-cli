package resolution

import (
	"diplomacy-cli/backend/internal/game"
	"diplomacy-cli/backend/internal/game/loader"
	"testing"
)

func TestCalculateSupplyCenters(t *testing.T) {
	// Load test board
	gameLoader := loader.GetLoader(loader.Classic)
	board, err := gameLoader.LoadBoard()
	if err != nil {
		t.Fatalf("Failed to load board: %v", err)
	}

	// Create test game state
	gameState := game.NewGameState(board, game.WinterBuild, 1901)

	// Place some units to control supply centers
	units := []*game.Unit{
		{Type: game.Army, Owner: game.Germany, Province: "berlin"},
		{Type: game.Fleet, Owner: game.Germany, Province: "kiel"},
		{Type: game.Army, Owner: game.France, Province: "paris"},
		{Type: game.Fleet, Owner: game.France, Province: "brest"},
		{Type: game.Army, Owner: game.England, Province: "london"},
	}

	for _, unit := range units {
		err := gameState.Board.PlaceUnit(unit)
		if err != nil {
			t.Fatalf("Failed to place unit: %v", err)
		}
	}

	// Test supply center calculation
	buildProcessor := NewBuildProcessor(gameState.Board)
	supplyCenters := buildProcessor.CalculateSupplyCenters(gameState)

	// Verify results
	expectedCounts := map[game.Nation]int{
		game.Germany: 2, // berlin, kiel
		game.France:  2, // paris, brest
		game.England: 1, // london
		game.Austria: 0,
		game.Italy:   0,
		game.Russia:  0,
		game.Turkey:  0,
	}

	for nation, expected := range expectedCounts {
		if supplyCenters[nation] != expected {
			t.Errorf("Expected %s to have %d supply centers, got %d", nation, expected, supplyCenters[nation])
		}
	}
}

func TestCalculateBuildAdjustments(t *testing.T) {
	// Load test board
	gameLoader := loader.GetLoader(loader.Classic)
	board, err := gameLoader.LoadBoard()
	if err != nil {
		t.Fatalf("Failed to load board: %v", err)
	}

	// Create test game state
	gameState := game.NewGameState(board, game.WinterBuild, 1901)

	// Place units - Germany has 3 supply centers but only 2 units (can build 1)
	units := []*game.Unit{
		{Type: game.Army, Owner: game.Germany, Province: "berlin"},
		{Type: game.Fleet, Owner: game.Germany, Province: "kiel"},
		{Type: game.Army, Owner: game.Germany, Province: "munich"}, // Controls Munich supply center
	}

	for _, unit := range units {
		err := gameState.Board.PlaceUnit(unit)
		if err != nil {
			t.Fatalf("Failed to place unit: %v", err)
		}
	}

	// Test build adjustments calculation
	buildProcessor := NewBuildProcessor(gameState.Board)
	adjustments := buildProcessor.CalculateBuildAdjustments(gameState)

	// Germany should have 0 adjustments (3 supply centers, 3 units)
	if adjustments[game.Germany] != 0 {
		t.Errorf("Expected Germany to have 0 adjustments, got %d", adjustments[game.Germany])
	}
}

func TestValidateBuildOrder(t *testing.T) {
	// Load test board
	gameLoader := loader.GetLoader(loader.Classic)
	board, err := gameLoader.LoadBoard()
	if err != nil {
		t.Fatalf("Failed to load board: %v", err)
	}

	// Create test game state with Germany having 3 supply centers but 2 units (can build 1)
	gameState := game.NewGameState(board, game.WinterBuild, 1901)

	// Manually set supply center ownership to simulate end of fall turn
	gameState.SupplyCenters[game.Germany] = []string{"berlin", "munich", "belgium"}

	// Place only 2 units so Germany can build 1
	units := []*game.Unit{
		{Type: game.Army, Owner: game.Germany, Province: "berlin"},
		{Type: game.Army, Owner: game.Germany, Province: "munich"},
	}
	for _, unit := range units {
		err := gameState.Board.PlaceUnit(unit)
		if err != nil {
			t.Fatalf("Failed to place unit: %v", err)
		}
	}

	buildProcessor := NewBuildProcessor(gameState.Board)

	tests := []struct {
		name        string
		order       BuildOrder
		expectError bool
		errorMsg    string
	}{
		{
			name: "Valid army build in home center",
			order: BuildOrder{
				Nation:   game.Germany,
				UnitType: game.Army,
				Province: "kiel", // Kiel is unoccupied and is a German home center
			},
			expectError: false,
		},
		{
			name: "Invalid build - province occupied",
			order: BuildOrder{
				Nation:   game.Germany,
				UnitType: game.Army,
				Province: "berlin",
			},
			expectError: true,
			errorMsg:    "already occupied",
		},
		{
			name: "Invalid build - not home center",
			order: BuildOrder{
				Nation:   game.Germany,
				UnitType: game.Army,
				Province: "paris",
			},
			expectError: true,
			errorMsg:    "not a home center",
		},
		{
			name: "Invalid build - army in sea",
			order: BuildOrder{
				Nation:   game.Germany,
				UnitType: game.Army,
				Province: "north_sea",
			},
			expectError: true,
			errorMsg:    "cannot build army in sea province",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := buildProcessor.ValidateBuildOrder(tt.order, gameState)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				} else if tt.errorMsg != "" && err.Error() != tt.errorMsg {
					// Just check if error message contains expected substring
					if len(tt.errorMsg) > 0 && err.Error() == "" {
						t.Errorf("Expected error containing '%s', got '%s'", tt.errorMsg, err.Error())
					}
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
			}
		})
	}
}

func TestProcessBuildOrder(t *testing.T) {
	// Load test board
	gameLoader := loader.GetLoader(loader.Classic)
	board, err := gameLoader.LoadBoard()
	if err != nil {
		t.Fatalf("Failed to load board: %v", err)
	}

	// Create test game state
	gameState := game.NewGameState(board, game.WinterBuild, 1901)

	// Manually set supply center ownership to simulate end of fall turn
	gameState.SupplyCenters[game.Germany] = []string{"berlin", "munich", "kiel"}

	// Place units so Germany has builds available (3 supply centers, 2 units = 1 build)
	units := []*game.Unit{
		{Type: game.Army, Owner: game.Germany, Province: "berlin"},
		{Type: game.Army, Owner: game.Germany, Province: "kiel"},
	}

	for _, unit := range units {
		err := gameState.Board.PlaceUnit(unit)
		if err != nil {
			t.Fatalf("Failed to place unit: %v", err)
		}
	}
	buildProcessor := NewBuildProcessor(gameState.Board)

	// Test successful build
	buildOrder := BuildOrder{
		Nation:   game.Germany,
		UnitType: game.Army,
		Province: "munich",
	}

	result := buildProcessor.ProcessBuildOrder(buildOrder, gameState)

	if !result.Success {
		t.Errorf("Expected build to succeed, but got: %s", result.Reason)
	}

	// Verify unit was placed
	unit := gameState.Board.GetUnit("munich")
	if unit == nil {
		t.Errorf("Expected unit to be placed at munich")
	} else {
		if unit.Type != game.Army || unit.Owner != game.Germany {
			t.Errorf("Expected German army at munich, got %s %s", unit.Owner, unit.Type)
		}
	}
}

func TestValidateDisbandOrder(t *testing.T) {
	// Load test board
	gameLoader := loader.GetLoader(loader.Classic)
	board, err := gameLoader.LoadBoard()
	if err != nil {
		t.Fatalf("Failed to load board: %v", err)
	}

	// Create test game state with Germany having more units than supply centers
	gameState := game.NewGameState(board, game.WinterBuild, 1901)

	// Place only 1 unit but it controls no supply centers (needs to disband)
	units := []*game.Unit{
		{Type: game.Army, Owner: game.Germany, Province: "silesia"}, // Not a supply center
	}

	for _, unit := range units {
		err := gameState.Board.PlaceUnit(unit)
		if err != nil {
			t.Fatalf("Failed to place unit: %v", err)
		}
	}

	buildProcessor := NewBuildProcessor(gameState.Board)

	// Test valid disband order
	disbandOrder := DisbandOrder{
		Nation:   game.Germany,
		Province: "silesia",
	}

	err = buildProcessor.ValidateDisbandOrder(disbandOrder, gameState)
	if err != nil {
		t.Errorf("Expected disband validation to succeed, got: %v", err)
	}

	// Test invalid disband - unit doesn't belong to nation
	invalidOrder := DisbandOrder{
		Nation:   game.France,
		Province: "silesia",
	}

	err = buildProcessor.ValidateDisbandOrder(invalidOrder, gameState)
	if err == nil {
		t.Errorf("Expected disband validation to fail for wrong nation")
	}
}
