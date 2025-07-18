package resolution

// OrderType defines the type of order (Move, Support, Convoy).
type OrderType int

const (
	Move OrderType = iota
	Support
	Convoy
	Hold
	Retreat
)

// String returns the string representation of an OrderType using interned strings
func (ot OrderType) String() string {
	switch ot {
	case Move:
		return MoveString
	case Support:
		return SupportString
	case Convoy:
		return ConvoyString
	case Hold:
		return HoldString
	case Retreat:
		return RetreatString
	default:
		return UnknownString
	}
}

// Order represents a single unit's order with resolution state.
type Order struct {
	Unit        string // e.g., "A Berlin"
	Type        OrderType
	Source      string // Source territory
	Destination string // Destination territory (for moves)
	Auxiliary   string // For supports/convoys, the location being supported/convoyed
	Owner       string // Nation that owns this unit

	// Resolution state - managed by adjudicator
	isResolved bool
	resolution bool // true for success, false for failure
	isVisited  bool // for cycle detection

	// Swap detection state
	isSwap      bool   // true if this order is part of a unit swap
	swapPartner *Order // pointer to the other order in the swap

	// Circular movement detection state
	isCircular    bool     // true if this order is part of a circular movement
	circularGroup []*Order // all orders in the same circular movement
}

// AdjudicationResult represents the outcome of an order after adjudication.
type AdjudicationResult struct {
	Order       Order
	Success     bool
	Reason      string // Human-readable explanation of the result
	Destination string // Actual destination (same as order.Destination for normal moves)
	Dislodged   bool   // True if this unit was dislodged
}

// OrderOutcome provides compatibility with the pipeline interface
type OrderOutcome struct {
	Result      string // "move_success", "move_bounced", etc.
	Destination string // Actual destination (same as order.To for normal moves)
	Dislodged   bool   // True if this unit was dislodged
}

// Order result constants for compatibility
const (
	MoveSuccess      string = "move_success"
	MoveBounced      string = "move_bounced"
	MoveNoConvoy     string = "move_no_convoy"
	SupportSuccess   string = "support_success"
	SupportCut       string = "support_cut"
	ConvoySuccess    string = "convoy_success"
	ConvoyDisrupted  string = "convoy_disrupted"
	HoldSuccess      string = "hold_success"
	Dislodged        string = "dislodged"
	RetreatSuccess   string = "retreat_success"
	RetreatFailed    string = "retreat_failed"
	RetreatDisbanded string = "retreat_disbanded"
)

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
	case Retreat:
		return o.Unit + " R " + o.Destination
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
	o.isCircular = false
	o.circularGroup = nil
}

// NewOrder creates a new order with interned strings for memory efficiency
func NewOrder(unit, source, destination, auxiliary, owner string, orderType OrderType) Order {
	return Order{
		Unit:        InternUnit(unit),
		Type:        orderType,
		Source:      InternTerritory(source),
		Destination: InternTerritory(destination),
		Auxiliary:   InternString(auxiliary),
		Owner:       InternOwner(owner),
	}
}

// InternStrings interns all string fields in the order for memory efficiency
func (o *Order) InternStrings() {
	o.Unit = InternUnit(o.Unit)
	o.Source = InternTerritory(o.Source)
	o.Destination = InternTerritory(o.Destination)
	o.Auxiliary = InternString(o.Auxiliary)
	o.Owner = InternOwner(o.Owner)
}
