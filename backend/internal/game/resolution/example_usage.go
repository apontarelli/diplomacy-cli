package resolution

import (
	"diplomacy-cli/backend/internal/game"
	"fmt"
	"log"
)

// ExampleUsage demonstrates how to use the Resolution Engine Interface
func ExampleUsage() {
	// Create a new DATC-compliant engine with default configuration
	engine := NewDefaultDATCEngine()

	// Get engine information
	info := engine.GetEngineInfo()
	fmt.Printf("Using engine: %s v%s\n", info.Name, info.Version)
	fmt.Printf("Features: %v\n", info.Features)

	// Create a simple board setup
	board := createExampleBoard()

	// Define some orders
	orders := []game.Order{
		{
			ID:       "order_1",
			UnitType: game.Army,
			From:     "munich",
			Type:     game.Move,
			To:       "berlin",
			Owner:    game.Germany,
		},
		{
			ID:                 "order_2",
			UnitType:           game.Army,
			From:               "paris",
			Type:               game.Support,
			SupportTarget:      "munich",
			SupportDestination: "berlin",
			Owner:              game.France,
		},
		{
			ID:       "order_3",
			UnitType: game.Fleet,
			From:     "north_sea",
			Type:     game.Hold,
			Owner:    game.England,
		},
	}

	// First, validate the orders
	fmt.Println("\n=== Order Validation ===")
	validationErrors, err := engine.ValidateOrders(orders, board)
	if err != nil {
		log.Fatalf("Validation failed: %v", err)
	}

	if len(validationErrors) > 0 {
		fmt.Printf("Found %d validation errors:\n", len(validationErrors))
		for _, validationError := range validationErrors {
			fmt.Printf("  - %s\n", validationError.Error())
		}
	} else {
		fmt.Println("All orders are valid!")
	}

	// Resolve the orders
	fmt.Println("\n=== Order Resolution ===")
	result, err := engine.Resolve(orders, board)
	if err != nil {
		log.Fatalf("Resolution failed: %v", err)
	}

	// Display the results
	fmt.Printf("Resolution completed for turn %d, phase %s\n", result.TurnNumber(), result.Phase())

	unitOutcomes := result.UnitOutcomes()
	fmt.Printf("Unit outcomes (%d total):\n", len(unitOutcomes))

	for unitID, outcome := range unitOutcomes {
		fmt.Printf("  - %s %s at %s: %s",
			unitID.Type, unitID.Owner, unitID.Province,
			formatUnitStatus(outcome.FinalStatus()))

		if outcome.ToProvince() != nil {
			fmt.Printf(" -> %s", *outcome.ToProvince())
		}

		fmt.Printf(" (%s)\n", outcome.Reason())
	}

	// Display conflicts if any
	conflicts := result.Conflicts()
	if len(conflicts) > 0 {
		fmt.Printf("\nConflicts (%d total):\n", len(conflicts))
		for _, conflict := range conflicts {
			fmt.Printf("  - Province %s: %s\n", conflict.Province(), conflict.Explanation())
		}
	}

	// Display support outcomes
	supportOutcomes := result.SupportOutcomes()
	if len(supportOutcomes) > 0 {
		fmt.Printf("\nSupport outcomes (%d total):\n", len(supportOutcomes))
		for supportID, outcome := range supportOutcomes {
			status := "successful"
			if outcome.IsCut() {
				status = "cut"
			} else if !outcome.IsSuccessful() {
				status = "failed"
			}
			fmt.Printf("  - %s %s supporting %s %s: %s (%s)\n",
				supportID.SupportingUnit.Type, supportID.SupportingUnit.Owner,
				supportID.SupportTarget.Type, supportID.SupportTarget.Owner,
				status, outcome.Reason())
		}
	}
}

// ExampleWithValidationErrors demonstrates validation error handling
func ExampleWithValidationErrors() {
	engine := NewDefaultDATCEngine()
	board := createExampleBoard()

	// Create orders with various validation errors
	invalidOrders := []game.Order{
		{
			// Missing destination for move order
			ID:       "invalid_1",
			UnitType: game.Army,
			From:     "munich",
			Type:     game.Move,
			Owner:    game.Germany,
		},
		{
			// Unit type mismatch
			ID:       "invalid_2",
			UnitType: game.Fleet, // Unit at paris is an army
			From:     "paris",
			Type:     game.Hold,
			Owner:    game.France,
		},
		{
			// Non-existent unit
			ID:       "invalid_3",
			UnitType: game.Army,
			From:     "empty_province",
			Type:     game.Move,
			To:       "berlin",
			Owner:    game.Germany,
		},
	}

	fmt.Println("\n=== Validation Error Example ===")
	validationErrors, err := engine.ValidateOrders(invalidOrders, board)
	if err != nil {
		log.Fatalf("Validation failed: %v", err)
	}

	fmt.Printf("Found %d validation errors:\n", len(validationErrors))
	for _, validationError := range validationErrors {
		fmt.Printf("  - Order %d [%s]: %s\n",
			validationError.OrderIndex,
			validationError.ErrorType.String(),
			validationError.Message)

		if len(validationError.Context) > 0 {
			fmt.Printf("    Context: %v\n", validationError.Context)
		}
	}
}

// ExampleCustomConfiguration demonstrates custom engine configuration
func ExampleCustomConfiguration() {
	// Create engine with custom configuration
	config := EngineConfig{
		StrictDATCCompliance: false, // Allow some validation errors
		ParadoxResolution:    OptimisticResolution,
		ValidationLevel:      ComprehensiveValidation,
		EnableDebugLogging:   true,
		MaxRecursionDepth:    50,
	}

	engine := NewDATCCompliantEngine(config)

	fmt.Println("\n=== Custom Configuration Example ===")
	info := engine.GetEngineInfo()
	fmt.Printf("Engine: %s\n", info.Name)
	fmt.Printf("Configuration:\n")
	for key, value := range info.Metadata {
		fmt.Printf("  - %s: %s\n", key, value)
	}
}

// Helper functions

func createExampleBoard() *game.Board {
	board := game.NewBoard()

	// Add provinces
	provinces := []*game.Province{
		{
			Name:           "munich",
			ShortCode:      "mun",
			DisplayName:    "Munich",
			Type:           game.Land,
			SupplyCenter:   true,
			ArmyNeighbors:  []string{"berlin", "vienna", "tyrolia"},
			FleetNeighbors: []string{},
		},
		{
			Name:           "berlin",
			ShortCode:      "ber",
			DisplayName:    "Berlin",
			Type:           game.Land,
			SupplyCenter:   true,
			ArmyNeighbors:  []string{"munich", "prussia", "silesia"},
			FleetNeighbors: []string{},
		},
		{
			Name:           "paris",
			ShortCode:      "par",
			DisplayName:    "Paris",
			Type:           game.Land,
			SupplyCenter:   true,
			ArmyNeighbors:  []string{"burgundy", "picardy", "gascony"},
			FleetNeighbors: []string{},
		},
		{
			Name:           "north_sea",
			ShortCode:      "nth",
			DisplayName:    "North Sea",
			Type:           game.Sea,
			SupplyCenter:   false,
			ArmyNeighbors:  []string{},
			FleetNeighbors: []string{"london", "edinburgh", "norway", "holland", "helgoland_bight"},
		},
	}

	for _, province := range provinces {
		board.AddProvince(province)
	}

	// Add units
	units := []*game.Unit{
		{
			Type:     game.Army,
			Owner:    game.Germany,
			Province: "munich",
		},
		{
			Type:     game.Army,
			Owner:    game.France,
			Province: "paris",
		},
		{
			Type:     game.Fleet,
			Owner:    game.England,
			Province: "north_sea",
		},
	}

	for _, unit := range units {
		if err := board.PlaceUnit(unit); err != nil {
			log.Fatalf("Failed to place unit: %v", err)
		}
	}

	return board
}

func formatUnitStatus(status UnitStatus) string {
	switch status {
	case UnitMoved:
		return "moved"
	case UnitHeld:
		return "held"
	case UnitDislodged:
		return "dislodged"
	case UnitBounced:
		return "bounced"
	case UnitRetreated:
		return "retreated"
	case UnitDisbanded:
		return "disbanded"
	default:
		return "unknown"
	}
}
