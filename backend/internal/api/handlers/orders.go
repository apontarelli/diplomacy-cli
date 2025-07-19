package handlers

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"diplomacy-cli/backend/internal/api/middleware"
	"diplomacy-cli/backend/internal/api/websocket"
	"diplomacy-cli/backend/internal/storage"
)

// OrderHandler handles order-related HTTP requests
type OrderHandler struct {
	db        *storage.Database
	wsHandler *websocket.Handler
}

// NewOrderHandler creates a new order handler
func NewOrderHandler(db *storage.Database, wsHandler *websocket.Handler) *OrderHandler {
	return &OrderHandler{
		db:        db,
		wsHandler: wsHandler,
	}
}

// SubmitOrders handles POST /api/games/{id}/orders
func (oh *OrderHandler) SubmitOrders(w http.ResponseWriter, r *http.Request) {
	// Get authenticated user from context
	user := middleware.MustGetUserFromContext(r.Context())

	// Extract game ID from path
	gameIDStr := r.PathValue("id")
	gameID, err := strconv.ParseInt(gameIDStr, 10, 64)
	if err != nil {
		apiErr := middleware.NewValidationError("Invalid game ID", "INVALID_GAME_ID", nil)
		middleware.WriteError(w, apiErr)
		return
	}

	// Parse request
	var req middleware.SubmitOrdersRequest
	if err := middleware.ReadJSON(r, &req); err != nil {
		apiErr := middleware.NewValidationError("Invalid request body", "INVALID_JSON", nil)
		middleware.WriteError(w, apiErr)
		return
	}

	// Validate request
	if len(req.Orders) == 0 {
		apiErr := middleware.NewValidationError("At least one order is required", "NO_ORDERS", nil)
		middleware.WriteError(w, apiErr)
		return
	}

	// Check if game exists and user is a player
	var playerID int64
	var gameStatus string
	query := `
		SELECT p.id, g.status 
		FROM players p 
		JOIN games g ON p.game_id = g.id 
		WHERE p.game_id = ? AND p.user_id = ?
	`
	err = oh.db.QueryRow(query, gameID, strconv.FormatInt(user.ID, 10)).Scan(&playerID, &gameStatus)
	if err != nil {
		apiErr := middleware.NewNotFoundError("Game not found or you are not a player in this game", "NOT_PLAYER")
		middleware.WriteError(w, apiErr)
		return
	}

	// Check if game is in a state that accepts orders
	if gameStatus != "in_progress" && gameStatus != "waiting_for_players" {
		apiErr := middleware.NewValidationError("Game is not accepting orders", "GAME_NOT_ACCEPTING_ORDERS", nil)
		middleware.WriteError(w, apiErr)
		return
	}

	// Validate each order
	validOrderTypes := []string{"move", "hold", "support", "convoy"}
	for i, order := range req.Orders {
		// Validate order type
		order.Type = strings.ToLower(strings.TrimSpace(order.Type))
		if order.Type == "" {
			context := map[string]interface{}{
				"order_index": i,
				"valid_types": validOrderTypes,
			}
			apiErr := middleware.NewValidationError("Order type is required", "MISSING_ORDER_TYPE", context)
			middleware.WriteError(w, apiErr)
			return
		}

		isValidType := false
		for _, validType := range validOrderTypes {
			if order.Type == validType {
				isValidType = true
				break
			}
		}
		if !isValidType {
			context := map[string]interface{}{
				"order_index": i,
				"valid_types": validOrderTypes,
			}
			apiErr := middleware.NewValidationError("Invalid order type", "INVALID_ORDER_TYPE", context)
			middleware.WriteError(w, apiErr)
			return
		}

		// Validate unit province
		order.UnitProvince = strings.ToLower(strings.TrimSpace(order.UnitProvince))
		if order.UnitProvince == "" {
			context := map[string]interface{}{
				"order_index": i,
			}
			apiErr := middleware.NewValidationError("Unit province is required", "MISSING_UNIT_PROVINCE", context)
			middleware.WriteError(w, apiErr)
			return
		}

		// Validate target for move orders
		if order.Type == "move" || order.Type == "support" || order.Type == "convoy" {
			order.Target = strings.ToLower(strings.TrimSpace(order.Target))
			if order.Target == "" {
				context := map[string]interface{}{
					"order_index": i,
					"order_type":  order.Type,
				}
				apiErr := middleware.NewValidationError("Target province is required for this order type", "MISSING_TARGET", context)
				middleware.WriteError(w, apiErr)
				return
			}
		}

		// Clean up coast if provided
		if order.Coast != "" {
			order.Coast = strings.ToLower(strings.TrimSpace(order.Coast))
		}

		// Update the order in the slice with cleaned values
		req.Orders[i] = order
	}

	// Clear existing orders for this player in this game
	deleteQuery := `DELETE FROM orders WHERE game_id = ? AND player_id = ?`
	_, err = oh.db.Exec(deleteQuery, gameID, playerID)
	if err != nil {
		middleware.WriteInternalError(w, err)
		return
	}

	// Insert new orders
	var submittedOrders []middleware.OrderResponse
	for _, order := range req.Orders {
		orderRecord := &storage.Order{
			GameID:    gameID,
			PlayerID:  playerID,
			Type:      order.Type,
			Target:    order.Target,
			Coast:     order.Coast,
			Resolved:  false,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		insertQuery := `
			INSERT INTO orders (game_id, player_id, type, target, coast, resolved, created_at, updated_at)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		`
		result, err := oh.db.Exec(insertQuery,
			orderRecord.GameID, orderRecord.PlayerID, orderRecord.Type,
			orderRecord.Target, orderRecord.Coast, orderRecord.Resolved,
			orderRecord.CreatedAt, orderRecord.UpdatedAt)
		if err != nil {
			middleware.WriteInternalError(w, err)
			return
		}

		orderID, err := result.LastInsertId()
		if err != nil {
			middleware.WriteInternalError(w, err)
			return
		}

		submittedOrders = append(submittedOrders, middleware.OrderResponse{
			ID:           orderID,
			UnitProvince: order.UnitProvince,
			Type:         order.Type,
			Target:       order.Target,
			Coast:        order.Coast,
		})
	}

	// Broadcast orders submitted event via WebSocket
	if oh.wsHandler != nil {
		orderEvent := websocket.NewOrdersSubmittedEvent(gameID, playerID, user.ID, len(submittedOrders), submittedOrders)
		oh.wsHandler.BroadcastGameUpdate(gameID, websocket.EventOrdersSubmitted, orderEvent)
	}

	// Create response
	response := middleware.SubmitOrdersResponse{
		Submitted: len(submittedOrders),
		Orders:    submittedOrders,
	}

	middleware.WriteJSON(w, http.StatusCreated, response)
}
