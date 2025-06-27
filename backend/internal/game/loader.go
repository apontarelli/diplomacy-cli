package game

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type DataLoader struct {
	dataPath string
}

func NewDataLoader(dataPath string) *DataLoader {
	return &DataLoader{dataPath: dataPath}
}

func (dl *DataLoader) LoadClassicGameData() (*GameData, error) {
	nations, err := dl.loadNations()
	if err != nil {
		return nil, fmt.Errorf("failed to load nations: %w", err)
	}

	territories, err := dl.loadTerritories()
	if err != nil {
		return nil, fmt.Errorf("failed to load territories: %w", err)
	}

	edges, err := dl.loadEdges()
	if err != nil {
		return nil, fmt.Errorf("failed to load edges: %w", err)
	}

	return &GameData{
		Nations:     nations,
		Territories: territories,
		Edges:       edges,
	}, nil
}

func (dl *DataLoader) loadNations() ([]Nation, error) {
	path := filepath.Join(dl.dataPath, "world", "nations.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var nations []Nation
	if err := json.Unmarshal(data, &nations); err != nil {
		return nil, err
	}

	return nations, nil
}

func (dl *DataLoader) loadTerritories() (map[string]Territory, error) {
	path := filepath.Join(dl.dataPath, "world", "territories.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var territories map[string]Territory
	if err := json.Unmarshal(data, &territories); err != nil {
		return nil, err
	}

	// Set the ID field for each territory
	for id, territory := range territories {
		territory.ID = id
		territories[id] = territory
	}

	return territories, nil
}

func (dl *DataLoader) loadEdges() ([]Edge, error) {
	path := filepath.Join(dl.dataPath, "world", "edges.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var edges []Edge
	if err := json.Unmarshal(data, &edges); err != nil {
		return nil, err
	}

	return edges, nil
}