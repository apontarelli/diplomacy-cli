package resolution

import (
	"diplomacy-cli/backend/internal/game"
	"testing"
)

func TestResolutionEngine_BasicResolve(t *testing.T) {
	board := game.NewBoard()

	board.AddProvince(&game.Province{
		Name:           "paris",
		ShortCode:      "par",
		DisplayName:    "Paris",
		Type:           game.Land,
		SupplyCenter:   true,
		ArmyNeighbors:  []string{"burgundy"},
		FleetNeighbors: []string{},
	})

	board.AddProvince(&game.Province{
		Name:           "burgundy",
		ShortCode:      "bur",
		DisplayName:    "Burgundy",
		Type:           game.Land,
		SupplyCenter:   false,
		ArmyNeighbors:  []string{"paris"},
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
	err := engine.Resolve()

	if err != nil {
		t.Fatalf("Resolution failed: %v", err)
	}

	results := engine.GetResults()

	if len(results) != 2 {
		t.Fatalf("Expected 2 results, got %d", len(results))
	}

	parisResult := results[0]
	if parisResult.Result != HoldSuccess {
		t.Errorf("Expected Paris hold to succeed, got %v", parisResult.Result)
	}

	burgundyResult := results[1]
	if burgundyResult.Result != MoveBounced {
		t.Errorf("Expected Burgundy move to bounce, got %v", burgundyResult.Result)
	}

	for i, result := range results {
		if result.Dislodged {
			t.Errorf("Order %d should not be dislodged", i)
		}
	}
}

func TestResolutionEngine_SupportOrder(t *testing.T) {
	board := game.NewBoard()

	provinces := []*game.Province{
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
			Name:           "burgundy",
			ShortCode:      "bur",
			DisplayName:    "Burgundy",
			Type:           game.Land,
			SupplyCenter:   false,
			ArmyNeighbors:  []string{"paris", "picardy"},
			FleetNeighbors: []string{},
		},
		{
			Name:           "picardy",
			ShortCode:      "pic",
			DisplayName:    "Picardy",
			Type:           game.Land,
			SupplyCenter:   false,
			ArmyNeighbors:  []string{"paris", "burgundy"},
			FleetNeighbors: []string{},
		},
	}

	for _, province := range provinces {
		board.AddProvince(province)
	}

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
		{
			ID:                 "3",
			UnitType:           game.Army,
			From:               "picardy",
			Type:               game.Support,
			SupportTarget:      "burgundy",
			SupportDestination: "paris",
			Owner:              game.Germany,
		},
	}

	engine := NewResolutionEngine(orders, board)
	err := engine.Resolve()

	if err != nil {
		t.Fatalf("Resolution failed: %v", err)
	}

	results := engine.GetResults()

	if len(results) != 3 {
		t.Fatalf("Expected 3 results, got %d", len(results))
	}

	supportResult := results[2]
	if supportResult.Result == SupportInvalid {
		t.Errorf("Support order should be valid, got %v with reason: %s",
			supportResult.Result, supportResult.FailureReason)
	}
}

func TestResolutionEngine_InvalidSupport(t *testing.T) {
	board := game.NewBoard()

	board.AddProvince(&game.Province{
		Name:           "paris",
		ShortCode:      "par",
		DisplayName:    "Paris",
		Type:           game.Land,
		SupplyCenter:   true,
		ArmyNeighbors:  []string{"burgundy"},
		FleetNeighbors: []string{},
	})

	orders := []*game.Order{
		{
			ID:                 "1",
			UnitType:           game.Army,
			From:               "paris",
			Type:               game.Support,
			SupportTarget:      "burgundy",
			SupportDestination: "picardy",
			Owner:              game.France,
		},
	}

	engine := NewResolutionEngine(orders, board)
	err := engine.Resolve()

	if err != nil {
		t.Fatalf("Resolution failed: %v", err)
	}

	results := engine.GetResults()

	if len(results) != 1 {
		t.Fatalf("Expected 1 result, got %d", len(results))
	}

	result := results[0]
	if result.Result != SupportInvalid {
		t.Errorf("Expected support to be invalid, got %v", result.Result)
	}

	if result.FailureReason == "" {
		t.Error("Expected failure reason to be set for invalid support")
	}
}
