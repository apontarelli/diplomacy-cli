package rules

import (
	"embed"
	"encoding/json"
	"fmt"
	"strings"
)

//go:embed data/classic/world/*.json
var classicData embed.FS

// LoadRules loads the rules for a given variant
func LoadRules(variant string) (*Rules, error) {
	switch variant {
	case "classic":
		return loadClassicRules()
	default:
		return nil, fmt.Errorf("unknown variant: %s", variant)
	}
}

// MustLoadRules loads rules and panics on error (for initialization)
func MustLoadRules(variant string) *Rules {
	rules, err := LoadRules(variant)
	if err != nil {
		panic(fmt.Sprintf("failed to load rules for variant %s: %v", variant, err))
	}
	return rules
}

func loadClassicRules() (*Rules, error) {
	// Load territories
	territoriesData, err := classicData.ReadFile("data/classic/world/territories.json")
	if err != nil {
		return nil, fmt.Errorf("failed to read territories.json: %w", err)
	}
	
	var territories map[string]Territory
	if err := json.Unmarshal(territoriesData, &territories); err != nil {
		return nil, fmt.Errorf("failed to parse territories.json: %w", err)
	}
	
	// Load edges
	edgesData, err := classicData.ReadFile("data/classic/world/edges.json")
	if err != nil {
		return nil, fmt.Errorf("failed to read edges.json: %w", err)
	}
	
	var edges []Edge
	if err := json.Unmarshal(edgesData, &edges); err != nil {
		return nil, fmt.Errorf("failed to parse edges.json: %w", err)
	}
	
	// Load nations
	nationsData, err := classicData.ReadFile("data/classic/world/nations.json")
	if err != nil {
		return nil, fmt.Errorf("failed to read nations.json: %w", err)
	}
	
	var nations []Nation
	if err := json.Unmarshal(nationsData, &nations); err != nil {
		return nil, fmt.Errorf("failed to parse nations.json: %w", err)
	}
	
	return buildRules(territories, edges, nations)
}

func buildRules(territories map[string]Territory, edges []Edge, nations []Nation) (*Rules, error) {
	rules := &Rules{
		TerritoryIDs:          make(map[string]bool),
		SupplyCenters:         make(map[string]bool),
		HomeCenters:           make(map[string][]string),
		TerritoryDisplayNames: make(map[string]string),
		NationDisplayNames:    make(map[string]string),
		TerritoryTypes:        make(map[string]string),
		HasCoast:              make(map[string]bool),
		ParentCoasts:          make(map[string][]string),
		ParentToCoast:         make(map[string][]string),
		CoastToParent:         make(map[string]string),
		Edges:                 edges,
		AdjacencyMap:          make(map[string][]Adjacency),
		ArmyAdjacency:         make(map[string]map[string]bool),
		FleetAdjacency:        make(map[string]map[string]bool),
	}
	
	// Process territories
	for tid, territory := range territories {
		rules.TerritoryIDs[tid] = true
		rules.TerritoryDisplayNames[tid] = territory.DisplayName
		rules.TerritoryTypes[tid] = territory.Type
		
		if territory.IsSupplyCenter {
			rules.SupplyCenters[tid] = true
		}
		
		if territory.HasCoast {
			rules.HasCoast[tid] = true
		}
		
		if territory.HomeCountry != "" {
			rules.HomeCenters[territory.HomeCountry] = append(rules.HomeCenters[territory.HomeCountry], tid)
		}
		
		// Process coasts
		for _, coast := range territory.Coasts {
			coastID := fmt.Sprintf("%s_%s", tid, coast)
			rules.TerritoryIDs[coastID] = true
			rules.TerritoryTypes[coastID] = "coast"
			rules.TerritoryDisplayNames[coastID] = fmt.Sprintf("%s (%s)", territory.DisplayName, strings.ToUpper(coast))
			rules.ParentCoasts[tid] = append(rules.ParentCoasts[tid], coastID)
			rules.ParentToCoast[tid] = append(rules.ParentToCoast[tid], coastID)
			rules.CoastToParent[coastID] = tid
		}
	}
	
	// Process nations
	for _, nation := range nations {
		rules.NationDisplayNames[nation.ID] = nation.DisplayName
	}
	
	// Build adjacency maps
	for _, edge := range edges {
		// Add both directions
		rules.AdjacencyMap[edge.From] = append(rules.AdjacencyMap[edge.From], Adjacency{
			To:   edge.To,
			Mode: edge.Mode,
		})
		rules.AdjacencyMap[edge.To] = append(rules.AdjacencyMap[edge.To], Adjacency{
			To:   edge.From,
			Mode: edge.Mode,
		})
	}
	
	// Build pre-computed adjacency maps for O(1) validation
	buildUnitAdjacencyMaps(rules)
	
	return rules, nil
}

func buildUnitAdjacencyMaps(rules *Rules) {
	// Initialize maps for all territories
	for territory := range rules.TerritoryIDs {
		rules.ArmyAdjacency[territory] = make(map[string]bool)
		rules.FleetAdjacency[territory] = make(map[string]bool)
	}
	
	// Populate adjacency based on edge modes
	for territory, adjacencies := range rules.AdjacencyMap {
		for _, adj := range adjacencies {
			switch adj.Mode {
			case "land":
				rules.ArmyAdjacency[territory][adj.To] = true
			case "sea":
				rules.FleetAdjacency[territory][adj.To] = true
			case "both":
				rules.ArmyAdjacency[territory][adj.To] = true
				rules.FleetAdjacency[territory][adj.To] = true
			}
		}
	}
}