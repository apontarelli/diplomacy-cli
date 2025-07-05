package validation

import (
	"diplomacy-cli/backend/internal/game"
	"diplomacy-cli/backend/internal/game/loader"
	"testing"
)

func TestProvinceResolver_MultipleFormats(t *testing.T) {
	// Load the classic map for real data
	mapLoader := loader.NewJSONLoader("../../../data/classic")
	board, err := mapLoader.LoadBoard()
	if err != nil {
		t.Fatalf("Failed to load board: %v", err)
	}

	resolver := NewProvinceResolver(board)

	// Test all three formats for North Sea
	testCases := []struct {
		format   string
		input    string
		expected string
	}{
		{"DATC Standard", "North Sea", "north_sea"},
		{"Shortcode", "nth", "north_sea"},
		{"Snake Case", "north_sea", "north_sea"},
		{"Mixed Case", "NORTH SEA", "north_sea"},
		{"Mixed Case", "NTH", "north_sea"},
	}

	for _, tc := range testCases {
		t.Run(tc.format+"_"+tc.input, func(t *testing.T) {
			canonical, exists := resolver.ResolveProvince(tc.input)
			if !exists {
				t.Errorf("Format %s: Input '%s' should resolve to a province", tc.format, tc.input)
				return
			}
			if canonical != tc.expected {
				t.Errorf("Format %s: Input '%s' resolved to '%s', expected '%s'",
					tc.format, tc.input, canonical, tc.expected)
			}
		})
	}
}

func TestProvinceResolver_CoastFormats(t *testing.T) {
	// Load the classic map for real data
	mapLoader := loader.NewJSONLoader("../../../data/classic")
	board, err := mapLoader.LoadBoard()
	if err != nil {
		t.Fatalf("Failed to load board: %v", err)
	}

	resolver := NewProvinceResolver(board)

	// Test coast parsing with different formats
	coastCases := []struct {
		format           string
		input            string
		expectedProvince string
		expectedCoast    string
	}{
		{"DATC Standard", "St. Petersburg/nc", "st_petersburg", "nc"},
		{"Shortcode", "stp/nc", "st_petersburg", "nc"},
		{"Snake Case", "st_petersburg/nc", "st_petersburg", "nc"},
		{"Mixed Case", "ST. PETERSBURG/NC", "st_petersburg", "nc"},
		{"Mixed Case", "STP/NC", "st_petersburg", "nc"},
	}

	for _, tc := range coastCases {
		t.Run(tc.format+"_"+tc.input, func(t *testing.T) {
			province, coast, exists := resolver.ParseCoast(tc.input)
			if !exists {
				t.Errorf("Format %s: Input '%s' should resolve to a province with coast", tc.format, tc.input)
				return
			}
			if province != tc.expectedProvince {
				t.Errorf("Format %s: Input '%s' resolved to province '%s', expected '%s'",
					tc.format, tc.input, province, tc.expectedProvince)
			}
			if coast != tc.expectedCoast {
				t.Errorf("Format %s: Input '%s' resolved to coast '%s', expected '%s'",
					tc.format, tc.input, coast, tc.expectedCoast)
			}
		})
	}
}

func TestProvinceResolver_InvalidInputs(t *testing.T) {
	// Create a minimal board for testing
	board := game.NewBoard()
	board.AddProvince(&game.Province{
		Name:        "paris",
		ShortCode:   "par",
		DisplayName: "Paris",
		Type:        game.Land,
	})

	resolver := NewProvinceResolver(board)

	invalidCases := []string{
		"nonexistent",
		"atlantis",
		"",
		"   ",
	}

	for _, input := range invalidCases {
		t.Run("Invalid_"+input, func(t *testing.T) {
			_, exists := resolver.ResolveProvince(input)
			if exists {
				t.Errorf("Input '%s' should not resolve to any province", input)
			}
		})
	}
}

func TestProvinceResolver_CaseInsensitive(t *testing.T) {
	// Create a minimal board for testing
	board := game.NewBoard()
	board.AddProvince(&game.Province{
		Name:        "paris",
		ShortCode:   "par",
		DisplayName: "Paris",
		Type:        game.Land,
	})

	resolver := NewProvinceResolver(board)

	// Test case insensitivity
	caseCases := []struct {
		input    string
		expected string
	}{
		{"paris", "paris"},
		{"PARIS", "paris"},
		{"Paris", "paris"},
		{"pArIs", "paris"},
		{"par", "paris"},
		{"PAR", "paris"},
		{"Par", "paris"},
	}

	for _, tc := range caseCases {
		t.Run("Case_"+tc.input, func(t *testing.T) {
			canonical, exists := resolver.ResolveProvince(tc.input)
			if !exists {
				t.Errorf("Input '%s' should resolve to a province", tc.input)
				return
			}
			if canonical != tc.expected {
				t.Errorf("Input '%s' resolved to '%s', expected '%s'", tc.input, canonical, tc.expected)
			}
		})
	}
}
