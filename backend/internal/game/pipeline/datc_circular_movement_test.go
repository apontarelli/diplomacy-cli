package pipeline

import (
	"diplomacy-cli/backend/internal/game"
	"testing"
)

// DATC 6.C - CIRCULAR MOVEMENT Test Cases
// Tests for unit swaps and circular dependencies

func TestDATCC1_ThreeArmyCircularMovement(t *testing.T) {
	// 6.C.1 - THREE ARMY CIRCULAR MOVEMENT
	// Three units can change place, even in spring 1901.
	gameState := CreateDATCGameState(t)

	// Add units based on orders
	gameState.Board.PlaceUnit(&game.Unit{Type: game.Fleet, Owner: game.Turkey, Province: "ankara"})
	gameState.Board.PlaceUnit(&game.Unit{Type: game.Army, Owner: game.Turkey, Province: "constantinople"})
	gameState.Board.PlaceUnit(&game.Unit{Type: game.Army, Owner: game.Turkey, Province: "smyrna"})

	// Add orders
	gameState.AddRawOrder(game.Turkey, "F Ankara - Constantinople")
	gameState.AddRawOrder(game.Turkey, "A Constantinople - Smyrna")
	gameState.AddRawOrder(game.Turkey, "A Smyrna - Ankara")

	// Process orders
	ProcessDATCTest(t, gameState)

	// Expected: All three units will move
	// Verify units moved to their destinations
	if gameState.Board.GetUnit("constantinople") == nil || gameState.Board.GetUnit("constantinople").Owner != game.Turkey {
		t.Errorf("❌ 6.C.1: Fleet should have moved to Constantinople")
	}
	if gameState.Board.GetUnit("smyrna") == nil || gameState.Board.GetUnit("smyrna").Owner != game.Turkey {
		t.Errorf("❌ 6.C.1: Army should have moved to Smyrna")
	}
	if gameState.Board.GetUnit("ankara") == nil || gameState.Board.GetUnit("ankara").Owner != game.Turkey {
		t.Errorf("❌ 6.C.1: Army should have moved to Ankara")
	}
	t.Logf("✅ 6.C.1: Three army circular movement succeeded")
}

func TestDATCC2_ThreeArmyCircularMovementWithSupport(t *testing.T) {
	// 6.C.2 - THREE ARMY CIRCULAR MOVEMENT WITH SUPPORT
	// Three units can change place, even when one gets support.
	gameState := CreateDATCGameState(t)

	// Add units based on orders
	gameState.Board.PlaceUnit(&game.Unit{Type: game.Fleet, Owner: game.Turkey, Province: "ankara"})
	gameState.Board.PlaceUnit(&game.Unit{Type: game.Army, Owner: game.Turkey, Province: "constantinople"})
	gameState.Board.PlaceUnit(&game.Unit{Type: game.Army, Owner: game.Turkey, Province: "smyrna"})
	gameState.Board.PlaceUnit(&game.Unit{Type: game.Army, Owner: game.Turkey, Province: "bulgaria"})

	// Add orders
	gameState.AddRawOrder(game.Turkey, "F Ankara - Constantinople")
	gameState.AddRawOrder(game.Turkey, "A Constantinople - Smyrna")
	gameState.AddRawOrder(game.Turkey, "A Smyrna - Ankara")
	gameState.AddRawOrder(game.Turkey, "a bulgaria s ankara - constantinople")

	// Process orders
	ProcessDATCTest(t, gameState)

	// Expected: All three units will move (support doesn't prevent circular movement)
	if gameState.Board.GetUnit("constantinople") == nil || gameState.Board.GetUnit("constantinople").Owner != game.Turkey {
		t.Errorf("❌ 6.C.2: Fleet should have moved to Constantinople")
	}
	if gameState.Board.GetUnit("smyrna") == nil || gameState.Board.GetUnit("smyrna").Owner != game.Turkey {
		t.Errorf("❌ 6.C.2: Army should have moved to Smyrna")
	}
	if gameState.Board.GetUnit("ankara") == nil || gameState.Board.GetUnit("ankara").Owner != game.Turkey {
		t.Errorf("❌ 6.C.2: Army should have moved to Ankara")
	}
	t.Logf("✅ 6.C.2: Three army circular movement with support succeeded")
}

func TestDATCC3_DisruptedThreeArmyCircularMovement(t *testing.T) {
	// 6.C.3 - A DISRUPTED THREE ARMY CIRCULAR MOVEMENT
	// When one of the units bounces, the whole circular movement will hold.
	gameState := CreateDATCGameState(t)

	// Add units based on orders
	gameState.Board.PlaceUnit(&game.Unit{Type: game.Fleet, Owner: game.Turkey, Province: "ankara"})
	gameState.Board.PlaceUnit(&game.Unit{Type: game.Army, Owner: game.Turkey, Province: "constantinople"})
	gameState.Board.PlaceUnit(&game.Unit{Type: game.Army, Owner: game.Turkey, Province: "smyrna"})
	gameState.Board.PlaceUnit(&game.Unit{Type: game.Army, Owner: game.Turkey, Province: "bulgaria"})

	// Add orders
	gameState.AddRawOrder(game.Turkey, "F Ankara - Constantinople")
	gameState.AddRawOrder(game.Turkey, "A Constantinople - Smyrna")
	gameState.AddRawOrder(game.Turkey, "A Smyrna - Ankara")
	gameState.AddRawOrder(game.Turkey, "A Bulgaria - Constantinople")

	// Process orders
	ProcessDATCTest(t, gameState)

	// Expected: Every unit will keep its place (circular movement disrupted by bounce)
	if gameState.Board.GetUnit("ankara") == nil || gameState.Board.GetUnit("ankara").Type != game.Fleet {
		t.Errorf("❌ 6.C.3: Fleet should have stayed at Ankara")
	}
	if gameState.Board.GetUnit("constantinople") == nil || gameState.Board.GetUnit("constantinople").Type != game.Army {
		t.Errorf("❌ 6.C.3: Army should have stayed at Constantinople")
	}
	if gameState.Board.GetUnit("smyrna") == nil || gameState.Board.GetUnit("smyrna").Type != game.Army {
		t.Errorf("❌ 6.C.3: Army should have stayed at Smyrna")
	}
	if gameState.Board.GetUnit("bulgaria") == nil || gameState.Board.GetUnit("bulgaria").Type != game.Army {
		t.Errorf("❌ 6.C.3: Army should have stayed at Bulgaria")
	}
	t.Logf("✅ 6.C.3: Disrupted circular movement correctly held all units")
}

func TestDATCC4_CircularMovementWithAttackedConvoy(t *testing.T) {
	// 6.C.4 - A CIRCULAR MOVEMENT WITH ATTACKED CONVOY
	// When the circular movement contains an attacked convoy, the circular movement succeeds.
	gameState := CreateDATCGameState(t)

	// Add units based on orders
	gameState.Board.PlaceUnit(&game.Unit{Type: game.Army, Owner: game.Austria, Province: "trieste"})
	gameState.Board.PlaceUnit(&game.Unit{Type: game.Army, Owner: game.Austria, Province: "serbia"})
	gameState.Board.PlaceUnit(&game.Unit{Type: game.Army, Owner: game.Turkey, Province: "bulgaria"})
	gameState.Board.PlaceUnit(&game.Unit{Type: game.Fleet, Owner: game.Turkey, Province: "aegean_sea"})
	gameState.Board.PlaceUnit(&game.Unit{Type: game.Fleet, Owner: game.Turkey, Province: "ionian_sea"})
	gameState.Board.PlaceUnit(&game.Unit{Type: game.Fleet, Owner: game.Turkey, Province: "adriatic_sea"})
	gameState.Board.PlaceUnit(&game.Unit{Type: game.Fleet, Owner: game.Italy, Province: "naples"})

	// Add orders
	gameState.AddRawOrder(game.Austria, "A Trieste - Serbia")
	gameState.AddRawOrder(game.Austria, "A Serbia - Bulgaria")
	gameState.AddRawOrder(game.Turkey, "A Bulgaria - Trieste")
	gameState.AddRawOrder(game.Turkey, "F Aegean Sea Convoys A Bulgaria - Trieste")
	gameState.AddRawOrder(game.Turkey, "F Ionian Sea Convoys A Bulgaria - Trieste")
	gameState.AddRawOrder(game.Turkey, "F Adriatic Sea Convoys A Bulgaria - Trieste")
	gameState.AddRawOrder(game.Italy, "F Naples - Ionian Sea")

	// Process orders
	result := ProcessDATCTest(t, gameState)

	// Expected: Circular movement succeeds despite convoy attack
	// The convoy attack should be resolved before circular movement calculation
	finalState := result.NewGameState
	if finalState == nil {
		t.Fatalf("❌ 6.C.4: Processing failed: %v", result.ProcessError)
	}

	bulgUnit := finalState.Board.GetUnit("bulgaria")
	if bulgUnit == nil || bulgUnit.Owner != game.Austria {
		t.Errorf("❌ 6.C.4: Austrian army should have moved to Bulgaria")
	}

	triesteUnit := finalState.Board.GetUnit("trieste")
	if triesteUnit == nil || triesteUnit.Owner != game.Turkey {
		t.Errorf("❌ 6.C.4: Turkish army should have moved to Trieste")
	}
	t.Logf("✅ 6.C.4: Circular movement with attacked convoy succeeded")
}

func TestDATCC5_DisruptedCircularMovementDueToDislodgedConvoy(t *testing.T) {
	// 6.C.5 - A DISRUPTED CIRCULAR MOVEMENT DUE TO DISLODGED CONVOY
	// When the circular movement contains a convoy, the circular movement is disrupted when the convoying fleet is dislodged.
	gameState := CreateDATCGameState(t)

	// Add units based on orders
	gameState.Board.PlaceUnit(&game.Unit{Type: game.Army, Owner: game.Austria, Province: "trieste"})
	gameState.Board.PlaceUnit(&game.Unit{Type: game.Army, Owner: game.Austria, Province: "serbia"})
	gameState.Board.PlaceUnit(&game.Unit{Type: game.Army, Owner: game.Turkey, Province: "bulgaria"})
	gameState.Board.PlaceUnit(&game.Unit{Type: game.Fleet, Owner: game.Turkey, Province: "aegean_sea"})
	gameState.Board.PlaceUnit(&game.Unit{Type: game.Fleet, Owner: game.Turkey, Province: "ionian_sea"})
	gameState.Board.PlaceUnit(&game.Unit{Type: game.Fleet, Owner: game.Turkey, Province: "adriatic_sea"})
	gameState.Board.PlaceUnit(&game.Unit{Type: game.Fleet, Owner: game.Italy, Province: "naples"})
	gameState.Board.PlaceUnit(&game.Unit{Type: game.Fleet, Owner: game.Italy, Province: "tunis"})

	// Add orders
	gameState.AddRawOrder(game.Austria, "A Trieste - Serbia")
	gameState.AddRawOrder(game.Austria, "A Serbia - Bulgaria")
	gameState.AddRawOrder(game.Turkey, "A Bulgaria - Trieste")
	gameState.AddRawOrder(game.Turkey, "F Aegean Sea Convoys A Bulgaria - Trieste")
	gameState.AddRawOrder(game.Turkey, "F Ionian Sea Convoys A Bulgaria - Trieste")
	gameState.AddRawOrder(game.Turkey, "F Adriatic Sea Convoys A Bulgaria - Trieste")
	gameState.AddRawOrder(game.Italy, "F Naples - Ionian Sea")
	gameState.AddRawOrder(game.Italy, "f tunis s naples - ionian_sea")

	// Process orders
	ProcessDATCTest(t, gameState)

	// Expected: Circular movement disrupted due to dislodged convoy fleet
	// All units should stay in place
	if gameState.Board.GetUnit("trieste") == nil || gameState.Board.GetUnit("trieste").Owner != game.Austria {
		t.Errorf("❌ 6.C.5: Austrian army should have stayed at Trieste")
	}
	if gameState.Board.GetUnit("serbia") == nil || gameState.Board.GetUnit("serbia").Owner != game.Austria {
		t.Errorf("❌ 6.C.5: Austrian army should have stayed at Serbia")
	}
	if gameState.Board.GetUnit("bulgaria") == nil || gameState.Board.GetUnit("bulgaria").Owner != game.Turkey {
		t.Errorf("❌ 6.C.5: Turkish army should have stayed at Bulgaria")
	}
	t.Logf("✅ 6.C.5: Circular movement correctly disrupted by dislodged convoy")
}

func TestDATCC6_TwoArmiesWithTwoConvoys(t *testing.T) {
	// 6.C.6 - TWO ARMIES WITH TWO CONVOYS
	// Two armies can swap places even when they are not adjacent.
	gameState := CreateDATCGameState(t)

	// Add units based on orders
	gameState.Board.PlaceUnit(&game.Unit{Type: game.Fleet, Owner: game.England, Province: "north_sea"})
	gameState.Board.PlaceUnit(&game.Unit{Type: game.Army, Owner: game.England, Province: "london"})
	gameState.Board.PlaceUnit(&game.Unit{Type: game.Fleet, Owner: game.France, Province: "english_channel"})
	gameState.Board.PlaceUnit(&game.Unit{Type: game.Army, Owner: game.France, Province: "belgium"})

	// Add orders
	gameState.AddRawOrder(game.England, "F North Sea Convoys A London - Belgium")
	gameState.AddRawOrder(game.England, "A London - Belgium")
	gameState.AddRawOrder(game.France, "F English Channel Convoys A Belgium - London")
	gameState.AddRawOrder(game.France, "A Belgium - London")

	// Process orders
	result := ProcessDATCTest(t, gameState)

	// Expected: Both convoys should succeed (unit swap via convoy)
	// Debug: Check what units are actually on the board
	finalState := result.NewGameState
	if finalState == nil {
		t.Fatalf("❌ 6.C.6: Processing failed: %v", result.ProcessError)
	}

	t.Logf("🔍 6.C.6 Debug - Units after processing:")
	for province, unit := range finalState.Board.Units {
		if unit != nil {
			t.Logf("  %s: %s %s", province, unit.Owner, unit.Type)
		}
	}

	if finalState.Board.GetUnit("belgium") == nil || finalState.Board.GetUnit("belgium").Owner != game.England {
		t.Errorf("❌ 6.C.6: English army should have moved to Belgium")
	}
	if finalState.Board.GetUnit("london") == nil || finalState.Board.GetUnit("london").Owner != game.France {
		t.Errorf("❌ 6.C.6: French army should have moved to London")
	}
	t.Logf("✅ 6.C.6: Two armies with two convoys succeeded")
}

func TestDATCC7_DisruptedUnitSwap(t *testing.T) {
	// 6.C.7 - DISRUPTED UNIT SWAP
	// If in a swap one of the unit bounces, then the swap fails.
	gameState := CreateDATCGameState(t)

	// Add units based on orders
	gameState.Board.PlaceUnit(&game.Unit{Type: game.Fleet, Owner: game.England, Province: "north_sea"})
	gameState.Board.PlaceUnit(&game.Unit{Type: game.Army, Owner: game.England, Province: "london"})
	gameState.Board.PlaceUnit(&game.Unit{Type: game.Fleet, Owner: game.France, Province: "english_channel"})
	gameState.Board.PlaceUnit(&game.Unit{Type: game.Army, Owner: game.France, Province: "belgium"})
	gameState.Board.PlaceUnit(&game.Unit{Type: game.Army, Owner: game.France, Province: "burgundy"})

	// Add orders
	gameState.AddRawOrder(game.England, "F North Sea Convoys A London - Belgium")
	gameState.AddRawOrder(game.England, "A London - Belgium")
	gameState.AddRawOrder(game.France, "F English Channel Convoys A Belgium - London")
	gameState.AddRawOrder(game.France, "A Belgium - London")
	gameState.AddRawOrder(game.France, "A Burgundy - Belgium")

	// Process orders
	ProcessDATCTest(t, gameState)

	// Expected: None of the units will succeed to move (swap disrupted by bounce)
	if gameState.Board.GetUnit("london") == nil || gameState.Board.GetUnit("london").Owner != game.England {
		t.Errorf("❌ 6.C.7: English army should have stayed at London")
	}
	if gameState.Board.GetUnit("belgium") == nil || gameState.Board.GetUnit("belgium").Owner != game.France {
		t.Errorf("❌ 6.C.7: French army should have stayed at Belgium")
	}
	if gameState.Board.GetUnit("burgundy") == nil || gameState.Board.GetUnit("burgundy").Owner != game.France {
		t.Errorf("❌ 6.C.7: French army should have stayed at Burgundy")
	}
	t.Logf("✅ 6.C.7: Disrupted unit swap correctly failed")
}

func TestDATCC8_NoSelfDislodgementInDisruptedCircularMovement(t *testing.T) {
	// 6.C.8 - NO SELF DISLODGEMENT IN DISRUPTED CIRCULAR MOVEMENT
	// Self dislodgement is prohibited as usual in circular movement.
	gameState := CreateDATCGameState(t)

	// Add units based on orders
	gameState.Board.PlaceUnit(&game.Unit{Type: game.Fleet, Owner: game.Turkey, Province: "constantinople"})
	gameState.Board.PlaceUnit(&game.Unit{Type: game.Army, Owner: game.Turkey, Province: "bulgaria"})
	gameState.Board.PlaceUnit(&game.Unit{Type: game.Army, Owner: game.Turkey, Province: "smyrna"})
	gameState.Board.PlaceUnit(&game.Unit{Type: game.Fleet, Owner: game.Russia, Province: "black_sea"})
	gameState.Board.PlaceUnit(&game.Unit{Type: game.Army, Owner: game.Austria, Province: "serbia"})

	// Add orders
	gameState.AddRawOrder(game.Turkey, "F Constantinople - Black Sea")
	gameState.AddRawOrder(game.Turkey, "A Bulgaria - Constantinople")
	gameState.AddRawOrder(game.Turkey, "a smyrna s bulgaria - constantinople")
	gameState.AddRawOrder(game.Russia, "F Black Sea - Bulgaria/ec")
	gameState.AddRawOrder(game.Austria, "A Serbia - Bulgaria")

	// Process orders
	ProcessDATCTest(t, gameState)

	// Expected: None of the units will succeed to move (self-dislodgement prevented)
	if gameState.Board.GetUnit("constantinople") == nil || gameState.Board.GetUnit("constantinople").Owner != game.Turkey {
		t.Errorf("❌ 6.C.8: Turkish fleet should have stayed at Constantinople")
	}
	if gameState.Board.GetUnit("bulgaria") == nil || gameState.Board.GetUnit("bulgaria").Owner != game.Turkey {
		t.Errorf("❌ 6.C.8: Turkish army should have stayed at Bulgaria")
	}
	if gameState.Board.GetUnit("black_sea") == nil || gameState.Board.GetUnit("black_sea").Owner != game.Russia {
		t.Errorf("❌ 6.C.8: Russian fleet should have stayed at Black Sea")
	}
	t.Logf("✅ 6.C.8: Self-dislodgement correctly prevented in circular movement")
}

func TestDATCC9_NoHelpInDislodgementOfOwnUnitInDisruptedCircularMovement(t *testing.T) {
	// 6.C.9 - NO HELP IN DISLODGEMENT OF OWN UNIT IN DISRUPTED CIRCULAR MOVEMENT
	// Helping to dislodge your own unit is prohibited as usual in circular movement.
	gameState := CreateDATCGameState(t)

	// Add units based on orders
	gameState.Board.PlaceUnit(&game.Unit{Type: game.Fleet, Owner: game.Turkey, Province: "constantinople"})
	gameState.Board.PlaceUnit(&game.Unit{Type: game.Army, Owner: game.Turkey, Province: "smyrna"})
	gameState.Board.PlaceUnit(&game.Unit{Type: game.Fleet, Owner: game.Russia, Province: "black_sea"})
	gameState.Board.PlaceUnit(&game.Unit{Type: game.Army, Owner: game.Austria, Province: "serbia"})
	gameState.Board.PlaceUnit(&game.Unit{Type: game.Army, Owner: game.Russia, Province: "bulgaria"})

	// Add orders
	gameState.AddRawOrder(game.Turkey, "F Constantinople - Black Sea")
	gameState.AddRawOrder(game.Turkey, "a smyrna s bulgaria - constantinople")
	gameState.AddRawOrder(game.Russia, "F Black Sea - Bulgaria/ec")
	gameState.AddRawOrder(game.Austria, "A Serbia - Bulgaria")
	gameState.AddRawOrder(game.Russia, "A Bulgaria - Constantinople")

	// Process orders
	result := ProcessDATCTest(t, gameState)

	// Expected: None of the units will succeed to move (helping own dislodgement prevented)
	if result.ProcessError != nil {
		t.Fatalf("❌ 6.C.9: ProcessDATCTest failed with error: %v", result.ProcessError)
	}
	finalState := result.NewGameState
	if finalState.Board.GetUnit("constantinople") == nil || finalState.Board.GetUnit("constantinople").Owner != game.Turkey {
		t.Errorf("❌ 6.C.9: Turkish fleet should have stayed at Constantinople")
	}
	if finalState.Board.GetUnit("bulgaria") == nil || finalState.Board.GetUnit("bulgaria").Owner != game.Russia {
		t.Errorf("❌ 6.C.9: Russian army should have stayed at Bulgaria")
	}
	if finalState.Board.GetUnit("black_sea") == nil || finalState.Board.GetUnit("black_sea").Owner != game.Russia {
		t.Errorf("❌ 6.C.9: Russian fleet should have stayed at Black Sea")
	}
	t.Logf("✅ 6.C.9: Help in dislodgement of own unit correctly prevented")
}
