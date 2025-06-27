package rules

import "diplomacy-backend/internal/game"

// Edge represents a connection between two territories with movement restrictions
type Edge struct {
	From string `json:"from"`
	To   string `json:"to"`
	Mode string `json:"mode"` // "land", "sea", "both"
}

// Territory represents a territory from the map data
type Territory struct {
	DisplayName    string   `json:"display_name"`
	Type           string   `json:"type"` // "land", "sea"
	HomeCountry    string   `json:"home_country,omitempty"`
	HasCoast       bool     `json:"has_coast,omitempty"`
	IsSupplyCenter bool     `json:"is_supply_center,omitempty"`
	Coasts         []string `json:"coasts,omitempty"`
}

// Nation represents a playable nation
type Nation struct {
	ID          string `json:"id"`
	DisplayName string `json:"display_name"`
}

// Adjacency represents a valid movement from one territory to another
type Adjacency struct {
	To   string
	Mode string // "land", "sea", "both"
}

// Rules contains all the game rules and map data for a variant
type Rules struct {
	// Territory data
	TerritoryIDs         map[string]bool
	SupplyCenters        map[string]bool
	HomeCenters          map[string][]string // nation -> territories
	TerritoryDisplayNames map[string]string
	NationDisplayNames   map[string]string
	TerritoryTypes       map[string]string // territory -> "land"/"sea"/"coast"
	HasCoast             map[string]bool
	
	// Coast handling
	ParentCoasts    map[string][]string // parent -> coast IDs
	ParentToCoast   map[string][]string // parent -> coast IDs  
	CoastToParent   map[string]string   // coast -> parent
	
	// Adjacency data
	Edges         []Edge
	AdjacencyMap  map[string][]Adjacency // territory -> adjacent territories
	
	// Pre-computed adjacency maps for O(1) validation
	ArmyAdjacency  map[string]map[string]bool
	FleetAdjacency map[string]map[string]bool
}

// CanMove checks if a unit can move from one territory to another
func (r *Rules) CanMove(unitType game.UnitType, from, to string) bool {
	switch unitType {
	case game.UnitTypeArmy:
		if adjacents, exists := r.ArmyAdjacency[from]; exists {
			return adjacents[to]
		}
	case game.UnitTypeFleet:
		if adjacents, exists := r.FleetAdjacency[from]; exists {
			return adjacents[to]
		}
	}
	return false
}

// IsValidTerritory checks if a territory ID exists
func (r *Rules) IsValidTerritory(territory string) bool {
	return r.TerritoryIDs[territory]
}

// IsSupplyCenter checks if a territory is a supply center
func (r *Rules) IsSupplyCenter(territory string) bool {
	return r.SupplyCenters[territory]
}

// GetTerritoryType returns the type of a territory ("land", "sea", "coast")
func (r *Rules) GetTerritoryType(territory string) string {
	return r.TerritoryTypes[territory]
}

// CanUnitOccupy checks if a unit type can occupy a territory type
func (r *Rules) CanUnitOccupy(unitType game.UnitType, territory string) bool {
	territoryType := r.GetTerritoryType(territory)
	
	switch unitType {
	case game.UnitTypeArmy:
		return territoryType == "land"
	case game.UnitTypeFleet:
		return territoryType == "sea" || territoryType == "coast"
	}
	return false
}