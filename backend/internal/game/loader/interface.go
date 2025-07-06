package loader

import "diplomacy-cli/backend/internal/game"

type MapLoader interface {
	LoadBoard() (*game.Board, error)
	LoadStartingUnits() ([]*game.Unit, error)
	LoadStartingSupplyCenters() (map[game.Nation][]string, error)
	LoadVariantMetadata() (*VariantMetadata, error)
	GetMapName() string
}

type VariantMetadata struct {
	Name        string   `json:"name"`
	ShortName   string   `json:"shortName"`
	Description string   `json:"description"`
	Version     string   `json:"version"`
	Author      string   `json:"author"`
	PlayerCount int      `json:"playerCount"`
	StartYear   int      `json:"startYear"`
	Nations     []string `json:"nations"`
}

type MapVariant string

const (
	Classic MapVariant = "classic"
)

func GetLoader(variant MapVariant) MapLoader {
	switch variant {
	case Classic:
		return NewJSONLoader("../../../data/classic")
	default:
		return NewJSONLoader("../../../data/classic")
	}
}
