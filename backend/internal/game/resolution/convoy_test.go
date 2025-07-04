package resolution

import (
	"diplomacy-cli/backend/internal/game"
	"testing"
)

func TestConvoyPathDiscovery(t *testing.T) {
	board := &game.Board{
		Provinces: map[string]*game.Province{
			"london": {
				Name:           "london",
				Type:           game.Land,
				FleetNeighbors: []string{"english_channel"},
			},
			"english_channel": {
				Name:           "english_channel",
				Type:           game.Sea,
				FleetNeighbors: []string{"london", "brest", "mid_atlantic_ocean"},
			},
			"mid_atlantic_ocean": {
				Name:           "mid_atlantic_ocean",
				Type:           game.Sea,
				FleetNeighbors: []string{"english_channel", "brest"},
			},
			"brest": {
				Name:           "brest",
				Type:           game.Land,
				FleetNeighbors: []string{"english_channel", "mid_atlantic_ocean"},
			},
		},
	}

	orders := []*game.Order{
		{
			ID:       "1",
			UnitType: game.Army,
			From:     "london",
			Type:     game.Move,
			To:       "brest",
			Owner:    game.England,
		},
		{
			ID:           "2",
			UnitType:     game.Fleet,
			From:         "english_channel",
			Type:         game.Convoy,
			ConvoyTarget: "london",
			Owner:        game.England,
		},
		{
			ID:           "3",
			UnitType:     game.Fleet,
			From:         "mid_atlantic_ocean",
			Type:         game.Convoy,
			ConvoyTarget: "london",
			Owner:        game.England,
		},
	}

	engine := NewResolutionEngine(orders, board)
	engine.processConvoys()

	origin := NewProvinceCoast("london", "")
	destination := NewProvinceCoast("brest", "")
	convoyKey := ConvoyKey{Origin: origin, Destination: destination}

	path, exists := engine.convoys[convoyKey]
	if !exists {
		t.Fatal("Expected convoy path to be found")
	}

	if len(path) < 2 {
		t.Fatalf("Expected convoy path with at least 2 steps, got %d", len(path))
	}

	if path[0].Province != "london" {
		t.Errorf("Expected path to start at london, got %s", path[0].Province)
	}

	if path[len(path)-1].Province != "brest" {
		t.Errorf("Expected path to end at brest, got %s", path[len(path)-1].Province)
	}
}

func TestConvoyPathNotFound(t *testing.T) {
	board := &game.Board{
		Provinces: map[string]*game.Province{
			"london": {
				Name:           "london",
				Type:           game.Land,
				FleetNeighbors: []string{"english_channel"},
			},
			"english_channel": {
				Name:           "english_channel",
				Type:           game.Sea,
				FleetNeighbors: []string{"london"},
			},
			"brest": {
				Name: "brest",
				Type: game.Land,
			},
		},
	}

	orders := []*game.Order{
		{
			ID:       "1",
			UnitType: game.Army,
			From:     "london",
			Type:     game.Move,
			To:       "brest",
			Owner:    game.England,
		},
	}

	engine := NewResolutionEngine(orders, board)
	engine.processConvoys()

	origin := NewProvinceCoast("london", "")
	destination := NewProvinceCoast("brest", "")
	convoyKey := ConvoyKey{Origin: origin, Destination: destination}

	_, exists := engine.convoys[convoyKey]
	if exists {
		t.Error("Expected no convoy path to be found")
	}
}

func TestParseProvinceCoastString(t *testing.T) {
	tests := []struct {
		name             string
		input            string
		expectedProvince string
		expectedCoast    string
	}{
		{
			name:             "Coast specification",
			input:            "st_petersburg/nc",
			expectedProvince: "st_petersburg",
			expectedCoast:    "nc",
		},
		{
			name:             "No coast specification",
			input:            "london",
			expectedProvince: "london",
			expectedCoast:    "",
		},
		{
			name:             "Another coast specification",
			input:            "spain/sc",
			expectedProvince: "spain",
			expectedCoast:    "sc",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pc := parseProvinceCoastString(tt.input)

			if pc.Province != tt.expectedProvince {
				t.Errorf("Expected province %s, got %s", tt.expectedProvince, pc.Province)
			}
			if pc.Coast != tt.expectedCoast {
				t.Errorf("Expected coast %s, got %s", tt.expectedCoast, pc.Coast)
			}
		})
	}
}

func TestProvinceCoastMethods(t *testing.T) {
	t.Run("String method", func(t *testing.T) {
		pc1 := NewProvinceCoast("london", "")
		if pc1.String() != "london" {
			t.Errorf("Expected 'london', got '%s'", pc1.String())
		}

		pc2 := NewProvinceCoast("st_petersburg", "nc")
		if pc2.String() != "st_petersburg/nc" {
			t.Errorf("Expected 'st_petersburg/nc', got '%s'", pc2.String())
		}
	})

	t.Run("HasCoast method", func(t *testing.T) {
		pc1 := NewProvinceCoast("london", "")
		if pc1.HasCoast() {
			t.Error("Expected HasCoast() to return false for province without coast")
		}

		pc2 := NewProvinceCoast("st_petersburg", "nc")
		if !pc2.HasCoast() {
			t.Error("Expected HasCoast() to return true for province with coast")
		}
	})
}
