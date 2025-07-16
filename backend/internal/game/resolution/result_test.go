package resolution

import (
	"diplomacy-cli/backend/internal/game"
	"testing"
)

func TestResolutionResult(t *testing.T) {
	result := NewResolutionResult(1, game.SpringMovement)
	
	if result.TurnNumber() != 1 {
		t.Errorf("Expected turn number 1, got %d", result.TurnNumber())
	}
	
	if result.Phase() != game.SpringMovement {
		t.Errorf("Expected phase SpringMovement, got %s", result.Phase())
	}
	
	if err := result.Validate(); err != nil {
		t.Errorf("Validation failed: %v", err)
	}
}

func TestUnitID(t *testing.T) {
	unitID := UnitID{
		Type:     game.Army,
		Owner:    game.England,
		Province: "London",
	}
	
	if unitID.Type != game.Army {
		t.Errorf("Expected Army type, got %s", unitID.Type)
	}
	
	if unitID.Owner != game.England {
		t.Errorf("Expected England owner, got %s", unitID.Owner)
	}
	
	if unitID.Province != "London" {
		t.Errorf("Expected London province, got %s", unitID.Province)
	}
}
