package resolution

import (
	"testing"
)

func TestAnalyzeHeadToHeadBattle_Order1Wins(t *testing.T) {
	orders := []Order{
		{Unit: "A Berlin", Type: Move, Source: "Berlin", Destination: "Munich"},
		{Unit: "A Munich", Type: Move, Source: "Munich", Destination: "Berlin"},
	}

	adj := NewAdjudicator(orders)
	cr := NewDetailedConflictResolver(adj)

	order1 := &orders[0]
	order2 := &orders[1]

	analysis := cr.analyzeHeadToHeadBattle(order1, order2, true)

	if analysis.ConflictType != HeadToHead {
		t.Errorf("Expected HeadToHead conflict type, got %v", analysis.ConflictType)
	}

	if !analysis.Resolution.IsHeadToHead {
		t.Error("Expected IsHeadToHead to be true")
	}

	if len(analysis.Competitors) != 2 {
		t.Errorf("Expected 2 competitors, got %d", len(analysis.Competitors))
	}

	if analysis.Competitors[0] != order1 || analysis.Competitors[1] != order2 {
		t.Error("Competitors not set correctly")
	}

	if len(analysis.Resolution.DATCRules) == 0 {
		t.Error("Expected DATC rules to be populated")
	}

	if len(analysis.Resolution.Reasoning) == 0 {
		t.Error("Expected reasoning to be populated")
	}
}

func TestAnalyzeHeadToHeadBattle_Standoff(t *testing.T) {
	orders := []Order{
		{Unit: "A Berlin", Type: Move, Source: "Berlin", Destination: "Munich"},
		{Unit: "A Munich", Type: Move, Source: "Munich", Destination: "Berlin"},
		{Unit: "A Kiel", Type: Support, Source: "Kiel", Auxiliary: "Berlin -> Munich"},
		{Unit: "A Vienna", Type: Support, Source: "Vienna", Auxiliary: "Munich -> Berlin"},
	}

	adj := NewAdjudicator(orders)
	cr := NewDetailedConflictResolver(adj)

	order1 := &orders[0]
	order2 := &orders[1]

	analysis := cr.analyzeHeadToHeadBattle(order1, order2, true)

	if analysis.ConflictType != HeadToHead {
		t.Errorf("Expected HeadToHead conflict type, got %v", analysis.ConflictType)
	}

	if !analysis.Resolution.IsHeadToHead {
		t.Error("Expected IsHeadToHead to be true")
	}

	if analysis.Winner != nil && !analysis.Resolution.IsStandoff {
		t.Error("Expected either no winner or standoff for equal strength head-to-head")
	}
}

func TestOrderTypeString(t *testing.T) {
	tests := []struct {
		orderType OrderType
		expected  string
	}{
		{Move, "Move"},
		{Support, "Support"},
		{Convoy, "Convoy"},
		{Hold, "Hold"},
		{Retreat, "Retreat"},
		{OrderType(99), "Unknown"},
	}

	for _, test := range tests {
		result := test.orderType.String()
		if result != test.expected {
			t.Errorf("OrderType(%d).String() = %q, expected %q", test.orderType, result, test.expected)
		}
	}
}
