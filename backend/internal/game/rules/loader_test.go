package rules

import (
	"testing"

	"diplomacy-backend/internal/game"
)

func TestLoadClassicRules(t *testing.T) {
	rules, err := LoadRules("classic")
	if err != nil {
		t.Fatalf("Failed to load classic rules: %v", err)
	}
	
	// Test basic territory loading
	if !rules.IsValidTerritory("ber") {
		t.Error("Berlin should be a valid territory")
	}
	
	if !rules.IsValidTerritory("nth") {
		t.Error("North Sea should be a valid territory")
	}
	
	// Test supply centers
	if !rules.IsSupplyCenter("ber") {
		t.Error("Berlin should be a supply center")
	}
	
	// Test territory types
	if rules.GetTerritoryType("ber") != "land" {
		t.Error("Berlin should be a land territory")
	}
	
	if rules.GetTerritoryType("nth") != "sea" {
		t.Error("North Sea should be a sea territory")
	}
	
	// Test unit occupation rules
	if !rules.CanUnitOccupy(game.UnitTypeArmy, "ber") {
		t.Error("Army should be able to occupy Berlin (land)")
	}
	
	if rules.CanUnitOccupy(game.UnitTypeArmy, "nth") {
		t.Error("Army should not be able to occupy North Sea (sea)")
	}
	
	if !rules.CanUnitOccupy(game.UnitTypeFleet, "nth") {
		t.Error("Fleet should be able to occupy North Sea (sea)")
	}
	
	if rules.CanUnitOccupy(game.UnitTypeFleet, "mun") {
		t.Error("Fleet should not be able to occupy Munich (land, no coast)")
	}
}

func TestAdjacencyValidation(t *testing.T) {
	rules, err := LoadRules("classic")
	if err != nil {
		t.Fatalf("Failed to load classic rules: %v", err)
	}
	
	// Test valid army movement (land to land)
	if !rules.CanMove(game.UnitTypeArmy, "ber", "mun") {
		t.Error("Army should be able to move from Berlin to Munich")
	}
	
	// Test invalid army movement (land to sea)
	if rules.CanMove(game.UnitTypeArmy, "lvp", "iri") {
		t.Error("Army should not be able to move from Liverpool to Irish Sea")
	}
	
	// Test valid fleet movement (sea to sea)
	if !rules.CanMove(game.UnitTypeFleet, "nth", "eng") {
		t.Error("Fleet should be able to move from North Sea to English Channel")
	}
	
	// Test invalid fleet movement (sea to land)
	if rules.CanMove(game.UnitTypeFleet, "kie", "mun") {
		t.Error("Fleet should not be able to move from Kiel to Munich")
	}
	
	// Test invalid movement (non-adjacent)
	if rules.CanMove(game.UnitTypeFleet, "nth", "pic") {
		t.Error("Fleet should not be able to move from North Sea to Picardy (non-adjacent)")
	}
}

func TestCoastHandling(t *testing.T) {
	rules, err := LoadRules("classic")
	if err != nil {
		t.Fatalf("Failed to load classic rules: %v", err)
	}
	
	// Test that Spain has coasts
	if !rules.HasCoast["spa"] {
		t.Error("Spain should have coasts")
	}
	
	// Test that coast territories exist
	if !rules.IsValidTerritory("spa_nc") {
		t.Error("Spain North Coast should be a valid territory")
	}
	
	if !rules.IsValidTerritory("spa_sc") {
		t.Error("Spain South Coast should be a valid territory")
	}
	
	// Test coast to parent mapping
	if rules.CoastToParent["spa_nc"] != "spa" {
		t.Error("Spain North Coast should map to Spain as parent")
	}
}