package resolution

import (
	"diplomacy-cli/backend/internal/game"
	"fmt"
)

func (re *ResolutionEngine) Resolve() error {
	if err := re.validateOrderRelationships(); err != nil {
		return fmt.Errorf("order validation failed: %w", err)
	}

	maxIterations := 10
	for iteration := range maxIterations {
		prevConvoyPaths := re.copyConvoyPaths()

		re.executeResolutionPass()

		if re.convoyPathsEqual(prevConvoyPaths, re.convoys) {
			break
		}

		if iteration == maxIterations-1 {
			return fmt.Errorf("resolution failed to converge after %d iterations", maxIterations)
		}
	}

	re.assignFinalOutcomes()

	return nil
}

func (re *ResolutionEngine) executeResolutionPass() {
	re.processConvoys()
	re.processMoves()
	re.cutSupports()
	re.calculateStrength()
	re.resolveConflicts()
	re.detectDislodgements()
}

func (re *ResolutionEngine) validateOrderRelationships() error {
	for i, order := range re.orders {
		switch order.Type {
		case game.Support:
			supportedIndex := re.FindOrderByTerritory(order.SupportTarget)

			if supportedIndex == -1 {
				re.outcomes[i].Result = SupportInvalid
				re.outcomes[i].FailureReason = fmt.Sprintf("no unit found at supported location %s", order.SupportTarget)
				continue
			}

			supportedOrder := re.orders[supportedIndex]

			if order.SupportDestination != "" {
				if supportedOrder.Type != game.Move || supportedOrder.To != order.SupportDestination {
					re.outcomes[i].Result = SupportInvalid
					re.outcomes[i].FailureReason = fmt.Sprintf("supported unit is not moving to %s", order.SupportDestination)
				}
			} else {
				if supportedOrder.Type != game.Hold {
					re.outcomes[i].Result = SupportInvalid
					re.outcomes[i].FailureReason = "supported unit is not holding"
				}
			}

		case game.Convoy:
			convoyedIndex := re.FindOrderByTerritory(order.ConvoyTarget)

			if convoyedIndex == -1 {
				re.outcomes[i].Result = ConvoyInvalid
				re.outcomes[i].FailureReason = fmt.Sprintf("no unit found at convoyed location %s", order.ConvoyTarget)
				continue
			}

			convoyedOrder := re.orders[convoyedIndex]

			if convoyedOrder.UnitType != game.Army {
				re.outcomes[i].Result = ConvoyInvalid
				re.outcomes[i].FailureReason = "can only convoy armies"
				continue
			}

			if convoyedOrder.Type != game.Move {
				re.outcomes[i].Result = ConvoyInvalid
				re.outcomes[i].FailureReason = "convoyed unit is not moving"
				continue
			}
		}
	}

	return nil
}

func (re *ResolutionEngine) processMoves() {
	for i, order := range re.orders {
		if order.Type != game.Move {
			continue
		}

		destination := order.To

		isAdjacent := re.IsAdjacent(order.OrigTerritory, destination, order.UnitType)

		if order.UnitType == game.Army {
			convoyKey := ConvoyKey{Origin: order.OrigTerritory, Destination: destination}
			hasConvoyPath := len(re.convoys[convoyKey]) > 0

			if isAdjacent || hasConvoyPath {
				order.NewTerritory = destination
				if hasConvoyPath {
					re.outcomes[i].ConvoyPath = re.convoys[convoyKey]
				}
			} else {
				re.outcomes[i].Result = MoveNoConvoy
				re.outcomes[i].FailureReason = "no adjacent path or convoy available"
			}
		} else if order.UnitType == game.Fleet {
			if isAdjacent {
				order.NewTerritory = destination
			} else {
				re.outcomes[i].Result = MoveInvalid
				re.outcomes[i].FailureReason = "destination not adjacent"
			}
		}
	}

	re.rebuildConflictMap()
}

func (re *ResolutionEngine) calculateStrength() {
	for i := range re.strength {
		re.strength[i] = 1
		re.outcomes[i].Strength = 1
	}

	for supportedUnitKey, supporterIndices := range re.supports {
		supportedIndex := re.FindOrderByUnit(supportedUnitKey, "")
		if supportedIndex == -1 {
			continue
		}

		for _, supporterIndex := range supporterIndices {
			if !re.outcomes[supporterIndex].SupportCut && re.outcomes[supporterIndex].Result != SupportInvalid {
				re.strength[supportedIndex]++
				re.outcomes[supportedIndex].Strength++
			}
		}
	}
}

func (re *ResolutionEngine) assignFinalOutcomes() {
	for i, order := range re.orders {
		if re.outcomes[i].Result != "" {
			continue
		}

		if re.outcomes[i].Dislodged {
			re.outcomes[i].Result = Dislodged
			continue
		}

		switch order.Type {
		case game.Move:
			if order.NewTerritory == order.To {
				re.outcomes[i].Result = MoveSuccess
			} else {
				re.outcomes[i].Result = MoveBounced
			}

		case game.Hold:
			re.outcomes[i].Result = HoldSuccess

		case game.Support:
			if re.outcomes[i].SupportCut {
				re.outcomes[i].Result = SupportCut
			} else {
				re.outcomes[i].Result = SupportSuccess
			}

		case game.Convoy:
			re.outcomes[i].Result = ConvoySuccess
		}
	}
}

func (re *ResolutionEngine) copyConvoyPaths() map[ConvoyKey][]string {
	result := make(map[ConvoyKey][]string)
	for key, path := range re.convoys {
		pathCopy := make([]string, len(path))
		copy(pathCopy, path)
		result[key] = pathCopy
	}
	return result
}

func (re *ResolutionEngine) convoyPathsEqual(a, b map[ConvoyKey][]string) bool {
	if len(a) != len(b) {
		return false
	}

	for key, pathA := range a {
		pathB, exists := b[key]
		if !exists || len(pathA) != len(pathB) {
			return false
		}

		for i, territory := range pathA {
			if territory != pathB[i] {
				return false
			}
		}
	}

	return true
}

func (re *ResolutionEngine) processConvoys() {
	re.convoys = make(map[ConvoyKey][]string)
}

func (re *ResolutionEngine) cutSupports() {
	for i := range re.outcomes {
		re.outcomes[i].SupportCut = false
	}
}

func (re *ResolutionEngine) resolveConflicts() {
	re.rebuildConflictMap()

	for _, orderIndices := range re.conflicts {
		if len(orderIndices) <= 1 {
			continue
		}

		maxStrength := 0
		for _, idx := range orderIndices {
			if re.strength[idx] > maxStrength {
				maxStrength = re.strength[idx]
			}
		}

		winners := []int{}
		for _, idx := range orderIndices {
			if re.strength[idx] == maxStrength {
				winners = append(winners, idx)
			}
		}

		if len(winners) > 1 {
			for _, idx := range orderIndices {
				if re.orders[idx].NewTerritory != re.orders[idx].OrigTerritory {
					re.orders[idx].NewTerritory = re.orders[idx].OrigTerritory
				}
			}
		} else {
			winner := winners[0]
			for _, idx := range orderIndices {
				if idx != winner && re.orders[idx].NewTerritory != re.orders[idx].OrigTerritory {
					re.orders[idx].NewTerritory = re.orders[idx].OrigTerritory
				}
			}
		}
	}
}

func (re *ResolutionEngine) detectDislodgements() {
	for i := range re.outcomes {
		re.outcomes[i].Dislodged = false
	}
}
