package resolution

import (
	"diplomacy-cli/backend/internal/game"
	"testing"
)

func TestSelfAttackSupportCutting(t *testing.T) {
	board := game.NewBoard()
	board.AddProvince(&game.Province{
		Name:           "london",
		ShortCode:      "lon",
		Type:           game.Land,
		ArmyNeighbors:  []string{"wales"},
		FleetNeighbors: []string{},
	})

	board.AddProvince(&game.Province{
		Name:           "wales",
		ShortCode:      "wal",
		Type:           game.Land,
		ArmyNeighbors:  []string{"london", "yorkshire"},
		FleetNeighbors: []string{},
	})

	board.AddProvince(&game.Province{
		Name:           "yorkshire",
		ShortCode:      "yor",
		Type:           game.Land,
		ArmyNeighbors:  []string{"wales"},
		FleetNeighbors: []string{},
	})

	tests := []struct {
		name     string
		orders   []*game.Order
		expected map[int]bool // orderIndex -> expectedSupportCut
	}{
		{
			name: "Basic support cut should work",
			orders: []*game.Order{
				{Type: game.Move, From: "lon", To: "wal", UnitType: game.Army, Owner: "england"},
				{Type: game.Support, From: "wal", SupportTarget: "yor", UnitType: game.Army, Owner: "england"},
				{Type: game.Hold, From: "yor", UnitType: game.Army, Owner: "england"},
			},
			expected: map[int]bool{
				1: true, // Support from wal should be cut by move from lon
			},
		},
		{
			name: "Self-attack support cut should be prevented",
			orders: []*game.Order{
				{Type: game.Move, From: "yor", To: "lon", UnitType: game.Army, Owner: "england"},               // A attacking B
				{Type: game.Support, From: "wal", SupportTarget: "yor", UnitType: game.Army, Owner: "england"}, // C supporting A
				{Type: game.Move, From: "lon", To: "wal", UnitType: game.Army, Owner: "france"},                // B attacking C
			},
			expected: map[int]bool{
				1: false, // Support should NOT be cut (self-attack rule)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			engine := NewResolutionEngine(tt.orders, board)

			engine.executeResolutionPass()

			for orderIndex, expectedCut := range tt.expected {
				if orderIndex >= len(engine.outcomes) {
					t.Errorf("Order index %d out of range", orderIndex)
					continue
				}

				actualCut := engine.outcomes[orderIndex].SupportCut
				if actualCut != expectedCut {
					t.Errorf("Order %d: expected SupportCut=%v, got %v",
						orderIndex, expectedCut, actualCut)
				}
			}
		})
	}
}
