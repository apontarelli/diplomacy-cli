package resolution

import (
	"diplomacy-cli/backend/internal/game"
	"fmt"
)

// ResolutionEngine defines the interface for order resolution engines
// This interface provides a clean abstraction for different resolution implementations
// while maintaining DATC compliance and extensibility
type ResolutionEngine interface {
	// Resolve processes a set of orders and returns the complete resolution result
	Resolve(orders []game.Order, board *game.Board) (*ResolutionResult, error)

	// ValidateOrders performs comprehensive validation of orders before resolution
	ValidateOrders(orders []game.Order, board *game.Board) ([]ValidationError, error)

	// GetEngineInfo returns metadata about this engine implementation
	GetEngineInfo() EngineInfo
}

// ValidationError represents a single validation error with context
type ValidationError struct {
	OrderIndex int                    `json:"order_index"`
	ErrorType  ValidationErrorType    `json:"error_type"`
	Message    string                 `json:"message"`
	Context    map[string]interface{} `json:"context"`
}

// ValidationErrorType categorizes different types of validation errors
type ValidationErrorType int

const (
	SyntaxError ValidationErrorType = iota
	SemanticError
	GameRuleError
	ContextError
)

// String returns a human-readable representation of the validation error type
func (vet ValidationErrorType) String() string {
	switch vet {
	case SyntaxError:
		return "syntax_error"
	case SemanticError:
		return "semantic_error"
	case GameRuleError:
		return "game_rule_error"
	case ContextError:
		return "context_error"
	default:
		return "unknown_error"
	}
}

// Error implements the error interface for ValidationError
func (ve ValidationError) Error() string {
	return fmt.Sprintf("[%s] Order %d: %s", ve.ErrorType.String(), ve.OrderIndex, ve.Message)
}

// EngineInfo provides metadata about a resolution engine implementation
type EngineInfo struct {
	Name        string            `json:"name"`
	Version     string            `json:"version"`
	Description string            `json:"description"`
	Features    []string          `json:"features"`
	Metadata    map[string]string `json:"metadata"`
}

// EngineConfig configures the behavior of a resolution engine
type EngineConfig struct {
	StrictDATCCompliance bool                  `json:"strict_datc_compliance"`
	ParadoxResolution    ParadoxResolutionMode `json:"paradox_resolution"`
	ValidationLevel      ValidationLevel       `json:"validation_level"`
	EnableDebugLogging   bool                  `json:"enable_debug_logging"`
	MaxRecursionDepth    int                   `json:"max_recursion_depth"`
}

// ParadoxResolutionMode defines how paradoxes should be resolved
type ParadoxResolutionMode int

const (
	OptimisticResolution ParadoxResolutionMode = iota
	PessimisticResolution
	BackupRuleResolution
	StrictDATCResolution
)

// ValidationLevel defines the depth of validation to perform
type ValidationLevel int

const (
	BasicValidation ValidationLevel = iota
	StandardValidation
	StrictValidation
	ComprehensiveValidation
)

// DATCCompliantEngine implements the ResolutionEngine interface with DATC compliance
type DATCCompliantEngine struct {
	validator    *OrderValidator
	resolver     *ConflictResolver
	convoyEngine *ConvoyEngine
	config       EngineConfig
}

// OrderValidator handles the multi-stage validation pipeline
type OrderValidator struct {
	config EngineConfig
}

// ConflictResolver handles move conflicts and strength calculations
type ConflictResolver struct {
	config EngineConfig
}

// ConvoyEngine handles convoy path validation and resolution
type ConvoyEngine struct {
	config EngineConfig
}

// NewDATCCompliantEngine creates a new DATC-compliant resolution engine
func NewDATCCompliantEngine(config EngineConfig) *DATCCompliantEngine {
	// Set default values for configuration
	if config.MaxRecursionDepth == 0 {
		config.MaxRecursionDepth = 100
	}

	return &DATCCompliantEngine{
		validator:    &OrderValidator{config: config},
		resolver:     &ConflictResolver{config: config},
		convoyEngine: &ConvoyEngine{config: config},
		config:       config,
	}
}

// NewDefaultDATCEngine creates a DATC engine with standard configuration
func NewDefaultDATCEngine() *DATCCompliantEngine {
	config := EngineConfig{
		StrictDATCCompliance: true,
		ParadoxResolution:    BackupRuleResolution,
		ValidationLevel:      StandardValidation,
		EnableDebugLogging:   false,
		MaxRecursionDepth:    100,
	}
	return NewDATCCompliantEngine(config)
}

// Resolve implements the ResolutionEngine interface
func (engine *DATCCompliantEngine) Resolve(orders []game.Order, board *game.Board) (*ResolutionResult, error) {
	// First validate all orders
	validationErrors, err := engine.ValidateOrders(orders, board)
	if err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// If there are validation errors and strict compliance is enabled, fail
	if len(validationErrors) > 0 && engine.config.StrictDATCCompliance {
		return nil, fmt.Errorf("validation errors found: %d errors", len(validationErrors))
	}

	// Convert game.Order to resolution.Order for compatibility with existing adjudicator
	resolutionOrders := engine.convertToResolutionOrders(orders)

	// Use existing adjudicator for resolution
	adjudicator := NewAdjudicator(resolutionOrders)
	adjudicationResults := adjudicator.ResolveAll()

	// Convert results to ResolutionResult format
	result := engine.buildResolutionResult(adjudicationResults, orders, board)

	return result, nil
}

// ValidateOrders implements comprehensive order validation
func (engine *DATCCompliantEngine) ValidateOrders(orders []game.Order, board *game.Board) ([]ValidationError, error) {
	var errors []ValidationError

	for i, order := range orders {
		// Syntax validation
		if syntaxErrors := engine.validator.validateSyntax(order, i); len(syntaxErrors) > 0 {
			errors = append(errors, syntaxErrors...)
		}

		// Semantic validation
		if semanticErrors := engine.validator.validateSemantics(order, board, i); len(semanticErrors) > 0 {
			errors = append(errors, semanticErrors...)
		}

		// Game rule validation
		if ruleErrors := engine.validator.validateGameRules(order, board, i); len(ruleErrors) > 0 {
			errors = append(errors, ruleErrors...)
		}

		// Context validation
		if contextErrors := engine.validator.validateContext(order, orders, board, i); len(contextErrors) > 0 {
			errors = append(errors, contextErrors...)
		}
	}

	return errors, nil
}

// GetEngineInfo returns metadata about this engine
func (engine *DATCCompliantEngine) GetEngineInfo() EngineInfo {
	return EngineInfo{
		Name:        "DATC Compliant Engine",
		Version:     "1.0.0",
		Description: "A Diplomacy Adjudicator Test Cases (DATC) compliant resolution engine",
		Features: []string{
			"DATC compliance",
			"Paradox resolution",
			"Convoy path validation",
			"Multi-stage order validation",
			"Detailed resolution logging",
		},
		Metadata: map[string]string{
			"strict_compliance": fmt.Sprintf("%t", engine.config.StrictDATCCompliance),
			"paradox_mode":      fmt.Sprintf("%d", engine.config.ParadoxResolution),
			"validation_level":  fmt.Sprintf("%d", engine.config.ValidationLevel),
		},
	}
}

// convertToResolutionOrders converts game.Order to resolution.Order for compatibility
func (engine *DATCCompliantEngine) convertToResolutionOrders(orders []game.Order) []Order {
	resolutionOrders := make([]Order, len(orders))

	for i, order := range orders {
		resolutionOrder := Order{
			Unit:        fmt.Sprintf("%s %s", order.UnitType, order.From),
			Source:      order.From,
			Destination: order.To,
			Owner:       string(order.Owner),
		}

		// Convert order type
		switch order.Type {
		case game.Move:
			resolutionOrder.Type = Move
		case game.Hold:
			resolutionOrder.Type = Hold
		case game.Support:
			resolutionOrder.Type = Support
			if order.SupportDestination != "" {
				resolutionOrder.Auxiliary = fmt.Sprintf("%s -> %s", order.SupportTarget, order.SupportDestination)
			} else {
				resolutionOrder.Auxiliary = order.SupportTarget
			}
		case game.Convoy:
			resolutionOrder.Type = Convoy
			resolutionOrder.Auxiliary = order.ConvoyTarget
		}

		resolutionOrders[i] = resolutionOrder
	}

	return resolutionOrders
}

// buildResolutionResult converts adjudication results to ResolutionResult
func (engine *DATCCompliantEngine) buildResolutionResult(adjResults []AdjudicationResult, originalOrders []game.Order, board *game.Board) *ResolutionResult {
	// For now, create a basic result structure
	// This would be expanded to include all the rich domain model information
	result := NewResolutionResult(1, game.SpringMovement) // TODO: Get actual turn/phase

	for i, adjResult := range adjResults {
		if i < len(originalOrders) {
			unitID := UnitID{
				Type:     game.UnitType(originalOrders[i].UnitType),
				Owner:    originalOrders[i].Owner,
				Province: originalOrders[i].From,
			}

			var finalStatus UnitStatus
			var toProvince *string
			if adjResult.Success {
				if originalOrders[i].Type == game.Move {
					finalStatus = UnitMoved
					toProvince = &originalOrders[i].To
				} else {
					finalStatus = UnitHeld
				}
			} else {
				finalStatus = UnitBounced
			}

			outcome := UnitOutcome{
				unit: game.Unit{
					Type:     game.UnitType(originalOrders[i].UnitType),
					Owner:    originalOrders[i].Owner,
					Province: originalOrders[i].From,
				},
				orderGiven:   originalOrders[i],
				finalStatus:  finalStatus,
				fromProvince: originalOrders[i].From,
				toProvince:   toProvince,
				reason:       adjResult.Reason,
			}

			result.AddUnitOutcome(unitID, outcome)
		}
	}

	return result
}

// Validation methods for OrderValidator

// validateSyntax performs basic syntax validation
func (ov *OrderValidator) validateSyntax(order game.Order, index int) []ValidationError {
	var errors []ValidationError

	// Check required fields
	if order.From == "" {
		errors = append(errors, ValidationError{
			OrderIndex: index,
			ErrorType:  SyntaxError,
			Message:    "order missing source province",
			Context:    map[string]interface{}{"field": "from"},
		})
	}

	if order.UnitType == "" {
		errors = append(errors, ValidationError{
			OrderIndex: index,
			ErrorType:  SyntaxError,
			Message:    "order missing unit type",
			Context:    map[string]interface{}{"field": "unit_type"},
		})
	}

	if order.Type == "" {
		errors = append(errors, ValidationError{
			OrderIndex: index,
			ErrorType:  SyntaxError,
			Message:    "order missing order type",
			Context:    map[string]interface{}{"field": "type"},
		})
	}

	// Validate order-specific syntax
	switch order.Type {
	case game.Move:
		if order.To == "" {
			errors = append(errors, ValidationError{
				OrderIndex: index,
				ErrorType:  SyntaxError,
				Message:    "move order missing destination",
				Context:    map[string]interface{}{"field": "to"},
			})
		}
	case game.Support:
		if order.SupportTarget == "" {
			errors = append(errors, ValidationError{
				OrderIndex: index,
				ErrorType:  SyntaxError,
				Message:    "support order missing target",
				Context:    map[string]interface{}{"field": "support_target"},
			})
		}
	case game.Convoy:
		if order.ConvoyTarget == "" {
			errors = append(errors, ValidationError{
				OrderIndex: index,
				ErrorType:  SyntaxError,
				Message:    "convoy order missing target",
				Context:    map[string]interface{}{"field": "convoy_target"},
			})
		}
	}

	return errors
}

// validateSemantics performs semantic validation
func (ov *OrderValidator) validateSemantics(order game.Order, board *game.Board, index int) []ValidationError {
	var errors []ValidationError

	// Check if unit exists at source province
	unit := board.GetUnit(order.From)
	if unit == nil {
		errors = append(errors, ValidationError{
			OrderIndex: index,
			ErrorType:  SemanticError,
			Message:    fmt.Sprintf("no unit found at province %s", order.From),
			Context:    map[string]interface{}{"province": order.From},
		})
		return errors // Can't continue validation without a unit
	}

	// Check unit type matches
	if unit.Type != order.UnitType {
		errors = append(errors, ValidationError{
			OrderIndex: index,
			ErrorType:  SemanticError,
			Message:    fmt.Sprintf("unit type mismatch: expected %s, got %s", unit.Type, order.UnitType),
			Context: map[string]interface{}{
				"expected": unit.Type,
				"actual":   order.UnitType,
			},
		})
	}

	// Check owner matches
	if unit.Owner != order.Owner {
		errors = append(errors, ValidationError{
			OrderIndex: index,
			ErrorType:  SemanticError,
			Message:    fmt.Sprintf("unit owner mismatch: expected %s, got %s", unit.Owner, order.Owner),
			Context: map[string]interface{}{
				"expected": unit.Owner,
				"actual":   order.Owner,
			},
		})
	}

	// Validate destination province exists (for move orders)
	if order.Type == game.Move && order.To != "" {
		if board.GetProvince(order.To) == nil {
			errors = append(errors, ValidationError{
				OrderIndex: index,
				ErrorType:  SemanticError,
				Message:    fmt.Sprintf("destination province %s does not exist", order.To),
				Context:    map[string]interface{}{"province": order.To},
			})
		}
	}

	return errors
}

// validateGameRules performs game rule validation
func (ov *OrderValidator) validateGameRules(order game.Order, board *game.Board, index int) []ValidationError {
	var errors []ValidationError

	unit := board.GetUnit(order.From)
	if unit == nil {
		return errors // Already caught in semantic validation
	}

	switch order.Type {
	case game.Move:
		if order.To != "" {
			// Check adjacency for direct moves
			if !board.IsAdjacent(order.From, order.To) {
				// This might be a convoy move, so we don't immediately error
				// The convoy validation will handle this case
			}

			// Check unit type can move to destination
			destProvince := board.GetProvince(order.To)
			if destProvince != nil {
				if unit.Type == game.Army && destProvince.Type == game.Sea {
					errors = append(errors, ValidationError{
						OrderIndex: index,
						ErrorType:  GameRuleError,
						Message:    "army cannot move to sea province without convoy",
						Context: map[string]interface{}{
							"unit_type": unit.Type,
							"province":  order.To,
						},
					})
				}
			}
		}

	case game.Support:
		// Validate support target exists
		if order.SupportTarget != "" {
			supportedUnit := board.GetUnit(order.SupportTarget)
			if supportedUnit == nil {
				errors = append(errors, ValidationError{
					OrderIndex: index,
					ErrorType:  GameRuleError,
					Message:    fmt.Sprintf("cannot support non-existent unit at %s", order.SupportTarget),
					Context:    map[string]interface{}{"target": order.SupportTarget},
				})
			}
		}

	case game.Convoy:
		// Only fleets can convoy
		if unit.Type != game.Fleet {
			errors = append(errors, ValidationError{
				OrderIndex: index,
				ErrorType:  GameRuleError,
				Message:    "only fleets can convoy",
				Context:    map[string]interface{}{"unit_type": unit.Type},
			})
		}
	}

	return errors
}

// validateContext performs context validation
func (ov *OrderValidator) validateContext(order game.Order, allOrders []game.Order, board *game.Board, index int) []ValidationError {
	var errors []ValidationError

	// Check for duplicate orders from same unit
	for i, otherOrder := range allOrders {
		if i != index && otherOrder.From == order.From && otherOrder.Owner == order.Owner {
			errors = append(errors, ValidationError{
				OrderIndex: index,
				ErrorType:  ContextError,
				Message:    fmt.Sprintf("duplicate order for unit at %s", order.From),
				Context: map[string]interface{}{
					"province":    order.From,
					"other_index": i,
				},
			})
		}
	}

	// Additional context validations can be added here
	// such as phase-appropriate orders, etc.

	return errors
}
