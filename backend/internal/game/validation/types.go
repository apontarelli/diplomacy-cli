package validation

import (
	"diplomacy-cli/backend/internal/game"
	"strings"
)

type TokenType int

const (
	PROVINCE TokenType = iota
	UNIT_TYPE
	DASH
	SUPPORT
	CONVOY
	HOLD
	BUILD
	DISBAND
	EOF
	INVALID
)

type Token struct {
	Type     TokenType
	Value    string
	Position int
}

type SyntaxError struct {
	Position int
	Token    string
	Expected []string
	Message  string
}

type SyntaxResult struct {
	Order      *game.Order
	Errors     []SyntaxError
	Valid      bool
	Raw        string
	Normalized string
}

type OrderParser interface {
	CanParse(tokens []Token) bool
	Parse(tokens []Token, resolver *ProvinceResolver) (*game.Order, error)
	OrderType() game.OrderType
}

type ProvinceResolver struct {
	provinces map[string]*game.Province
	shortcuts map[string]string
}

func NewProvinceResolver(board *game.Board) *ProvinceResolver {
	resolver := &ProvinceResolver{
		provinces: board.Provinces,
		shortcuts: make(map[string]string),
	}

	for name, province := range board.Provinces {
		if province.ShortCode != "" {
			resolver.shortcuts[province.ShortCode] = name
		}
		resolver.shortcuts[name] = name
		if province.DisplayName != "" {
			resolver.shortcuts[strings.ToLower(province.DisplayName)] = name
		}
	}

	return resolver
}

func (pr *ProvinceResolver) ResolveProvince(identifier string) (string, bool) {
	if _, exists := pr.provinces[identifier]; exists {
		return identifier, true
	}

	withUnderscores := strings.ReplaceAll(identifier, " ", "_")
	if _, exists := pr.provinces[withUnderscores]; exists {
		return withUnderscores, true
	}

	if canonical, exists := pr.shortcuts[identifier]; exists {
		return canonical, true
	}

	normalized := strings.ToLower(identifier)
	if canonical, exists := pr.shortcuts[normalized]; exists {
		return canonical, true
	}

	return "", false
}
func (pr *ProvinceResolver) ValidateProvince(identifier string) (*game.Province, bool) {
	canonical, exists := pr.ResolveProvince(identifier)
	if !exists {
		return nil, false
	}

	province := pr.provinces[canonical]
	return province, province != nil
}

func (pr *ProvinceResolver) ParseCoast(identifier string) (string, string, bool) {
	for i, char := range identifier {
		if char == '/' {
			provincePart := identifier[:i]
			coastPart := identifier[i+1:]

			canonical, exists := pr.ResolveProvince(provincePart)
			if !exists {
				return "", "", false
			}

			province := pr.provinces[canonical]
			if province != nil {
				// Normalize coast part to lowercase for comparison
				normalizedCoast := strings.ToLower(coastPart)
				if _, hasCoast := province.CoastNeighbors[normalizedCoast]; hasCoast {
					return canonical, normalizedCoast, true
				}
			}
			return "", "", false
		}
	}

	canonical, exists := pr.ResolveProvince(identifier)
	return canonical, "", exists
}
