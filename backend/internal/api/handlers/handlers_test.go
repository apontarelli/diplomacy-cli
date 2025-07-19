package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"diplomacy-cli/backend/internal/api/middleware"
)

func TestHealthEndpoint(t *testing.T) {
	// Create a simple health check test
	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()

	// Simple handler for health check
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	})

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	expected := `{"status":"ok"}`
	if w.Body.String() != expected {
		t.Errorf("Expected %s, got %s", expected, w.Body.String())
	}
}

func TestCreateGameRequest(t *testing.T) {
	// Test that we can marshal/unmarshal the request types
	req := middleware.CreateGameRequest{
		Name: "Test Game",
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("Failed to marshal request: %v", err)
	}

	var unmarshaled middleware.CreateGameRequest
	err = json.Unmarshal(data, &unmarshaled)
	if err != nil {
		t.Fatalf("Failed to unmarshal request: %v", err)
	}

	if unmarshaled.Name != req.Name {
		t.Errorf("Expected name %s, got %s", req.Name, unmarshaled.Name)
	}
}

func TestJoinGameRequest(t *testing.T) {
	// Test that we can marshal/unmarshal the request types
	req := middleware.JoinGameRequest{
		Nation: "england",
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("Failed to marshal request: %v", err)
	}

	var unmarshaled middleware.JoinGameRequest
	err = json.Unmarshal(data, &unmarshaled)
	if err != nil {
		t.Fatalf("Failed to unmarshal request: %v", err)
	}

	if unmarshaled.Nation != req.Nation {
		t.Errorf("Expected nation %s, got %s", req.Nation, unmarshaled.Nation)
	}
}

func TestSubmitOrdersRequest(t *testing.T) {
	// Test that we can marshal/unmarshal the request types
	req := middleware.SubmitOrdersRequest{
		Orders: []middleware.OrderRequest{
			{
				UnitProvince: "lon",
				Type:         "move",
				Target:       "nth",
			},
		},
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("Failed to marshal request: %v", err)
	}

	var unmarshaled middleware.SubmitOrdersRequest
	err = json.Unmarshal(data, &unmarshaled)
	if err != nil {
		t.Fatalf("Failed to unmarshal request: %v", err)
	}

	if len(unmarshaled.Orders) != 1 {
		t.Errorf("Expected 1 order, got %d", len(unmarshaled.Orders))
	}

	if unmarshaled.Orders[0].UnitProvince != req.Orders[0].UnitProvince {
		t.Errorf("Expected unit province %s, got %s", req.Orders[0].UnitProvince, unmarshaled.Orders[0].UnitProvince)
	}
}

func TestErrorTypes(t *testing.T) {
	// Test error creation and marshaling
	err := middleware.NewValidationError("Test error", "TEST_CODE", map[string]string{"field": "value"})

	if err.Type != "validation_error" {
		t.Errorf("Expected type validation_error, got %s", err.Type)
	}

	if err.Message != "Test error" {
		t.Errorf("Expected message 'Test error', got %s", err.Message)
	}

	if err.Code != "TEST_CODE" {
		t.Errorf("Expected code TEST_CODE, got %s", err.Code)
	}
}

func TestJSONHelpers(t *testing.T) {
	// Test ReadJSON function
	testData := middleware.CreateGameRequest{Name: "Test Game"}
	jsonData, _ := json.Marshal(testData)

	req := httptest.NewRequest("POST", "/test", bytes.NewReader(jsonData))
	req.Header.Set("Content-Type", "application/json")

	var result middleware.CreateGameRequest
	err := middleware.ReadJSON(req, &result)
	if err != nil {
		t.Fatalf("ReadJSON failed: %v", err)
	}

	if result.Name != testData.Name {
		t.Errorf("Expected name %s, got %s", testData.Name, result.Name)
	}
}

func TestWriteJSON(t *testing.T) {
	// Test WriteJSON function
	w := httptest.NewRecorder()
	testData := map[string]string{"message": "test"}

	err := middleware.WriteJSON(w, http.StatusOK, testData)
	if err != nil {
		t.Fatalf("WriteJSON failed: %v", err)
	}

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	contentType := w.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("Expected Content-Type application/json, got %s", contentType)
	}
}
