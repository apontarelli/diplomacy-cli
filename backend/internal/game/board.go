package game

import "fmt"

type ProvinceType string

const (
	Land ProvinceType = "land"
	Sea  ProvinceType = "sea"
)

type UnitType string

const (
	Army  UnitType = "army"
	Fleet UnitType = "fleet"
)

type Nation string

const (
	Austria Nation = "austria"
	England Nation = "england"
	France  Nation = "france"
	Germany Nation = "germany"
	Italy   Nation = "italy"
	Russia  Nation = "russia"
	Turkey  Nation = "turkey"
)

type Province struct {
	Name           string // Primary key: "st_petersburg", "english_channel" (snake_case)
	ShortCode      string // Compact form: "stp", "eng" (for user input/display)
	DisplayName    string // Human readable: "St. Petersburg", "English Channel"
	Type           ProvinceType
	SupplyCenter   bool
	CoastNeighbors map[string][]string // coast name -> adjacent provinces/coasts
	ArmyNeighbors  []string            // provinces armies can reach
	FleetNeighbors []string            // provinces fleets can reach
}

type Unit struct {
	Type      UnitType
	Owner     Nation
	Province  string
	Coast     string
	Dislodged bool
}

type Board struct {
	Provinces map[string]*Province
	Units     map[string]*Unit
}

func NewBoard() *Board {
	return &Board{
		Provinces: make(map[string]*Province),
		Units:     make(map[string]*Unit),
	}
}

func (b *Board) AddProvince(province *Province) {
	b.Provinces[province.Name] = province
}

func (b *Board) PlaceUnit(unit *Unit) error {
	province, exists := b.Provinces[unit.Province]
	if !exists {
		return fmt.Errorf("province %s does not exist", unit.Province)
	}

	if unit.Type == Army && province.Type == Sea {
		return fmt.Errorf("cannot place army in sea province %s", unit.Province)
	}
	if unit.Type == Fleet && !b.canPlaceFleet(unit.Province) {
		return fmt.Errorf("cannot place fleet in province %s", unit.Province)
	}

	if b.Units[unit.Province] != nil {
		return fmt.Errorf("province %s is already occupied", unit.Province)
	}

	b.Units[unit.Province] = unit
	return nil
}

func (b *Board) RemoveUnit(provinceName string) {
	delete(b.Units, provinceName)
}

func (b *Board) GetUnit(provinceName string) *Unit {
	return b.Units[provinceName]
}

func (b *Board) GetProvince(name string) *Province {
	return b.Provinces[name]
}

func (b *Board) IsAdjacent(from, to string) bool {
	province := b.GetProvince(from)
	if province == nil {
		return false
	}

	for _, neighbor := range province.ArmyNeighbors {
		if neighbor == to {
			return true
		}
	}
	for _, neighbor := range province.FleetNeighbors {
		if neighbor == to {
			return true
		}
	}
	return false
}

func (p *Province) hasCoasts() bool {
	return len(p.CoastNeighbors) > 0
}

func (b *Board) canPlaceFleet(provinceName string) bool {
	province := b.GetProvince(provinceName)
	if province == nil {
		return false
	}

	return province.Type == Sea || len(province.FleetNeighbors) > 0
}

func (b *Board) GetSupplyCenters() []*Province {
	var centers []*Province
	for _, province := range b.Provinces {
		if province.SupplyCenter {
			centers = append(centers, province)
		}
	}
	return centers
}

func (b *Board) GetUnitsByOwner(owner Nation) []*Unit {
	var units []*Unit
	for _, unit := range b.Units {
		if unit != nil && unit.Owner == owner {
			units = append(units, unit)
		}
	}
	return units
}
