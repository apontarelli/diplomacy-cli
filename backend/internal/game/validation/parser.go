package validation

import (
	"diplomacy-cli/backend/internal/game"
	"fmt"
	"strings"
)

type OrderParserFunc func(tokens []Token, resolver *ProvinceResolver) (*game.Order, error)

type OrderParserRegistry struct {
	MovementParsers   []OrderParserFunc
	RetreatParsers    []OrderParserFunc
	AdjustmentParsers []OrderParserFunc
}

func NewOrderParserRegistry() *OrderParserRegistry {
	return &OrderParserRegistry{
		MovementParsers: []OrderParserFunc{
			parseConvoy,
			parseSupportMove,
			parseSupportHold,
			parseMove,
			parseHold,
		},
		RetreatParsers: []OrderParserFunc{
			parseRetreat,
		},
		AdjustmentParsers: []OrderParserFunc{
			parseBuild,
			parseDisband,
		},
	}
}

func (opr *OrderParserRegistry) ParseOrder(tokens []Token, phase game.Phase, resolver *ProvinceResolver) (*game.Order, error) {
	// Combine adjacent province tokens to handle multi-word province names
	tokens = combineProvinceTokens(tokens)

	var parsers []OrderParserFunc

	switch phase {
	case game.SpringMovement, game.FallMovement:
		parsers = opr.MovementParsers
	case game.SpringRetreat, game.FallRetreat:
		parsers = opr.RetreatParsers
	case game.WinterBuild:
		parsers = opr.AdjustmentParsers
	default:
		return nil, fmt.Errorf("unknown phase: %s", phase)
	}

	var errors []error
	for _, parser := range parsers {
		order, err := parser(tokens, resolver)
		if err == nil {
			return order, nil
		}
		errors = append(errors, err)
	}

	// Return the most informative error
	bestError := selectBestError(errors)
	return nil, fmt.Errorf("failed to parse order: %v", bestError)
}

// combineProvinceTokens combines adjacent PROVINCE tokens to handle multi-word province names
func combineProvinceTokens(tokens []Token) []Token {
	if len(tokens) <= 1 {
		return tokens
	}

	var result []Token
	i := 0

	for i < len(tokens) {
		if tokens[i].Type == PROVINCE {
			// Start combining province tokens
			combined := tokens[i].Value
			j := i + 1

			// Combine consecutive PROVINCE tokens
			for j < len(tokens) && tokens[j].Type == PROVINCE {
				combined += " " + tokens[j].Value
				j++
			}

			// Create a new combined token
			result = append(result, Token{
				Type:     PROVINCE,
				Value:    combined,
				Position: tokens[i].Position,
			})
			i = j
		} else {
			// Non-province token, add as-is
			result = append(result, tokens[i])
			i++
		}
	}

	return result
}

func parseMove(tokens []Token, resolver *ProvinceResolver) (*game.Order, error) {
	if len(tokens) > 0 && tokens[len(tokens)-1].Type == EOF {
		tokens = tokens[:len(tokens)-1]
	}

	var unitType game.UnitType
	var fromToken, toToken Token

	if len(tokens) == 4 && tokens[0].Type == UNIT_TYPE && tokens[1].Type == PROVINCE &&
		tokens[2].Type == DASH && tokens[3].Type == PROVINCE {

		switch tokens[0].Value {
		case "army", "a":
			unitType = game.Army
		case "fleet", "f":
			unitType = game.Fleet
		default:
			return nil, fmt.Errorf("invalid unit type: %s", tokens[0].Value)
		}
		fromToken = tokens[1]
		toToken = tokens[3]
	} else if len(tokens) == 3 && tokens[0].Type == PROVINCE &&
		tokens[1].Type == DASH && tokens[2].Type == PROVINCE {
		fromToken = tokens[0]
		toToken = tokens[2]
	} else {
		return nil, fmt.Errorf("invalid move order format")
	}

	fromProvince, fromCoast, fromValid := resolver.ParseCoast(fromToken.Value)
	if !fromValid {
		return nil, fmt.Errorf("invalid from province: %s", fromToken.Value)
	}

	toProvince, toCoast, toValid := resolver.ParseCoast(toToken.Value)
	if !toValid {
		return nil, fmt.Errorf("invalid to province: %s", toToken.Value)
	}

	if err := validateAdjacency(fromProvince, toProvince, fromCoast, toCoast, resolver); err != nil {
		return nil, err
	}

	return &game.Order{
		Type:      game.Move,
		UnitType:  unitType,
		From:      fromProvince,
		FromCoast: fromCoast,
		To:        toProvince,
		ToCoast:   toCoast,
	}, nil
}

func parseHold(tokens []Token, resolver *ProvinceResolver) (*game.Order, error) {
	if len(tokens) > 0 && tokens[len(tokens)-1].Type == EOF {
		tokens = tokens[:len(tokens)-1]
	}

	var unitType game.UnitType
	var provinceToken Token

	if len(tokens) == 3 && tokens[0].Type == UNIT_TYPE && tokens[1].Type == PROVINCE && tokens[2].Type == HOLD {
		switch tokens[0].Value {
		case "army", "a":
			unitType = game.Army
		case "fleet", "f":
			unitType = game.Fleet
		default:
			return nil, fmt.Errorf("invalid unit type: %s", tokens[0].Value)
		}
		provinceToken = tokens[1]
	} else if len(tokens) == 2 && tokens[0].Type == PROVINCE && tokens[1].Type == HOLD {
		provinceToken = tokens[0]
	} else {
		return nil, fmt.Errorf("invalid hold order format")
	}

	province, coast, valid := resolver.ParseCoast(provinceToken.Value)
	if !valid {
		return nil, fmt.Errorf("invalid province: %s", provinceToken.Value)
	}

	return &game.Order{
		Type:      game.Hold,
		UnitType:  unitType,
		From:      province,
		FromCoast: coast,
	}, nil
}

func parseSupportHold(tokens []Token, resolver *ProvinceResolver) (*game.Order, error) {
	if len(tokens) > 0 && tokens[len(tokens)-1].Type == EOF {
		tokens = tokens[:len(tokens)-1]
	}

	var unitType game.UnitType
	var fromToken, supportToken Token

	if len(tokens) == 4 && tokens[0].Type == UNIT_TYPE && tokens[1].Type == PROVINCE &&
		tokens[2].Type == SUPPORT && tokens[3].Type == PROVINCE {

		switch tokens[0].Value {
		case "army", "a":
			unitType = game.Army
		case "fleet", "f":
			unitType = game.Fleet
		default:
			return nil, fmt.Errorf("invalid unit type: %s", tokens[0].Value)
		}
		fromToken = tokens[1]
		supportToken = tokens[3]
	} else if len(tokens) == 3 && tokens[0].Type == PROVINCE &&
		tokens[1].Type == SUPPORT && tokens[2].Type == PROVINCE {
		fromToken = tokens[0]
		supportToken = tokens[2]
	} else {
		return nil, fmt.Errorf("invalid support hold order format")
	}

	fromProvince, fromCoast, fromValid := resolver.ParseCoast(fromToken.Value)
	if !fromValid {
		return nil, fmt.Errorf("invalid from province: %s", fromToken.Value)
	}

	supportProvince, _, supportValid := resolver.ParseCoast(supportToken.Value)
	if !supportValid {
		return nil, fmt.Errorf("invalid support target province: %s", supportToken.Value)
	}

	return &game.Order{
		Type:          game.Support,
		UnitType:      unitType,
		From:          fromProvince,
		FromCoast:     fromCoast,
		SupportTarget: supportProvince,
	}, nil
}

func parseSupportMove(tokens []Token, resolver *ProvinceResolver) (*game.Order, error) {
	if len(tokens) > 0 && tokens[len(tokens)-1].Type == EOF {
		tokens = tokens[:len(tokens)-1]
	}

	var unitType game.UnitType
	var fromToken, supportFromToken, supportToToken Token

	if len(tokens) == 6 && tokens[0].Type == UNIT_TYPE && tokens[1].Type == PROVINCE &&
		tokens[2].Type == SUPPORT && tokens[3].Type == PROVINCE &&
		tokens[4].Type == DASH && tokens[5].Type == PROVINCE {

		switch tokens[0].Value {
		case "army", "a":
			unitType = game.Army
		case "fleet", "f":
			unitType = game.Fleet
		default:
			return nil, fmt.Errorf("invalid unit type: %s", tokens[0].Value)
		}
		fromToken = tokens[1]
		supportFromToken = tokens[3]
		supportToToken = tokens[5]
	} else if len(tokens) == 5 && tokens[0].Type == PROVINCE &&
		tokens[1].Type == SUPPORT && tokens[2].Type == PROVINCE &&
		tokens[3].Type == DASH && tokens[4].Type == PROVINCE {
		fromToken = tokens[0]
		supportFromToken = tokens[2]
		supportToToken = tokens[4]
	} else {
		return nil, fmt.Errorf("invalid support move order format")
	}

	fromProvince, fromCoast, fromValid := resolver.ParseCoast(fromToken.Value)
	if !fromValid {
		return nil, fmt.Errorf("invalid from province: %s", fromToken.Value)
	}

	supportFromProvince, _, supportFromValid := resolver.ParseCoast(supportFromToken.Value)
	if !supportFromValid {
		return nil, fmt.Errorf("invalid support origin province: %s", supportFromToken.Value)
	}

	supportToProvince, _, supportToValid := resolver.ParseCoast(supportToToken.Value)
	if !supportToValid {
		return nil, fmt.Errorf("invalid support destination province: %s", supportToToken.Value)
	}

	return &game.Order{
		Type:               game.Support,
		UnitType:           unitType,
		From:               fromProvince,
		FromCoast:          fromCoast,
		SupportTarget:      supportFromProvince,
		SupportDestination: supportToProvince,
	}, nil
}

func parseConvoy(tokens []Token, resolver *ProvinceResolver) (*game.Order, error) {
	if len(tokens) > 0 && tokens[len(tokens)-1].Type == EOF {
		tokens = tokens[:len(tokens)-1]
	}

	var unitType game.UnitType
	var fromToken, convoyFromToken, convoyToToken Token

	if len(tokens) == 7 && tokens[0].Type == UNIT_TYPE && tokens[1].Type == PROVINCE &&
		tokens[2].Type == CONVOY && tokens[3].Type == UNIT_TYPE && tokens[4].Type == PROVINCE &&
		tokens[5].Type == DASH && tokens[6].Type == PROVINCE {
		// DATC format: "F North Sea Convoys A Yorkshire - Yorkshire"
		switch tokens[0].Value {
		case "fleet", "f":
			unitType = game.Fleet
		default:
			return nil, fmt.Errorf("only fleets can convoy, got: %s", tokens[0].Value)
		}
		fromToken = tokens[1]
		convoyFromToken = tokens[4]
		convoyToToken = tokens[6]
	} else if len(tokens) == 6 && tokens[0].Type == UNIT_TYPE && tokens[1].Type == PROVINCE &&
		tokens[2].Type == CONVOY && tokens[3].Type == PROVINCE &&
		tokens[4].Type == DASH && tokens[5].Type == PROVINCE {
		// Standard format: "F North Sea C Yorkshire - Yorkshire"
		switch tokens[0].Value {
		case "fleet", "f":
			unitType = game.Fleet
		default:
			return nil, fmt.Errorf("only fleets can convoy, got: %s", tokens[0].Value)
		}
		fromToken = tokens[1]
		convoyFromToken = tokens[3]
		convoyToToken = tokens[5]
	} else if len(tokens) == 5 && tokens[0].Type == PROVINCE &&
		tokens[1].Type == CONVOY && tokens[2].Type == PROVINCE &&
		tokens[3].Type == DASH && tokens[4].Type == PROVINCE {
		// No unit type format: "North Sea C Yorkshire - Yorkshire"
		fromToken = tokens[0]
		convoyFromToken = tokens[2]
		convoyToToken = tokens[4]
	} else {
		return nil, fmt.Errorf("invalid convoy order format")
	}

	fromProvince, fromCoast, fromValid := resolver.ParseCoast(fromToken.Value)
	if !fromValid {
		return nil, fmt.Errorf("invalid from province: %s", fromToken.Value)
	}

	convoyFromProvince, _, convoyFromValid := resolver.ParseCoast(convoyFromToken.Value)
	if !convoyFromValid {
		return nil, fmt.Errorf("invalid convoy origin province: %s", convoyFromToken.Value)
	}

	convoyToProvince, _, convoyToValid := resolver.ParseCoast(convoyToToken.Value)
	if !convoyToValid {
		return nil, fmt.Errorf("invalid convoy destination province: %s", convoyToToken.Value)
	}

	return &game.Order{
		Type:         game.Convoy,
		UnitType:     unitType,
		From:         fromProvince,
		FromCoast:    fromCoast,
		ConvoyTarget: convoyFromProvince,
		To:           convoyToProvince,
	}, nil
}

func parseBuild(tokens []Token, resolver *ProvinceResolver) (*game.Order, error) {
	if len(tokens) > 0 && tokens[len(tokens)-1].Type == EOF {
		tokens = tokens[:len(tokens)-1]
	}

	if len(tokens) != 3 || tokens[0].Type != BUILD || tokens[1].Type != UNIT_TYPE || tokens[2].Type != PROVINCE {
		return nil, fmt.Errorf("invalid build order format")
	}

	var unitType game.UnitType
	switch tokens[1].Value {
	case "army", "a":
		unitType = game.Army
	case "fleet", "f":
		unitType = game.Fleet
	default:
		return nil, fmt.Errorf("invalid unit type: %s", tokens[1].Value)
	}

	province, coast, valid := resolver.ParseCoast(tokens[2].Value)
	if !valid {
		return nil, fmt.Errorf("invalid province: %s", tokens[2].Value)
	}

	return &game.Order{
		Type:     game.Move,
		UnitType: unitType,
		To:       province,
		ToCoast:  coast,
	}, nil
}

func parseDisband(tokens []Token, resolver *ProvinceResolver) (*game.Order, error) {
	if len(tokens) > 0 && tokens[len(tokens)-1].Type == EOF {
		tokens = tokens[:len(tokens)-1]
	}

	if len(tokens) != 3 || tokens[0].Type != DISBAND || tokens[1].Type != UNIT_TYPE || tokens[2].Type != PROVINCE {
		return nil, fmt.Errorf("invalid disband order format")
	}

	var unitType game.UnitType
	switch tokens[1].Value {
	case "army", "a":
		unitType = game.Army
	case "fleet", "f":
		unitType = game.Fleet
	default:
		return nil, fmt.Errorf("invalid unit type: %s", tokens[1].Value)
	}

	province, coast, valid := resolver.ParseCoast(tokens[2].Value)
	if !valid {
		return nil, fmt.Errorf("invalid province: %s", tokens[2].Value)
	}

	return &game.Order{
		Type:      game.Hold,
		UnitType:  unitType,
		From:      province,
		FromCoast: coast,
	}, nil
}

func parseRetreat(tokens []Token, resolver *ProvinceResolver) (*game.Order, error) {
	if len(tokens) > 0 && tokens[len(tokens)-1].Type == EOF {
		tokens = tokens[:len(tokens)-1]
	}

	if len(tokens) != 3 || tokens[0].Type != PROVINCE || tokens[1].Type != DASH || tokens[2].Type != PROVINCE {
		return nil, fmt.Errorf("invalid retreat order format")
	}

	fromProvince, fromCoast, fromValid := resolver.ParseCoast(tokens[0].Value)
	if !fromValid {
		return nil, fmt.Errorf("invalid from province: %s", tokens[0].Value)
	}

	toProvince, toCoast, toValid := resolver.ParseCoast(tokens[2].Value)
	if !toValid {
		return nil, fmt.Errorf("invalid to province: %s", tokens[2].Value)
	}

	return &game.Order{
		Type:      game.Move,
		From:      fromProvince,
		FromCoast: fromCoast,
		To:        toProvince,
		ToCoast:   toCoast,
	}, nil
}

func validateAdjacency(fromProvince, toProvince, fromCoast, toCoast string, resolver *ProvinceResolver) error {
	fromProv, fromExists := resolver.ValidateProvince(fromProvince)
	if !fromExists {
		return fmt.Errorf("from province %s does not exist", fromProvince)
	}

	toProv, toExists := resolver.ValidateProvince(toProvince)
	if !toExists {
		return fmt.Errorf("to province %s does not exist", toProvince)
	}

	if !areAdjacent(fromProv, toProvince, fromCoast, toCoast) {
		return fmt.Errorf("provinces %s and %s are not adjacent", fromProvince, toProvince)
	}

	if fromCoast != "" {
		if _, hasCoast := fromProv.CoastNeighbors[fromCoast]; !hasCoast {
			return fmt.Errorf("province %s does not have coast %s", fromProvince, fromCoast)
		}
	}

	if toCoast != "" {
		if _, hasCoast := toProv.CoastNeighbors[toCoast]; !hasCoast {
			return fmt.Errorf("province %s does not have coast %s", toProvince, toCoast)
		}
	}

	return nil
}

func areAdjacent(fromProv *game.Province, toProvince, fromCoast, toCoast string) bool {
	if fromCoast != "" {
		if coastNeighbors, exists := fromProv.CoastNeighbors[fromCoast]; exists {
			for _, neighbor := range coastNeighbors {
				if toCoast != "" {
					targetWithCoast := toProvince + "/" + toCoast
					if neighbor == targetWithCoast {
						return true
					}
				}
				if neighbor == toProvince {
					return true
				}
			}
		}
		return false
	}

	for _, neighbor := range fromProv.ArmyNeighbors {
		if neighbor == toProvince {
			return true
		}
	}

	for _, neighbor := range fromProv.FleetNeighbors {
		if neighbor == toProvince {
			return true
		}
	}

	return false
}

// selectBestError returns the most informative error from a list of parser errors
func selectBestError(errors []error) error {
	if len(errors) == 0 {
		return fmt.Errorf("no errors to select from")
	}

	// Priority order for error types (most informative first)
	priorities := []string{
		"not adjacent",      // Adjacency errors are most informative
		"does not exist",    // Province existence errors
		"invalid unit type", // Unit type errors
		"invalid province",  // Province validation errors
		"invalid.*format",   // Format errors are least informative
	}

	// Find the highest priority error
	for _, priority := range priorities {
		for _, err := range errors {
			if strings.Contains(err.Error(), priority) {
				return err
			}
		}
	}

	// If no priority match, return the first error
	return errors[0]
}
