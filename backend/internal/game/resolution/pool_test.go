package resolution

import (
	"diplomacy-cli/backend/internal/game"
	"runtime"
	"testing"
)

func TestObjectPool_Order(t *testing.T) {
	pool := NewObjectPool()

	// Get an order from the pool
	order1 := pool.GetOrder()
	if order1 == nil {
		t.Fatal("Expected non-nil order from pool")
	}

	// Modify the order
	order1.Unit = "A Berlin"
	order1.Type = Move
	order1.Source = "berlin"
	order1.Destination = "munich"
	order1.Owner = "Germany"

	// Return it to the pool
	pool.PutOrder(order1)

	// Get another order - should be the same instance but reset
	order2 := pool.GetOrder()
	if order2 != order1 {
		t.Error("Expected to get the same order instance from pool")
	}

	// Verify it was reset
	if order2.Unit != "" || order2.Source != "" || order2.Destination != "" {
		t.Error("Order was not properly reset when returned from pool")
	}
}

func TestObjectPool_UnitOutcome(t *testing.T) {
	pool := NewObjectPool()

	outcome1 := pool.GetUnitOutcome()
	if outcome1 == nil {
		t.Fatal("Expected non-nil unit outcome from pool")
	}

	// Modify the outcome
	outcome1.fromProvince = "berlin"
	toProvince := "munich"
	outcome1.toProvince = &toProvince
	outcome1.strength = 5

	// Return to pool
	pool.PutUnitOutcome(outcome1)

	// Get another outcome
	outcome2 := pool.GetUnitOutcome()
	if outcome2 != outcome1 {
		t.Error("Expected to get the same outcome instance from pool")
	}

	// Verify it was reset
	if outcome2.fromProvince != "" || outcome2.toProvince != nil || outcome2.strength != 0 {
		t.Error("UnitOutcome was not properly reset when returned from pool")
	}
}

func TestObjectPool_Conflict(t *testing.T) {
	pool := NewObjectPool()

	conflict1 := pool.GetConflict()
	if conflict1 == nil {
		t.Fatal("Expected non-nil conflict from pool")
	}

	// Modify the conflict
	conflict1.province = "berlin"
	conflict1.competitors = append(conflict1.competitors, UnitID{Type: game.Army, Owner: game.Germany, Province: "berlin"})
	conflict1.strengths[UnitID{Type: game.Army, Owner: game.Germany, Province: "berlin"}] = 1

	// Return to pool
	pool.PutConflict(conflict1)

	// Get another conflict
	conflict2 := pool.GetConflict()
	if conflict2 != conflict1 {
		t.Error("Expected to get the same conflict instance from pool")
	}

	// Verify it was reset
	if conflict2.province != "" || len(conflict2.competitors) != 0 || len(conflict2.strengths) != 0 {
		t.Error("Conflict was not properly reset when returned from pool")
	}
}

func TestObjectPool_Slices(t *testing.T) {
	pool := NewObjectPool()

	// Test UnitID slice
	slice1 := pool.GetUnitIDSlice()
	if slice1 == nil {
		t.Fatal("Expected non-nil UnitID slice from pool")
	}

	slice1 = append(slice1, UnitID{Type: game.Army, Owner: game.Germany, Province: "berlin"})
	pool.PutUnitIDSlice(slice1)

	slice2 := pool.GetUnitIDSlice()
	if len(slice2) != 0 {
		t.Error("UnitID slice was not properly reset")
	}

	// Test string slice
	strSlice1 := pool.GetStringSlice()
	strSlice1 = append(strSlice1, "berlin", "munich")
	pool.PutStringSlice(strSlice1)

	strSlice2 := pool.GetStringSlice()
	if len(strSlice2) != 0 {
		t.Error("String slice was not properly reset")
	}
}

func TestPooledResolutionResult(t *testing.T) {
	pool := NewObjectPool()

	result := pool.NewPooledResolutionResult(1, game.SpringMovement)
	if result == nil {
		t.Fatal("Expected non-nil pooled resolution result")
	}

	// Test adding pooled unit outcome
	unitID := UnitID{Type: game.Army, Owner: game.Germany, Province: "berlin"}
	unit := game.Unit{Type: game.Army, Owner: game.Germany, Province: "berlin"}
	order := game.Order{From: "berlin", To: "munich", Type: "move"}

	result.AddPooledUnitOutcome(unitID, unit, order, UnitMoved, "berlin", nil, nil, "successful move", 1, 0)

	outcomes := result.UnitOutcomes()
	if len(outcomes) != 1 {
		t.Error("Expected 1 unit outcome")
	}

	outcome, exists := outcomes[unitID]
	if !exists {
		t.Error("Expected unit outcome to exist")
	}

	if outcome.FromProvince() != "berlin" {
		t.Error("Expected from province to be berlin")
	}
}

// BenchmarkObjectPool_Order benchmarks order allocation with and without pooling
func BenchmarkObjectPool_Order(b *testing.B) {
	pool := NewObjectPool()

	b.Run("WithPool", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			order := pool.GetOrder()
			order.Unit = "A Berlin"
			order.Type = Move
			order.Source = "berlin"
			order.Destination = "munich"
			pool.PutOrder(order)
		}
	})

	b.Run("WithoutPool", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			order := &Order{
				Unit:        "A Berlin",
				Type:        Move,
				Source:      "berlin",
				Destination: "munich",
			}
			_ = order // Prevent optimization
		}
	})
}

// BenchmarkObjectPool_ResolutionResult benchmarks resolution result creation
func BenchmarkObjectPool_ResolutionResult(b *testing.B) {
	pool := NewObjectPool()

	b.Run("WithPool", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			result := pool.NewPooledResolutionResult(1, game.SpringMovement)

			// Add some outcomes using pooled objects
			unitID := UnitID{Type: game.Army, Owner: game.Germany, Province: "berlin"}
			unit := game.Unit{Type: game.Army, Owner: game.Germany, Province: "berlin"}
			order := game.Order{From: "berlin", To: "munich", Type: "move"}

			result.AddPooledUnitOutcome(unitID, unit, order, UnitMoved, "berlin", nil, nil, "successful move", 1, 0)
			result.Release()
		}
	})

	b.Run("WithoutPool", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			result := NewResolutionResult(1, game.SpringMovement)

			// Add outcomes without pooling
			unitID := UnitID{Type: game.Army, Owner: game.Germany, Province: "berlin"}
			unit := game.Unit{Type: game.Army, Owner: game.Germany, Province: "berlin"}
			order := game.Order{From: "berlin", To: "munich", Type: "move"}

			outcome := UnitOutcome{
				unit:         unit,
				orderGiven:   order,
				finalStatus:  UnitMoved,
				fromProvince: "berlin",
				reason:       "successful move",
				strength:     1,
				supportCount: 0,
			}

			result.AddUnitOutcome(unitID, outcome)
		}
	})
}

// BenchmarkMemoryPressure tests pool behavior under memory pressure
func BenchmarkMemoryPressure(b *testing.B) {
	pool := NewObjectPool()

	b.Run("HighAllocationWithPool", func(b *testing.B) {
		var m1, m2 runtime.MemStats
		runtime.GC()
		runtime.ReadMemStats(&m1)

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			// Simulate high allocation scenario
			orders := make([]*Order, 100)
			for j := 0; j < 100; j++ {
				orders[j] = pool.GetOrder()
				orders[j].Unit = "A Berlin"
				orders[j].Type = Move
			}

			// Return all to pool
			for j := 0; j < 100; j++ {
				pool.PutOrder(orders[j])
			}
		}
		b.StopTimer()

		runtime.GC()
		runtime.ReadMemStats(&m2)
		b.Logf("Memory allocated: %d bytes, GC cycles: %d", m2.TotalAlloc-m1.TotalAlloc, m2.NumGC-m1.NumGC)
	})

	b.Run("HighAllocationWithoutPool", func(b *testing.B) {
		var m1, m2 runtime.MemStats
		runtime.GC()
		runtime.ReadMemStats(&m1)

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			// Simulate high allocation scenario without pooling
			orders := make([]*Order, 100)
			for j := 0; j < 100; j++ {
				orders[j] = &Order{
					Unit: "A Berlin",
					Type: Move,
				}
			}
			_ = orders // Prevent optimization
		}
		b.StopTimer()

		runtime.GC()
		runtime.ReadMemStats(&m2)
		b.Logf("Memory allocated: %d bytes, GC cycles: %d", m2.TotalAlloc-m1.TotalAlloc, m2.NumGC-m1.NumGC)
	})
}
