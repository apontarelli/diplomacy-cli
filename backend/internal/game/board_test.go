package game

import (
	"testing"
)

func TestNewBoard(t *testing.T) {
	board := NewBoard()

	if board == nil {
		t.Fatal("NewBoard() returned nil")
	}

	if board.Provinces == nil {
		t.Error("Provinces map not initialized")
	}

	if board.Units == nil {
		t.Error("Units map not initialized")
	}

	if len(board.Provinces) != 0 {
		t.Error("Expected empty provinces map")
	}

	if len(board.Units) != 0 {
		t.Error("Expected empty units map")
	}
}

func TestAddProvince(t *testing.T) {
	board := NewBoard()

	province := &Province{
		Name:           "paris",
		ShortCode:      "par",
		DisplayName:    "Paris",
		Type:           Land,
		SupplyCenter:   true,
		CoastNeighbors: make(map[string][]string),
		ArmyNeighbors:  []string{"burgundy"},
		FleetNeighbors: []string{},
	}
	board.AddProvince(province)

	if len(board.Provinces) != 1 {
		t.Errorf("Expected 1 province, got %d", len(board.Provinces))
	}

	retrieved := board.GetProvince("paris")
	if retrieved == nil {
		t.Error("Province not found after adding")
	}

	if retrieved.Name != "paris" {
		t.Errorf("Expected province name 'paris', got '%s'", retrieved.Name)
	}
}

func TestPlaceUnit_ValidPlacements(t *testing.T) {
	board := NewBoard()

	landProvince := &Province{
		Name:           "paris",
		Type:           Land,
		SupplyCenter:   true,
		CoastNeighbors: make(map[string][]string),
		ArmyNeighbors:  []string{"burgundy"},
		FleetNeighbors: []string{},
	}
	board.AddProvince(landProvince)

	seaProvince := &Province{
		Name:           "english_channel",
		Type:           Sea,
		SupplyCenter:   false,
		CoastNeighbors: make(map[string][]string),
		ArmyNeighbors:  []string{},
		FleetNeighbors: []string{"london", "brest"},
	}
	board.AddProvince(seaProvince)

	coastalProvince := &Province{
		Name:           "london",
		Type:           Land,
		SupplyCenter:   true,
		CoastNeighbors: make(map[string][]string),
		ArmyNeighbors:  []string{"wales", "yorkshire"},
		FleetNeighbors: []string{"english_channel", "north_sea"},
	}
	board.AddProvince(coastalProvince)

	tests := []struct {
		name string
		unit *Unit
	}{
		{
			name: "Army in land province",
			unit: &Unit{Type: Army, Owner: France, Province: "paris"},
		},
		{
			name: "Fleet in sea province",
			unit: &Unit{Type: Fleet, Owner: England, Province: "english_channel"},
		},
		{
			name: "Fleet in coastal province",
			unit: &Unit{Type: Fleet, Owner: England, Province: "london"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := board.PlaceUnit(tt.unit)
			if err != nil {
				t.Errorf("PlaceUnit() failed: %v", err)
			}

			placed := board.GetUnit(tt.unit.Province)
			if placed == nil {
				t.Error("Unit not found after placement")
			}

			if placed.Type != tt.unit.Type {
				t.Errorf("Expected unit type %s, got %s", tt.unit.Type, placed.Type)
			}
		})
	}
}

func TestPlaceUnit_InvalidPlacements(t *testing.T) {
	board := NewBoard()

	inlandProvince := &Province{
		Name:           "moscow",
		Type:           Land,
		SupplyCenter:   true,
		CoastNeighbors: make(map[string][]string),
		ArmyNeighbors:  []string{"warsaw", "ukraine"},
		FleetNeighbors: []string{}, // No fleet connections
	}
	board.AddProvince(inlandProvince)

	seaProvince := &Province{
		Name:           "black_sea",
		Type:           Sea,
		SupplyCenter:   false,
		CoastNeighbors: make(map[string][]string),
		ArmyNeighbors:  []string{},
		FleetNeighbors: []string{"ankara", "sevastopol"},
	}
	board.AddProvince(seaProvince)

	tests := []struct {
		name        string
		unit        *Unit
		expectedErr string
	}{
		{
			name:        "Army in sea province",
			unit:        &Unit{Type: Army, Owner: Russia, Province: "black_sea"},
			expectedErr: "cannot place army in sea province",
		},
		{
			name:        "Fleet in inland province",
			unit:        &Unit{Type: Fleet, Owner: Russia, Province: "moscow"},
			expectedErr: "cannot place fleet in province",
		},
		{
			name:        "Unit in nonexistent province",
			unit:        &Unit{Type: Army, Owner: France, Province: "atlantis"},
			expectedErr: "province atlantis does not exist",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := board.PlaceUnit(tt.unit)
			if err == nil {
				t.Error("Expected error but got none")
			}

			if err != nil && tt.expectedErr != "" {
				if len(err.Error()) == 0 || err.Error()[:len(tt.expectedErr)] != tt.expectedErr {
					t.Errorf("Expected error containing '%s', got '%s'", tt.expectedErr, err.Error())
				}
			}
		})
	}
}

func TestPlaceUnit_OccupiedProvince(t *testing.T) {
	board := NewBoard()

	province := &Province{
		Name:           "paris",
		Type:           Land,
		SupplyCenter:   true,
		CoastNeighbors: make(map[string][]string),
		ArmyNeighbors:  []string{"burgundy"},
		FleetNeighbors: []string{},
	}
	board.AddProvince(province)

	unit1 := &Unit{Type: Army, Owner: France, Province: "paris"}
	err := board.PlaceUnit(unit1)
	if err != nil {
		t.Fatalf("Failed to place first unit: %v", err)
	}

	unit2 := &Unit{Type: Army, Owner: Germany, Province: "paris"}
	err = board.PlaceUnit(unit2)
	if err == nil {
		t.Error("Expected error when placing unit in occupied province")
	}

	if err != nil && err.Error() != "province paris is already occupied" {
		t.Errorf("Expected 'province paris is already occupied', got '%s'", err.Error())
	}
}

func TestRemoveUnit(t *testing.T) {
	board := NewBoard()

	province := &Province{
		Name:           "paris",
		Type:           Land,
		SupplyCenter:   true,
		CoastNeighbors: make(map[string][]string),
		ArmyNeighbors:  []string{"burgundy"},
		FleetNeighbors: []string{},
	}
	board.AddProvince(province)

	unit := &Unit{Type: Army, Owner: France, Province: "paris"}
	board.PlaceUnit(unit)

	if board.GetUnit("paris") == nil {
		t.Error("Unit should be present before removal")
	}

	board.RemoveUnit("paris")

	if board.GetUnit("paris") != nil {
		t.Error("Unit should be removed")
	}

	board.RemoveUnit("nonexistent")
}

func TestGetUnit(t *testing.T) {
	board := NewBoard()

	province := &Province{
		Name:           "paris",
		Type:           Land,
		SupplyCenter:   true,
		CoastNeighbors: make(map[string][]string),
		ArmyNeighbors:  []string{"burgundy"},
		FleetNeighbors: []string{},
	}
	board.AddProvince(province)

	unit := board.GetUnit("paris")
	if unit != nil {
		t.Error("Expected nil for empty province")
	}

	placedUnit := &Unit{Type: Army, Owner: France, Province: "paris"}
	board.PlaceUnit(placedUnit)

	retrieved := board.GetUnit("paris")
	if retrieved == nil {
		t.Error("Expected unit but got nil")
	}

	if retrieved.Type != Army || retrieved.Owner != France {
		t.Error("Retrieved unit has wrong properties")
	}

	nonexistent := board.GetUnit("atlantis")
	if nonexistent != nil {
		t.Error("Expected nil for nonexistent province")
	}
}

func TestGetProvince(t *testing.T) {
	board := NewBoard()

	province := &Province{
		Name:           "paris",
		Type:           Land,
		SupplyCenter:   true,
		CoastNeighbors: make(map[string][]string),
		ArmyNeighbors:  []string{"burgundy"},
		FleetNeighbors: []string{},
	}
	board.AddProvince(province)

	retrieved := board.GetProvince("paris")
	if retrieved == nil {
		t.Error("Expected province but got nil")
	}

	if retrieved.Name != "paris" || retrieved.Type != Land {
		t.Error("Retrieved province has wrong properties")
	}

	nonexistent := board.GetProvince("atlantis")
	if nonexistent != nil {
		t.Error("Expected nil for nonexistent province")
	}
}

func TestIsAdjacent(t *testing.T) {
	board := NewBoard()

	paris := &Province{
		Name:           "paris",
		Type:           Land,
		SupplyCenter:   true,
		CoastNeighbors: make(map[string][]string),
		ArmyNeighbors:  []string{"burgundy", "picardy"},
		FleetNeighbors: []string{},
	}
	board.AddProvince(paris)

	london := &Province{
		Name:           "london",
		Type:           Land,
		SupplyCenter:   true,
		CoastNeighbors: make(map[string][]string),
		ArmyNeighbors:  []string{"wales", "yorkshire"},
		FleetNeighbors: []string{"english_channel", "north_sea"},
	}
	board.AddProvince(london)

	tests := []struct {
		name     string
		from     string
		to       string
		expected bool
	}{
		{
			name:     "Army neighbor exists",
			from:     "paris",
			to:       "burgundy",
			expected: true,
		},
		{
			name:     "Fleet neighbor exists",
			from:     "london",
			to:       "english_channel",
			expected: true,
		},
		{
			name:     "No adjacency",
			from:     "paris",
			to:       "london",
			expected: false,
		},
		{
			name:     "Nonexistent from province",
			from:     "atlantis",
			to:       "paris",
			expected: false,
		},
		{
			name:     "Reverse adjacency check",
			from:     "burgundy",
			to:       "paris",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := board.IsAdjacent(tt.from, tt.to)
			if result != tt.expected {
				t.Errorf("IsAdjacent(%s, %s) = %v, expected %v", tt.from, tt.to, result, tt.expected)
			}
		})
	}
}

func TestGetSupplyCenters(t *testing.T) {
	board := NewBoard()

	paris := &Province{Name: "paris", Type: Land, SupplyCenter: true}
	board.AddProvince(paris)

	burgundy := &Province{Name: "burgundy", Type: Land, SupplyCenter: false}
	board.AddProvince(burgundy)

	london := &Province{Name: "london", Type: Land, SupplyCenter: true}
	board.AddProvince(london)

	centers := board.GetSupplyCenters()

	if len(centers) != 2 {
		t.Errorf("Expected 2 supply centers, got %d", len(centers))
	}

	found := make(map[string]bool)
	for _, center := range centers {
		found[center.Name] = true
	}

	if !found["paris"] || !found["london"] {
		t.Error("Missing expected supply centers")
	}

	if found["burgundy"] {
		t.Error("Non-supply center included in results")
	}
}

func TestGetUnitsByOwner(t *testing.T) {
	board := NewBoard()

	paris := &Province{Name: "paris", Type: Land, ArmyNeighbors: []string{}, FleetNeighbors: []string{}}
	london := &Province{Name: "london", Type: Land, ArmyNeighbors: []string{}, FleetNeighbors: []string{}}
	berlin := &Province{Name: "berlin", Type: Land, ArmyNeighbors: []string{}, FleetNeighbors: []string{}}

	board.AddProvince(paris)
	board.AddProvince(london)
	board.AddProvince(berlin)

	frenchArmy := &Unit{Type: Army, Owner: France, Province: "paris"}
	englishArmy := &Unit{Type: Army, Owner: England, Province: "london"}
	germanArmy := &Unit{Type: Army, Owner: Germany, Province: "berlin"}

	board.PlaceUnit(frenchArmy)
	board.PlaceUnit(englishArmy)
	board.PlaceUnit(germanArmy)

	frenchUnits := board.GetUnitsByOwner(France)
	if len(frenchUnits) != 1 {
		t.Errorf("Expected 1 French unit, got %d", len(frenchUnits))
	}
	if len(frenchUnits) > 0 && frenchUnits[0].Province != "paris" {
		t.Error("Wrong French unit returned")
	}

	englishUnits := board.GetUnitsByOwner(England)
	if len(englishUnits) != 1 {
		t.Errorf("Expected 1 English unit, got %d", len(englishUnits))
	}

	italianUnits := board.GetUnitsByOwner(Italy)
	if len(italianUnits) != 0 {
		t.Errorf("Expected 0 Italian units, got %d", len(italianUnits))
	}
}

func TestHasCoasts(t *testing.T) {
	stPetersburg := &Province{
		Name: "st_petersburg",
		Type: Land,
		CoastNeighbors: map[string][]string{
			"nc": {"barents_sea"},
			"sc": {"gulf_of_bothnia"},
		},
	}

	if !stPetersburg.hasCoasts() {
		t.Error("Expected province with coasts to return true")
	}

	moscow := &Province{
		Name:           "moscow",
		Type:           Land,
		CoastNeighbors: make(map[string][]string),
	}

	if moscow.hasCoasts() {
		t.Error("Expected province without coasts to return false")
	}
}

func TestCanPlaceFleet(t *testing.T) {
	board := NewBoard()

	seaProvince := &Province{
		Name:           "english_channel",
		Type:           Sea,
		FleetNeighbors: []string{"london", "brest"},
	}
	board.AddProvince(seaProvince)

	coastalProvince := &Province{
		Name:           "london",
		Type:           Land,
		FleetNeighbors: []string{"english_channel", "north_sea"},
	}
	board.AddProvince(coastalProvince)

	inlandProvince := &Province{
		Name:           "moscow",
		Type:           Land,
		FleetNeighbors: []string{},
	}
	board.AddProvince(inlandProvince)

	tests := []struct {
		name     string
		province string
		expected bool
	}{
		{
			name:     "Sea province allows fleets",
			province: "english_channel",
			expected: true,
		},
		{
			name:     "Coastal province allows fleets",
			province: "london",
			expected: true,
		},
		{
			name:     "Inland province does not allow fleets",
			province: "moscow",
			expected: false,
		},
		{
			name:     "Nonexistent province does not allow fleets",
			province: "atlantis",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := board.canPlaceFleet(tt.province)
			if result != tt.expected {
				t.Errorf("canPlaceFleet(%s) = %v, expected %v", tt.province, result, tt.expected)
			}
		})
	}
}
