package integration

import (
	"diplomacy-cli/backend/internal/game"
	"diplomacy-cli/backend/internal/game/loader"
	"diplomacy-cli/backend/internal/game/resolution"
	"diplomacy-cli/backend/internal/game/validation"
	"testing"
)

// TestEndToEndPipeline tests the complete pipeline from raw orders to final game state
func TestEndToEndPipeline(t *testing.T) {
	// Load the classic map
	mapLoader := loader.NewJSONLoader("../../data/classic")
	board, err := mapLoader.LoadBoard()
	if err != nil {
		t.Fatalf("Failed to load board: %v", err)
	}

	// Create initial game state with starting units
	gameState := game.NewGameState(board, game.SpringMovement, 1901)

	// Load starting units
	startingUnits, err := mapLoader.LoadStartingUnits()
	if err != nil {
		t.Fatalf("Failed to load starting units: %v", err)
	}

	// Place starting units on the board
	for _, unit := range startingUnits {
		gameState.Board.PlaceUnit(unit)
	}

	// Add some simple raw orders for testing
	rawOrders := map[game.Nation][]string{
		game.France: {
			"A Par H",
			"F Bre H",
			"A Mar H",
		},
		game.Germany: {
			"A Ber H",
			"F Kie H",
			"A Mun H",
		},
		game.England: {
			"F Lon H",
			"A Lvp H",
			"F Edi H",
		},
	}

	// Add raw orders to game state
	for nation, orders := range rawOrders {
		for _, orderText := range orders {
			err := gameState.AddRawOrder(nation, orderText)
			if err != nil {
				t.Errorf("Failed to add raw order %s for %s: %v", orderText, nation, err)
			}
		}
	}

	// Test the complete pipeline
	newState, err := ProcessTurn(gameState)
	if err != nil {
		t.Fatalf("Failed to process turn: %v", err)
	}

	// Verify the new state
	if newState == nil {
		t.Fatal("ProcessTurn returned nil state")
	}

	// Verify phase advancement logic
	if newState.Phase != game.FallMovement {
		t.Errorf("Expected phase %s, got %s", game.FallMovement, newState.Phase)
	}

	if newState.Year != 1901 {
		t.Errorf("Expected year 1901, got %d", newState.Year)
	}

	// Verify that raw orders were cleared
	if len(newState.RawOrders) != 0 {
		t.Error("Raw orders should be cleared after processing")
	}

	// Verify that units moved (basic check)
	// This is a simple check - in a real scenario we'd verify specific unit positions
	unitCount := 0
	for _, unit := range newState.Board.Units {
		if unit != nil {
			unitCount++
		}
	}

	if unitCount == 0 {
		t.Error("No units found after processing turn")
	}
}

// TestMultiTurnProgression tests advancing through multiple turns
func TestMultiTurnProgression(t *testing.T) {
	// Load the classic map
	mapLoader := loader.NewJSONLoader("../../data/classic")
	board, err := mapLoader.LoadBoard()
	if err != nil {
		t.Fatalf("Failed to load board: %v", err)
	}

	// Create initial game state
	gameState := game.NewGameState(board, game.SpringMovement, 1901)

	// Load starting units
	startingUnits, err := mapLoader.LoadStartingUnits()
	if err != nil {
		t.Fatalf("Failed to load starting units: %v", err)
	}

	// Place starting units
	for _, unit := range startingUnits {
		gameState.Board.PlaceUnit(unit)
	}

	// Test progression through multiple phases
	phases := []struct {
		expectedPhase game.Phase
		expectedYear  int
		orders        map[game.Nation][]string
	}{
		{
			expectedPhase: game.SpringMovement,
			expectedYear:  1901,
			orders: map[game.Nation][]string{
				game.France:  {"A Par H", "F Bre H", "A Mar H"},
				game.Germany: {"A Ber H", "F Kie H", "A Mun H"},
				game.England: {"F Lon H", "A Lvp H", "F Edi H"},
			},
		},
		{
			expectedPhase: game.FallMovement,
			expectedYear:  1901,
			orders: map[game.Nation][]string{
				game.France:  {"A Par H", "F Bre H", "A Mar H"},
				game.Germany: {"A Ber H", "F Kie H", "A Mun H"},
				game.England: {"F Lon H", "A Lvp H", "F Edi H"},
			},
		},
		{
			expectedPhase: game.WinterBuild,
			expectedYear:  1902,
			orders:        map[game.Nation][]string{}, // No orders for build phase in this test
		},
		{
			expectedPhase: game.SpringMovement,
			expectedYear:  1902,
			orders: map[game.Nation][]string{
				game.France:  {"A Par H", "F Bre H", "A Mar H"},
				game.Germany: {"A Ber H", "F Kie H", "A Mun H"},
				game.England: {"F Lon H", "A Lvp H", "F Edi H"},
			},
		},
	}

	currentState := gameState
	for i, phase := range phases {
		t.Run(string(phase.expectedPhase), func(t *testing.T) {
			// Verify current phase and year
			if currentState.Phase != phase.expectedPhase {
				t.Errorf("Turn %d: Expected phase %s, got %s", i, phase.expectedPhase, currentState.Phase)
			}

			if currentState.Year != phase.expectedYear {
				t.Errorf("Turn %d: Expected year %d, got %d", i, phase.expectedYear, currentState.Year)
			}

			// Add orders if this is a movement phase
			if phase.expectedPhase == game.SpringMovement || phase.expectedPhase == game.FallMovement {
				for nation, orders := range phase.orders {
					for _, orderText := range orders {
						err := currentState.AddRawOrder(nation, orderText)
						if err != nil {
							t.Errorf("Turn %d: Failed to add order %s for %s: %v", i, orderText, nation, err)
						}
					}
				}

				// Process the turn
				newState, err := ProcessTurn(currentState)
				if err != nil {
					t.Fatalf("Turn %d: Failed to process turn: %v", i, err)
				}

				currentState = newState
			} else {
				// For non-movement phases, just advance without processing orders
				currentState.AdvanceToNextPhase([]*game.Unit{}) // No dislodged units for this test
			}
		})
	}
}

// ProcessTurn implements the complete pipeline from raw orders to new game state
func ProcessTurn(gameState *game.GameState) (*game.GameState, error) {
	// Step 1: Parse all raw orders
	var allOrders []*game.Order
	resolver := validation.NewProvinceResolver(gameState.Board)
	registry := validation.NewOrderParserRegistry()

	for nation, rawOrders := range gameState.RawOrders {
		for _, rawOrder := range rawOrders {
			// Tokenize the order
			tokens := validation.Tokenize(rawOrder)

			// Parse the order
			order, err := registry.ParseOrder(tokens, gameState.Phase, resolver)
			if err != nil {
				return nil, err
			}

			// Set the owner
			order.Owner = nation
			allOrders = append(allOrders, order)
		}
	}

	// Step 2: Resolve orders using the resolution engine
	engine := resolution.NewResolutionEngine(allOrders, gameState.Board)
	err := engine.Resolve()
	if err != nil {
		return nil, err
	}

	// Step 3: Apply results to create new game state
	newState := gameState.Clone()

	// Get resolution results
	results := engine.GetResults()

	// Apply unit movements and dislodgements
	var dislodgedUnits []*game.Unit
	for i, order := range allOrders {
		if i >= len(results) {
			continue
		}

		result := results[i]

		// Handle successful moves
		if order.Type == game.Move && result.Result == resolution.MoveSuccess {
			unit := newState.Board.GetUnit(order.From)
			if unit != nil {
				// Remove unit from old position
				newState.Board.RemoveUnit(order.From)

				// Place unit in new position
				unit.Province = order.To
				unit.Coast = order.ToCoast
				newState.Board.PlaceUnit(unit)
			}
		}

		// Handle dislodgements
		if result.Dislodged {
			unit := newState.Board.GetUnit(order.From)
			if unit != nil {
				dislodgedUnits = append(dislodgedUnits, unit)
				newState.Board.RemoveUnit(order.From)
			}
		}
	}

	// Step 4: Advance to next phase
	newState.AdvanceToNextPhase(dislodgedUnits)

	return newState, nil
}

// TestProcessTurnWithConflicts tests the pipeline with conflicting orders
func TestProcessTurnWithConflicts(t *testing.T) {
	// Create a simple test board
	board := game.NewBoard()

	// Add test provinces
	board.AddProvince(&game.Province{
		Name:          "paris",
		ShortCode:     "par",
		DisplayName:   "Paris",
		Type:          game.Land,
		SupplyCenter:  true,
		ArmyNeighbors: []string{"burgundy", "picardy"},
	})

	board.AddProvince(&game.Province{
		Name:          "burgundy",
		ShortCode:     "bur",
		DisplayName:   "Burgundy",
		Type:          game.Land,
		SupplyCenter:  false,
		ArmyNeighbors: []string{"paris", "munich"},
	})

	board.AddProvince(&game.Province{
		Name:          "munich",
		ShortCode:     "mun",
		DisplayName:   "Munich",
		Type:          game.Land,
		SupplyCenter:  true,
		ArmyNeighbors: []string{"burgundy"},
	})

	// Place units
	board.PlaceUnit(&game.Unit{Type: game.Army, Owner: game.France, Province: "paris"})
	board.PlaceUnit(&game.Unit{Type: game.Army, Owner: game.Germany, Province: "munich"})

	// Create game state
	gameState := game.NewGameState(board, game.SpringMovement, 1901)

	// Add conflicting orders (both armies move to burgundy)
	gameState.AddRawOrder(game.France, "A Par-Bur")
	gameState.AddRawOrder(game.Germany, "A Mun-Bur")

	// Process the turn
	newState, err := ProcessTurn(gameState)
	if err != nil {
		t.Fatalf("Failed to process turn with conflicts: %v", err)
	}

	// Verify that both units bounced (stayed in original positions)
	frenchUnit := newState.Board.GetUnit("paris")
	germanUnit := newState.Board.GetUnit("munich")
	burgundyUnit := newState.Board.GetUnit("burgundy")

	if frenchUnit == nil {
		t.Error("French unit should still be in Paris after bounce")
	}

	if germanUnit == nil {
		t.Error("German unit should still be in Munich after bounce")
	}

	if burgundyUnit != nil {
		t.Error("No unit should be in Burgundy after bounce")
	}
}
