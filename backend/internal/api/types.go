package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// Request types for API endpoints

// CreateGameRequest represents the request to create a new game
type CreateGameRequest struct {
	Name string `json:"name"`
}

// CreateGameResponse represents the response when creating a game
type CreateGameResponse struct {
	ID           int64     `json:"id"`
	Name         string    `json:"name"`
	Status       string    `json:"status"`
	Phase        string    `json:"phase"`
	Year         int       `json:"year"`
	CreatorID    int64     `json:"creator_id"`
	PlayersCount int       `json:"players_count"`
	MaxPlayers   int       `json:"max_players"`
	CreatedAt    time.Time `json:"created_at"`
}

// JoinGameRequest represents the request to join a game
type JoinGameRequest struct {
	Nation string `json:"nation"`
}

// JoinGameResponse represents the response when joining a game
type JoinGameResponse struct {
	PlayerID int64     `json:"player_id"`
	GameID   int64     `json:"game_id"`
	Nation   string    `json:"nation"`
	JoinedAt time.Time `json:"joined_at"`
}

// OrderRequest represents a single order in the orders submission
type OrderRequest struct {
	UnitProvince string `json:"unit_province"`
	Type         string `json:"type"`
	Target       string `json:"target"`
	Coast        string `json:"coast,omitempty"`
}

// SubmitOrdersRequest represents the request to submit orders
type SubmitOrdersRequest struct {
	Orders []OrderRequest `json:"orders"`
}

// OrderResponse represents a submitted order with ID
type OrderResponse struct {
	ID           int64  `json:"id"`
	UnitProvince string `json:"unit_province"`
	Type         string `json:"type"`
	Target       string `json:"target"`
	Coast        string `json:"coast,omitempty"`
}

// SubmitOrdersResponse represents the response when submitting orders
type SubmitOrdersResponse struct {
	Submitted int             `json:"submitted"`
	Orders    []OrderResponse `json:"orders"`
}

// Error handling types

// APIError represents a structured API error
type APIError struct {
	Type    string      `json:"type"`
	Message string      `json:"message"`
	Code    string      `json:"code"`
	Context interface{} `json:"context,omitempty"`
}

// Error implements the error interface
func (e APIError) Error() string {
	return e.Message
}

// Common error constructors

// NewValidationError creates a validation error
func NewValidationError(message, code string, context interface{}) APIError {
	return APIError{
		Type:    "validation_error",
		Message: message,
		Code:    code,
		Context: context,
	}
}

// NewAuthError creates an authentication error
func NewAuthError(message, code string) APIError {
	return APIError{
		Type:    "auth_error",
		Message: message,
		Code:    code,
	}
}

// NewNotFoundError creates a not found error
func NewNotFoundError(message, code string) APIError {
	return APIError{
		Type:    "not_found_error",
		Message: message,
		Code:    code,
	}
}

// NewInternalError creates an internal server error
func NewInternalError(message, code string) APIError {
	return APIError{
		Type:    "internal_error",
		Message: message,
		Code:    code,
	}
}

// NewRateLimitError creates a rate limit error
func NewRateLimitError(message, code string, context interface{}) APIError {
	return APIError{
		Type:    "rate_limit_error",
		Message: message,
		Code:    code,
		Context: context,
	}
}

// JSON helper functions

// WriteJSON writes a JSON response with the given status code
func WriteJSON(w http.ResponseWriter, status int, data interface{}) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(data)
}

// ReadJSON reads JSON from the request body into the destination
func ReadJSON(r *http.Request, dst interface{}) error {
	if r.Header.Get("Content-Type") != "application/json" {
		return fmt.Errorf("content-type must be application/json")
	}

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(dst); err != nil {
		return fmt.Errorf("invalid JSON: %w", err)
	}

	return nil
}

// WriteError writes a structured error response
func WriteError(w http.ResponseWriter, err APIError) error {
	var status int
	switch err.Type {
	case "validation_error":
		status = http.StatusBadRequest
	case "auth_error":
		status = http.StatusUnauthorized
	case "not_found_error":
		status = http.StatusNotFound
	case "rate_limit_error":
		status = http.StatusTooManyRequests
	case "internal_error":
		status = http.StatusInternalServerError
	default:
		status = http.StatusInternalServerError
	}

	response := map[string]interface{}{
		"error": err,
	}

	return WriteJSON(w, status, response)
}

// WriteInternalError writes a generic internal server error
func WriteInternalError(w http.ResponseWriter, err error) error {
	apiErr := NewInternalError("Internal server error", "INTERNAL_ERROR")
	return WriteError(w, apiErr)
}
