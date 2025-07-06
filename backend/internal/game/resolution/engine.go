package resolution

import (
	"diplomacy-cli/backend/internal/game"
)

// engine.go provides compatibility with the existing pipeline interface
// while using the new recursive adjudication system internally.

// OrderResult represents the type of resolution result
type OrderResult string

const (
	MoveSuccess     OrderResult = "move_success"
	MoveBounced     OrderResult = "move_bounced"
	MoveNoConvoy    OrderResult = "move_no_convoy"
	SupportSuccess  OrderResult = "support_success"
	SupportCut      OrderResult = "support_cut"
	ConvoySuccess   OrderResult = "convoy_success"
	ConvoyDisrupted OrderResult = "convoy_disrupted"
	HoldSuccess     OrderResult = "hold_success"
	Dislodged       OrderResult = "dislodged"
)

// OrderOutcome represents the complete result of order resolution
type OrderOutcome struct {
	Result        OrderResult
	Dislodged     bool
	SupportCut    bool
	ConvoyPath    []string
	Strength      int
	FailureReason string
	Destination   string // The actual destination
	IsCircular    bool   // True if part of circular movement
}

// ResolutionEngine provides the interface expected by the pipeline
type ResolutionEngine struct {
	orders  []*game.Order
	board   *game.Board
	results []OrderOutcome
}

// NewResolutionEngine creates a new resolution engine
func NewResolutionEngine(orders []*game.Order, board *game.Board) *ResolutionEngine {
	return &ResolutionEngine{
		orders:  orders,
		board:   board,
		results: make([]OrderOutcome, len(orders)),
	}
}

// Resolve executes the resolution algorithm
func (e *ResolutionEngine) Resolve() error {
	// Convert game.Order to our internal Order type
	internalOrders := make([]Order, len(e.orders))
	for i, order := range e.orders {
		internalOrders[i] = convertToInternalOrder(order)
	}

	// Run the recursive adjudication
	adj := NewAdjudicator(internalOrders)
	results := adj.ResolveAll()

	// Convert results back to legacy format
	for i, result := range results {
		outcome := OrderOutcome{
			Destination: e.orders[i].To,
			IsCircular:  false, // TODO: Detect circular movements
			Dislodged:   false, // TODO: Detect dislodgements
		}

		if result.Success {
			switch e.orders[i].Type {
			case game.Move:
				outcome.Result = MoveSuccess
			case game.Support:
				outcome.Result = SupportSuccess
			case game.Convoy:
				outcome.Result = ConvoySuccess
			case game.Hold:
				outcome.Result = HoldSuccess
			}
		} else {
			switch e.orders[i].Type {
			case game.Move:
				outcome.Result = MoveBounced
			case game.Support:
				outcome.Result = SupportCut
			case game.Convoy:
				outcome.Result = ConvoyDisrupted
			case game.Hold:
				outcome.Result = Dislodged
			}
		}

		e.results[i] = outcome
	}

	return nil
}

// GetResults returns the resolution results
func (e *ResolutionEngine) GetResults() []OrderOutcome {
	return e.results
}

// convertToInternalOrder converts a game.Order to our internal Order type
func convertToInternalOrder(order *game.Order) Order {
	var orderType OrderType

	// Map order types
	switch order.Type {
	case game.Move:
		orderType = Move
	case game.Support:
		orderType = Support
	case game.Convoy:
		orderType = Convoy
	case game.Hold:
		orderType = Hold
	default:
		orderType = Hold
	}

	return Order{
		Unit:        string(order.UnitType) + " " + order.From,
		Type:        orderType,
		Source:      order.From,
		Destination: order.To,
		Auxiliary:   order.SupportTarget, // For support orders
	}
}
