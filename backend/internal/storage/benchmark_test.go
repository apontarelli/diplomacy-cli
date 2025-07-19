package storage

import (
	"testing"

	"diplomacy-cli/backend/internal/game"
)

func BenchmarkDeserializeGameState(b *testing.B) {
	// Create a complex game state for benchmarking
	gameState := createComplexBenchmarkGameState()

	// Serialize it once
	data, err := SerializeGameState(gameState)
	if err != nil {
		b.Fatalf("Failed to serialize: %v", err)
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, err := DeserializeGameState(data)
		if err != nil {
			b.Fatalf("Failed to deserialize: %v", err)
		}
	}
}

func BenchmarkSerializeGameState(b *testing.B) {
	gameState := createComplexBenchmarkGameState()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, err := SerializeGameState(gameState)
		if err != nil {
			b.Fatalf("Failed to serialize: %v", err)
		}
	}
}

func createComplexBenchmarkGameState() *game.GameState {
	board := game.NewBoard()

	// Add all 75 Diplomacy provinces (representative sample)
	provinces := []struct {
		name, shortCode, displayName string
		pType                        game.ProvinceType
		supplyCenter                 bool
		armyNeighbors                []string
		fleetNeighbors               []string
	}{
		{"vienna", "vie", "Vienna", game.Land, true, []string{"bud", "gal", "boh", "tyr", "tri"}, []string{}},
		{"budapest", "bud", "Budapest", game.Land, true, []string{"vie", "gal", "rum", "ser", "tri"}, []string{}},
		{"trieste", "tri", "Trieste", game.Land, true, []string{"vie", "bud", "ser", "alb", "adr", "ven"}, []string{"adr", "alb", "ven"}},
		{"london", "lon", "London", game.Land, true, []string{"yor", "wal"}, []string{"nth", "eng"}},
		{"liverpool", "lvp", "Liverpool", game.Land, true, []string{"yor", "wal", "cly", "edi"}, []string{"iri", "nao", "cly"}},
		{"edinburgh", "edi", "Edinburgh", game.Land, true, []string{"lvp", "cly", "yor"}, []string{"nth", "nwg", "cly"}},
		{"paris", "par", "Paris", game.Land, true, []string{"pic", "bur", "gas", "bre"}, []string{}},
		{"marseilles", "mar", "Marseilles", game.Land, true, []string{"gas", "bur", "pie", "spa"}, []string{"lyo", "spa", "pie"}},
		{"brest", "bre", "Brest", game.Land, true, []string{"par", "pic", "gas"}, []string{"eng", "mao", "gas"}},
		{"berlin", "ber", "Berlin", game.Land, true, []string{"pru", "sil", "mun", "kie"}, []string{"bal", "kie", "pru"}},
		{"munich", "mun", "Munich", game.Land, true, []string{"ber", "sil", "boh", "tyr", "bur", "ruh", "kie"}, []string{}},
		{"kiel", "kie", "Kiel", game.Land, true, []string{"ber", "mun", "ruh", "hol", "den"}, []string{"bal", "hel", "den", "hol"}},
		// Add some sea provinces
		{"english_channel", "eng", "English Channel", game.Sea, false, []string{}, []string{"lon", "wal", "bre", "pic", "bel", "nth"}},
		{"north_sea", "nth", "North Sea", game.Sea, false, []string{}, []string{"lon", "yor", "edi", "nwy", "ska", "den", "hel", "hol", "bel", "eng"}},
		{"adriatic_sea", "adr", "Adriatic Sea", game.Sea, false, []string{}, []string{"tri", "alb", "apu", "ven", "ion"}},
	}

	for _, p := range provinces {
		province := &game.Province{
			Name:           p.name,
			ShortCode:      p.shortCode,
			DisplayName:    p.displayName,
			Type:           p.pType,
			SupplyCenter:   p.supplyCenter,
			CoastNeighbors: make(map[string][]string),
			ArmyNeighbors:  p.armyNeighbors,
			FleetNeighbors: p.fleetNeighbors,
		}
		board.AddProvince(province)
	}

	// Add units for all 7 nations
	units := []*game.Unit{
		{Type: game.Army, Owner: game.Austria, Province: "vienna", Coast: ""},
		{Type: game.Army, Owner: game.Austria, Province: "budapest", Coast: ""},
		{Type: game.Fleet, Owner: game.Austria, Province: "trieste", Coast: ""},
		{Type: game.Fleet, Owner: game.England, Province: "london", Coast: ""},
		{Type: game.Army, Owner: game.England, Province: "liverpool", Coast: ""},
		{Type: game.Fleet, Owner: game.England, Province: "edinburgh", Coast: ""},
		{Type: game.Army, Owner: game.France, Province: "paris", Coast: ""},
		{Type: game.Army, Owner: game.France, Province: "marseilles", Coast: ""},
		{Type: game.Fleet, Owner: game.France, Province: "brest", Coast: ""},
		{Type: game.Army, Owner: game.Germany, Province: "berlin", Coast: ""},
		{Type: game.Army, Owner: game.Germany, Province: "munich", Coast: ""},
		{Type: game.Fleet, Owner: game.Germany, Province: "kiel", Coast: ""},
	}

	for _, unit := range units {
		board.PlaceUnit(unit)
	}

	// Create game state with complex data
	gameState := game.NewGameState(board, game.SpringMovement, 1901)

	// Add raw orders for all nations
	nations := []game.Nation{game.Austria, game.England, game.France, game.Germany, game.Italy, game.Russia, game.Turkey}
	supplyCenterMap := map[game.Nation][]string{
		game.Austria: {"vienna"},
		game.England: {"london"},
		game.France:  {"paris"},
		game.Germany: {"berlin"},
		game.Italy:   {}, // No supply centers for simplicity
		game.Russia:  {}, // No supply centers for simplicity
		game.Turkey:  {}, // No supply centers for simplicity
	}

	for _, nation := range nations {
		gameState.RawOrders[nation] = []string{
			"A vie-bud",
			"F tri S A vie-bud",
			"A boh-mun",
		}
		if centers, exists := supplyCenterMap[nation]; exists {
			gameState.SupplyCenters[nation] = centers
		}
	}

	// Add some dislodged units
	dislodgedProvinces := []string{"vienna", "london", "paris"}
	for i, nation := range nations[:3] {
		unit := &game.Unit{
			Type:      game.Army,
			Owner:     nation,
			Province:  dislodgedProvinces[i],
			Coast:     "",
			Dislodged: true,
		}
		gameState.DislodgedUnits = append(gameState.DislodgedUnits, unit)
	}

	return gameState
}
