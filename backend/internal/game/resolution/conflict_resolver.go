package resolution

import (
	"fmt"
)

// DetailedConflictResolver handles move conflicts and determines resolution outcomes
// This implements the core conflict resolution logic for DATC compliance
type DetailedConflictResolver struct {
	adjudicator *Adjudicator
}

// NewDetailedConflictResolver creates a new detailed conflict resolver
func NewDetailedConflictResolver(adj *Adjudicator) *DetailedConflictResolver {
	return &DetailedConflictResolver{
		adjudicator: adj,
	}
}

// ConflictAnalysis represents a detailed analysis of a move conflict
type ConflictAnalysis struct {
	Province     string                         `json:"province"`
	Competitors  []*Order                       `json:"competitors"`
	Winner       *Order                         `json:"winner"`
	ConflictType ConflictType                   `json:"conflict_type"`
	Explanation  string                         `json:"explanation"`
	Strengths    map[string]StrengthCalculation `json:"strengths"`
	Resolution   ResolutionDecision             `json:"resolution"`
}

// ResolutionDecision captures the final decision and reasoning
type ResolutionDecision struct {
	Decision     string   `json:"decision"`
	Reasoning    []string `json:"reasoning"`
	DATCRules    []string `json:"datc_rules"`
	IsStandoff   bool     `json:"is_standoff"`
	IsHeadToHead bool     `json:"is_head_to_head"`
}

// AnalyzeConflicts identifies and analyzes all move conflicts
func (cr *DetailedConflictResolver) AnalyzeConflicts(optimistic bool) []ConflictAnalysis {
	conflicts := make([]ConflictAnalysis, 0)

	// Group moves by destination
	movesByDestination := make(map[string][]*Order)
	for _, order := range cr.adjudicator.orders {
		if order.Type == Move {
			dest := order.Destination
			movesByDestination[dest] = append(movesByDestination[dest], order)
		}
	}

	// Analyze each contested province
	for province, competitors := range movesByDestination {
		if len(competitors) > 1 {
			analysis := cr.analyzeProvinceConflict(province, competitors, optimistic)
			conflicts = append(conflicts, analysis)
		} else if len(competitors) == 1 {
			// Single move - check against hold strength
			analysis := cr.analyzeSingleMove(province, competitors[0], optimistic)
			conflicts = append(conflicts, analysis)
		}
	}

	return conflicts
}

// analyzeProvinceConflict analyzes a conflict where multiple units move to the same province
func (cr *DetailedConflictResolver) analyzeProvinceConflict(province string, competitors []*Order, optimistic bool) ConflictAnalysis {
	analysis := ConflictAnalysis{
		Province:    province,
		Competitors: competitors,
		Strengths:   make(map[string]StrengthCalculation),
		Resolution: ResolutionDecision{
			Reasoning: make([]string, 0),
			DATCRules: make([]string, 0),
		},
	}

	analysis.Resolution.Reasoning = append(analysis.Resolution.Reasoning,
		fmt.Sprintf("Analyzing conflict for province %s with %d competing moves", province, len(competitors)))

	// Check for head-to-head battles first
	headToHeadPairs := cr.findHeadToHeadBattles(competitors)
	if len(headToHeadPairs) > 0 {
		analysis.ConflictType = HeadToHead
		analysis.Resolution.IsHeadToHead = true
		analysis.Resolution.DATCRules = append(analysis.Resolution.DATCRules, "DATC 6.A.1: Head-to-head battles")
		return cr.resolveHeadToHeadConflict(analysis, headToHeadPairs, optimistic)
	}

	// Calculate strengths for all competitors
	maxStrength := 0
	var strongestMoves []*Order

	for _, competitor := range competitors {
		strengthCalc := cr.adjudicator.calculateAttackStrengthDetailed(competitor, optimistic)
		analysis.Strengths[competitor.Source] = strengthCalc

		analysis.Resolution.Reasoning = append(analysis.Resolution.Reasoning,
			fmt.Sprintf("Move %s -> %s: attack strength %d",
				competitor.Source, competitor.Destination, strengthCalc.TotalStrength))

		if strengthCalc.TotalStrength > maxStrength {
			maxStrength = strengthCalc.TotalStrength
			strongestMoves = []*Order{competitor}
		} else if strengthCalc.TotalStrength == maxStrength {
			strongestMoves = append(strongestMoves, competitor)
		}
	}

	// Check hold strength of destination
	holdStrengthCalc := cr.adjudicator.calculateHoldStrengthDetailed(province, !optimistic)
	analysis.Strengths[province+"_hold"] = holdStrengthCalc

	analysis.Resolution.Reasoning = append(analysis.Resolution.Reasoning,
		fmt.Sprintf("Province %s hold strength: %d", province, holdStrengthCalc.TotalStrength))

	// Determine outcome
	if maxStrength > holdStrengthCalc.TotalStrength {
		if len(strongestMoves) == 1 {
			// Single strongest move succeeds
			analysis.Winner = strongestMoves[0]
			analysis.ConflictType = SimpleMove
			analysis.Resolution.Decision = fmt.Sprintf("Move %s -> %s succeeds",
				strongestMoves[0].Source, strongestMoves[0].Destination)
			analysis.Resolution.Reasoning = append(analysis.Resolution.Reasoning,
				fmt.Sprintf("Attack strength %d > hold strength %d", maxStrength, holdStrengthCalc.TotalStrength))
		} else {
			// Multiple moves with equal strength - standoff
			analysis.ConflictType = Standoff
			analysis.Resolution.IsStandoff = true
			analysis.Resolution.Decision = "Standoff - no move succeeds"
			analysis.Resolution.Reasoning = append(analysis.Resolution.Reasoning,
				fmt.Sprintf("Multiple moves with equal strength %d - standoff occurs", maxStrength))
			analysis.Resolution.DATCRules = append(analysis.Resolution.DATCRules, "DATC 6.A.2: Equal strength standoff")
		}
	} else {
		// All moves fail against hold strength
		analysis.ConflictType = SimpleMove
		analysis.Resolution.Decision = "All moves fail - province holds"
		analysis.Resolution.Reasoning = append(analysis.Resolution.Reasoning,
			fmt.Sprintf("Maximum attack strength %d <= hold strength %d", maxStrength, holdStrengthCalc.TotalStrength))
	}

	analysis.Explanation = analysis.Resolution.Decision
	return analysis
}

// analyzeSingleMove analyzes a single move against hold strength
func (cr *DetailedConflictResolver) analyzeSingleMove(province string, move *Order, optimistic bool) ConflictAnalysis {
	analysis := ConflictAnalysis{
		Province:    province,
		Competitors: []*Order{move},
		Strengths:   make(map[string]StrengthCalculation),
		Resolution: ResolutionDecision{
			Reasoning: make([]string, 0),
			DATCRules: make([]string, 0),
		},
	}

	// Calculate attack strength
	attackStrengthCalc := cr.adjudicator.calculateAttackStrengthDetailed(move, optimistic)
	analysis.Strengths[move.Source] = attackStrengthCalc

	// Calculate hold strength
	holdStrengthCalc := cr.adjudicator.calculateHoldStrengthDetailed(province, !optimistic)
	analysis.Strengths[province+"_hold"] = holdStrengthCalc

	analysis.Resolution.Reasoning = append(analysis.Resolution.Reasoning,
		fmt.Sprintf("Single move %s -> %s: attack strength %d vs hold strength %d",
			move.Source, move.Destination, attackStrengthCalc.TotalStrength, holdStrengthCalc.TotalStrength))

	if attackStrengthCalc.TotalStrength > holdStrengthCalc.TotalStrength {
		analysis.Winner = move
		analysis.ConflictType = SimpleMove
		analysis.Resolution.Decision = fmt.Sprintf("Move %s -> %s succeeds", move.Source, move.Destination)
		analysis.Resolution.Reasoning = append(analysis.Resolution.Reasoning,
			fmt.Sprintf("Attack strength %d > hold strength %d",
				attackStrengthCalc.TotalStrength, holdStrengthCalc.TotalStrength))
	} else {
		analysis.ConflictType = SimpleMove
		analysis.Resolution.Decision = "Move fails - province holds"
		analysis.Resolution.Reasoning = append(analysis.Resolution.Reasoning,
			fmt.Sprintf("Attack strength %d <= hold strength %d",
				attackStrengthCalc.TotalStrength, holdStrengthCalc.TotalStrength))
	}

	analysis.Explanation = analysis.Resolution.Decision
	return analysis
}

// findHeadToHeadBattles identifies head-to-head battles among competitors
func (cr *DetailedConflictResolver) findHeadToHeadBattles(competitors []*Order) [][]*Order {
	var pairs [][]*Order

	for i, order1 := range competitors {
		for j, order2 := range competitors {
			if i < j && cr.adjudicator.wouldCreateHeadToHead(order1, order2) {
				pairs = append(pairs, []*Order{order1, order2})
			}
		}
	}

	return pairs
}

// resolveHeadToHeadConflict resolves head-to-head battles using defend strength
func (cr *DetailedConflictResolver) resolveHeadToHeadConflict(analysis ConflictAnalysis, headToHeadPairs [][]*Order, optimistic bool) ConflictAnalysis {
	analysis.Resolution.Reasoning = append(analysis.Resolution.Reasoning, "Head-to-head battle detected")

	// For now, handle simple case of one head-to-head pair
	if len(headToHeadPairs) == 1 {
		pair := headToHeadPairs[0]
		order1, order2 := pair[0], pair[1]

		// Calculate attack vs defend strengths
		attack1 := cr.adjudicator.calculateAttackStrengthDetailed(order1, optimistic)
		defend2 := cr.adjudicator.calculateDefendStrengthDetailed(order2, !optimistic)

		attack2 := cr.adjudicator.calculateAttackStrengthDetailed(order2, optimistic)
		defend1 := cr.adjudicator.calculateDefendStrengthDetailed(order1, !optimistic)

		analysis.Strengths[order1.Source+"_attack"] = attack1
		analysis.Strengths[order1.Source+"_defend"] = defend1
		analysis.Strengths[order2.Source+"_attack"] = attack2
		analysis.Strengths[order2.Source+"_defend"] = defend2

		analysis.Resolution.Reasoning = append(analysis.Resolution.Reasoning,
			fmt.Sprintf("Head-to-head: %s attack %d vs %s defend %d",
				order1.Source, attack1.TotalStrength, order2.Source, defend2.TotalStrength))
		analysis.Resolution.Reasoning = append(analysis.Resolution.Reasoning,
			fmt.Sprintf("Head-to-head: %s attack %d vs %s defend %d",
				order2.Source, attack2.TotalStrength, order1.Source, defend1.TotalStrength))

		// Determine winner based on attack vs defend comparison
		order1Succeeds := attack1.TotalStrength > defend2.TotalStrength
		order2Succeeds := attack2.TotalStrength > defend1.TotalStrength

		if order1Succeeds && !order2Succeeds {
			analysis.Winner = order1
			analysis.Resolution.Decision = fmt.Sprintf("Head-to-head: %s -> %s succeeds", order1.Source, order1.Destination)
		} else if order2Succeeds && !order1Succeeds {
			analysis.Winner = order2
			analysis.Resolution.Decision = fmt.Sprintf("Head-to-head: %s -> %s succeeds", order2.Source, order2.Destination)
		} else {
			// Both succeed or both fail - standoff
			analysis.ConflictType = Standoff
			analysis.Resolution.IsStandoff = true
			analysis.Resolution.Decision = "Head-to-head standoff - both moves fail"
			analysis.Resolution.DATCRules = append(analysis.Resolution.DATCRules, "DATC 6.A.3: Head-to-head standoff")
		}
	}

	analysis.Explanation = analysis.Resolution.Decision
	return analysis
}
