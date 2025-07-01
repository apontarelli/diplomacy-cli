package loader

import (
	"testing"
)

func TestJSONLoader_LoadBoard(t *testing.T) {
	loader := NewJSONLoader("../../../data/classic")

	board, err := loader.LoadBoard()
	if err != nil {
		t.Fatalf("Failed to load board: %v", err)
	}

	if board == nil {
		t.Fatal("Board is nil")
	}

	if len(board.Provinces) == 0 {
		t.Fatal("No provinces loaded")
	}

	testProvinces := []string{"london", "paris", "berlin", "vienna", "rome", "moscow", "constantinople"}
	for _, provinceName := range testProvinces {
		province := board.GetProvince(provinceName)
		if province == nil {
			t.Errorf("Province %s not found", provinceName)
		}
	}

	spain := board.GetProvince("spain")
	if spain != nil && spain.CoastNeighbors != nil {
		if len(spain.CoastNeighbors) == 0 {
			t.Error("Spain should have coast neighbors")
		}
		if _, hasNC := spain.CoastNeighbors["nc"]; !hasNC {
			t.Error("Spain should have north coast")
		}
		if _, hasSC := spain.CoastNeighbors["sc"]; !hasSC {
			t.Error("Spain should have south coast")
		}
	}

	stPetersburg := board.GetProvince("st_petersburg")
	if stPetersburg != nil && stPetersburg.CoastNeighbors != nil {
		if len(stPetersburg.CoastNeighbors) == 0 {
			t.Error("St. Petersburg should have coast neighbors")
		}
		if _, hasNC := stPetersburg.CoastNeighbors["nc"]; !hasNC {
			t.Error("St. Petersburg should have north coast")
		}
		if _, hasSC := stPetersburg.CoastNeighbors["sc"]; !hasSC {
			t.Error("St. Petersburg should have south coast")
		}
	}
}

func TestJSONLoader_LoadStartingUnits(t *testing.T) {
	loader := NewJSONLoader("../../../data/classic")

	units, err := loader.LoadStartingUnits()
	if err != nil {
		t.Fatalf("Failed to load starting units: %v", err)
	}

	if len(units) == 0 {
		t.Fatal("No starting units loaded")
	}

	expectedUnits := 22
	if len(units) != expectedUnits {
		t.Errorf("Expected %d units, got %d", expectedUnits, len(units))
	}
}

func TestJSONLoader_LoadStartingSupplyCenters(t *testing.T) {
	loader := NewJSONLoader("../../../data/classic")

	supplyCenters, err := loader.LoadStartingSupplyCenters()
	if err != nil {
		t.Fatalf("Failed to load starting supply centers: %v", err)
	}

	if len(supplyCenters) == 0 {
		t.Fatal("No starting supply centers loaded")
	}

	expectedNations := 7
	if len(supplyCenters) != expectedNations {
		t.Errorf("Expected %d nations, got %d", expectedNations, len(supplyCenters))
	}
	expectedCounts := map[string]int{
		"england": 3,
		"france":  3,
		"germany": 3,
		"italy":   3,
		"austria": 3,
		"russia":  4,
		"turkey":  3,
	}

	for nation, centers := range supplyCenters {
		expected := expectedCounts[string(nation)]
		if len(centers) != expected {
			t.Errorf("Nation %s should have %d supply centers, got %d", nation, expected, len(centers))
		}
	}
}

func TestParseProvinceAndCoast(t *testing.T) {
	loader := NewJSONLoader("")

	tests := []struct {
		input            string
		expectedProvince string
		expectedCoast    string
	}{
		{"london", "london", ""},
		{"st_petersburg_nc", "st_petersburg", "nc"},
		{"st_petersburg_sc", "st_petersburg", "sc"},
		{"bulgaria_ec", "bulgaria", "ec"},
		{"bulgaria_sc", "bulgaria", "sc"},
		{"spain_nc", "spain", "nc"},
		{"spain_sc", "spain", "sc"},
		{"english_channel", "english_channel", ""},
	}

	for _, test := range tests {
		province, coast := loader.parseProvinceAndCoast(test.input)
		if province != test.expectedProvince {
			t.Errorf("For input %s, expected province %s, got %s", test.input, test.expectedProvince, province)
		}
		if coast != test.expectedCoast {
			t.Errorf("For input %s, expected coast %s, got %s", test.input, test.expectedCoast, coast)
		}
	}
}
