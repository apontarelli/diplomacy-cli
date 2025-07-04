package resolution

import (
	"diplomacy-cli/backend/internal/game"
	"testing"
)

func TestNewResolutionEngine(t *testing.T) {
	board := game.NewBoard()

	board.AddProvince(&game.Province{
		Name:           "paris",
		ShortCode:      "par",
		DisplayName:    "Paris",
		Type:           game.Land,
		SupplyCenter:   true,
		ArmyNeighbors:  []string{"burgundy", "picardy"},
		FleetNeighbors: []string{},
	})

	orders := []*game.Order{
		{
			ID:       "1",
			UnitType: game.Army,
			From:     "paris",
			Type:     game.Hold,
			Owner:    game.France,
		},
		{
			ID:       "2",
			UnitType: game.Army,
			From:     "burgundy",
			Type:     game.Move,
			To:       "paris",
			Owner:    game.Germany,
		},
	}

	engine := NewResolutionEngine(orders, board)

	if engine == nil {
		t.Fatal("Expected engine to be created")
	}

	if len(engine.orders) != 2 {
		t.Errorf("Expected 2 orders, got %d", len(engine.orders))
	}

	if len(engine.outcomes) != 2 {
		t.Errorf("Expected 2 outcomes, got %d", len(engine.outcomes))
	}

	if len(engine.strength) != 2 {
		t.Errorf("Expected 2 strength values, got %d", len(engine.strength))
	}

	parisOrder := engine.FindOrderByTerritory("paris")
	if parisOrder != 0 {
		t.Errorf("Expected paris order at index 0, got %d", parisOrder)
	}

	for i, strength := range engine.strength {
		if strength != 1 {
			t.Errorf("Expected initial strength 1 for order %d, got %d", i, strength)
		}
	}
}

func TestMakeUnitKey(t *testing.T) {
	tests := []struct {
		territory string
		coast     string
		expected  string
	}{
		{"paris", "", "paris"},
		{"st_petersburg", "nc", "st_petersburg/nc"},
		{"spain", "sc", "spain/sc"},
	}

	for _, test := range tests {
		result := makeUnitKey(test.territory, test.coast)
		if result != test.expected {
			t.Errorf("makeUnitKey(%q, %q) = %q, expected %q",
				test.territory, test.coast, result, test.expected)
		}
	}
}

func TestOrderResult(t *testing.T) {
	results := []OrderResult{
		MoveSuccess,
		MoveBounced,
		MoveNoConvoy,
		SupportSuccess,
		SupportCut,
		ConvoySuccess,
		HoldSuccess,
		Dislodged,
	}

	for _, result := range results {
		if string(result) == "" {
			t.Errorf("Order result should not be empty: %v", result)
		}
	}
}
