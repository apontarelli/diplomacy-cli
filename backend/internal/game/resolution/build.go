package resolution

import (
	"diplomacy-cli/backend/internal/game"
	"fmt"
)

// BuildProcessor handles build and adjustment phase logic
type BuildProcessor struct {
	board *game.Board
}

// NewBuildProcessor creates a new build processor
func NewBuildProcessor(board *game.Board) *BuildProcessor {
	return &BuildProcessor{
		board: board,
	}
}

// CalculateSupplyCenters calculates the number of supply centers controlled by each player
func (bp *BuildProcessor) CalculateSupplyCenters(gameState *game.GameState) map[game.Nation]int {
	supplyCenterCounts := make(map[game.Nation]int)

	// Initialize all nations to 0
	nations := []game.Nation{
		game.Austria, game.England, game.France, game.Germany,
		game.Italy, game.Russia, game.Turkey,
	}
	for _, nation := range nations {
		supplyCenterCounts[nation] = 0
	}

	// In winter build phase, use the supply center ownership from game state
	// (which should be set from the previous turn's results)
	if gameState.Phase == game.WinterBuild && len(gameState.SupplyCenters) > 0 {
		for nation, centers := range gameState.SupplyCenters {
			supplyCenterCounts[nation] = len(centers)
		}
		return supplyCenterCounts
	}

	// For other phases, count supply centers by checking which nation controls each supply center province
	for provinceName, province := range gameState.Board.Provinces {
		if province.SupplyCenter {
			// Check if there's a unit in this province
			unit := gameState.Board.GetUnit(provinceName)
			if unit != nil {
				supplyCenterCounts[unit.Owner]++
			}
		}
	}

	return supplyCenterCounts
}

// CalculateUnitCounts calculates the number of units each player currently has
func (bp *BuildProcessor) CalculateUnitCounts(gameState *game.GameState) map[game.Nation]int {
	unitCounts := make(map[game.Nation]int)

	// Initialize all nations to 0
	nations := []game.Nation{
		game.Austria, game.England, game.France, game.Germany,
		game.Italy, game.Russia, game.Turkey,
	}
	for _, nation := range nations {
		unitCounts[nation] = 0
	}

	// Count units for each nation
	for _, unit := range gameState.Board.Units {
		if unit != nil {
			unitCounts[unit.Owner]++
		}
	}

	return unitCounts
}

// CalculateBuildAdjustments calculates how many builds/disbands each player needs
// Positive values = builds allowed, negative values = disbands required
func (bp *BuildProcessor) CalculateBuildAdjustments(gameState *game.GameState) map[game.Nation]int {
	supplyCenterCounts := bp.CalculateSupplyCenters(gameState)
	unitCounts := bp.CalculateUnitCounts(gameState)

	adjustments := make(map[game.Nation]int)

	for nation := range supplyCenterCounts {
		adjustments[nation] = supplyCenterCounts[nation] - unitCounts[nation]
	}

	return adjustments
}

// BuildOrder represents a build order during winter build phase
type BuildOrder struct {
	Nation   game.Nation
	UnitType game.UnitType
	Province string
	Coast    string // For fleet builds on coastal provinces with multiple coasts
}

// DisbandOrder represents a disband order during winter build phase
type DisbandOrder struct {
	Nation   game.Nation
	Province string
}

// BuildResult represents the outcome of a build order
type BuildResult struct {
	Order   BuildOrder
	Success bool
	Reason  string
}

// DisbandResult represents the outcome of a disband order
type DisbandResult struct {
	Order   DisbandOrder
	Success bool
	Reason  string
}

// ValidateBuildOrder validates a build order against game rules
func (bp *BuildProcessor) ValidateBuildOrder(order BuildOrder, gameState *game.GameState) error {
	// Check if the nation has builds available
	adjustments := bp.CalculateBuildAdjustments(gameState)
	if adjustments[order.Nation] <= 0 {
		return fmt.Errorf("nation %s has no builds available", order.Nation)
	}

	// Check if the province exists
	province := gameState.Board.GetProvince(order.Province)
	if province == nil {
		return fmt.Errorf("province %s does not exist", order.Province)
	}

	// Check if the province is already occupied
	if gameState.Board.GetUnit(order.Province) != nil {
		return fmt.Errorf("province %s is already occupied", order.Province)
	}

	// Check if the province is a home center for this nation
	if !bp.isHomeCenter(order.Province, order.Nation) {
		return fmt.Errorf("province %s is not a home center for %s", order.Province, order.Nation)
	}

	// Validate unit type can be placed in this province
	if order.UnitType == game.Army && province.Type == game.Sea {
		return fmt.Errorf("cannot build army in sea province %s", order.Province)
	}

	if order.UnitType == game.Fleet && !bp.canBuildFleet(order.Province) {
		return fmt.Errorf("cannot build fleet in province %s", order.Province)
	}

	return nil
}

// ValidateDisbandOrder validates a disband order against game rules
func (bp *BuildProcessor) ValidateDisbandOrder(order DisbandOrder, gameState *game.GameState) error {
	// Check if the nation has disbands required
	adjustments := bp.CalculateBuildAdjustments(gameState)
	if adjustments[order.Nation] >= 0 {
		return fmt.Errorf("nation %s has no disbands required", order.Nation)
	}

	// Check if there's a unit at the specified province
	unit := gameState.Board.GetUnit(order.Province)
	if unit == nil {
		return fmt.Errorf("no unit found at province %s", order.Province)
	}

	// Check if the unit belongs to the nation
	if unit.Owner != order.Nation {
		return fmt.Errorf("unit at %s belongs to %s, not %s", order.Province, unit.Owner, order.Nation)
	}

	return nil
}

// ProcessBuildOrder executes a validated build order
func (bp *BuildProcessor) ProcessBuildOrder(order BuildOrder, gameState *game.GameState) BuildResult {
	err := bp.ValidateBuildOrder(order, gameState)
	if err != nil {
		return BuildResult{
			Order:   order,
			Success: false,
			Reason:  err.Error(),
		}
	}

	// Create and place the new unit
	newUnit := &game.Unit{
		Type:     order.UnitType,
		Owner:    order.Nation,
		Province: order.Province,
		Coast:    order.Coast,
	}

	err = gameState.Board.PlaceUnit(newUnit)
	if err != nil {
		return BuildResult{
			Order:   order,
			Success: false,
			Reason:  fmt.Sprintf("failed to place unit: %v", err),
		}
	}

	return BuildResult{
		Order:   order,
		Success: true,
		Reason:  "build successful",
	}
}

// ProcessDisbandOrder executes a validated disband order
func (bp *BuildProcessor) ProcessDisbandOrder(order DisbandOrder, gameState *game.GameState) DisbandResult {
	err := bp.ValidateDisbandOrder(order, gameState)
	if err != nil {
		return DisbandResult{
			Order:   order,
			Success: false,
			Reason:  err.Error(),
		}
	}

	// Remove the unit
	gameState.Board.RemoveUnit(order.Province)

	return DisbandResult{
		Order:   order,
		Success: true,
		Reason:  "disband successful",
	}
}

// isHomeCenter checks if a province is a home center for the given nation
func (bp *BuildProcessor) isHomeCenter(provinceName string, nation game.Nation) bool {
	// This would typically be loaded from game data, but for now we'll hardcode
	// the classic Diplomacy home centers
	homeCenters := map[game.Nation][]string{
		game.Austria: {"vienna", "budapest", "trieste"},
		game.England: {"london", "liverpool", "edinburgh"},
		game.France:  {"paris", "marseilles", "brest"},
		game.Germany: {"berlin", "munich", "kiel"},
		game.Italy:   {"rome", "venice", "naples"},
		game.Russia:  {"moscow", "st_petersburg", "sevastopol", "warsaw"},
		game.Turkey:  {"constantinople", "ankara", "smyrna"},
	}

	centers, exists := homeCenters[nation]
	if !exists {
		return false
	}

	for _, center := range centers {
		if center == provinceName {
			return true
		}
	}

	return false
}

// canBuildFleet checks if a fleet can be built in the given province
func (bp *BuildProcessor) canBuildFleet(provinceName string) bool {
	province := bp.board.GetProvince(provinceName)
	if province == nil {
		return false
	}

	// Fleet can be built in sea provinces or coastal land provinces
	return province.Type == game.Sea || len(province.FleetNeighbors) > 0
}
