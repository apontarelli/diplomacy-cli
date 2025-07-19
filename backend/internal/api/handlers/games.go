package handlers

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"diplomacy-cli/backend/internal/api/middleware"
	"diplomacy-cli/backend/internal/storage"
)

// GameHandler handles game-related HTTP requests
type GameHandler struct {
	db *storage.Database
}

// NewGameHandler creates a new game handler
func NewGameHandler(db *storage.Database) *GameHandler {
	return &GameHandler{db: db}
}

// CreateGame handles POST /api/games
func (gh *GameHandler) CreateGame(w http.ResponseWriter, r *http.Request) {
	// Get authenticated user from context
	user := middleware.MustGetUserFromContext(r.Context())

	// Parse request
	var req middleware.CreateGameRequest
	if err := middleware.ReadJSON(r, &req); err != nil {
		apiErr := middleware.NewValidationError("Invalid request body", "INVALID_JSON", nil)
		middleware.WriteError(w, apiErr)
		return
	}

	// Validate request
	if strings.TrimSpace(req.Name) == "" {
		apiErr := middleware.NewValidationError("Game name is required", "MISSING_GAME_NAME", nil)
		middleware.WriteError(w, apiErr)
		return
	}

	// Create game in database
	game := &storage.Game{
		Name:      strings.TrimSpace(req.Name),
		Phase:     "spring_movement",
		Year:      1901,
		Status:    "waiting_for_players",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Insert game
	query := `
		INSERT INTO games (name, phase, year, status, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?)
	`
	result, err := gh.db.Exec(query, game.Name, game.Phase, game.Year, game.Status, game.CreatedAt, game.UpdatedAt)
	if err != nil {
		middleware.WriteInternalError(w, err)
		return
	}

	// Get the inserted game ID
	gameID, err := result.LastInsertId()
	if err != nil {
		middleware.WriteInternalError(w, err)
		return
	}
	game.ID = gameID

	// Create response
	response := middleware.CreateGameResponse{
		ID:           game.ID,
		Name:         game.Name,
		Status:       game.Status,
		Phase:        game.Phase,
		Year:         game.Year,
		CreatorID:    user.ID,
		PlayersCount: 0,
		MaxPlayers:   7, // Standard Diplomacy has 7 players
		CreatedAt:    game.CreatedAt,
	}

	middleware.WriteJSON(w, http.StatusCreated, response)
}

// JoinGame handles POST /api/games/{id}/join
func (gh *GameHandler) JoinGame(w http.ResponseWriter, r *http.Request) {
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
	var req middleware.JoinGameRequest
	if err := middleware.ReadJSON(r, &req); err != nil {
		apiErr := middleware.NewValidationError("Invalid request body", "INVALID_JSON", nil)
		middleware.WriteError(w, apiErr)
		return
	}

	// Validate nation
	validNations := []string{"austria", "england", "france", "germany", "italy", "russia", "turkey"}
	req.Nation = strings.ToLower(strings.TrimSpace(req.Nation))
	if req.Nation == "" {
		apiErr := middleware.NewValidationError("Nation is required", "MISSING_NATION", nil)
		middleware.WriteError(w, apiErr)
		return
	}

	isValidNation := false
	for _, nation := range validNations {
		if req.Nation == nation {
			isValidNation = true
			break
		}
	}
	if !isValidNation {
		context := map[string]interface{}{
			"valid_nations": validNations,
		}
		apiErr := middleware.NewValidationError("Invalid nation", "INVALID_NATION", context)
		middleware.WriteError(w, apiErr)
		return
	}

	// Check if game exists and is joinable
	var game storage.Game
	query := `SELECT id, name, status, phase, year, created_at, updated_at FROM games WHERE id = ?`
	err = gh.db.QueryRow(query, gameID).Scan(&game.ID, &game.Name, &game.Status, &game.Phase, &game.Year, &game.CreatedAt, &game.UpdatedAt)
	if err != nil {
		apiErr := middleware.NewNotFoundError("Game not found", "GAME_NOT_FOUND")
		middleware.WriteError(w, apiErr)
		return
	}

	if game.Status != "waiting_for_players" {
		apiErr := middleware.NewValidationError("Game is not accepting new players", "GAME_NOT_JOINABLE", nil)
		middleware.WriteError(w, apiErr)
		return
	}

	// Check if user is already in this game
	var existingPlayerID int64
	checkQuery := `SELECT id FROM players WHERE game_id = ? AND user_id = ?`
	err = gh.db.QueryRow(checkQuery, gameID, strconv.FormatInt(user.ID, 10)).Scan(&existingPlayerID)
	if err == nil {
		apiErr := middleware.NewValidationError("You are already in this game", "ALREADY_JOINED", nil)
		middleware.WriteError(w, apiErr)
		return
	}

	// Check if nation is already taken
	var existingNationPlayerID int64
	nationQuery := `SELECT id FROM players WHERE game_id = ? AND nation = ?`
	err = gh.db.QueryRow(nationQuery, gameID, req.Nation).Scan(&existingNationPlayerID)
	if err == nil {
		// Get available nations
		availableNations := gh.getAvailableNations(gameID, validNations)
		context := map[string]interface{}{
			"available_nations": availableNations,
		}
		apiErr := middleware.NewValidationError("Nation already taken", "NATION_TAKEN", context)
		middleware.WriteError(w, apiErr)
		return
	}

	// Create player
	player := &storage.Player{
		GameID:    gameID,
		UserID:    strconv.FormatInt(user.ID, 10),
		Nation:    req.Nation,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	insertQuery := `
		INSERT INTO players (game_id, user_id, nation, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?)
	`
	result, err := gh.db.Exec(insertQuery, player.GameID, player.UserID, player.Nation, player.CreatedAt, player.UpdatedAt)
	if err != nil {
		middleware.WriteInternalError(w, err)
		return
	}

	playerID, err := result.LastInsertId()
	if err != nil {
		middleware.WriteInternalError(w, err)
		return
	}
	player.ID = playerID

	// Create response
	response := middleware.JoinGameResponse{
		PlayerID: player.ID,
		GameID:   gameID,
		Nation:   req.Nation,
		JoinedAt: player.CreatedAt,
	}

	middleware.WriteJSON(w, http.StatusCreated, response)
}

// getAvailableNations returns the list of nations not yet taken in a game
func (gh *GameHandler) getAvailableNations(gameID int64, allNations []string) []string {
	// Get taken nations
	query := `SELECT nation FROM players WHERE game_id = ?`
	rows, err := gh.db.Query(query, gameID)
	if err != nil {
		return allNations // Return all if query fails
	}
	defer rows.Close()

	takenNations := make(map[string]bool)
	for rows.Next() {
		var nation string
		if err := rows.Scan(&nation); err == nil {
			takenNations[nation] = true
		}
	}

	// Filter available nations
	var available []string
	for _, nation := range allNations {
		if !takenNations[nation] {
			available = append(available, nation)
		}
	}

	return available
}
