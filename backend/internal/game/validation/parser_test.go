package validation

import (
	"diplomacy-cli/backend/internal/game"
	"diplomacy-cli/backend/internal/game/loader"
	"fmt"
	"testing"
)

func setupTestBoard(t *testing.T) (*game.Board, *ProvinceResolver) {
	jsonLoader := loader.NewJSONLoader("../../../data/classic")
	board, err := jsonLoader.LoadBoard()
	if err != nil {
		t.Fatalf("Failed to load test board: %v", err)
	}

	resolver := NewProvinceResolver(board)
	return board, resolver
}

func TestCombineProvinceTokens(t *testing.T) {
	tests := []struct {
		name     string
		input    []Token
		expected []Token
	}{
		{
			name: "Single province token",
			input: []Token{
				{Type: PROVINCE, Value: "paris", Position: 0},
			},
			expected: []Token{
				{Type: PROVINCE, Value: "paris", Position: 0},
			},
		},
		{
			name: "Two adjacent province tokens",
			input: []Token{
				{Type: PROVINCE, Value: "north", Position: 0},
				{Type: PROVINCE, Value: "sea", Position: 6},
			},
			expected: []Token{
				{Type: PROVINCE, Value: "north sea", Position: 0},
			},
		},
		{
			name: "Two adjacent province tokens",
			input: []Token{
				{Type: PROVINCE, Value: "north", Position: 0},
				{Type: PROVINCE, Value: "sea", Position: 6},
			},
			expected: []Token{
				{Type: PROVINCE, Value: "north sea", Position: 0},
			},
		},
		{
			name: "Multi-word province with other tokens",
			input: []Token{
				{Type: UNIT_TYPE, Value: "f", Position: 0},
				{Type: PROVINCE, Value: "north", Position: 2},
				{Type: PROVINCE, Value: "sea", Position: 8},
				{Type: DASH, Value: "-", Position: 12},
				{Type: PROVINCE, Value: "picardy", Position: 14},
			},
			expected: []Token{
				{Type: UNIT_TYPE, Value: "f", Position: 0},
				{Type: PROVINCE, Value: "north sea", Position: 2},
				{Type: DASH, Value: "-", Position: 12},
				{Type: PROVINCE, Value: "picardy", Position: 14},
			},
		},
		{
			name: "Three adjacent province tokens",
			input: []Token{
				{Type: PROVINCE, Value: "st", Position: 0},
				{Type: PROVINCE, Value: "petersburg", Position: 3},
				{Type: PROVINCE, Value: "north", Position: 13},
			},
			expected: []Token{
				{Type: PROVINCE, Value: "st petersburg north", Position: 0},
			},
		},
		{
			name: "Multiple separate multi-word provinces",
			input: []Token{
				{Type: PROVINCE, Value: "north", Position: 0},
				{Type: PROVINCE, Value: "sea", Position: 6},
				{Type: DASH, Value: "-", Position: 10},
				{Type: PROVINCE, Value: "english", Position: 12},
				{Type: PROVINCE, Value: "channel", Position: 20},
			},
			expected: []Token{
				{Type: PROVINCE, Value: "north sea", Position: 0},
				{Type: DASH, Value: "-", Position: 10},
				{Type: PROVINCE, Value: "english channel", Position: 12},
			},
		},
		{
			name: "No province tokens",
			input: []Token{
				{Type: UNIT_TYPE, Value: "f", Position: 0},
				{Type: DASH, Value: "-", Position: 2},
				{Type: SUPPORT, Value: "supports", Position: 4},
			},
			expected: []Token{
				{Type: UNIT_TYPE, Value: "f", Position: 0},
				{Type: DASH, Value: "-", Position: 2},
				{Type: SUPPORT, Value: "supports", Position: 4},
			},
		},
		{
			name:     "Empty token list",
			input:    []Token{},
			expected: []Token{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := combineProvinceTokens(tt.input)

			if len(result) != len(tt.expected) {
				t.Errorf("Expected %d tokens, got %d", len(tt.expected), len(result))
				return
			}

			for i, token := range result {
				expected := tt.expected[i]
				if token.Type != expected.Type {
					t.Errorf("Token %d: expected type %v, got %v", i, expected.Type, token.Type)
				}
				if token.Value != expected.Value {
					t.Errorf("Token %d: expected value %q, got %q", i, expected.Value, token.Value)
				}
				if token.Position != expected.Position {
					t.Errorf("Token %d: expected position %d, got %d", i, expected.Position, token.Position)
				}
			}
		})
	}
}

func TestDATCFormatParsing(t *testing.T) {
	_, resolver := setupTestBoard(t)
	registry := NewOrderParserRegistry()

	tests := []struct {
		name        string
		input       string
		expectError bool
		description string
	}{
		{
			name:        "DATC standard move format",
			input:       "F North Sea - Picardy",
			expectError: true, // Should fail due to adjacency, but parsing should work
			description: "DATC format with multi-word province names",
		},
		{
			name:        "DATC format with valid move",
			input:       "F North Sea - Norway",
			expectError: false,
			description: "DATC format with valid adjacent provinces",
		},
		{
			name:        "DATC format support order",
			input:       "F North Sea Supports A Yorkshire - Liverpool",
			expectError: true, // TODO: Support orders with multi-word provinces not yet implemented
			description: "DATC format support order with multi-word provinces",
		},
		{
			name:        "DATC format with coast",
			input:       "F St. Petersburg/nc - Barents Sea",
			expectError: true, // TODO: Coast parsing in DATC format not yet implemented
			description: "DATC format with coast specification",
		},
		{
			name:        "DATC format convoy",
			input:       "F North Sea Convoys A London - Belgium",
			expectError: false,
			description: "DATC format convoy with multi-word provinces",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tokens := Tokenize(tt.input)
			order, err := registry.ParseOrder(tokens, game.SpringMovement, resolver)

			if tt.expectError && err == nil {
				t.Errorf("Expected error for %s, but parsing succeeded", tt.description)
			}
			if !tt.expectError && err != nil {
				t.Errorf("Expected success for %s, but got error: %v", tt.description, err)
			}
			if !tt.expectError && order == nil {
				t.Errorf("Expected order for %s, but got nil", tt.description)
			}
		})
	}
}
func TestParseMove(t *testing.T) {
	_, resolver := setupTestBoard(t)

	tests := []struct {
		name          string
		input         string
		expectError   bool
		expectedOrder *game.Order
	}{
		{
			name:        "Simple move",
			input:       "par - bur",
			expectError: false,
			expectedOrder: &game.Order{
				Type: game.Move,
				From: "paris",
				To:   "burgundy",
			},
		},
		{
			name:        "Move with unit type",
			input:       "a par - bur",
			expectError: false,
			expectedOrder: &game.Order{
				Type:     game.Move,
				UnitType: game.Army,
				From:     "paris",
				To:       "burgundy",
			},
		},
		{
			name:        "Fleet move with coast",
			input:       "f stp/sc - bot",
			expectError: false,
			expectedOrder: &game.Order{
				Type:      game.Move,
				UnitType:  game.Fleet,
				From:      "st_petersburg",
				FromCoast: "sc",
				To:        "gulf_of_bothnia",
			},
		},
		{
			name:        "Invalid province",
			input:       "xyz - abc",
			expectError: true,
		},
		{
			name:        "Missing dash",
			input:       "par bur",
			expectError: true,
		},
		{
			name:        "Too few tokens",
			input:       "par",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tokens := Tokenize(tt.input)
			order, err := parseMove(tokens, resolver)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if order.Type != tt.expectedOrder.Type {
				t.Errorf("Expected type %v, got %v", tt.expectedOrder.Type, order.Type)
			}
			if order.From != tt.expectedOrder.From {
				t.Errorf("Expected from %v, got %v", tt.expectedOrder.From, order.From)
			}
			if order.To != tt.expectedOrder.To {
				t.Errorf("Expected to %v, got %v", tt.expectedOrder.To, order.To)
			}
			if tt.expectedOrder.UnitType != "" && order.UnitType != tt.expectedOrder.UnitType {
				t.Errorf("Expected unit type %v, got %v", tt.expectedOrder.UnitType, order.UnitType)
			}
			if tt.expectedOrder.FromCoast != "" && order.FromCoast != tt.expectedOrder.FromCoast {
				t.Errorf("Expected from coast %v, got %v", tt.expectedOrder.FromCoast, order.FromCoast)
			}
		})
	}
}

func TestParseHold(t *testing.T) {
	_, resolver := setupTestBoard(t)

	tests := []struct {
		name          string
		input         string
		expectError   bool
		expectedOrder *game.Order
	}{
		{
			name:        "Simple hold",
			input:       "par hold",
			expectError: false,
			expectedOrder: &game.Order{
				Type: game.Hold,
				From: "paris",
			},
		},
		{
			name:        "Hold with unit type",
			input:       "a par h",
			expectError: false,
			expectedOrder: &game.Order{
				Type:     game.Hold,
				UnitType: game.Army,
				From:     "paris",
			},
		},
		{
			name:        "Fleet hold with coast",
			input:       "f stp/sc hold",
			expectError: false,
			expectedOrder: &game.Order{
				Type:      game.Hold,
				UnitType:  game.Fleet,
				From:      "st_petersburg",
				FromCoast: "sc",
			},
		},
		{
			name:        "Invalid province",
			input:       "xyz hold",
			expectError: true,
		},
		{
			name:        "Missing hold keyword",
			input:       "par",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tokens := Tokenize(tt.input)
			order, err := parseHold(tokens, resolver)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if order.Type != tt.expectedOrder.Type {
				t.Errorf("Expected type %v, got %v", tt.expectedOrder.Type, order.Type)
			}
			if order.From != tt.expectedOrder.From {
				t.Errorf("Expected from %v, got %v", tt.expectedOrder.From, order.From)
			}
			if tt.expectedOrder.UnitType != "" && order.UnitType != tt.expectedOrder.UnitType {
				t.Errorf("Expected unit type %v, got %v", tt.expectedOrder.UnitType, order.UnitType)
			}
			if tt.expectedOrder.FromCoast != "" && order.FromCoast != tt.expectedOrder.FromCoast {
				t.Errorf("Expected from coast %v, got %v", tt.expectedOrder.FromCoast, order.FromCoast)
			}
		})
	}
}

func TestParseSupportHold(t *testing.T) {
	_, resolver := setupTestBoard(t)

	tests := []struct {
		name          string
		input         string
		expectError   bool
		expectedOrder *game.Order
	}{
		{
			name:        "Simple support hold",
			input:       "bur s par",
			expectError: false,
			expectedOrder: &game.Order{
				Type:          game.Support,
				From:          "burgundy",
				SupportTarget: "paris",
			},
		},
		{
			name:        "Support hold with unit type",
			input:       "a bur s par",
			expectError: false,
			expectedOrder: &game.Order{
				Type:          game.Support,
				UnitType:      game.Army,
				From:          "burgundy",
				SupportTarget: "paris",
			},
		},
		{
			name:        "Invalid support target",
			input:       "bur s xyz",
			expectError: true,
		},
		{
			name:        "Missing support keyword",
			input:       "bur par",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tokens := Tokenize(tt.input)
			order, err := parseSupportHold(tokens, resolver)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if order.Type != tt.expectedOrder.Type {
				t.Errorf("Expected type %v, got %v", tt.expectedOrder.Type, order.Type)
			}
			if order.From != tt.expectedOrder.From {
				t.Errorf("Expected from %v, got %v", tt.expectedOrder.From, order.From)
			}
			if order.SupportTarget != tt.expectedOrder.SupportTarget {
				t.Errorf("Expected support target %v, got %v", tt.expectedOrder.SupportTarget, order.SupportTarget)
			}
		})
	}
}

func TestParseSupportMove(t *testing.T) {
	_, resolver := setupTestBoard(t)

	tests := []struct {
		name          string
		input         string
		expectError   bool
		expectedOrder *game.Order
	}{
		{
			name:        "Simple support move",
			input:       "bur s par - pic",
			expectError: false,
			expectedOrder: &game.Order{
				Type:               game.Support,
				From:               "burgundy",
				SupportTarget:      "paris",
				SupportDestination: "picardy",
			},
		},
		{
			name:        "Support move with unit type",
			input:       "a bur s par - pic",
			expectError: false,
			expectedOrder: &game.Order{
				Type:               game.Support,
				UnitType:           game.Army,
				From:               "burgundy",
				SupportTarget:      "paris",
				SupportDestination: "picardy",
			},
		},
		{
			name:        "Invalid support destination",
			input:       "bur s par - xyz",
			expectError: true,
		},
		{
			name:        "Missing dash",
			input:       "bur s par pic",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tokens := Tokenize(tt.input)
			order, err := parseSupportMove(tokens, resolver)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if order.Type != tt.expectedOrder.Type {
				t.Errorf("Expected type %v, got %v", tt.expectedOrder.Type, order.Type)
			}
			if order.From != tt.expectedOrder.From {
				t.Errorf("Expected from %v, got %v", tt.expectedOrder.From, order.From)
			}
			if order.SupportTarget != tt.expectedOrder.SupportTarget {
				t.Errorf("Expected support target %v, got %v", tt.expectedOrder.SupportTarget, order.SupportTarget)
			}
			if order.SupportDestination != tt.expectedOrder.SupportDestination {
				t.Errorf("Expected support destination %v, got %v", tt.expectedOrder.SupportDestination, order.SupportDestination)
			}
		})
	}
}

func TestParseConvoy(t *testing.T) {
	_, resolver := setupTestBoard(t)

	tests := []struct {
		name          string
		input         string
		expectError   bool
		expectedOrder *game.Order
	}{
		{
			name:        "Simple convoy",
			input:       "eng c lon - bel",
			expectError: false,
			expectedOrder: &game.Order{
				Type:         game.Convoy,
				From:         "english_channel",
				ConvoyTarget: "london",
				To:           "belgium",
			},
		},
		{
			name:        "Convoy with unit type",
			input:       "f eng c lon - bel",
			expectError: false,
			expectedOrder: &game.Order{
				Type:         game.Convoy,
				UnitType:     game.Fleet,
				From:         "english_channel",
				ConvoyTarget: "london",
				To:           "belgium",
			},
		},
		{
			name:        "Army cannot convoy",
			input:       "a eng c lon - cal",
			expectError: true,
		},
		{
			name:        "Missing dash",
			input:       "eng c lon cal",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tokens := Tokenize(tt.input)
			order, err := parseConvoy(tokens, resolver)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if order.Type != tt.expectedOrder.Type {
				t.Errorf("Expected type %v, got %v", tt.expectedOrder.Type, order.Type)
			}
			if order.From != tt.expectedOrder.From {
				t.Errorf("Expected from %v, got %v", tt.expectedOrder.From, order.From)
			}
			if order.ConvoyTarget != tt.expectedOrder.ConvoyTarget {
				t.Errorf("Expected convoy target %v, got %v", tt.expectedOrder.ConvoyTarget, order.ConvoyTarget)
			}
			if order.To != tt.expectedOrder.To {
				t.Errorf("Expected to %v, got %v", tt.expectedOrder.To, order.To)
			}
		})
	}
}

func TestOrderParserRegistry(t *testing.T) {
	_, resolver := setupTestBoard(t)
	registry := NewOrderParserRegistry()

	tests := []struct {
		name         string
		input        string
		phase        game.Phase
		expectError  bool
		expectedType game.OrderType
	}{
		{
			name:         "Move order in movement phase",
			input:        "par - bur",
			phase:        game.SpringMovement,
			expectError:  false,
			expectedType: game.Move,
		},
		{
			name:         "Hold order in movement phase",
			input:        "par hold",
			phase:        game.SpringMovement,
			expectError:  false,
			expectedType: game.Hold,
		},
		{
			name:         "Support order in movement phase",
			input:        "bur s par",
			phase:        game.SpringMovement,
			expectError:  false,
			expectedType: game.Support,
		},
		{
			name:         "Convoy order in movement phase",
			input:        "eng c lon - bel",
			phase:        game.SpringMovement,
			expectError:  false,
			expectedType: game.Convoy,
		},
		{
			name:        "Build order in wrong phase",
			input:       "build army paris",
			phase:       game.SpringMovement,
			expectError: true,
		},
		{
			name:        "Invalid order format",
			input:       "invalid order",
			phase:       game.SpringMovement,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tokens := Tokenize(tt.input)
			order, err := registry.ParseOrder(tokens, tt.phase, resolver)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if order.Type != tt.expectedType {
				t.Errorf("Expected type %v, got %v", tt.expectedType, order.Type)
			}
		})
	}
}

func TestParseOrderIntegration(t *testing.T) {
	_, resolver := setupTestBoard(t)
	registry := NewOrderParserRegistry()

	rawOrders := []string{
		"a par - bur",
		"f lon hold",
		"a ber s mun - tyr",
		"f eng c lon - bel",
	}

	for i, rawOrder := range rawOrders {
		t.Run(fmt.Sprintf("Order_%d", i), func(t *testing.T) {
			tokens := Tokenize(rawOrder)
			order, err := registry.ParseOrder(tokens, game.SpringMovement, resolver)

			if err != nil {
				t.Errorf("Failed to parse order '%s': %v", rawOrder, err)
				return
			}

			if order == nil {
				t.Errorf("Got nil order for '%s'", rawOrder)
			}
		})
	}
}
