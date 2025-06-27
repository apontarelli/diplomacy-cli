package api

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"time"

	"diplomacy-backend/internal/game"
	"diplomacy-backend/internal/storage"
)

type Server struct {
	db       *storage.DB
	gameData *game.GameData
}

func NewServer(db *storage.DB, gameData *game.GameData) *Server {
	return &Server{
		db:       db,
		gameData: gameData,
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
	for _, assigned := range assignedNations {
		if assigned == nation {
			return true
		}
	}
	return false
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

func (s *Server) SetupRoutes() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", s.HealthHandler)
	mux.HandleFunc("/game-data", s.GameDataHandler)
	mux.HandleFunc("/games", s.CreateGameHandler)
	mux.HandleFunc("/game", s.GetGameHandler)
	mux.HandleFunc("/game-state", s.GetGameStateHandler)
	mux.HandleFunc("/register-player", s.RegisterPlayerHandler)
	return mux
}