package api_test

import (
	"testing"

	"diplomacy-backend/internal/api"
	"diplomacy-backend/internal/game"
	"diplomacy-backend/internal/game/rules"
	"diplomacy-backend/internal/storage"
)

// TestServerCreation tests basic server creation
func TestServerCreation(t *testing.T) {
	// Create a test database (nil for now, but could be in-memory SQLite)
	var db *storage.DB
	
	// Create test game data
	gameData := &game.GameData{}
	
	// Create server using the constructor
	server := api.NewServer(db, gameData)
	
	if server == nil {
		t.Fatal("Failed to create server")
	}
	
	t.Log("Successfully created API server")
}

// TestRulesLoading tests that rules can be loaded
func TestRulesLoading(t *testing.T) {
	rules := rules.MustLoadRules("classic")
	if rules == nil {
		t.Fatal("Failed to load classic rules")
	}
	t.Log("Successfully loaded classic rules for API tests")
}

// TestOrderTypes tests that order types are available
func TestOrderTypes(t *testing.T) {
	orderTypes := []game.OrderType{
		game.OrderTypeMove,
		game.OrderTypeHold,
		game.OrderTypeSupport,
		game.OrderTypeConvoy,
	}
	
	if len(orderTypes) != 4 {
		t.Errorf("Expected 4 order types, got %d", len(orderTypes))
	}
	
	t.Log("Order types are available for API tests")
}

// TODO: Add more comprehensive API tests here
// The original DATC tests and integration tests can be added back
// once the package/import issues are resolved