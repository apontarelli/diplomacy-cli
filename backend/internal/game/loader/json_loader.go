package loader

import (
	"diplomacy-cli/backend/internal/game"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"slices"
)

type JSONLoader struct {
	dataPath string
}

func NewJSONLoader(dataPath string) *JSONLoader {
	return &JSONLoader{
		dataPath: dataPath,
	}
}

func (jl *JSONLoader) GetMapName() string {
	metadata, err := jl.LoadVariantMetadata()
	if err != nil {
		return "Classic Diplomacy (JSON)"
	}
	return metadata.Name
}

func (jl *JSONLoader) LoadVariantMetadata() (*VariantMetadata, error) {
	variantPath := filepath.Join(jl.dataPath, "variant.json")
	data, err := os.ReadFile(variantPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read variant metadata file: %w", err)
	}

	var metadata VariantMetadata
	if err := json.Unmarshal(data, &metadata); err != nil {
		return nil, fmt.Errorf("failed to parse variant metadata JSON: %w", err)
	}

	return &metadata, nil
}

type Province struct {
	ShortCode      string   `json:"shortCode"`
	DisplayName    string   `json:"displayName"`
	Type           string   `json:"type"`
	HomeCountry    string   `json:"homeCountry"`
	HasCoast       bool     `json:"hasCoast"`
	IsSupplyCenter bool     `json:"isSupplyCenter"`
	Coasts         []string `json:"coasts"`
}

type Edge struct {
	From string `json:"from"`
	To   string `json:"to"`
	Mode string `json:"mode"`
}

type Unit struct {
	Owner      string `json:"owner"`
	Type       string `json:"type"`
	LocationId string `json:"locationId"`
}

type Ownership struct {
	ProvinceId string `json:"provinceId"`
	Owner      string `json:"owner"`
}

func (jl *JSONLoader) parseProvinceAndCoast(location string) (string, string) {
	coastSuffixes := []string{"_nc", "_sc", "_ec", "_wc"}
	for _, suffix := range coastSuffixes {
		if len(location) > len(suffix) && location[len(location)-len(suffix):] == suffix {
			province := location[:len(location)-len(suffix)]
			coast := suffix[1:]
			return province, coast
		}
	}
	return location, ""
}

func addUniqueNeighbor(neighbors []string, neighbor string) []string {
	if !slices.Contains(neighbors, neighbor) {
		return append(neighbors, neighbor)
	}
	return neighbors
}

func (jl *JSONLoader) LoadBoard() (*game.Board, error) {
	board := game.NewBoard()

	provinces, err := jl.loadProvinces()
	if err != nil {
		return nil, fmt.Errorf("failed to load provinces: %w", err)
	}

	edges, err := jl.loadEdges()
	if err != nil {
		return nil, fmt.Errorf("failed to load edges: %w", err)
	}

	armyNeighbors := make(map[string][]string)
	fleetNeighbors := make(map[string][]string)
	coastNeighbors := make(map[string]map[string][]string)

	for name, jsonProvince := range provinces {
		if jsonProvince.HasCoast && len(jsonProvince.Coasts) > 0 {
			coastNeighbors[name] = make(map[string][]string)
			for _, coast := range jsonProvince.Coasts {
				coastNeighbors[name][coast] = []string{}
			}
		}
	}

	for _, edge := range edges {
		fromProvince, fromCoast := jl.parseProvinceAndCoast(edge.From)
		toProvince, toCoast := jl.parseProvinceAndCoast(edge.To)

		switch edge.Mode {
		case "land":
			armyNeighbors[fromProvince] = addUniqueNeighbor(armyNeighbors[fromProvince], toProvince)
			armyNeighbors[toProvince] = addUniqueNeighbor(armyNeighbors[toProvince], fromProvince)
		case "sea":
			fleetNeighbors[fromProvince] = addUniqueNeighbor(fleetNeighbors[fromProvince], toProvince)
			fleetNeighbors[toProvince] = addUniqueNeighbor(fleetNeighbors[toProvince], fromProvince)

			if fromCoast != "" && coastNeighbors[fromProvince] != nil {
				coastNeighbors[fromProvince][fromCoast] = addUniqueNeighbor(coastNeighbors[fromProvince][fromCoast], edge.To)
			}
			if toCoast != "" && coastNeighbors[toProvince] != nil {
				coastNeighbors[toProvince][toCoast] = addUniqueNeighbor(coastNeighbors[toProvince][toCoast], edge.From)
			}
		case "coast":
			armyNeighbors[fromProvince] = addUniqueNeighbor(armyNeighbors[fromProvince], toProvince)
			armyNeighbors[toProvince] = addUniqueNeighbor(armyNeighbors[toProvince], fromProvince)
			fleetNeighbors[fromProvince] = addUniqueNeighbor(fleetNeighbors[fromProvince], toProvince)
			fleetNeighbors[toProvince] = addUniqueNeighbor(fleetNeighbors[toProvince], fromProvince)
		}
	}

	for name, jsonProvince := range provinces {
		var provinceType game.ProvinceType
		switch jsonProvince.Type {
		case "land":
			provinceType = game.Land
		case "sea":
			provinceType = game.Sea
		default:
			return nil, fmt.Errorf("unknown province type: %s", jsonProvince.Type)
		}

		province := &game.Province{
			Name:           name,
			ShortCode:      jsonProvince.ShortCode,
			DisplayName:    jsonProvince.DisplayName,
			Type:           provinceType,
			SupplyCenter:   jsonProvince.IsSupplyCenter,
			CoastNeighbors: coastNeighbors[name],
			ArmyNeighbors:  armyNeighbors[name],
			FleetNeighbors: fleetNeighbors[name],
		}

		board.AddProvince(province)
	}

	return board, nil
}

func (jl *JSONLoader) LoadStartingUnits() ([]*game.Unit, error) {
	unitsPath := filepath.Join(jl.dataPath, "start", "starting_units.json")
	data, err := os.ReadFile(unitsPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read starting units file: %w", err)
	}

	var jsonUnits []Unit
	if err := json.Unmarshal(data, &jsonUnits); err != nil {
		return nil, fmt.Errorf("failed to parse starting units JSON: %w", err)
	}

	var units []*game.Unit
	for _, jsonUnit := range jsonUnits {
		var unitType game.UnitType
		switch jsonUnit.Type {
		case "army":
			unitType = game.Army
		case "fleet":
			unitType = game.Fleet
		default:
			return nil, fmt.Errorf("unknown unit type: %s", jsonUnit.Type)
		}

		unit := &game.Unit{
			Type:     unitType,
			Owner:    game.Nation(jsonUnit.Owner),
			Province: jsonUnit.LocationId,
		}

		units = append(units, unit)
	}

	return units, nil
}

func (jl *JSONLoader) LoadStartingSupplyCenters() (map[game.Nation][]string, error) {
	ownershipsPath := filepath.Join(jl.dataPath, "start", "starting_ownerships.json")
	data, err := os.ReadFile(ownershipsPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read starting ownerships file: %w", err)
	}

	var jsonOwnerships []Ownership
	if err := json.Unmarshal(data, &jsonOwnerships); err != nil {
		return nil, fmt.Errorf("failed to parse starting ownerships JSON: %w", err)
	}

	supplyCenters := make(map[game.Nation][]string)
	for _, ownership := range jsonOwnerships {
		nation := game.Nation(ownership.Owner)
		supplyCenters[nation] = append(supplyCenters[nation], ownership.ProvinceId)
	}

	return supplyCenters, nil
}

func (jl *JSONLoader) loadProvinces() (map[string]Province, error) {
	provincesPath := filepath.Join(jl.dataPath, "world", "provinces.json")
	data, err := os.ReadFile(provincesPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read provinces file: %w", err)
	}

	var provinces map[string]Province
	if err := json.Unmarshal(data, &provinces); err != nil {
		return nil, fmt.Errorf("failed to parse provinces JSON: %w", err)
	}

	return provinces, nil
}

func (jl *JSONLoader) loadEdges() ([]Edge, error) {
	edgesPath := filepath.Join(jl.dataPath, "world", "edges.json")
	data, err := os.ReadFile(edgesPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read edges file: %w", err)
	}

	var edges []Edge
	if err := json.Unmarshal(data, &edges); err != nil {
		return nil, fmt.Errorf("failed to parse edges JSON: %w", err)
	}

	return edges, nil
}
