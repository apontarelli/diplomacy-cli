package resolution

import (
	"diplomacy-cli/backend/internal/game"
	"testing"
)

func TestDATCCompliantEngine_GetEngineInfo(t *testing.T) {
	engine := NewDefaultDATCEngine()
	info := engine.GetEngineInfo()

	if info.Name == "" {
		t.Error("Engine name should not be empty")
	}

	if info.Version == "" {
		t.Error("Engine version should not be empty")
	}

	if len(info.Features) == 0 {
		t.Error("Engine should have at least one feature")
	}

	expectedFeatures := []string{
		"DATC compliance",
		"Paradox resolution",
		"Convoy path validation",
		"Multi-stage order validation",
		"Detailed resolution logging",
	}

	for _, expectedFeature := range expectedFeatures {
		found := false
		for _, feature := range info.Features {
			if feature == expectedFeature {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected feature '%s' not found in engine features", expectedFeature)
		}
	}
}

func TestDATCCompliantEngine_ValidateOrders_SyntaxErrors(t *testing.T) {
	engine := NewDefaultDATCEngine()
	board := createTestBoard()

	tests := []struct {
		name          string
		order         game.Order
		expectedError ValidationErrorType
		expectedMsg   string
	}{
		{
			name: "missing source province",
			order: game.Order{
				UnitType: game.Army,
				Type:     game.Move,
				To:       "berlin",
				Owner:    game.Germany,
			},
			expectedError: SyntaxError,
			expectedMsg:   "order missing source province",
		},
		{
			name: "missing unit type",
			order: game.Order{
				From:  "munich",
				Type:  game.Move,
				To:    "berlin",
				Owner: game.Germany,
			},
			expectedError: SyntaxError,
			expectedMsg:   "order missing unit type",
		},
		{
			name: "missing order type",
			order: game.Order{
				UnitType: game.Army,
				From:     "munich",
				To:       "berlin",
				Owner:    game.Germany,
			},
			expectedError: SyntaxError,
			expectedMsg:   "order missing order type",
		},
		{
			name: "move order missing destination",
			order: game.Order{
				UnitType: game.Army,
				From:     "munich",
				Type:     game.Move,
				Owner:    game.Germany,
			},
			expectedError: SyntaxError,
			expectedMsg:   "move order missing destination",
		},
		{
			name: "support order missing target",
			order: game.Order{
				UnitType: game.Army,
				From:     "munich",
				Type:     game.Support,
				Owner:    game.Germany,
			},
			expectedError: SyntaxError,
			expectedMsg:   "support order missing target",
		},
		{
			name: "convoy order missing target",
			order: game.Order{
				UnitType: game.Fleet,
				From:     "north_sea",
				Type:     game.Convoy,
				Owner:    game.England,
			},
			expectedError: SyntaxError,
			expectedMsg:   "convoy order missing target",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errors, err := engine.ValidateOrders([]game.Order{tt.order}, board)
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if len(errors) == 0 {
				t.Fatal("Expected validation error but got none")
			}

			if errors[0].ErrorType != tt.expectedError {
				t.Errorf("Expected error type %v, got %v", tt.expectedError, errors[0].ErrorType)
			}

			if errors[0].Message != tt.expectedMsg {
				t.Errorf("Expected error message '%s', got '%s'", tt.expectedMsg, errors[0].Message)
			}
		})
	}
}

func TestDATCCompliantEngine_ValidateOrders_SemanticErrors(t *testing.T) {
	engine := NewDefaultDATCEngine()
	board := createTestBoard()

	tests := []struct {
		name          string
		order         game.Order
		expectedError ValidationErrorType
		expectedMsg   string
	}{
		{
			name: "no unit at source province",
			order: game.Order{
				UnitType: game.Army,
				From:     "empty_province",
				Type:     game.Move,
				To:       "berlin",
				Owner:    game.Germany,
			},
			expectedError: SemanticError,
			expectedMsg:   "no unit found at province empty_province",
		},
		{
			name: "unit type mismatch",
			order: game.Order{
				UnitType: game.Fleet, // Unit at munich is an army
				From:     "munich",
				Type:     game.Move,
				To:       "berlin",
				Owner:    game.Germany,
			},
			expectedError: SemanticError,
			expectedMsg:   "unit type mismatch: expected army, got fleet",
		},
		{
			name: "unit owner mismatch",
			order: game.Order{
				UnitType: game.Army,
				From:     "munich", // German unit
				Type:     game.Move,
				To:       "berlin",
				Owner:    game.France, // Wrong owner
			},
			expectedError: SemanticError,
			expectedMsg:   "unit owner mismatch: expected germany, got france",
		},
		{
			name: "destination province does not exist",
			order: game.Order{
				UnitType: game.Army,
				From:     "munich",
				Type:     game.Move,
				To:       "nonexistent_province",
				Owner:    game.Germany,
			},
			expectedError: SemanticError,
			expectedMsg:   "destination province nonexistent_province does not exist",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errors, err := engine.ValidateOrders([]game.Order{tt.order}, board)
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if len(errors) == 0 {
				t.Fatal("Expected validation error but got none")
			}

			found := false
			for _, validationError := range errors {
				if validationError.ErrorType == tt.expectedError && validationError.Message == tt.expectedMsg {
					found = true
					break
				}
			}

			if !found {
				t.Errorf("Expected error type %v with message '%s' not found in errors: %v",
					tt.expectedError, tt.expectedMsg, errors)
			}
		})
	}
}

func TestDATCCompliantEngine_ValidateOrders_GameRuleErrors(t *testing.T) {
	engine := NewDefaultDATCEngine()
	board := createTestBoard()

	tests := []struct {
		name          string
		order         game.Order
		expectedError ValidationErrorType
		expectedMsg   string
	}{
		{
			name: "army cannot move to sea province without convoy",
			order: game.Order{
				UnitType: game.Army,
				From:     "munich",
				Type:     game.Move,
				To:       "north_sea", // Sea province
				Owner:    game.Germany,
			},
			expectedError: GameRuleError,
			expectedMsg:   "army cannot move to sea province without convoy",
		},
		{
			name: "cannot support non-existent unit",
			order: game.Order{
				UnitType:      game.Army,
				From:          "munich",
				Type:          game.Support,
				SupportTarget: "empty_province",
				Owner:         game.Germany,
			},
			expectedError: GameRuleError,
			expectedMsg:   "cannot support non-existent unit at empty_province",
		},
		{
			name: "only fleets can convoy",
			order: game.Order{
				UnitType:     game.Army,
				From:         "munich",
				Type:         game.Convoy,
				ConvoyTarget: "london -> calais",
				Owner:        game.Germany,
			},
			expectedError: GameRuleError,
			expectedMsg:   "only fleets can convoy",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errors, err := engine.ValidateOrders([]game.Order{tt.order}, board)
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if len(errors) == 0 {
				t.Fatal("Expected validation error but got none")
			}

			found := false
			for _, validationError := range errors {
				if validationError.ErrorType == tt.expectedError && validationError.Message == tt.expectedMsg {
					found = true
					break
				}
			}

			if !found {
				t.Errorf("Expected error type %v with message '%s' not found in errors: %v",
					tt.expectedError, tt.expectedMsg, errors)
			}
		})
	}
}

func TestDATCCompliantEngine_ValidateOrders_ContextErrors(t *testing.T) {
	engine := NewDefaultDATCEngine()
	board := createTestBoard()

	// Test duplicate orders from same unit
	orders := []game.Order{
		{
			UnitType: game.Army,
			From:     "munich",
			Type:     game.Move,
			To:       "berlin",
			Owner:    game.Germany,
		},
		{
			UnitType: game.Army,
			From:     "munich",
			Type:     game.Hold,
			Owner:    game.Germany,
		},
	}

	errors, err := engine.ValidateOrders(orders, board)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(errors) == 0 {
		t.Fatal("Expected validation error for duplicate orders but got none")
	}

	found := false
	for _, validationError := range errors {
		if validationError.ErrorType == ContextError &&
			validationError.Message == "duplicate order for unit at munich" {
			found = true
			break
		}
	}

	if !found {
		t.Errorf("Expected context error for duplicate orders not found in errors: %v", errors)
	}
}

func TestDATCCompliantEngine_ValidateOrders_ValidOrders(t *testing.T) {
	engine := NewDefaultDATCEngine()
	board := createTestBoard()

	validOrders := []game.Order{
		{
			UnitType: game.Army,
			From:     "munich",
			Type:     game.Move,
			To:       "berlin",
			Owner:    game.Germany,
		},
		{
			UnitType: game.Fleet,
			From:     "north_sea",
			Type:     game.Hold,
			Owner:    game.England,
		},
		{
			UnitType:           game.Army,
			From:               "paris",
			Type:               game.Support,
			SupportTarget:      "munich",
			SupportDestination: "berlin",
			Owner:              game.France,
		},
	}

	errors, err := engine.ValidateOrders(validOrders, board)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(errors) > 0 {
		t.Errorf("Expected no validation errors for valid orders, but got: %v", errors)
	}
}

func TestDATCCompliantEngine_Resolve_BasicResolution(t *testing.T) {
	engine := NewDefaultDATCEngine()
	board := createTestBoard()

	orders := []game.Order{
		{
			ID:       "1",
			UnitType: game.Army,
			From:     "munich",
			Type:     game.Move,
			To:       "berlin",
			Owner:    game.Germany,
		},
		{
			ID:       "2",
			UnitType: game.Fleet,
			From:     "north_sea",
			Type:     game.Hold,
			Owner:    game.England,
		},
	}

	result, err := engine.Resolve(orders, board)
	if err != nil {
		t.Fatalf("Unexpected error during resolution: %v", err)
	}

	if result == nil {
		t.Fatal("Expected resolution result but got nil")
	}

	unitOutcomes := result.UnitOutcomes()
	if len(unitOutcomes) != len(orders) {
		t.Errorf("Expected %d unit outcomes, got %d", len(orders), len(unitOutcomes))
	}

	// Check that we have outcomes for both units
	germanArmyID := UnitID{Type: game.Army, Owner: game.Germany, Province: "munich"}
	englishFleetID := UnitID{Type: game.Fleet, Owner: game.England, Province: "north_sea"}

	if _, exists := unitOutcomes[germanArmyID]; !exists {
		t.Error("Expected outcome for German army not found")
	}

	if _, exists := unitOutcomes[englishFleetID]; !exists {
		t.Error("Expected outcome for English fleet not found")
	}
}

func TestValidationErrorType_String(t *testing.T) {
	tests := []struct {
		errorType ValidationErrorType
		expected  string
	}{
		{SyntaxError, "syntax_error"},
		{SemanticError, "semantic_error"},
		{GameRuleError, "game_rule_error"},
		{ContextError, "context_error"},
		{ValidationErrorType(999), "unknown_error"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if got := tt.errorType.String(); got != tt.expected {
				t.Errorf("ValidationErrorType.String() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestValidationError_Error(t *testing.T) {
	err := ValidationError{
		OrderIndex: 5,
		ErrorType:  SyntaxError,
		Message:    "test error message",
		Context:    map[string]interface{}{"field": "test"},
	}

	expected := "[syntax_error] Order 5: test error message"
	if got := err.Error(); got != expected {
		t.Errorf("ValidationError.Error() = %v, want %v", got, expected)
	}
}

// Helper function to create a test board with some units
func createTestBoard() *game.Board {
	board := game.NewBoard()

	// Add some provinces
	provinces := []*game.Province{
		{
			Name:           "munich",
			ShortCode:      "mun",
			DisplayName:    "Munich",
			Type:           game.Land,
			SupplyCenter:   true,
			ArmyNeighbors:  []string{"berlin", "vienna"},
			FleetNeighbors: []string{},
		},
		{
			Name:           "berlin",
			ShortCode:      "ber",
			DisplayName:    "Berlin",
			Type:           game.Land,
			SupplyCenter:   true,
			ArmyNeighbors:  []string{"munich", "prussia"},
			FleetNeighbors: []string{},
		},
		{
			Name:           "north_sea",
			ShortCode:      "nth",
			DisplayName:    "North Sea",
			Type:           game.Sea,
			SupplyCenter:   false,
			ArmyNeighbors:  []string{},
			FleetNeighbors: []string{"london", "edinburgh", "norway"},
		},
		{
			Name:           "paris",
			ShortCode:      "par",
			DisplayName:    "Paris",
			Type:           game.Land,
			SupplyCenter:   true,
			ArmyNeighbors:  []string{"burgundy", "picardy"},
			FleetNeighbors: []string{},
		},
		{
			Name:           "london",
			ShortCode:      "lon",
			DisplayName:    "London",
			Type:           game.Land,
			SupplyCenter:   true,
			ArmyNeighbors:  []string{"wales", "yorkshire"},
			FleetNeighbors: []string{"north_sea", "english_channel"},
		},
	}

	for _, province := range provinces {
		board.AddProvince(province)
	}

	// Add some units
	units := []*game.Unit{
		{
			Type:     game.Army,
			Owner:    game.Germany,
			Province: "munich",
		},
		{
			Type:     game.Fleet,
			Owner:    game.England,
			Province: "north_sea",
		},
		{
			Type:     game.Army,
			Owner:    game.France,
			Province: "paris",
		},
	}

	for _, unit := range units {
		board.PlaceUnit(unit)
	}

	return board
}
