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

func (s *Server) SetupRoutes() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", s.HealthHandler)
	mux.HandleFunc("/game-data", s.GameDataHandler)
	mux.HandleFunc("/games", s.CreateGameHandler)
	mux.HandleFunc("/game", s.GetGameHandler)
	return mux
}