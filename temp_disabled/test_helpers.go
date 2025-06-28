package api_test

import (
	"testing"

	"diplomacy-backend/internal/api"
	"diplomacy-backend/internal/game"
	"diplomacy-backend/internal/game/rules"
)

// TestServer creates a new server instance for testing
func TestServer() *api.Server {
	gameRules := rules.MustLoadRules("classic")
	return api.NewServer(gameRules)
}

// CreateMoveOrder creates a move order with the given parameters
func CreateMoveOrder(id, unitID, from, to string) game.Order {
	return game.Order{
		ID:            id,
		UnitID:        unitID,
		OrderType:     game.OrderTypeMove,
		FromTerritory: from,
		ToTerritory:   to,
		Status:        game.OrderStatusSubmitted,
	}
}

// CreateSupportOrder creates a support order with the given parameters
func CreateSupportOrder(id, unitID, supportUnit, toTerritory string) game.Order {
	return game.Order{
		ID:          id,
		UnitID:      unitID,
		OrderType:   game.OrderTypeSupport,
		SupportUnit: supportUnit,
		ToTerritory: toTerritory,
		Status:      game.OrderStatusSubmitted,
	}
}

// CreateHoldOrder creates a hold order with the given parameters
func CreateHoldOrder(id, unitID, territory string) game.Order {
	return game.Order{
		ID:            id,
		UnitID:        unitID,
		OrderType:     game.OrderTypeHold,
		FromTerritory: territory,
		Status:        game.OrderStatusSubmitted,
	}
}

// CreateConvoyOrder creates a convoy order with the given parameters
func CreateConvoyOrder(id, unitID, supportUnit, toTerritory string) game.Order {
	return game.Order{
		ID:          id,
		UnitID:      unitID,
		OrderType:   game.OrderTypeConvoy,
		SupportUnit: supportUnit,
		ToTerritory: toTerritory,
		Status:      game.OrderStatusSubmitted,
	}
}

// AssertOrderResult checks that an order result matches expected values
func AssertOrderResult(t *testing.T, result game.OrderResult, expectedResult game.ResolutionResult, expectedPosition string) {
	t.Helper()
	
	if result.Result != expectedResult {
		t.Errorf("Expected result %v, got %v for order %s", expectedResult, result.Result, result.OrderID)
	}
	
	if expectedPosition != "" && result.NewPosition != expectedPosition {
		t.Errorf("Expected new position '%s', got '%s' for order %s", expectedPosition, result.NewPosition, result.OrderID)
	}
}

// AssertAllOrdersResult checks that all orders have the same expected result
func AssertAllOrdersResult(t *testing.T, results []game.OrderResult, expectedResult game.ResolutionResult) {
	t.Helper()
	
	for _, result := range results {
		if result.Result != expectedResult {
			t.Errorf("Expected %v for all orders, got %v for order %s", expectedResult, result.Result, result.OrderID)
		}
	}
}

// AssertResultCount checks that the number of results matches expected
func AssertResultCount(t *testing.T, results []game.OrderResult, expected int) {
	t.Helper()
	
	if len(results) != expected {
		t.Fatalf("Expected %d results, got %d", expected, len(results))
	}
}

// CreateUnitPositions creates a unit position map from pairs of unitID, territory
func CreateUnitPositions(pairs ...string) map[string]string {
	if len(pairs)%2 != 0 {
		panic("CreateUnitPositions requires an even number of arguments (unitID, territory pairs)")
	}
	
	positions := make(map[string]string)
	for i := 0; i < len(pairs); i += 2 {
		positions[pairs[i]] = pairs[i+1]
	}
	return positions
}

// TestScenario represents a complete test scenario with setup and expectations
type TestScenario struct {
	Name          string
	MoveOrders    []game.Order
	SupportOrders []game.Order
	HoldOrders    []game.Order
	ConvoyOrders  []game.Order
	UnitPositions map[string]string
	Expectations  []OrderExpectation
}

// OrderExpectation represents expected result for a specific order
type OrderExpectation struct {
	OrderID         string
	ExpectedResult  game.ResolutionResult
	ExpectedPosition string // Optional, only for successful moves
}

// RunTestScenario executes a complete test scenario
func RunTestScenario(t *testing.T, scenario TestScenario) {
	t.Helper()
	
	server := TestServer()
	
	// Resolve movement orders
	moveResults := server.resolveMovementOrders(scenario.MoveOrders, scenario.SupportOrders, scenario.UnitPositions)
	
	// Resolve other order types
	supportResults := server.resolveSupportOrders(scenario.SupportOrders, moveResults)
	holdResults := server.resolveHoldOrders(scenario.HoldOrders, moveResults)
	convoyResults := server.resolveConvoyOrders(scenario.ConvoyOrders)
	
	// Combine all results
	allResults := append(moveResults, supportResults...)
	allResults = append(allResults, holdResults...)
	allResults = append(allResults, convoyResults...)
	
	// Check expectations
	for _, expectation := range scenario.Expectations {
		found := false
		for _, result := range allResults {
			if result.OrderID == expectation.OrderID {
				AssertOrderResult(t, result, expectation.ExpectedResult, expectation.ExpectedPosition)
				found = true
				break
			}
		}
		if !found {
			t.Errorf("No result found for order %s", expectation.OrderID)
		}
	}
}

// Common unit positions for classic Diplomacy starting positions
var ClassicStartingPositions = map[string]string{
	// Austria
	"austria_army_vie": "vie",
	"austria_army_bud": "bud",
	"austria_fleet_tri": "tri",
	
	// England  
	"england_army_lvp": "lvp",
	"england_fleet_lon": "lon",
	"england_fleet_edi": "edi",
	
	// France
	"france_army_par": "par",
	"france_army_mar": "mar", 
	"france_fleet_bre": "bre",
	
	// Germany
	"germany_army_ber": "ber",
	"germany_army_mun": "mun",
	"germany_fleet_kie": "kie",
	
	// Italy
	"italy_army_rom": "rom",
	"italy_army_ven": "ven",
	"italy_fleet_nap": "nap",
	
	// Russia
	"russia_army_war": "war",
	"russia_army_mos": "mos",
	"russia_fleet_stp": "stp",
	"russia_fleet_sev": "sev",
	
	// Turkey
	"turkey_army_con": "con",
	"turkey_army_smy": "smy",
	"turkey_fleet_ank": "ank",
}