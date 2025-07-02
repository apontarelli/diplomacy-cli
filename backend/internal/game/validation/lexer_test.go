package validation

import (
	"testing"
)

func TestNormalizeInput(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"A par - bur", "a par - bur"},
		{"  A  PAR  —  BUR  ", "a par - bur"},
		{"F stp/sc - bot", "f stp/sc - bot"},
		{"army paris hold", "army paris hold"},
		{"Fleet London supports Army Wales - Belgium", "fleet london supports army wales - belgium"},
		{"A par–bur", "a par - bur"},
		{"F eng / nth", "f eng/nth"},
		{"build army paris!!!", "build army paris!!!"},
		{"", ""},
	}

	for _, test := range tests {
		result := normalizeInput(test.input)
		if result != test.expected {
			t.Errorf("normalizeInput(%q) = %q, expected %q", test.input, result, test.expected)
		}
	}
}

func TestTokenize(t *testing.T) {
	tests := []struct {
		input    string
		expected []TokenType
	}{
		{
			"a par - bur",
			[]TokenType{UNIT_TYPE, PROVINCE, DASH, PROVINCE, EOF},
		},
		{
			"par hold",
			[]TokenType{PROVINCE, HOLD, EOF},
		},
		{
			"bur s par - pic",
			[]TokenType{PROVINCE, SUPPORT, PROVINCE, DASH, PROVINCE, EOF},
		},
		{
			"f stp/sc - bot",
			[]TokenType{UNIT_TYPE, PROVINCE, DASH, PROVINCE, EOF},
		},
		{
			"build fleet london",
			[]TokenType{BUILD, UNIT_TYPE, PROVINCE, EOF},
		},
		{
			"disband army paris",
			[]TokenType{DISBAND, UNIT_TYPE, PROVINCE, EOF},
		},
		{
			"eng c lon - cal",
			[]TokenType{PROVINCE, CONVOY, PROVINCE, DASH, PROVINCE, EOF},
		},
	}

	for _, test := range tests {
		tokens := Tokenize(test.input)
		if len(tokens) != len(test.expected) {
			t.Errorf("Tokenize(%q) returned %d tokens, expected %d", test.input, len(tokens), len(test.expected))
			continue
		}

		for i, token := range tokens {
			if token.Type != test.expected[i] {
				t.Errorf("Tokenize(%q) token %d: got %v, expected %v", test.input, i, token.Type, test.expected[i])
			}
		}
	}
}

func TestTokenizeValues(t *testing.T) {
	input := "a par - bur"
	tokens := Tokenize(input)

	expectedValues := []string{"a", "par", "-", "bur", ""}
	for i, token := range tokens {
		if token.Value != expectedValues[i] {
			t.Errorf("Token %d: got value %q, expected %q", i, token.Value, expectedValues[i])
		}
	}
}

func TestTokenizePositions(t *testing.T) {
	input := "a par - bur"
	tokens := Tokenize(input)

	expectedPositions := []int{0, 2, 6, 8, 11}
	for i, token := range tokens {
		if token.Position != expectedPositions[i] {
			t.Errorf("Token %d: got position %d, expected %d", i, token.Position, expectedPositions[i])
		}
	}
}

func TestTokenizeWithCoasts(t *testing.T) {
	tests := []struct {
		input         string
		expectedValue string
		expectedType  TokenType
	}{
		{"stp/sc", "stp/sc", PROVINCE},
		{"spa/nc", "spa/nc", PROVINCE},
		{"bul/ec", "bul/ec", PROVINCE},
		{"london", "london", PROVINCE},
	}

	for _, test := range tests {
		tokens := Tokenize(test.input)
		if len(tokens) < 1 {
			t.Errorf("Tokenize(%q) returned no tokens", test.input)
			continue
		}

		token := tokens[0]
		if token.Type != test.expectedType {
			t.Errorf("Tokenize(%q): got type %v, expected %v", test.input, token.Type, test.expectedType)
		}
		if token.Value != test.expectedValue {
			t.Errorf("Tokenize(%q): got value %q, expected %q", test.input, token.Value, test.expectedValue)
		}
	}
}

func TestTokenizeWithErrors(t *testing.T) {
	tests := []struct {
		input        string
		expectErrors bool
	}{
		{"a par - bur", false},
		{"a par @ bur", true},
		{"a par # bur", true},
		{"normal order", false},
		{"order with ! symbols", true},
	}

	for _, test := range tests {
		tokens, errors := TokenizeWithErrors(test.input)
		hasErrors := len(errors) > 0

		if hasErrors != test.expectErrors {
			t.Errorf("TokenizeWithErrors(%q): got errors=%v, expected errors=%v", test.input, hasErrors, test.expectErrors)
		}

		if test.expectErrors {
			for _, token := range tokens {
				if token.Type == INVALID {
					found := false
					for _, err := range errors {
						if err.Position == token.Position {
							found = true
							break
						}
					}
					if !found {
						t.Errorf("INVALID token at position %d has no corresponding error", token.Position)
					}
				}
			}
		}
	}
}

func TestClassifyWord(t *testing.T) {
	tests := []struct {
		word     string
		expected TokenType
	}{
		{"hold", HOLD},
		{"h", HOLD},
		{"support", SUPPORT},
		{"supports", SUPPORT},
		{"s", SUPPORT},
		{"convoy", CONVOY},
		{"convoys", CONVOY},
		{"c", CONVOY},
		{"build", BUILD},
		{"b", BUILD},
		{"disband", DISBAND},
		{"remove", DISBAND},
		{"d", DISBAND},
		{"r", DISBAND},
		{"army", UNIT_TYPE},
		{"a", UNIT_TYPE},
		{"fleet", UNIT_TYPE},
		{"f", UNIT_TYPE},
		{"paris", PROVINCE},
		{"london", PROVINCE},
		{"stp", PROVINCE},
		{"unknown", PROVINCE},
	}

	for _, test := range tests {
		result := classifyWord(test.word)
		if result != test.expected {
			t.Errorf("classifyWord(%q) = %v, expected %v", test.word, result, test.expected)
		}
	}
}
