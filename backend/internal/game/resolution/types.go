package resolution

// OrderType defines the type of order (Move, Support, Convoy).
type OrderType int

const (
	Move OrderType = iota
	Support
	Convoy
	Hold
)

// Order represents a single unit's order with resolution state.
type Order struct {
	Unit        string // e.g., "A Berlin"
	Type        OrderType
	Source      string // Source territory
	Destination string // Destination territory (for moves)
	Auxiliary   string // For supports/convoys, the location being supported/convoyed

	// Resolution state - managed by adjudicator
	isResolved bool
	resolution bool // true for success, false for failure
	isVisited  bool // for cycle detection
}

// InternalOrderResult represents the outcome of an order after adjudication.
type InternalOrderResult struct {
	Order   Order
	Success bool
	Reason  string // Human-readable explanation of the result
}

// String returns a human-readable representation of the order.
func (o Order) String() string {
	switch o.Type {
	case Move:
		return o.Unit + " -> " + o.Destination
	case Support:
		return o.Unit + " S " + o.Auxiliary
	case Convoy:
		return o.Unit + " C " + o.Auxiliary
	case Hold:
		return o.Unit + " H"
	default:
		return o.Unit + " (unknown order)"
	}
}

// IsResolved returns whether this order has been definitively resolved.
func (o *Order) IsResolved() bool {
	return o.isResolved
}

// Resolution returns the resolution result (only valid if IsResolved() is true).
func (o *Order) Resolution() bool {
	return o.resolution
}

// setResolution marks the order as resolved with the given result.
func (o *Order) setResolution(success bool) {
	o.isResolved = true
	o.resolution = success
}

// reset clears the resolution state (used between adjudication runs).
func (o *Order) reset() {
	o.isResolved = false
	o.resolution = false
	o.isVisited = false
}
