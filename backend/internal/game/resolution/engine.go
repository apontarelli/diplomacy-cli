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
		prevSupportCuts := re.copySupportCuts()

		re.executeResolutionPass()

		if re.convoyPathsEqual(prevConvoyPaths, re.convoys) && re.supportCutsEqual(prevSupportCuts) {
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
			origin := NewProvinceCoast(order.OrigTerritory, order.FromCoast)
			dest := NewProvinceCoast(destination, order.ToCoast)
			convoyKey := ConvoyKey{Origin: origin, Destination: dest}
			convoyPath, hasConvoyPath := re.convoys[convoyKey]

			if isAdjacent || hasConvoyPath {
				order.NewTerritory = destination
				if hasConvoyPath {

					stringPath := make([]string, len(convoyPath))
					for j, pc := range convoyPath {
						stringPath[j] = pc.String()
					}
					re.outcomes[i].ConvoyPath = stringPath
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

func (re *ResolutionEngine) copyConvoyPaths() map[ConvoyKey][]ProvinceCoast {
	result := make(map[ConvoyKey][]ProvinceCoast)
	for key, path := range re.convoys {
		pathCopy := make([]ProvinceCoast, len(path))
		copy(pathCopy, path)
		result[key] = pathCopy
	}
	return result
}

func (re *ResolutionEngine) convoyPathsEqual(a, b map[ConvoyKey][]ProvinceCoast) bool {
	if len(a) != len(b) {
		return false
	}

	for key, pathA := range a {
		pathB, exists := b[key]
		if !exists || len(pathA) != len(pathB) {
			return false
		}

		for i, pc := range pathA {
			if pc != pathB[i] {
				return false
			}
		}
	}

	return true
}

func (re *ResolutionEngine) copySupportCuts() []bool {
	result := make([]bool, len(re.outcomes))
	for i, outcome := range re.outcomes {
		result[i] = outcome.SupportCut
	}
	return result
}

func (re *ResolutionEngine) supportCutsEqual(prev []bool) bool {
	if len(prev) != len(re.outcomes) {
		return false
	}

	for i, prevCut := range prev {
		if prevCut != re.outcomes[i].SupportCut {
			return false
		}
	}

	return true
}

func (re *ResolutionEngine) cutSupports() {
	for i := range re.outcomes {
		re.outcomes[i].SupportCut = false
	}

	for i, order := range re.orders {
		if order.Type != game.Move {
			continue
		}

		if order.To == "" || order.To == order.From {
			continue
		}

		destination := order.To

		supporterIndex := re.FindOrderByTerritory(destination)
		if supporterIndex == -1 {
			continue
		}

		supporterOrder := re.orders[supporterIndex]
		if supporterOrder.Type != game.Support {
			continue
		}

		if re.isSelfAttackSupportCut(i, supporterIndex) {
			continue
		}

		re.outcomes[supporterIndex].SupportCut = true
	}
}

func (re *ResolutionEngine) isSelfAttackSupportCut(attackerIndex, supporterIndex int) bool {
	attacker := re.orders[attackerIndex]
	supporter := re.orders[supporterIndex]

	supportedUnitKey := makeUnitKey(supporter.SupportTarget, "")
	supportedIndex := re.FindOrderByUnit(supportedUnitKey, "")
	if supportedIndex == -1 {
		return false
	}

	supported := re.orders[supportedIndex]

	if supported.Type == game.Move && supported.To == attacker.From {
		return true
	}

	return false
}

func (re *ResolutionEngine) hasFriendlyProtection(territory string, conflictIndices []int, winnerIndex int) bool {
	defenderIndex := re.FindOrderByTerritory(territory)
	if defenderIndex == -1 {
		return false
	}

	defender := re.orders[defenderIndex]
	winner := re.orders[winnerIndex]

	if winnerIndex == defenderIndex {
		return false
	}

	if winner.Owner == defender.Owner {
		return true
	}

	for _, conflictIndex := range conflictIndices {
		if conflictIndex == winnerIndex || conflictIndex == defenderIndex {
			continue
		}

		order := re.orders[conflictIndex]
		if order.Type == game.Move && order.Owner == defender.Owner {
			return true
		}
	}

	return false
}

func (re *ResolutionEngine) resolveConflicts() {
	re.rebuildConflictMap()

	for territory, orderIndices := range re.conflicts {
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

		if len(winners) == 1 && re.hasFriendlyProtection(territory, orderIndices, winners[0]) {
			for _, idx := range orderIndices {
				if re.orders[idx].NewTerritory != re.orders[idx].OrigTerritory {
					re.orders[idx].NewTerritory = re.orders[idx].OrigTerritory
				}
			}
		} else if len(winners) > 1 {
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

	for territory, conflictIndices := range re.conflicts {
		if len(conflictIndices) <= 1 {
			continue
		}

		defenderIndex := re.FindOrderByTerritory(territory)
		if defenderIndex == -1 {
			continue
		}

		for _, conflictIndex := range conflictIndices {
			if conflictIndex == defenderIndex {
				continue
			}

			order := re.orders[conflictIndex]

			if order.Type == game.Move && order.NewTerritory == territory && order.OrigTerritory != territory {
				re.outcomes[defenderIndex].Dislodged = true
				break
			}
		}
	}
}
