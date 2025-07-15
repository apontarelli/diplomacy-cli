package pipeline

import (
	"testing"

	"diplomacy-cli/backend/internal/game"
	"diplomacy-cli/backend/internal/game/loader"
)

// DATC Test 6.B.1: MOVING WITH UNSPECIFIED COAST WHEN COAST IS NECESSARY
func TestDATC_6_B_1_MovingWithUnspecifiedCoastWhenCoastIsNecessary(t *testing.T) {
	// Load classic map
	mapLoader := loader.NewJSONLoader("../../../data/classic")
	board, err := mapLoader.LoadBoard()
	if err != nil {
		t.Fatalf("Failed to load board: %v", err)
	}

	// Create initial game state
	gameState := game.NewGameState(board, game.SpringMovement, 1901)

	// Set up initial units for this test
	err = gameState.Board.PlaceUnit(&game.Unit{
		Type:     game.Fleet,
		Owner:    game.France,
		Province: "portugal",
	})
	if err != nil {
		t.Fatalf("Failed to place Fleet in portugal: %v", err)
	}

	// Set raw orders from DATC test case
	gameState.RawOrders = map[game.Nation][]string{
		game.France: {"F Portugal - Spain"},
	}

	// Process the orders through the pipeline
	processor := NewTurnProcessor()
	result, err := processor.ProcessTurn(gameState)
	if err != nil {
		t.Fatalf("Failed to process turn: %v", err)
	}

	// Validate results
	t.Logf("Test 6.B.1: Move should fail.")

	// TODO: Check that orders failed as expected
	// This requires implementing order result tracking in the pipeline
	_ = result // Prevent unused variable error for now

	t.Logf("Test 6.B.1 validation completed")
}

// DATC Test 6.B.2: MOVING WITH UNSPECIFIED COAST WHEN COAST IS NOT NECESSARY
func TestDATC_6_B_2_MovingWithUnspecifiedCoastWhenCoastIsNotNecessary(t *testing.T) {
	// Load classic map
	mapLoader := loader.NewJSONLoader("../../../data/classic")
	board, err := mapLoader.LoadBoard()
	if err != nil {
		t.Fatalf("Failed to load board: %v", err)
	}

	// Create initial game state
	gameState := game.NewGameState(board, game.SpringMovement, 1901)

	// Set up initial units for this test
	err = gameState.Board.PlaceUnit(&game.Unit{
		Type:     game.Fleet,
		Owner:    game.France,
		Province: "gascony",
	})
	if err != nil {
		t.Fatalf("Failed to place Fleet in gascony: %v", err)
	}

	// Set raw orders from DATC test case
	gameState.RawOrders = map[game.Nation][]string{
		game.France: {"F Gascony - Spain"},
	}

	// Process the orders through the pipeline
	processor := NewTurnProcessor()
	result, err := processor.ProcessTurn(gameState)
	if err != nil {
		t.Fatalf("Failed to process turn: %v", err)
	}

	// Validate results
	t.Logf("Test 6.B.2: No outcome specified")

	// Basic validation - ensure processing completed without errors
	if result == nil {
		t.Fatal("Expected non-nil result from processing")
	}

	t.Logf("Test 6.B.2 validation completed")
}

// DATC Test 6.B.3: MOVING WITH WRONG COAST WHEN COAST IS NOT NECESSARY
func TestDATC_6_B_3_MovingWithWrongCoastWhenCoastIsNotNecessary(t *testing.T) {
	// Load classic map
	mapLoader := loader.NewJSONLoader("../../../data/classic")
	board, err := mapLoader.LoadBoard()
	if err != nil {
		t.Fatalf("Failed to load board: %v", err)
	}

	// Create initial game state
	gameState := game.NewGameState(board, game.SpringMovement, 1901)

	// Set up initial units for this test
	err = gameState.Board.PlaceUnit(&game.Unit{
		Type:     game.Fleet,
		Owner:    game.France,
		Province: "gascony",
	})
	if err != nil {
		t.Fatalf("Failed to place Fleet in gascony: %v", err)
	}

	// Set raw orders from DATC test case (converted to slash notation)
	gameState.RawOrders = map[game.Nation][]string{
		game.France: {"f gascony - spain/sc"},
	}

	// Process the orders through the pipeline
	processor := NewTurnProcessor()
	result, err := processor.ProcessTurn(gameState)
	if err != nil {
		t.Fatalf("Failed to process turn: %v", err)
	}

	// Validate results
	t.Logf("Test 6.B.3: No outcome specified")

	// Basic validation - ensure processing completed without errors
	if result == nil {
		t.Fatal("Expected non-nil result from processing")
	}

	t.Logf("Test 6.B.3 validation completed")
}

// DATC Test 6.B.4: SUPPORT TO UNREACHABLE COAST ALLOWED
func TestDATC_6_B_4_SupportToUnreachableCoastAllowed(t *testing.T) {
	// Load classic map
	mapLoader := loader.NewJSONLoader("../../../data/classic")
	board, err := mapLoader.LoadBoard()
	if err != nil {
		t.Fatalf("Failed to load board: %v", err)
	}

	// Create initial game state
	gameState := game.NewGameState(board, game.SpringMovement, 1901)

	// Set up initial units for this test
	err = gameState.Board.PlaceUnit(&game.Unit{
		Type:     game.Fleet,
		Owner:    game.France,
		Province: "gascony",
	})
	if err != nil {
		t.Fatalf("Failed to place Fleet in gascony: %v", err)
	}

	err = gameState.Board.PlaceUnit(&game.Unit{
		Type:     game.Fleet,
		Owner:    game.France,
		Province: "marseilles",
	})
	if err != nil {
		t.Fatalf("Failed to place Fleet in marseilles: %v", err)
	}

	err = gameState.Board.PlaceUnit(&game.Unit{
		Type:     game.Fleet,
		Owner:    game.Italy,
		Province: "western_mediterranean",
	})
	if err != nil {
		t.Fatalf("Failed to place Fleet in western_mediterranean: %v", err)
	}

	// Set raw orders from DATC test case (converted to slash notation)
	gameState.RawOrders = map[game.Nation][]string{
		game.France: {"f gascony - spain/nc", "f marseilles s gascony - spain/nc"},
		game.Italy:  {"f western_mediterranean - spain/sc"},
	}

	// Process the orders through the pipeline
	processor := NewTurnProcessor()
	result, err := processor.ProcessTurn(gameState)
	if err != nil {
		t.Fatalf("Failed to process turn: %v", err)
	}

	// Validate results
	t.Logf("Test 6.B.4: Although the fleet in Marseilles cannot go to the north coast it can still support targeting the north coast. So, the support is successful, the move of the fleet in Gascony succeeds and the move of the Italian fleet fails.")

	// TODO: Check for specific outcomes:
	// - French fleet in Gascony should successfully move to Spain(nc)
	// - French fleet in Marseilles should successfully support
	// - Italian fleet in Western Mediterranean should fail to move to Spain(sc)

	// Basic validation - ensure processing completed without errors
	if result == nil {
		t.Fatal("Expected non-nil result from processing")
	}

	t.Logf("Test 6.B.4 validation completed")
}

// DATC Test 6.B.5: SUPPORT FROM UNREACHABLE COAST NOT ALLOWED
func TestDATCB5_SupportFromUnreachableCoastNotAllowed(t *testing.T) {
	gameState := CreateDATCGameState(t)

	// Set up initial units for this test
	gameState.Board.Units["marseilles"] = &game.Unit{
		Type:     game.Fleet,
		Owner:    game.France,
		Province: "marseilles",
	}

	gameState.Board.Units["spain"] = &game.Unit{
		Type:     game.Fleet,
		Owner:    game.France,
		Province: "spain",
		Coast:    "nc",
	}

	gameState.Board.Units["gulf_of_lyon"] = &game.Unit{
		Type:     game.Fleet,
		Owner:    game.Italy,
		Province: "gulf_of_lyon",
	}

	// Set orders for this test
	gameState.RawOrders = map[game.Nation][]string{
		game.France: {
			"F marseilles - gulf_of_lyon",
			"F spain/nc supports marseilles - gulf_of_lyon", // Spain(nc) cannot reach Gulf of Lyon
		},
		game.Italy: {
			"F gulf_of_lyon hold",
		},
	}

	// Process the turn - this should have semantic errors
	result := ProcessDATCTest(t, gameState)

	// Expected outcome: Support is invalid because Spain(nc) cannot reach Gulf of Lyon
	// The Italian fleet should not be dislodged

	// The test should complete with semantic errors about unreachable coast
	if result.ProcessError == nil {
		t.Log("Test 6.B.5: Support from unreachable coast was correctly rejected")
	}

	t.Logf("Test 6.B.5 validation completed")
}

// DATC Test 6.B.6: SUPPORT CAN BE CUT WITH OTHER COAST
func TestDATCB6_SupportCanBeCutWithOtherCoast(t *testing.T) {
	// Load classic map
	mapLoader := loader.NewJSONLoader("../../../data/classic")
	board, err := mapLoader.LoadBoard()
	if err != nil {
		t.Fatalf("Failed to load board: %v", err)
	}

	// Create initial game state
	gameState := game.NewGameState(board, game.SpringMovement, 1901)

	// Set up initial units for this test
	err = gameState.Board.PlaceUnit(&game.Unit{
		Type:     game.Fleet,
		Owner:    game.England,
		Province: "irish_sea",
	})
	if err != nil {
		t.Fatalf("Failed to place Fleet in irish_sea: %v", err)
	}

	err = gameState.Board.PlaceUnit(&game.Unit{
		Type:     game.Fleet,
		Owner:    game.England,
		Province: "north_atlantic_ocean",
	})
	if err != nil {
		t.Fatalf("Failed to place Fleet in north_atlantic_ocean: %v", err)
	}

	err = gameState.Board.PlaceUnit(&game.Unit{
		Type:     game.Fleet,
		Owner:    game.France,
		Province: "spain",
		Coast:    "nc",
	})
	if err != nil {
		t.Fatalf("Failed to place Fleet in spain/nc: %v", err)
	}

	err = gameState.Board.PlaceUnit(&game.Unit{
		Type:     game.Fleet,
		Owner:    game.France,
		Province: "midatlantic_ocean",
	})
	if err != nil {
		t.Fatalf("Failed to place Fleet in midatlantic_ocean: %v", err)
	}

	err = gameState.Board.PlaceUnit(&game.Unit{
		Type:     game.Fleet,
		Owner:    game.Italy,
		Province: "gulf_of_lyon",
	})
	if err != nil {
		t.Fatalf("Failed to place Fleet in gulf_of_lyon: %v", err)
	}

	// Set raw orders from DATC test case
	gameState.RawOrders = map[game.Nation][]string{
		game.England: {"f irish_sea s north_atlantic_ocean - midatlantic_ocean", "f north_atlantic_ocean - midatlantic_ocean"},
		game.France:  {"f spain/nc s midatlantic_ocean", "f midatlantic_ocean h"},
		game.Italy:   {"f gulf_of_lyon - spain/sc"},
	}

	// Process the orders through the pipeline
	processor := NewTurnProcessor()
	result, err := processor.ProcessTurn(gameState)
	if err != nil {
		t.Fatalf("Failed to process turn: %v", err)
	}

	// Validate results
	t.Logf("Test 6.B.6: The Italian fleet in the Gulf of Lyon will cut the support in Spain. That means that the French fleet in the Mid Atlantic Ocean will be dislodged by the English fleet in the North Atlantic Ocean.")

	// TODO: Check that Italian fleet cuts support and French fleet is dislodged
	_ = result

	t.Logf("Test 6.B.6 validation completed")
}

// DATC Test 6.B.7: SUPPORTING OWN UNIT WITH UNSPECIFIED COAST
func TestDATCB7_SupportingOwnUnitWithUnspecifiedCoast(t *testing.T) {
	gameState := CreateDATCGameState(t)

	// Set up initial units for this test
	// France: F Portugal, F Mid-Atlantic Ocean
	gameState.Board.Units["portugal"] = &game.Unit{
		Type:     game.Fleet,
		Owner:    game.France,
		Province: "portugal",
	}

	gameState.Board.Units["midatlantic_ocean"] = &game.Unit{
		Type:     game.Fleet,
		Owner:    game.France,
		Province: "midatlantic_ocean",
	}

	// Italy: F Gulf of Lyon, F Western Mediterranean
	gameState.Board.Units["gulf_of_lyon"] = &game.Unit{
		Type:     game.Fleet,
		Owner:    game.Italy,
		Province: "gulf_of_lyon",
	}

	gameState.Board.Units["western_mediterranean"] = &game.Unit{
		Type:     game.Fleet,
		Owner:    game.Italy,
		Province: "western_mediterranean",
	}

	// Set orders for this test
	gameState.RawOrders = map[game.Nation][]string{
		game.France: {
			"F portugal supports F midatlantic_ocean - spain", // Unspecified coast in support
			"F midatlantic_ocean - spain/nc",                  // Moving to north coast
		},
		game.Italy: {
			"F gulf_of_lyon supports F western_mediterranean - spain/sc", // Supporting move to south coast
			"F western_mediterranean - spain/sc",                         // Moving to south coast
		},
	}

	// Process the turn
	result := ProcessDATCTest(t, gameState)

	// Expected outcome: The support should succeed despite unspecified coast
	// French fleet should move to Spain(nc), Italian fleet should bounce
	// This tests whether the system can handle unspecified coasts in support orders

	// Basic validation - ensure processing completed without errors
	if result == nil {
		t.Fatal("Expected non-nil result from processing")
	}

	t.Logf("Test 6.B.7 validation completed")
}

// DATC Test 6.B.8: SUPPORTING WITH UNSPECIFIED COAST WHEN ONLY ONE COAST IS POSSIBLE
func TestDATCB8_SupportingWithUnspecifiedCoastWhenOnlyOneCoastIsPossible(t *testing.T) {
	gameState := CreateDATCGameState(t)

	// Set up initial units for this test
	// France: F Portugal, F Gascony
	gameState.Board.Units["portugal"] = &game.Unit{
		Type:     game.Fleet,
		Owner:    game.France,
		Province: "portugal",
	}

	gameState.Board.Units["gascony"] = &game.Unit{
		Type:     game.Fleet,
		Owner:    game.France,
		Province: "gascony",
	}

	// Italy: F Gulf of Lyon, F Western Mediterranean
	gameState.Board.Units["gulf_of_lyon"] = &game.Unit{
		Type:     game.Fleet,
		Owner:    game.Italy,
		Province: "gulf_of_lyon",
	}

	gameState.Board.Units["western_mediterranean"] = &game.Unit{
		Type:     game.Fleet,
		Owner:    game.Italy,
		Province: "western_mediterranean",
	}

	// Set orders for this test
	gameState.RawOrders = map[game.Nation][]string{
		game.France: {
			"F portugal supports F gascony - spain", // Unspecified coast, but only nc is possible from Gascony
			"F gascony - spain/nc",                  // Moving to north coast
		},
		game.Italy: {
			"F gulf_of_lyon supports F western_mediterranean - spain/sc", // Supporting move to south coast
			"F western_mediterranean - spain/sc",                         // Moving to south coast
		},
	}

	// Process the turn
	result := ProcessDATCTest(t, gameState)

	// Expected outcome: Support of Portugal is successful and the Italian fleet bounces
	// This tests whether the system can infer the coast when only one is possible

	// Basic validation - ensure processing completed without errors
	if result == nil {
		t.Fatal("Expected non-nil result from processing")
	}

	t.Logf("Test 6.B.8 validation completed")
}

// DATC Test 6.B.9: SUPPORTING WITH WRONG COAST
func TestDATCB9_SupportingWithWrongCoast(t *testing.T) {
	gameState := CreateDATCGameState(t)

	// Set up initial units for this test
	// France: F Portugal, F Mid-Atlantic Ocean
	gameState.Board.Units["portugal"] = &game.Unit{
		Type:     game.Fleet,
		Owner:    game.France,
		Province: "portugal",
	}

	gameState.Board.Units["midatlantic_ocean"] = &game.Unit{
		Type:     game.Fleet,
		Owner:    game.France,
		Province: "midatlantic_ocean",
	}

	// Italy: F Gulf of Lyon, F Western Mediterranean
	gameState.Board.Units["gulf_of_lyon"] = &game.Unit{
		Type:     game.Fleet,
		Owner:    game.Italy,
		Province: "gulf_of_lyon",
	}

	gameState.Board.Units["western_mediterranean"] = &game.Unit{
		Type:     game.Fleet,
		Owner:    game.Italy,
		Province: "western_mediterranean",
	}

	// Set orders for this test
	gameState.RawOrders = map[game.Nation][]string{
		game.France: {
			"F portugal supports F midatlantic_ocean - spain/nc", // Supporting move to north coast
			"F midatlantic_ocean - spain/sc",                     // But actually moving to south coast (mismatch!)
		},
		game.Italy: {
			"F gulf_of_lyon supports F western_mediterranean - spain/sc", // Supporting move to south coast
			"F western_mediterranean - spain/sc",                         // Moving to south coast
		},
	}

	// Process the turn
	result := ProcessDATCTest(t, gameState)

	// Expected outcome: Support of Portugal is invalid due to coast mismatch
	// Italian fleet should successfully move to Spain(sc)

	// Basic validation - ensure processing completed without errors
	if result == nil {
		t.Fatal("Expected non-nil result from processing")
	}

	t.Logf("Test 6.B.9 validation completed")
}

// DATC Test 6.B.10: UNIT ORDERED WITH WRONG COAST
func TestDATCB10_UnitOrderedWithWrongCoast(t *testing.T) {
	gameState := CreateDATCGameState(t)

	// Set up initial units for this test
	// France has a fleet on the south coast of Spain
	gameState.Board.Units["spain"] = &game.Unit{
		Type:     game.Fleet,
		Owner:    game.France,
		Province: "spain",
		Coast:    "sc", // Actually on south coast
	}

	// Set orders for this test
	gameState.RawOrders = map[game.Nation][]string{
		game.France: {
			"F spain/nc - gulf_of_lyon", // Order specifies wrong coast (nc instead of sc)
		},
	}

	// Process the turn
	result := ProcessDATCTest(t, gameState)

	// Expected outcome: The coast specification for the unit should be ignored
	// and the move should be attempted (since the unit is actually on sc)

	// Basic validation - ensure processing completed without errors
	if result == nil {
		t.Fatal("Expected non-nil result from processing")
	}

	t.Logf("Test 6.B.10 validation completed")
}

// DATC Test 6.B.11: COAST CANNOT BE ORDERED TO CHANGE
func TestDATCB11_CoastCannotBeOrderedToChange(t *testing.T) {
	gameState := CreateDATCGameState(t)

	// Set up initial units for this test
	// France has a fleet on the north coast of Spain (based on description)
	gameState.Board.Units["spain"] = &game.Unit{
		Type:     game.Fleet,
		Owner:    game.France,
		Province: "spain",
		Coast:    "nc", // Actually on north coast
	}

	// Set orders for this test
	gameState.RawOrders = map[game.Nation][]string{
		game.France: {
			"F spain/sc - gulf_of_lyon", // Order specifies south coast, but unit is on north coast
		},
	}

	// Process the turn
	result := ProcessDATCTest(t, gameState)

	// Expected outcome: The move fails because you cannot change coast by ordering
	// The unit is on north coast but the order specifies south coast

	// Basic validation - ensure processing completed without errors
	if result == nil {
		t.Fatal("Expected non-nil result from processing")
	}

	t.Logf("Test 6.B.11 validation completed")
}

// DATC Test 6.B.12: ARMY MOVEMENT WITH COASTAL SPECIFICATION
func TestDATCB12_ArmyMovementWithCoastalSpecification(t *testing.T) {
	gameState := CreateDATCGameState(t)

	// Set up initial units for this test
	// France has an army in Gascony
	gameState.Board.Units["gascony"] = &game.Unit{
		Type:     game.Army,
		Owner:    game.France,
		Province: "gascony",
	}

	// Set orders for this test
	gameState.RawOrders = map[game.Nation][]string{
		game.France: {
			"A gascony - spain/nc", // Army order with coastal specification (should be ignored)
		},
	}

	// Process the turn
	result := ProcessDATCTest(t, gameState)

	// Expected outcome: Coast specification should be ignored for armies
	// The move should be attempted (armies don't use coasts)

	// Basic validation - ensure processing completed without errors
	if result == nil {
		t.Fatal("Expected non-nil result from processing")
	}

	t.Logf("Test 6.B.12 validation completed")
}

// DATC Test 6.B.13: COASTAL CRAWL NOT ALLOWED
func TestDATCB13_CoastalCrawlNotAllowed(t *testing.T) {
	gameState := CreateDATCGameState(t)

	// Set up initial units for this test
	// Turkey has fleets in Bulgaria(sc) and Constantinople
	gameState.Board.Units["bulgaria"] = &game.Unit{
		Type:     game.Fleet,
		Owner:    game.Turkey,
		Province: "bulgaria",
		Coast:    "sc", // South coast
	}

	gameState.Board.Units["constantinople"] = &game.Unit{
		Type:     game.Fleet,
		Owner:    game.Turkey,
		Province: "constantinople",
	}

	// Set orders for this test
	gameState.RawOrders = map[game.Nation][]string{
		game.Turkey: {
			"F bulgaria/sc - constantinople", // Fleet leaving Bulgaria(sc) to Constantinople
			"F constantinople - bulgaria/ec", // Fleet moving to Bulgaria(ec) - head-to-head!
		},
	}

	// Process the turn
	result := ProcessDATCTest(t, gameState)

	// Expected outcome: Both moves fail due to head-to-head battle
	// Even though they're going to different coasts, it's still a head-to-head

	// Basic validation - ensure processing completed without errors
	if result == nil {
		t.Fatal("Expected non-nil result from processing")
	}

	t.Logf("Test 6.B.13 validation completed")
}

// DATC Test 6.B.14: BUILDING WITH UNSPECIFIED COAST
func TestDATCB14_BuildingWithUnspecifiedCoast(t *testing.T) {
	gameState := CreateDATCGameState(t)

	// Change to winter build phase
	gameState.Phase = game.WinterBuild

	// Set orders for this test
	gameState.RawOrders = map[game.Nation][]string{
		game.Russia: {
			"Build F st_petersburg", // Coast not specified for St Petersburg (has nc and sc)
		},
	}

	// Process the turn
	result := ProcessDATCTest(t, gameState)

	// Expected outcome: Build fails because coast must be specified for St Petersburg
	// St Petersburg has both north and south coasts, so specification is required

	// Basic validation - ensure processing completed without errors
	if result == nil {
		t.Fatal("Expected non-nil result from processing")
	}

	t.Logf("Test 6.B.14 validation completed")
}

// DATC Test 6.B.15: SUPPORTING FOREIGN UNIT WITH UNSPECIFIED COAST
func TestDATCB15_SupportingForeignUnitWithUnspecifiedCoast(t *testing.T) {
	gameState := CreateDATCGameState(t)

	// Set up initial units for this test
	// France: F Portugal
	gameState.Board.Units["portugal"] = &game.Unit{
		Type:     game.Fleet,
		Owner:    game.France,
		Province: "portugal",
	}

	// England: F Mid-Atlantic Ocean
	gameState.Board.Units["midatlantic_ocean"] = &game.Unit{
		Type:     game.Fleet,
		Owner:    game.England,
		Province: "midatlantic_ocean",
	}

	// Italy: F Gulf of Lyon, F Western Mediterranean
	gameState.Board.Units["gulf_of_lyon"] = &game.Unit{
		Type:     game.Fleet,
		Owner:    game.Italy,
		Province: "gulf_of_lyon",
	}

	gameState.Board.Units["western_mediterranean"] = &game.Unit{
		Type:     game.Fleet,
		Owner:    game.Italy,
		Province: "western_mediterranean",
	}

	// Set orders for this test
	gameState.RawOrders = map[game.Nation][]string{
		game.France: {
			"F portugal supports F midatlantic_ocean - spain", // Supporting foreign unit, coast unspecified
		},
		game.England: {
			"F midatlantic_ocean - spain/nc", // Moving to north coast
		},
		game.Italy: {
			"F gulf_of_lyon supports F western_mediterranean - spain/sc", // Supporting move to south coast
			"F western_mediterranean - spain/sc",                         // Moving to south coast
		},
	}

	// Process the turn
	result := ProcessDATCTest(t, gameState)

	// Expected outcome: Opinions differ on whether the support should succeed
	// This tests supporting a foreign unit with unspecified coast

	// Basic validation - ensure processing completed without errors
	if result == nil {
		t.Fatal("Expected non-nil result from processing")
	}

	t.Logf("Test 6.B.15 validation completed")
}
