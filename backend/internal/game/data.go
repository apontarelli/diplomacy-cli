package game

type Nation struct {
	ID          string `json:"id"`
	DisplayName string `json:"display_name"`
}

type TerritoryType string

const (
	TerritoryTypeLand TerritoryType = "land"
	TerritoryTypeSea  TerritoryType = "sea"
)

type Territory struct {
	ID              string        `json:"id"`
	DisplayName     string        `json:"display_name"`
	Type            TerritoryType `json:"type"`
	HomeCountry     string        `json:"home_country,omitempty"`
	HasCoast        bool          `json:"has_coast,omitempty"`
	IsSupplyCenter  bool          `json:"is_supply_center,omitempty"`
}

type EdgeMode string

const (
	EdgeModeLand EdgeMode = "land"
	EdgeModeSea  EdgeMode = "sea"
)

type Edge struct {
	From string   `json:"from"`
	To   string   `json:"to"`
	Mode EdgeMode `json:"mode"`
}

type GameData struct {
	Nations     []Nation
	Territories map[string]Territory
	Edges       []Edge
}