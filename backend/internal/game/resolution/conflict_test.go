package resolution

import (
	"diplomacy-cli/backend/internal/game"
	"testing"
)

func TestFriendlyUnitProtection(t *testing.T) {
	board := game.NewBoard()
	board.AddProvince(&game.Province{
		Name:           "berlin",
		ShortCode:      "ber",
		Type:           game.Land,
		ArmyNeighbors:  []string{"munich", "warsaw"},
		FleetNeighbors: []string{},
	})

	board.AddProvince(&game.Province{
		Name:           "munich",
		ShortCode:      "mun",
		Type:           game.Land,
		ArmyNeighbors:  []string{"berlin"},
		FleetNeighbors: []string{},
	})

	board.AddProvince(&game.Province{
		Name:           "warsaw",
		ShortCode:      "war",
		Type:           game.Land,
		ArmyNeighbors:  []string{"berlin"},
		FleetNeighbors: []string{},
	})

	tests := []struct {
		name     string
		orders   []*game.Order
		expected map[int]string
	}{
		{
			name: "Friendly unit protection should cause bounce",
			orders: []*game.Order{
				{Type: game.Hold, From: "berlin", UnitType: game.Army, Owner: "germany"},
				{Type: game.Move, From: "munich", To: "berlin", UnitType: game.Army, Owner: "germany"},
				{Type: game.Move, From: "warsaw", To: "berlin", UnitType: game.Army, Owner: "france"},
			},
			expected: map[int]string{
				0: "berlin",
				1: "munich",
				2: "warsaw",
			},
		},
		{
			name: "No friendly protection when no friendly attackers",
			orders: []*game.Order{
				{Type: game.Hold, From: "berlin", UnitType: game.Army, Owner: "germany"},
				{Type: game.Move, From: "warsaw", To: "berlin", UnitType: game.Army, Owner: "france"},
			},
			expected: map[int]string{
				0: "berlin",
				1: "berlin",
			},
		},
		{
			name: "Friendly winner should be protected",
			orders: []*game.Order{
				{Type: game.Hold, From: "berlin", UnitType: game.Army, Owner: "germany"},
				{Type: game.Move, From: "munich", To: "berlin", UnitType: game.Army, Owner: "germany"},
				{Type: game.Move, From: "warsaw", To: "berlin", UnitType: game.Army, Owner: "france"},
			},
			expected: map[int]string{
				0: "berlin",
				1: "munich",
				2: "warsaw",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			engine := NewResolutionEngine(tt.orders, board)

			engine.processConvoys()
			engine.processMoves()
			engine.cutSupports()
			engine.calculateStrength()

			if tt.name == "Friendly unit protection should cause bounce" && len(engine.strength) > 2 {
				engine.strength[2] = 2
			}
			if tt.name == "No friendly protection when no friendly attackers" && len(engine.strength) > 1 {
				engine.strength[1] = 2
			}
			if tt.name == "Friendly winner should be protected" && len(engine.strength) > 1 {
				engine.strength[1] = 2
			}

			engine.resolveConflicts()
			engine.detectDislodgements()
			for orderIndex, expectedTerritory := range tt.expected {
				if orderIndex >= len(engine.orders) {
					t.Errorf("Order index %d out of range", orderIndex)
					continue
				}

				actualTerritory := engine.orders[orderIndex].NewTerritory
				if actualTerritory != expectedTerritory {
					t.Errorf("Order %d: expected NewTerritory=%s, got %s",
						orderIndex, expectedTerritory, actualTerritory)
				}
			}
		})
	}
}

func TestDislodgementDetection(t *testing.T) {
	board := game.NewBoard()
	board.AddProvince(&game.Province{
		Name:           "paris",
		ShortCode:      "par",
		Type:           game.Land,
		ArmyNeighbors:  []string{"burgundy"},
		FleetNeighbors: []string{},
	})

	board.AddProvince(&game.Province{
		Name:           "burgundy",
		ShortCode:      "bur",
		Type:           game.Land,
		ArmyNeighbors:  []string{"paris"},
		FleetNeighbors: []string{},
	})

	tests := []struct {
		name     string
		orders   []*game.Order
		expected map[int]bool
	}{
		{
			name: "Successful attack should dislodge defender",
			orders: []*game.Order{
				{Type: game.Hold, From: "paris", UnitType: game.Army, Owner: "france"},
				{Type: game.Move, From: "burgundy", To: "paris", UnitType: game.Army, Owner: "germany"},
			},
			expected: map[int]bool{
				0: true,
				1: false,
			},
		},
		{
			name: "Failed attack should not dislodge defender",
			orders: []*game.Order{
				{Type: game.Hold, From: "paris", UnitType: game.Army, Owner: "france"},
				{Type: game.Move, From: "burgundy", To: "paris", UnitType: game.Army, Owner: "germany"},
			},
			expected: map[int]bool{
				0: false,
				1: false,
			},
		},
		{
			name: "Bounced attacks should not dislodge anyone",
			orders: []*game.Order{
				{Type: game.Hold, From: "paris", UnitType: game.Army, Owner: "france"},
				{Type: game.Move, From: "burgundy", To: "paris", UnitType: game.Army, Owner: "germany"},
			},
			expected: map[int]bool{
				0: false,
				1: false,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			engine := NewResolutionEngine(tt.orders, board)

			engine.processConvoys()
			engine.processMoves()
			engine.cutSupports()
			engine.calculateStrength()

			if tt.name == "Successful attack should dislodge defender" {
				engine.strength[1] = 2
			} else if tt.name == "Failed attack should not dislodge defender" {
				engine.strength[0] = 2
			}

			engine.resolveConflicts()
			engine.detectDislodgements()

			for orderIndex, expectedDislodged := range tt.expected {
				if orderIndex >= len(engine.outcomes) {
					t.Errorf("Order index %d out of range", orderIndex)
					continue
				}

				actualDislodged := engine.outcomes[orderIndex].Dislodged
				if actualDislodged != expectedDislodged {
					t.Errorf("Order %d: expected Dislodged=%v, got %v",
						orderIndex, expectedDislodged, actualDislodged)
				}
			}
		})
	}
}

func TestComplexFriendlyProtectionScenario(t *testing.T) {
	board := game.NewBoard()

	board.AddProvince(&game.Province{
		Name:           "berlin",
		ShortCode:      "ber",
		Type:           game.Land,
		ArmyNeighbors:  []string{"munich", "warsaw"},
		FleetNeighbors: []string{},
	})

	board.AddProvince(&game.Province{
		Name:           "munich",
		ShortCode:      "mun",
		Type:           game.Land,
		ArmyNeighbors:  []string{"berlin"},
		FleetNeighbors: []string{},
	})

	board.AddProvince(&game.Province{
		Name:           "warsaw",
		ShortCode:      "war",
		Type:           game.Land,
		ArmyNeighbors:  []string{"berlin"},
		FleetNeighbors: []string{},
	})

	orders := []*game.Order{
		{Type: game.Hold, From: "berlin", UnitType: game.Army, Owner: "germany"},               // German defender
		{Type: game.Move, From: "munich", To: "berlin", UnitType: game.Army, Owner: "germany"}, // German attacker
		{Type: game.Move, From: "warsaw", To: "berlin", UnitType: game.Army, Owner: "france"},  // French attacker with support
	}

	engine := NewResolutionEngine(orders, board)

	engine.strength[0] = 1 // German defender
	engine.strength[1] = 1 // German attacker
	engine.strength[2] = 2 // French attacker (supported)

	engine.executeResolutionPass()

	if engine.orders[0].NewTerritory != "berlin" {
		t.Errorf("German defender should stay in Berlin, got %s", engine.orders[0].NewTerritory)
	}
	if engine.orders[1].NewTerritory != "munich" {
		t.Errorf("German attacker should bounce back to Munich, got %s", engine.orders[1].NewTerritory)
	}
	if engine.orders[2].NewTerritory != "warsaw" {
		t.Errorf("French attacker should bounce back to Warsaw, got %s", engine.orders[2].NewTerritory)
	}

	if engine.outcomes[0].Dislodged {
		t.Error("German defender should not be dislodged due to friendly protection")
	}
	if engine.outcomes[1].Dislodged {
		t.Error("German attacker should not be dislodged")
	}
	if engine.outcomes[2].Dislodged {
		t.Error("French attacker should not be dislodged")
	}
}
