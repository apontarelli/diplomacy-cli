package resolution

import (
	"testing"

	"diplomacy-cli/backend/internal/game"
)

func TestRetreatProcessor_ValidateRetreatOrder(t *testing.T) {
	// Create a simple board for testing
	board := game.NewBoard()

	// Add some provinces for testing
	board.Provinces["Berlin"] = &game.Province{
		Name:           "Berlin",
		ArmyNeighbors:  []string{"Munich", "Kiel", "Silesia"},
		FleetNeighbors: []string{},
	}
	board.Provinces["Munich"] = &game.Province{
		Name:           "Munich",
		ArmyNeighbors:  []string{"Berlin", "Tyrolia"},
		FleetNeighbors: []string{},
	}
	board.Provinces["Kiel"] = &game.Province{
		Name:           "Kiel",
		ArmyNeighbors:  []string{"Berlin", "Denmark"},
		FleetNeighbors: []string{},
	}

	rp := NewRetreatProcessor(board)

	tests := []struct {
		name        string
		order       RetreatOrder
		dislodged   DislodgedUnit
		expectError bool
		errorMsg    string
	}{
		{
			name: "Valid retreat order",
			order: RetreatOrder{
				Unit:        "A Berlin",
				From:        "Berlin",
				Destination: "Munich",
				Owner:       "Germany",
			},
			dislodged: DislodgedUnit{
				Unit:             "A Berlin",
				Owner:            "Germany",
				DislodgedFrom:    "Berlin",
				AttackerOrigin:   "Silesia",
				PossibleRetreats: []string{"Munich", "Kiel"},
			},
			expectError: false,
		},
		{
			name: "Invalid retreat - wrong from province",
			order: RetreatOrder{
				Unit:        "A Berlin",
				From:        "Munich",
				Destination: "Tyrolia",
				Owner:       "Germany",
			},
			dislodged: DislodgedUnit{
				Unit:             "A Berlin",
				Owner:            "Germany",
				DislodgedFrom:    "Berlin",
				AttackerOrigin:   "Silesia",
				PossibleRetreats: []string{"Munich", "Kiel"},
			},
			expectError: true,
			errorMsg:    "unit must retreat from Berlin, not Munich",
		},
		{
			name: "Invalid retreat - destination not in possible retreats",
			order: RetreatOrder{
				Unit:        "A Berlin",
				From:        "Berlin",
				Destination: "Tyrolia",
				Owner:       "Germany",
			},
			dislodged: DislodgedUnit{
				Unit:             "A Berlin",
				Owner:            "Germany",
				DislodgedFrom:    "Berlin",
				AttackerOrigin:   "Silesia",
				PossibleRetreats: []string{"Munich", "Kiel"},
			},
			expectError: true,
			errorMsg:    "invalid retreat destination Tyrolia, valid options: [Munich Kiel]",
		},
		{
			name: "Invalid retreat - to attacker's origin",
			order: RetreatOrder{
				Unit:        "A Berlin",
				From:        "Berlin",
				Destination: "Silesia",
				Owner:       "Germany",
			},
			dislodged: DislodgedUnit{
				Unit:             "A Berlin",
				Owner:            "Germany",
				DislodgedFrom:    "Berlin",
				AttackerOrigin:   "Silesia",
				PossibleRetreats: []string{"Munich", "Kiel", "Silesia"},
			},
			expectError: true,
			errorMsg:    "cannot retreat to attacker's origin province Silesia",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := rp.ValidateRetreatOrder(tt.order, tt.dislodged)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				} else if tt.errorMsg != "" && err.Error() != tt.errorMsg {
					t.Errorf("Expected error message '%s', got '%s'", tt.errorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
			}
		})
	}
}

func TestRetreatProcessor_ProcessRetreats_Conflicts(t *testing.T) {
	board := game.NewBoard()
	rp := NewRetreatProcessor(board)

	// Test case: Two units trying to retreat to the same province
	dislodgedUnits := []DislodgedUnit{
		{
			Unit:             "A Berlin",
			Owner:            "Germany",
			DislodgedFrom:    "Berlin",
			AttackerOrigin:   "Silesia",
			PossibleRetreats: []string{"Munich", "Kiel"},
		},
		{
			Unit:             "A Prussia",
			Owner:            "Germany",
			DislodgedFrom:    "Prussia",
			AttackerOrigin:   "Warsaw",
			PossibleRetreats: []string{"Munich", "Livonia"},
		},
	}

	retreatOrders := []RetreatOrder{
		{
			Unit:        "A Berlin",
			From:        "Berlin",
			Destination: "Munich",
			Owner:       "Germany",
		},
		{
			Unit:        "A Prussia",
			From:        "Prussia",
			Destination: "Munich",
			Owner:       "Germany",
		},
	}

	results := rp.ProcessRetreats(dislodgedUnits, retreatOrders)

	// Both units should be disbanded due to conflict
	if len(results) != 2 {
		t.Errorf("Expected 2 results, got %d", len(results))
	}

	for _, result := range results {
		if result.Success {
			t.Errorf("Expected retreat to fail due to conflict, but unit %s succeeded", result.Order.Unit)
		}
		if !result.Disbanded {
			t.Errorf("Expected unit %s to be disbanded due to conflict", result.Order.Unit)
		}
		if result.Reason != "Retreat conflict: 2 units attempting to retreat to Munich" {
			t.Errorf("Expected conflict reason, got: %s", result.Reason)
		}
	}
}

func TestRetreatProcessor_ProcessRetreats_NoOrder(t *testing.T) {
	board := game.NewBoard()
	rp := NewRetreatProcessor(board)

	// Test case: Dislodged unit with no retreat order
	dislodgedUnits := []DislodgedUnit{
		{
			Unit:             "A Berlin",
			Owner:            "Germany",
			DislodgedFrom:    "Berlin",
			AttackerOrigin:   "Silesia",
			PossibleRetreats: []string{"Munich", "Kiel"},
		},
	}

	retreatOrders := []RetreatOrder{} // No orders given

	results := rp.ProcessRetreats(dislodgedUnits, retreatOrders)

	if len(results) != 1 {
		t.Errorf("Expected 1 result, got %d", len(results))
	}

	result := results[0]
	if result.Success {
		t.Errorf("Expected retreat to fail when no order given")
	}
	if !result.Disbanded {
		t.Errorf("Expected unit to be disbanded when no order given")
	}
	if result.Reason != "No retreat order given" {
		t.Errorf("Expected 'No retreat order given' reason, got: %s", result.Reason)
	}
}

func TestRetreatProcessor_CalculatePossibleRetreats(t *testing.T) {
	// Create a board with some provinces
	board := game.NewBoard()

	board.Provinces["Berlin"] = &game.Province{
		Name:           "Berlin",
		ArmyNeighbors:  []string{"Munich", "Kiel", "Silesia"},
		FleetNeighbors: []string{},
	}
	board.Provinces["Munich"] = &game.Province{Name: "Munich"}
	board.Provinces["Kiel"] = &game.Province{Name: "Kiel"}
	board.Provinces["Silesia"] = &game.Province{Name: "Silesia"}

	// Add a unit occupying Munich
	board.Units["Munich"] = &game.Unit{
		Type:     game.Army,
		Owner:    game.Austria,
		Province: "Munich",
	}

	rp := NewRetreatProcessor(board)
	gameState := &game.GameState{Board: board}

	unit := &game.Unit{
		Type:     game.Army,
		Owner:    game.Germany,
		Province: "Berlin",
	}

	// Test with attacker from Silesia
	possibleRetreats := rp.CalculatePossibleRetreats(unit, "Silesia", gameState)

	// Should be able to retreat to Kiel but not Munich (occupied) or Silesia (attacker origin)
	expected := []string{"Kiel"}

	if len(possibleRetreats) != len(expected) {
		t.Errorf("Expected %d possible retreats, got %d: %v", len(expected), len(possibleRetreats), possibleRetreats)
	}

	for _, exp := range expected {
		found := false
		for _, actual := range possibleRetreats {
			if actual == exp {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected %s in possible retreats, but not found in %v", exp, possibleRetreats)
		}
	}
}
