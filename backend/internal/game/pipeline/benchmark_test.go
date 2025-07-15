package pipeline

import (
	"testing"

	"diplomacy-cli/backend/internal/game"
	"diplomacy-cli/backend/internal/game/loader"
)

func BenchmarkDATCTests(b *testing.B) {
	for i := 0; i < b.N; i++ {
		// Run a representative sample of DATC tests
		t := &testing.T{}
		TestDATCD1_SimplestSupportToHold(t)
		TestDATCE1_DislogedUnitHasNoEffectOnAttackersArea(t)
		TestDATCF1_NoConvoyInCoastalAreas(t)
		TestDATCH1_NoSupportsDuringRetreat(t)
		TestDATCI1_TooManyBuildOrders(t)
	}
}

func BenchmarkTurnProcessing(b *testing.B) {
	// Load test data
	mapLoader := loader.NewJSONLoader("../../../../backend/data/classic")
	board, err := mapLoader.LoadBoard()
	if err != nil {
		b.Fatalf("Failed to load board: %v", err)
	}

	gameState := game.NewGameState(board, game.SpringMovement, 1901)

	// Add test units
	gameState.Board.Units["london"] = &game.Unit{
		Type:     game.Fleet,
		Owner:    game.England,
		Province: "london",
	}
	gameState.Board.Units["wales"] = &game.Unit{
		Type:     game.Army,
		Owner:    game.England,
		Province: "wales",
	}

	gameState.RawOrders[game.England] = []string{
		"f london - english_channel",
		"wales s london - english_channel",
	}

	processor := NewTurnProcessor()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		processor.ProcessTurn(gameState)
	}
}
