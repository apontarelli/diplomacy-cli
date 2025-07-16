package resolution

import (
	"diplomacy-cli/backend/internal/game"
	"fmt"
)

// ResolutionResult represents the complete outcome of a turn's resolution
// This is the primary domain model for capturing all resolution information
// needed for DATC compliance and detailed game state analysis
type ResolutionResult struct {
	turnNumber     int
	phase          game.Phase
	unitOutcomes   map[UnitID]UnitOutcome
	conflicts      []Conflict
	supportResults map[SupportID]SupportOutcome
	convoyResults  map[ConvoyID]ConvoyOutcome
	resolutionLog  []ResolutionStep
}

// UnitID uniquely identifies a unit across the game
// Combines unit type, owner, and province for unambiguous identification
type UnitID struct {
	Type     game.UnitType
	Owner    game.Nation
	Province string
}

// SupportID uniquely identifies a support order
// Combines supporting unit and support target for tracking
type SupportID struct {
	SupportingUnit UnitID
	SupportTarget  UnitID
}

// ConvoyID uniquely identifies a convoy operation
// Combines convoying fleet and convoy target
type ConvoyID struct {
	ConvoyingFleet UnitID
	ConvoyTarget   UnitID
}

// BuildResult captures the result of a build order
// Used for Winter phases (Builds/Disbands)
type BuildResult struct {
	Order   BuildOrder
	Success bool
	Reason  string
}

// BuildType distinguishes between builds and disbands
type BuildType int

const (
	Build BuildType = iota
	Disband
)

// RetreatResult captures the result of a retreat order
// Used for Retreat phases
type RetreatResult struct {
	Order     RetreatOrder
	Success   bool
	Reason    string // Human-readable explanation
	Disbanded bool   // True if unit was disbanded instead of retreating
}

// ConvoyPathResult captures the result of a convoy path resolution
type ConvoyPathResult struct {
	PathID      string
	IsValid     bool
	Path        []string
	DisruptedBy []UnitID
	Reason      string
}

// OrderID uniquely identifies an order
type OrderID struct {
	UnitID UnitID
	Order  game.Order
}

// OrderResolution captures the complete resolution of a single order
type OrderResolution struct {
	OrderID       OrderID
	Success       bool
	FinalLocation string
	Reason        string
	Dependencies  []UnitID
}

// UnitOutcome captures the complete result for a single unit's order
// Provides detailed information about what happened to this unit

type UnitOutcome struct {
	unit         game.Unit
	orderGiven   game.Order
	finalStatus  UnitStatus
	fromProvince string
	toProvince   *string
	dislodgedBy  *UnitID
	reason       string
	strength     int
	supportCount int
}

// UnitStatus represents the final state of a unit after resolution
type UnitStatus int

const (
	UnitMoved UnitStatus = iota
	UnitHeld
	UnitDislodged
	UnitBounced
	UnitRetreated
	UnitDisbanded
)

// Conflict represents a contested province with competing units
type Conflict struct {
	province    string
	competitors []UnitID
	winner      *UnitID
	resolution  ConflictType
	explanation string
	strengths   map[UnitID]int
}

// ConflictType categorizes how a conflict was resolved
type ConflictType int

const (
	SimpleMove ConflictType = iota
	HeadToHead
	Standoff
	CircularMovement
	ConvoyDisruption
)

// SupportOutcome captures the result of a support order
type SupportOutcome struct {
	supportingUnit UnitID
	supportedUnit  UnitID
	supportType    SupportType
	isSuccessful   bool
	isCut          bool
	cutBy          []UnitID
	reason         string
	strength       int
}

// SupportType distinguishes between move support and hold support
type SupportType int

const (
	MoveSupport SupportType = iota
	HoldSupport
)

// ConvoyOutcome captures the result of a convoy operation
type ConvoyOutcome struct {
	convoyingFleet UnitID
	convoyTarget   UnitID
	isSuccessful   bool
	isDisrupted    bool
	disruptedBy    []UnitID
	reason         string
	path           []string
}

// ResolutionStep represents a single step in the resolution process
// Used for debugging and DATC compliance verification
type ResolutionStep struct {
	stepNumber   int
	description  string
	unitID       UnitID
	action       string
	result       string
	dependencies []UnitID
}

// RetreatOrder represents a retreat order for a dislodged unit
type RetreatOrder struct {
	Unit        string // e.g., "A Berlin"
	From        string // Province retreating from
	Destination string // Province retreating to
	Coast       string // Coast specification for coastal retreats
	Owner       string // Nation that owns this unit
}

// DisbandResult represents the outcome of a disband order
type DisbandResult struct {
	Order   DisbandOrder
	Success bool
	Reason  string
}

// ConvoyPathID uniquely identifies a convoy path
type ConvoyPathID string

// DislodgedUnit represents a unit that has been dislodged and needs to retreat
type DislodgedUnit struct {
	Unit             string   // e.g., "A Berlin"
	Owner            string   // Nation that owns this unit
	DislodgedFrom    string   // Province where unit was dislodged
	AttackerOrigin   string   // Province the successful attacker came from
	PossibleRetreats []string // Valid retreat destinations
}

// NewResolutionResult creates a new ResolutionResult for the given turn and phase
func NewResolutionResult(turn int, phase game.Phase) *ResolutionResult {
	return &ResolutionResult{
		turnNumber:     turn,
		phase:          phase,
		unitOutcomes:   make(map[UnitID]UnitOutcome),
		conflicts:      make([]Conflict, 0),
		supportResults: make(map[SupportID]SupportOutcome),
		convoyResults:  make(map[ConvoyID]ConvoyOutcome),
		resolutionLog:  make([]ResolutionStep, 0),
	}
}

// Immutable accessors for ResolutionResult

// TurnNumber returns the turn number for this resolution
func (r *ResolutionResult) TurnNumber() int {
	return r.turnNumber
}

// Phase returns the phase for this resolution
func (r *ResolutionResult) Phase() game.Phase {
	return r.phase
}

// UnitOutcomes returns a copy of all unit outcomes to maintain immutability
func (r *ResolutionResult) UnitOutcomes() map[UnitID]UnitOutcome {
	result := make(map[UnitID]UnitOutcome, len(r.unitOutcomes))
	for k, v := range r.unitOutcomes {
		result[k] = v
	}
	return result
}

// Conflicts returns a copy of all conflicts
func (r *ResolutionResult) Conflicts() []Conflict {
	result := make([]Conflict, len(r.conflicts))
	copy(result, r.conflicts)
	return result
}

// SupportOutcomes returns a copy of all support outcomes
func (r *ResolutionResult) SupportOutcomes() map[SupportID]SupportOutcome {
	result := make(map[SupportID]SupportOutcome, len(r.supportResults))
	for k, v := range r.supportResults {
		result[k] = v
	}
	return result
}

// ConvoyOutcomes returns a copy of all convoy outcomes
func (r *ResolutionResult) ConvoyOutcomes() map[ConvoyID]ConvoyOutcome {
	result := make(map[ConvoyID]ConvoyOutcome, len(r.convoyResults))
	for k, v := range r.convoyResults {
		result[k] = v
	}
	return result
}

// ResolutionLog returns a copy of the resolution log
func (r *ResolutionResult) ResolutionLog() []ResolutionStep {
	result := make([]ResolutionStep, len(r.resolutionLog))
	copy(result, r.resolutionLog)
	return result
}

// AddUnitOutcome adds a unit outcome to the resolution result
func (r *ResolutionResult) AddUnitOutcome(unitID UnitID, outcome UnitOutcome) {
	r.unitOutcomes[unitID] = outcome
}

// AddConflict adds a conflict to the resolution result
func (r *ResolutionResult) AddConflict(conflict Conflict) {
	r.conflicts = append(r.conflicts, conflict)
}

// AddSupportOutcome adds a support outcome to the resolution result
func (r *ResolutionResult) AddSupportOutcome(supportID SupportID, outcome SupportOutcome) {
	r.supportResults[supportID] = outcome
}

// AddConvoyOutcome adds a convoy outcome to the resolution result
func (r *ResolutionResult) AddConvoyOutcome(convoyID ConvoyID, outcome ConvoyOutcome) {
	r.convoyResults[convoyID] = outcome
}

// AddResolutionStep adds a step to the resolution log
func (r *ResolutionResult) AddResolutionStep(step ResolutionStep) {
	r.resolutionLog = append(r.resolutionLog, step)
}

// String returns a human-readable representation of the resolution result
func (r *ResolutionResult) String() string {
	return fmt.Sprintf("ResolutionResult{Turn: %d, Phase: %s, Units: %d, Conflicts: %d}",
		r.turnNumber, r.phase, len(r.unitOutcomes), len(r.conflicts))
}

// Immutable accessors for UnitOutcome
func (uo *UnitOutcome) Unit() game.Unit {
	return game.Unit{
		Type:      uo.unit.Type,
		Owner:     uo.unit.Owner,
		Province:  uo.unit.Province,
		Coast:     uo.unit.Coast,
		Dislodged: uo.unit.Dislodged,
	}
}

func (uo *UnitOutcome) OrderGiven() game.Order {
	return game.Order{
		ID:                 uo.orderGiven.ID,
		UnitType:           uo.orderGiven.UnitType,
		From:               uo.orderGiven.From,
		FromCoast:          uo.orderGiven.FromCoast,
		Type:               uo.orderGiven.Type,
		To:                 uo.orderGiven.To,
		ToCoast:            uo.orderGiven.ToCoast,
		Owner:              uo.orderGiven.Owner,
		Result:             uo.orderGiven.Result,
		FailureReason:      uo.orderGiven.FailureReason,
		SupportTarget:      uo.orderGiven.SupportTarget,
		SupportDestination: uo.orderGiven.SupportDestination,
		ConvoyTarget:       uo.orderGiven.ConvoyTarget,
	}
}

func (uo *UnitOutcome) FinalStatus() UnitStatus { return uo.finalStatus }
func (uo *UnitOutcome) FromProvince() string    { return uo.fromProvince }
func (uo *UnitOutcome) ToProvince() *string {
	if uo.toProvince == nil {
		return nil
	}
	toProvince := *uo.toProvince
	return &toProvince
}
func (uo *UnitOutcome) DislodgedBy() *UnitID {
	if uo.dislodgedBy == nil {
		return nil
	}
	dislodgedBy := *uo.dislodgedBy
	return &dislodgedBy
}
func (uo *UnitOutcome) Reason() string    { return uo.reason }
func (uo *UnitOutcome) Strength() int     { return uo.strength }
func (uo *UnitOutcome) SupportCount() int { return uo.supportCount }

// Immutable accessors for Conflict
func (c *Conflict) Province() string { return c.province }
func (c *Conflict) Competitors() []UnitID {
	result := make([]UnitID, len(c.competitors))
	copy(result, c.competitors)
	return result
}
func (c *Conflict) Winner() *UnitID {
	if c.winner == nil {
		return nil
	}
	winner := *c.winner
	return &winner
}
func (c *Conflict) Resolution() ConflictType { return c.resolution }
func (c *Conflict) Explanation() string      { return c.explanation }
func (c *Conflict) Strengths() map[UnitID]int {
	result := make(map[UnitID]int, len(c.strengths))
	for k, v := range c.strengths {
		result[k] = v
	}
	return result
}

// Immutable accessors for SupportOutcome
func (so *SupportOutcome) SupportingUnit() UnitID   { return so.supportingUnit }
func (so *SupportOutcome) SupportedUnit() UnitID    { return so.supportedUnit }
func (so *SupportOutcome) SupportType() SupportType { return so.supportType }
func (so *SupportOutcome) IsSuccessful() bool       { return so.isSuccessful }
func (so *SupportOutcome) IsCut() bool              { return so.isCut }
func (so *SupportOutcome) CutBy() []UnitID {
	result := make([]UnitID, len(so.cutBy))
	copy(result, so.cutBy)
	return result
}
func (so *SupportOutcome) Reason() string { return so.reason }
func (so *SupportOutcome) Strength() int  { return so.strength }

// Immutable accessors for ConvoyOutcome
func (co *ConvoyOutcome) ConvoyingFleet() UnitID { return co.convoyingFleet }
func (co *ConvoyOutcome) ConvoyTarget() UnitID   { return co.convoyTarget }
func (co *ConvoyOutcome) IsSuccessful() bool     { return co.isSuccessful }
func (co *ConvoyOutcome) IsDisrupted() bool      { return co.isDisrupted }
func (co *ConvoyOutcome) DisruptedBy() []UnitID {
	result := make([]UnitID, len(co.disruptedBy))
	copy(result, co.disruptedBy)
	return result
}
func (co *ConvoyOutcome) Reason() string { return co.reason }
func (co *ConvoyOutcome) Path() []string {
	result := make([]string, len(co.path))
	copy(result, co.path)
	return result
}

// Immutable accessors for ResolutionStep
func (rs *ResolutionStep) StepNumber() int     { return rs.stepNumber }
func (rs *ResolutionStep) Description() string { return rs.description }
func (rs *ResolutionStep) UnitID() UnitID      { return rs.unitID }
func (rs *ResolutionStep) Action() string      { return rs.action }
func (rs *ResolutionStep) Result() string      { return rs.result }
func (rs *ResolutionStep) Dependencies() []UnitID {
	result := make([]UnitID, len(rs.dependencies))
	copy(result, rs.dependencies)
	return result
}

// Validate checks the integrity of the resolution result
func (r *ResolutionResult) Validate() error {
	if r.turnNumber <= 0 {
		return fmt.Errorf("invalid turn number: %d", r.turnNumber)
	}

	if len(r.phase) == 0 {
		return fmt.Errorf("phase cannot be empty")
	}

	// Validate that all conflicts reference valid unit outcomes
	for _, conflict := range r.conflicts {
		for _, competitor := range conflict.competitors {
			if _, exists := r.unitOutcomes[competitor]; !exists {
				return fmt.Errorf("conflict references unknown unit %s", competitor)
			}
		}
		if conflict.winner != nil {
			if _, exists := r.unitOutcomes[*conflict.winner]; !exists {
				return fmt.Errorf("conflict winner references unknown unit %s", *conflict.winner)
			}
		}
	}

	return nil
}
