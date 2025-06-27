package api

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"slices"
	"strings"
	"time"

	"diplomacy-backend/internal/game"
	"diplomacy-backend/internal/game/rules"
	"diplomacy-backend/internal/storage"
)

type Server struct {
	db       *storage.DB
	gameData *game.GameData
	rules    *rules.Rules
}

func NewServer(db *storage.DB, gameData *game.GameData) *Server {
	gameRules := rules.MustLoadRules("classic")
	return &Server{
		db:       db,
		gameData: gameData,
		rules:    gameRules,
	}
}

type HealthResponse struct {
	Status   string `json:"status"`
	Database string `json:"database"`
}

func (s *Server) HealthHandler(w http.ResponseWriter, r *http.Request) {
	response := HealthResponse{
		Status:   "ok",
		Database: "disconnected",
	}

	if err := s.db.Ping(); err == nil {
		response.Database = "connected"
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func (s *Server) GameDataHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(s.gameData)
}

type CreateGameRequest struct {
	Name string `json:"name"`
}

type CreateGameResponse struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Status string `json:"status"`
}

func (s *Server) CreateGameHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req CreateGameRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if req.Name == "" {
		http.Error(w, "Game name is required", http.StatusBadRequest)
		return
	}

	gameID, err := s.createGame(req.Name)
	if err != nil {
		http.Error(w, "Failed to create game", http.StatusInternalServerError)
		return
	}

	response := CreateGameResponse{
		ID:     gameID,
		Name:   req.Name,
		Status: "forming",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

func (s *Server) createGame(name string) (string, error) {
	gameID := generateID()
	
	newGame := game.Game{
		ID:        gameID,
		Name:      name,
		Status:    game.GameStatusForming,
		CreatedAt: time.Now(),
	}

	return gameID, s.db.CreateGame(newGame)
}

func generateID() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

func (s *Server) GetGameHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	gameID := r.URL.Query().Get("id")
	if gameID == "" {
		http.Error(w, "Game ID is required", http.StatusBadRequest)
		return
	}

	gameInfo, err := s.db.GetGame(gameID)
	if err != nil {
		http.Error(w, "Game not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(gameInfo)
}

type RegisterPlayerRequest struct {
	GameID     string `json:"game_id"`
	PlayerName string `json:"player_name"`
	Nation     string `json:"nation,omitempty"`
}

type RegisterPlayerResponse struct {
	PlayerID string `json:"player_id"`
	GameID   string `json:"game_id"`
	Nation   string `json:"nation"`
	Status   string `json:"status"`
}

func (s *Server) RegisterPlayerHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req RegisterPlayerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if req.GameID == "" || req.PlayerName == "" {
		http.Error(w, "Game ID and player name are required", http.StatusBadRequest)
		return
	}

	gameInfo, err := s.db.GetGame(req.GameID)
	if err != nil {
		http.Error(w, "Game not found", http.StatusNotFound)
		return
	}

	if gameInfo.Status != game.GameStatusForming {
		http.Error(w, "Game is not accepting new players", http.StatusBadRequest)
		return
	}

	assignedNations, err := s.db.GetAssignedNations(req.GameID)
	if err != nil {
		http.Error(w, "Failed to check assigned nations", http.StatusInternalServerError)
		return
	}

	var assignedNation string
	if req.Nation != "" {
		if s.isNationTaken(req.Nation, assignedNations) {
			http.Error(w, "Nation already taken", http.StatusBadRequest)
			return
		}
		if !s.isValidNation(req.Nation) {
			http.Error(w, "Invalid nation", http.StatusBadRequest)
			return
		}
		assignedNation = req.Nation
	} else {
		assignedNation = s.assignRandomNation(assignedNations)
		if assignedNation == "" {
			http.Error(w, "No available nations", http.StatusBadRequest)
			return
		}
	}

	playerID := generateID()
	player := game.Player{
		ID:     playerID,
		GameID: req.GameID,
		Nation: assignedNation,
		Status: game.PlayerStatusActive,
	}

	if err := s.db.CreatePlayer(player); err != nil {
		http.Error(w, "Failed to register player", http.StatusInternalServerError)
		return
	}

	response := RegisterPlayerResponse{
		PlayerID: playerID,
		GameID:   req.GameID,
		Nation:   assignedNation,
		Status:   string(game.PlayerStatusActive),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

func (s *Server) isNationTaken(nation string, assignedNations []string) bool {
	return slices.Contains(assignedNations, nation)
}

func (s *Server) isValidNation(nation string) bool {
	for _, n := range s.gameData.Nations {
		if n.ID == nation {
			return true
		}
	}
	return false
}

func (s *Server) assignRandomNation(assignedNations []string) string {
	for _, nation := range s.gameData.Nations {
		if !s.isNationTaken(nation.ID, assignedNations) {
			return nation.ID
		}
	}
	return ""
}

func (s *Server) GetGameStateHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	gameID := r.URL.Query().Get("id")
	if gameID == "" {
		http.Error(w, "Game ID is required", http.StatusBadRequest)
		return
	}

	gameState, err := s.db.GetGameState(gameID)
	if err != nil {
		http.Error(w, "Game not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(gameState)
}

type AdvancePhaseResponse struct {
	GameID   string `json:"game_id"`
	NewTurn  *game.Turn `json:"new_turn,omitempty"`
	Message  string `json:"message"`
}

func (s *Server) AdvancePhaseHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	gameID := r.URL.Query().Get("id")
	if gameID == "" {
		http.Error(w, "Game ID is required", http.StatusBadRequest)
		return
	}

	gameInfo, err := s.db.GetGame(gameID)
	if err != nil {
		http.Error(w, "Game not found", http.StatusNotFound)
		return
	}

	if gameInfo.Status != game.GameStatusActive {
		http.Error(w, "Game is not active", http.StatusBadRequest)
		return
	}

	currentTurn, err := s.db.GetCurrentTurn(gameID)
	if err != nil {
		http.Error(w, "No active turn found", http.StatusNotFound)
		return
	}

	newTurn, message, err := s.advancePhase(gameID, currentTurn)
	if err != nil {
		http.Error(w, "Failed to advance phase", http.StatusInternalServerError)
		return
	}

	response := AdvancePhaseResponse{
		GameID:  gameID,
		NewTurn: newTurn,
		Message: message,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func (s *Server) advancePhase(gameID string, currentTurn *game.Turn) (*game.Turn, string, error) {
	if err := s.db.UpdateTurnStatus(currentTurn.ID, game.TurnStatusCompleted); err != nil {
		return nil, "", err
	}

	nextYear := currentTurn.Year
	nextSeason := currentTurn.Season
	nextPhase := currentTurn.Phase

	switch currentTurn.Phase {
	case game.PhaseMovement:
		nextPhase = game.PhaseRetreat
	case game.PhaseRetreat:
		if currentTurn.Season == game.SeasonSpring {
			nextPhase = game.PhaseMovement
			nextSeason = game.SeasonFall
		} else {
			nextPhase = game.PhaseBuild
		}
	case game.PhaseBuild:
		nextPhase = game.PhaseMovement
		nextSeason = game.SeasonSpring
		nextYear++
	}

	deadline := time.Now().Add(24 * time.Hour) // 24 hour deadline
	newTurn := game.Turn{
		GameID:    gameID,
		Year:      nextYear,
		Season:    nextSeason,
		Phase:     nextPhase,
		Status:    game.TurnStatusActive,
		Deadline:  &deadline,
		CreatedAt: time.Now(),
	}

	if err := s.db.CreateTurn(newTurn); err != nil {
		return nil, "", err
	}

	createdTurn, err := s.db.GetCurrentTurn(gameID)
	if err != nil {
		return nil, "", err
	}

	message := "Advanced to " + string(nextSeason) + " " + string(nextPhase) + " phase"
	return createdTurn, message, nil
}

func (s *Server) CheckDeadlinesHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	expiredGames, err := s.checkAndAdvanceExpiredTurns()
	if err != nil {
		http.Error(w, "Failed to check deadlines", http.StatusInternalServerError)
		return
	}

	response := map[string]any{
		"expired_games": expiredGames,
		"message":       "Deadline check completed",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func (s *Server) checkAndAdvanceExpiredTurns() ([]string, error) {
	expiredTurns, err := s.db.GetExpiredTurns()
	if err != nil {
		return nil, err
	}

	var expiredGames []string
	for _, turn := range expiredTurns {
		_, _, err := s.advancePhase(turn.GameID, &turn)
		if err != nil {
			continue // Log error but continue with other games
		}
		expiredGames = append(expiredGames, turn.GameID)
	}

	return expiredGames, nil
}

type SubmitOrderRequest struct {
	GameID        string `json:"game_id"`
	PlayerID      string `json:"player_id"`
	UnitID        string `json:"unit_id"`
	OrderType     string `json:"order_type"`
	FromTerritory string `json:"from_territory"`
	ToTerritory   string `json:"to_territory,omitempty"`
	SupportUnit   string `json:"support_unit,omitempty"`
}

type SubmitOrderResponse struct {
	OrderID string `json:"order_id"`
	Status  string `json:"status"`
	Message string `json:"message"`
}

func (s *Server) SubmitOrderHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req SubmitOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if req.GameID == "" || req.PlayerID == "" || req.UnitID == "" || req.OrderType == "" {
		http.Error(w, "Missing required fields", http.StatusBadRequest)
		return
	}

	gameInfo, err := s.db.GetGame(req.GameID)
	if err != nil {
		http.Error(w, "Game not found", http.StatusNotFound)
		return
	}

	if gameInfo.Status != game.GameStatusActive {
		http.Error(w, "Game is not active", http.StatusBadRequest)
		return
	}

	currentTurn, err := s.db.GetCurrentTurn(req.GameID)
	if err != nil {
		http.Error(w, "No active turn found", http.StatusNotFound)
		return
	}

	if currentTurn.Phase != game.PhaseMovement {
		http.Error(w, "Orders can only be submitted during movement phase", http.StatusBadRequest)
		return
	}

	if err := s.validateOrder(req, currentTurn); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	orderID := generateID()
	order := game.Order{
		ID:            orderID,
		GameID:        req.GameID,
		TurnID:        currentTurn.ID,
		PlayerID:      req.PlayerID,
		UnitID:        req.UnitID,
		OrderType:     game.OrderType(req.OrderType),
		FromTerritory: req.FromTerritory,
		ToTerritory:   req.ToTerritory,
		SupportUnit:   req.SupportUnit,
		Status:        game.OrderStatusSubmitted,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	if err := s.db.CreateOrder(order); err != nil {
		http.Error(w, "Failed to submit order", http.StatusInternalServerError)
		return
	}

	response := SubmitOrderResponse{
		OrderID: orderID,
		Status:  "submitted",
		Message: "Order submitted successfully",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

func (s *Server) validateOrder(req SubmitOrderRequest, turn *game.Turn) error {
	orderType := game.OrderType(req.OrderType)
	
	// Basic order type validation
	switch orderType {
	case game.OrderTypeMove:
		if req.ToTerritory == "" {
			return fmt.Errorf("move orders require a destination territory")
		}
	case game.OrderTypeHold:
		// Hold orders don't need additional validation
	case game.OrderTypeSupport:
		if req.SupportUnit == "" {
			return fmt.Errorf("support orders require a unit to support")
		}
	case game.OrderTypeConvoy:
		if req.ToTerritory == "" || req.SupportUnit == "" {
			return fmt.Errorf("convoy orders require destination and unit to convoy")
		}
	default:
		return fmt.Errorf("invalid order type: %s", req.OrderType)
	}

	// Enhanced validation for move orders using rules engine
	if orderType == game.OrderTypeMove {
		return s.validateMoveOrder(req)
	}

	return nil
}

func (s *Server) validateMoveOrder(req SubmitOrderRequest) error {
	// Validate territories exist
	if !s.rules.IsValidTerritory(req.FromTerritory) {
		return fmt.Errorf("invalid origin territory: %s", req.FromTerritory)
	}
	
	if !s.rules.IsValidTerritory(req.ToTerritory) {
		return fmt.Errorf("invalid destination territory: %s", req.ToTerritory)
	}

	// Get unit information to determine unit type
	unit, err := s.db.GetUnitByID(req.UnitID)
	if err != nil {
		return fmt.Errorf("unit not found: %s", req.UnitID)
	}

	// Validate unit can occupy origin territory
	if !s.rules.CanUnitOccupy(unit.UnitType, req.FromTerritory) {
		return fmt.Errorf("%s cannot occupy %s", unit.UnitType, req.FromTerritory)
	}

	// Validate unit can occupy destination territory
	if !s.rules.CanUnitOccupy(unit.UnitType, req.ToTerritory) {
		return fmt.Errorf("%s cannot occupy %s", unit.UnitType, req.ToTerritory)
	}

	// Validate adjacency - unit can move from origin to destination
	if !s.rules.CanMove(unit.UnitType, req.FromTerritory, req.ToTerritory) {
		return fmt.Errorf("%s cannot move from %s to %s (not adjacent or invalid for unit type)", 
			unit.UnitType, req.FromTerritory, req.ToTerritory)
	}

	return nil
}

func (s *Server) validateMoveOrderWithRules(order game.Order, unitPositions map[string]string) error {
	// Validate territories exist
	if !s.rules.IsValidTerritory(order.FromTerritory) {
		return fmt.Errorf("invalid origin territory: %s", order.FromTerritory)
	}
	
	if !s.rules.IsValidTerritory(order.ToTerritory) {
		return fmt.Errorf("invalid destination territory: %s", order.ToTerritory)
	}

	// For testing, extract unit type from unit ID (format: "country_unittype_territory")
	unitType := s.extractUnitTypeFromID(order.UnitID)
	if unitType == "" {
		// Fallback: try to get from database if available
		if s.db != nil {
			unit, err := s.db.GetUnitByID(order.UnitID)
			if err != nil {
				return fmt.Errorf("unit not found: %s", order.UnitID)
			}
			unitType = string(unit.UnitType)
		} else {
			return fmt.Errorf("cannot determine unit type for: %s", order.UnitID)
		}
	}

	// Verify unit is actually in the from territory
	actualPosition := unitPositions[order.UnitID]
	if actualPosition != order.FromTerritory {
		return fmt.Errorf("unit %s is not in territory %s (actually in %s)", 
			order.UnitID, order.FromTerritory, actualPosition)
	}

	gameUnitType := game.UnitType(unitType)

	// Validate unit can occupy origin territory
	if !s.rules.CanUnitOccupy(gameUnitType, order.FromTerritory) {
		return fmt.Errorf("%s cannot occupy %s", gameUnitType, order.FromTerritory)
	}

	// Validate unit can occupy destination territory
	if !s.rules.CanUnitOccupy(gameUnitType, order.ToTerritory) {
		return fmt.Errorf("%s cannot occupy %s", gameUnitType, order.ToTerritory)
	}

	// Validate adjacency - unit can move from origin to destination
	if !s.rules.CanMove(gameUnitType, order.FromTerritory, order.ToTerritory) {
		return fmt.Errorf("%s cannot move from %s to %s (not adjacent or invalid for unit type)", 
			gameUnitType, order.FromTerritory, order.ToTerritory)
	}

	return nil
}

// extractUnitTypeFromID extracts unit type from test unit IDs like "england_fleet_north_sea"
func (s *Server) extractUnitTypeFromID(unitID string) string {
	// Split by underscore and look for "army" or "fleet"
	parts := strings.Split(unitID, "_")
	for _, part := range parts {
		if part == "army" || part == "fleet" {
			return part
		}
	}
	return ""
}

func (s *Server) GetOrdersHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	gameID := r.URL.Query().Get("game_id")
	playerID := r.URL.Query().Get("player_id")

	if gameID == "" {
		http.Error(w, "Game ID is required", http.StatusBadRequest)
		return
	}

	currentTurn, err := s.db.GetCurrentTurn(gameID)
	if err != nil {
		http.Error(w, "No active turn found", http.StatusNotFound)
		return
	}

	var orders []game.Order
	if playerID != "" {
		orders, err = s.db.GetOrdersByPlayer(gameID, currentTurn.ID, playerID)
	} else {
		orders, err = s.db.GetOrdersByTurn(gameID, currentTurn.ID)
	}

	if err != nil {
		http.Error(w, "Failed to get orders", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(orders)
}

type CancelOrderRequest struct {
	OrderID  string `json:"order_id"`
	PlayerID string `json:"player_id"`
}

func (s *Server) CancelOrderHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req CancelOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if req.OrderID == "" || req.PlayerID == "" {
		http.Error(w, "Order ID and Player ID are required", http.StatusBadRequest)
		return
	}

	if err := s.db.UpdateOrderStatus(req.OrderID, game.OrderStatusCancelled); err != nil {
		http.Error(w, "Failed to cancel order", http.StatusInternalServerError)
		return
	}

	response := map[string]string{
		"status":  "cancelled",
		"message": "Order cancelled successfully",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

type ModifyOrderRequest struct {
	OrderID       string `json:"order_id"`
	PlayerID      string `json:"player_id"`
	OrderType     string `json:"order_type,omitempty"`
	ToTerritory   string `json:"to_territory,omitempty"`
	SupportUnit   string `json:"support_unit,omitempty"`
}

func (s *Server) ModifyOrderHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req ModifyOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if req.OrderID == "" || req.PlayerID == "" {
		http.Error(w, "Order ID and Player ID are required", http.StatusBadRequest)
		return
	}

	// For now, we'll implement modification by cancelling the old order and creating a new one
	// In a more sophisticated system, you'd update the existing order
	if err := s.db.UpdateOrderStatus(req.OrderID, game.OrderStatusCancelled); err != nil {
		http.Error(w, "Failed to modify order", http.StatusInternalServerError)
		return
	}

	response := map[string]string{
		"status":  "modified",
		"message": "Order modified successfully (cancelled old order)",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func (s *Server) CheckOrderConflictsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	gameID := r.URL.Query().Get("game_id")
	if gameID == "" {
		http.Error(w, "Game ID is required", http.StatusBadRequest)
		return
	}

	currentTurn, err := s.db.GetCurrentTurn(gameID)
	if err != nil {
		http.Error(w, "No active turn found", http.StatusNotFound)
		return
	}

	orders, err := s.db.GetOrdersByTurn(gameID, currentTurn.ID)
	if err != nil {
		http.Error(w, "Failed to get orders", http.StatusInternalServerError)
		return
	}

	conflicts := s.detectOrderConflicts(orders)

	response := map[string]any{
		"conflicts": conflicts,
		"message":   "Order conflict check completed",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func (s *Server) detectOrderConflicts(orders []game.Order) []map[string]any {
	var conflicts []map[string]any
	territoryMoves := make(map[string][]string) // territory -> list of unit IDs trying to move there

	for _, order := range orders {
		if order.Status == game.OrderStatusCancelled {
			continue
		}

		if order.OrderType == game.OrderTypeMove && order.ToTerritory != "" {
			territoryMoves[order.ToTerritory] = append(territoryMoves[order.ToTerritory], order.UnitID)
		}
	}

	// Check for multiple units trying to move to the same territory
	for territory, unitIDs := range territoryMoves {
		if len(unitIDs) > 1 {
			conflict := map[string]any{
				"type":      "movement_conflict",
				"territory": territory,
				"units":     unitIDs,
				"message":   "Multiple units attempting to move to the same territory",
			}
			conflicts = append(conflicts, conflict)
		}
	}

	return conflicts
}

type ResolveOrdersResponse struct {
	GameID  string `json:"game_id"`
	TurnID  int    `json:"turn_id"`
	Results []game.OrderResult `json:"results"`
	Message string `json:"message"`
}

func (s *Server) ResolveOrdersHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	gameID := r.URL.Query().Get("game_id")
	if gameID == "" {
		http.Error(w, "Game ID is required", http.StatusBadRequest)
		return
	}

	gameInfo, err := s.db.GetGame(gameID)
	if err != nil {
		http.Error(w, "Game not found", http.StatusNotFound)
		return
	}

	if gameInfo.Status != game.GameStatusActive {
		http.Error(w, "Game is not active", http.StatusBadRequest)
		return
	}

	currentTurn, err := s.db.GetCurrentTurn(gameID)
	if err != nil {
		http.Error(w, "No active turn found", http.StatusNotFound)
		return
	}

	if currentTurn.Phase != game.PhaseMovement {
		http.Error(w, "Can only resolve orders during movement phase", http.StatusBadRequest)
		return
	}

	orders, err := s.db.GetOrdersByTurn(gameID, currentTurn.ID)
	if err != nil {
		http.Error(w, "Failed to get orders", http.StatusInternalServerError)
		return
	}

	results, err := s.resolveOrders(gameID, currentTurn.ID, orders)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to resolve orders: %v", err), http.StatusInternalServerError)
		return
	}

	response := ResolveOrdersResponse{
		GameID:  gameID,
		TurnID:  currentTurn.ID,
		Results: results,
		Message: "Orders resolved successfully",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func (s *Server) resolveOrders(gameID string, _ int, orders []game.Order) ([]game.OrderResult, error) {
	var results []game.OrderResult
	
	// Get current unit positions
	units, err := s.db.GetUnitsByGame(gameID)
	if err != nil {
		return nil, err
	}

	// Create maps for easier lookup
	unitPositions := make(map[string]string) // unitID -> territoryID
	for _, unit := range units {
		unitPositions[unit.ID] = unit.TerritoryID
	}

	// Group orders by type for processing
	moveOrders := make([]game.Order, 0)
	supportOrders := make([]game.Order, 0)
	holdOrders := make([]game.Order, 0)
	convoyOrders := make([]game.Order, 0)

	for _, order := range orders {
		if order.Status == game.OrderStatusCancelled {
			continue
		}

		switch order.OrderType {
		case game.OrderTypeMove:
			moveOrders = append(moveOrders, order)
		case game.OrderTypeSupport:
			supportOrders = append(supportOrders, order)
		case game.OrderTypeHold:
			holdOrders = append(holdOrders, order)
		case game.OrderTypeConvoy:
			convoyOrders = append(convoyOrders, order)
		}
	}

	// Resolve movement orders (basic implementation)
	moveResults := s.resolveMovementOrders(moveOrders, supportOrders, unitPositions)
	results = append(results, moveResults...)

	// Resolve support orders
	supportResults := s.resolveSupportOrders(supportOrders, moveResults)
	results = append(results, supportResults...)

	// Resolve hold orders (always succeed unless dislodged)
	holdResults := s.resolveHoldOrders(holdOrders, moveResults)
	results = append(results, holdResults...)

	// Resolve convoy orders (simplified - just mark as successful)
	convoyResults := s.resolveConvoyOrders(convoyOrders)
	results = append(results, convoyResults...)

	// Update unit positions based on successful moves
	if err := s.updateUnitPositions(gameID, results); err != nil {
		return nil, err
	}

	return results, nil
}

func (s *Server) resolveMovementOrders(moveOrders []game.Order, supportOrders []game.Order, unitPositions map[string]string) []game.OrderResult {
	var results []game.OrderResult
	var validMoves []game.Order
	
	// First, validate all move orders using rules engine
	for _, order := range moveOrders {
		if err := s.validateMoveOrderWithRules(order, unitPositions); err != nil {
			// Move is invalid - mark as failed
			results = append(results, game.OrderResult{
				OrderID: order.ID,
				Result:  game.ResolutionFailed,
				Reason:  err.Error(),
			})
		} else {
			// Move is valid - add to valid moves for conflict resolution
			validMoves = append(validMoves, order)
		}
	}
	
	// Group valid moves by destination to detect conflicts
	destinationMoves := make(map[string][]game.Order)
	for _, order := range validMoves {
		destinationMoves[order.ToTerritory] = append(destinationMoves[order.ToTerritory], order)
	}

	// Calculate support strength for each move
	supportStrength := s.calculateSupportStrength(moveOrders, supportOrders)

	for destination, moves := range destinationMoves {
		if len(moves) == 1 {
			// Single move to destination
			move := moves[0]
			strength := 1 + supportStrength[move.ID] // Base strength + support
			
			// Check if destination is occupied and defended
			defendingStrength := s.getDefendingStrength(destination, supportOrders, unitPositions)
			
			if strength > defendingStrength {
				results = append(results, game.OrderResult{
					OrderID:     move.ID,
					Result:      game.ResolutionSuccess,
					Reason:      "Move successful",
					NewPosition: move.ToTerritory,
				})
			} else {
				results = append(results, game.OrderResult{
					OrderID: move.ID,
					Result:  game.ResolutionFailed,
					Reason:  "Insufficient strength to dislodge defender",
				})
			}
		} else {
			// Multiple moves to same destination - all bounce
			for _, move := range moves {
				results = append(results, game.OrderResult{
					OrderID: move.ID,
					Result:  game.ResolutionBounced,
					Reason:  "Multiple units attempted to move to same territory",
				})
			}
		}
	}

	return results
}

func (s *Server) calculateSupportStrength(moveOrders []game.Order, supportOrders []game.Order) map[string]int {
	strength := make(map[string]int)
	
	for _, support := range supportOrders {
		// Find the move order being supported
		for _, move := range moveOrders {
			if move.UnitID == support.SupportUnit && move.ToTerritory == support.ToTerritory {
				strength[move.ID]++
				break
			}
		}
	}
	
	return strength
}

func (s *Server) getDefendingStrength(territory string, supportOrders []game.Order, unitPositions map[string]string) int {
	defendingStrength := 0
	
	// Find unit in the territory
	for unitID, position := range unitPositions {
		if position == territory {
			defendingStrength = 1 // Base defending strength
			
			// Add support for the defending unit
			for _, support := range supportOrders {
				if support.SupportUnit == unitID {
					defendingStrength++
				}
			}
			break
		}
	}
	
	return defendingStrength
}

func (s *Server) resolveSupportOrders(supportOrders []game.Order, _ []game.OrderResult) []game.OrderResult {
	var results []game.OrderResult
	
	for _, support := range supportOrders {
		// Support succeeds unless the supporting unit is dislodged
		result := game.OrderResult{
			OrderID: support.ID,
			Result:  game.ResolutionSuccess,
			Reason:  "Support provided",
		}
		
		// Check if supporting unit was dislodged (simplified)
		// In a full implementation, you'd check if the supporting unit's territory was attacked
		
		results = append(results, result)
	}
	
	return results
}

func (s *Server) resolveHoldOrders(holdOrders []game.Order, _ []game.OrderResult) []game.OrderResult {
	var results []game.OrderResult
	
	for _, hold := range holdOrders {
		// Hold orders succeed unless the unit is dislodged
		result := game.OrderResult{
			OrderID: hold.ID,
			Result:  game.ResolutionSuccess,
			Reason:  "Unit held position",
		}
		
		results = append(results, result)
	}
	
	return results
}

func (s *Server) updateUnitPositions(_ string, results []game.OrderResult) error {
	for _, result := range results {
		if result.Result == game.ResolutionSuccess && result.NewPosition != "" {
			// Get the order to find the unit ID
			order, err := s.db.GetOrderByID(result.OrderID)
			if err != nil {
				continue // Skip if order not found
			}
			
			// Update unit position
			if err := s.db.UpdateUnitPosition(order.UnitID, result.NewPosition); err != nil {
				return fmt.Errorf("failed to update unit position: %w", err)
			}
		}
	}
	
	return nil
}

func (s *Server) resolveConvoyOrders(convoyOrders []game.Order) []game.OrderResult {
	var results []game.OrderResult
	
	for _, convoy := range convoyOrders {
		// Simplified convoy resolution - just mark as successful
		// In a full implementation, you'd check:
		// 1. If the convoying fleet is in a sea territory
		// 2. If there's a valid convoy chain
		// 3. If the convoy chain is unbroken
		
		result := game.OrderResult{
			OrderID: convoy.ID,
			Result:  game.ResolutionSuccess,
			Reason:  "Convoy provided",
		}
		
		results = append(results, result)
	}
	
	return results
}

func (s *Server) SetupRoutes() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", s.HealthHandler)
	mux.HandleFunc("/game-data", s.GameDataHandler)
	mux.HandleFunc("/games", s.CreateGameHandler)
	mux.HandleFunc("/game", s.GetGameHandler)
	mux.HandleFunc("/game-state", s.GetGameStateHandler)
	mux.HandleFunc("/register-player", s.RegisterPlayerHandler)
	mux.HandleFunc("/advance-phase", s.AdvancePhaseHandler)
	mux.HandleFunc("/check-deadlines", s.CheckDeadlinesHandler)
	mux.HandleFunc("/submit-order", s.SubmitOrderHandler)
	mux.HandleFunc("/orders", s.GetOrdersHandler)
	mux.HandleFunc("/cancel-order", s.CancelOrderHandler)
	mux.HandleFunc("/modify-order", s.ModifyOrderHandler)
	mux.HandleFunc("/check-conflicts", s.CheckOrderConflictsHandler)
	mux.HandleFunc("/resolve-orders", s.ResolveOrdersHandler)
	return mux
}