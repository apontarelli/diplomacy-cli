package resolution

import (
	"diplomacy-cli/backend/internal/game"
	"slices"
)

type OrderResult string

const (
	MoveSuccess     OrderResult = "move_success"
	MoveBounced     OrderResult = "move_bounced"
	MoveNoConvoy    OrderResult = "move_no_convoy"
	MoveInvalid     OrderResult = "move_invalid"
	SupportSuccess  OrderResult = "support_success"
	SupportCut      OrderResult = "support_cut"
	SupportInvalid  OrderResult = "support_invalid"
	ConvoySuccess   OrderResult = "convoy_success"
	ConvoyDisrupted OrderResult = "convoy_disrupted"
	ConvoyInvalid   OrderResult = "convoy_invalid"
	HoldSuccess     OrderResult = "hold_success"
	Dislodged       OrderResult = "dislodged"
)

type ResolvedOrder struct {
	*game.Order
	Index         int
	OrigTerritory string
	NewTerritory  string
	UnitKey       string
}

type OrderOutcome struct {
	Result        OrderResult
	Dislodged     bool
	SupportCut    bool
	ConvoyPath    []string
	Strength      int
	FailureReason string
}

type ProvinceCoast struct {
	Province string
	Coast    string
}

func NewProvinceCoast(province, coast string) ProvinceCoast {
	return ProvinceCoast{
		Province: province,
		Coast:    coast,
	}
}

func (pc ProvinceCoast) String() string {
	if pc.Coast == "" {
		return pc.Province
	}
	return pc.Province + "/" + pc.Coast
}

func (pc ProvinceCoast) IsEmpty() bool {
	return pc.Province == ""
}

func (pc ProvinceCoast) HasCoast() bool {
	return pc.Coast != ""
}

type ConvoyKey struct {
	Origin      ProvinceCoast
	Destination ProvinceCoast
}

type ResolutionEngine struct {
	orders            []*ResolvedOrder
	board             *game.Board
	outcomes          []OrderOutcome
	strength          []int
	conflicts         map[string][]int
	supports          map[string][]int
	convoys           map[ConvoyKey][]ProvinceCoast
	ordersByUnit      map[string]int
	ordersByTerritory map[string]int
}

func NewResolutionEngine(orders []*game.Order, board *game.Board) *ResolutionEngine {
	resolvedOrders := make([]*ResolvedOrder, len(orders))
	outcomes := make([]OrderOutcome, len(orders))
	strength := make([]int, len(orders))

	for i, order := range orders {
		unitKey := makeUnitKey(order.From, order.FromCoast)
		resolvedOrders[i] = &ResolvedOrder{
			Order:         order,
			Index:         i,
			OrigTerritory: order.From,
			NewTerritory:  order.From,
			UnitKey:       unitKey,
		}

		outcomes[i] = OrderOutcome{
			Result:   "",
			Strength: 1,
		}
		strength[i] = 1
	}

	engine := &ResolutionEngine{
		orders:            resolvedOrders,
		board:             board,
		outcomes:          outcomes,
		strength:          strength,
		conflicts:         make(map[string][]int),
		supports:          make(map[string][]int),
		convoys:           make(map[ConvoyKey][]ProvinceCoast),
		ordersByUnit:      make(map[string]int),
		ordersByTerritory: make(map[string]int),
	}

	engine.buildLookupMaps()
	return engine
}

func (re *ResolutionEngine) buildLookupMaps() {
	re.conflicts = make(map[string][]int)
	re.supports = make(map[string][]int)
	re.ordersByUnit = make(map[string]int)
	re.ordersByTerritory = make(map[string]int)

	for i, order := range re.orders {
		re.ordersByUnit[order.UnitKey] = i
		re.ordersByTerritory[order.OrigTerritory] = i

		if order.Type == game.Support {
			supportedUnitKey := makeUnitKey(order.SupportTarget, "")
			re.supports[supportedUnitKey] = append(re.supports[supportedUnitKey], i)
		}
	}
}

func (re *ResolutionEngine) rebuildConflictMap() {
	re.conflicts = make(map[string][]int)

	for i, order := range re.orders {
		territory := order.NewTerritory
		re.conflicts[territory] = append(re.conflicts[territory], i)
	}
}

func (re *ResolutionEngine) GetResults() []OrderOutcome {
	return re.outcomes
}

func (re *ResolutionEngine) GetOrder(index int) *ResolvedOrder {
	if index < 0 || index >= len(re.orders) {
		return nil
	}
	return re.orders[index]
}

func (re *ResolutionEngine) FindOrderByUnit(territory, coast string) int {
	unitKey := makeUnitKey(territory, coast)
	if index, exists := re.ordersByUnit[unitKey]; exists {
		return index
	}
	if coast != "" {
		if index, exists := re.ordersByUnit[territory]; exists {
			return index
		}
	}
	return -1
}

func (re *ResolutionEngine) FindOrderByTerritory(territory string) int {
	if index, exists := re.ordersByTerritory[territory]; exists {
		return index
	}
	return -1
}

func makeUnitKey(territory, coast string) string {
	if coast == "" {
		return territory
	}
	return territory + "/" + coast
}

func (re *ResolutionEngine) IsAdjacent(from, to string, unitType game.UnitType) bool {
	province := re.board.GetProvince(from)
	if province == nil {
		return false
	}

	switch unitType {
	case game.Army:
		return slices.Contains(province.ArmyNeighbors, to)
	case game.Fleet:
		return slices.Contains(province.FleetNeighbors, to)
	}

	return false
}
