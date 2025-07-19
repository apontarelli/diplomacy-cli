package handlers

import (
	"html/template"
	"net/http"
	"strconv"

	"diplomacy-cli/backend/internal/storage"
)

// TemplateHandler handles HTML template rendering for HTMX
type TemplateHandler struct {
	db        *storage.Database
	templates *template.Template
}

// NewTemplateHandler creates a new template handler
func NewTemplateHandler(db *storage.Database) *TemplateHandler {
	// Load templates from the frontend directory
	templates := template.New("main")

	// Parse base layout
	template.Must(templates.ParseGlob("../frontend/templates/layouts/*.html"))
	// Parse page templates
	template.Must(templates.ParseGlob("../frontend/templates/*.html"))
	// Parse partials
	template.Must(templates.ParseGlob("../frontend/templates/partials/*.html"))

	return &TemplateHandler{
		db:        db,
		templates: templates,
	}
}

// GameListPage renders the games list page
func (h *TemplateHandler) GameListPage(w http.ResponseWriter, r *http.Request) {
	// Check if this is an HTMX request
	if r.Header.Get("HX-Request") == "true" {
		h.renderGameListPartial(w, r)
		return
	}

	// Render full page
	h.renderGameListPage(w, r)
}

// GamePage renders a specific game page
func (h *TemplateHandler) GamePage(w http.ResponseWriter, r *http.Request) {
	gameIDStr := r.PathValue("id")
	gameID, err := strconv.Atoi(gameIDStr)
	if err != nil {
		http.Error(w, "Invalid game ID", http.StatusBadRequest)
		return
	}

	// Check if this is an HTMX request
	if r.Header.Get("HX-Request") == "true" {
		h.renderGamePartial(w, r, gameID)
		return
	}

	// Render full page
	h.renderGamePage(w, r, gameID)
}

// GameBoardPartial renders just the game board for HTMX updates
func (h *TemplateHandler) GameBoardPartial(w http.ResponseWriter, r *http.Request) {
	gameIDStr := r.PathValue("id")
	gameID, err := strconv.Atoi(gameIDStr)
	if err != nil {
		http.Error(w, "Invalid game ID", http.StatusBadRequest)
		return
	}

	h.renderGameBoardPartial(w, r, gameID)
}

// ProvinceDetailsPartial renders province details for HTMX
func (h *TemplateHandler) ProvinceDetailsPartial(w http.ResponseWriter, r *http.Request) {
	provinceID := r.URL.Query().Get("province")
	if provinceID == "" {
		http.Error(w, "Province ID required", http.StatusBadRequest)
		return
	}

	h.renderProvinceDetails(w, r, provinceID)
}

// OrderFormPartial renders the order submission form
func (h *TemplateHandler) OrderFormPartial(w http.ResponseWriter, r *http.Request) {
	gameIDStr := r.PathValue("id")
	gameID, err := strconv.Atoi(gameIDStr)
	if err != nil {
		http.Error(w, "Invalid game ID", http.StatusBadRequest)
		return
	}

	h.renderOrderForm(w, r, gameID)
}

// Helper methods for rendering

func (h *TemplateHandler) renderGameListPage(w http.ResponseWriter, r *http.Request) {
	data := struct {
		Title string
		Games []GameData
	}{
		Title: "Games",
		Games: h.getGamesData(),
	}

	w.Header().Set("Content-Type", "text/html")
	if err := h.templates.ExecuteTemplate(w, "base.html", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
func (h *TemplateHandler) renderGameListPartial(w http.ResponseWriter, r *http.Request) {
	data := struct {
		Games []GameData
	}{
		Games: h.getGamesData(),
	}

	w.Header().Set("Content-Type", "text/html")
	if err := h.templates.ExecuteTemplate(w, "partials/game-list.html", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (h *TemplateHandler) renderGamePage(w http.ResponseWriter, r *http.Request, gameID int) {
	gameData := h.getGameData(gameID)
	if gameData == nil {
		http.Error(w, "Game not found", http.StatusNotFound)
		return
	}

	data := struct {
		Title string
		Game  *GameData
	}{
		Title: "Game " + strconv.Itoa(gameID),
		Game:  gameData,
	}

	w.Header().Set("Content-Type", "text/html")
	if err := h.templates.ExecuteTemplate(w, "game.html", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (h *TemplateHandler) renderGamePartial(w http.ResponseWriter, r *http.Request, gameID int) {
	gameData := h.getGameData(gameID)
	if gameData == nil {
		http.Error(w, "Game not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	if err := h.templates.ExecuteTemplate(w, "partials/game-content.html", gameData); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (h *TemplateHandler) renderGameBoardPartial(w http.ResponseWriter, r *http.Request, gameID int) {
	gameData := h.getGameData(gameID)
	if gameData == nil {
		http.Error(w, "Game not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	if err := h.templates.ExecuteTemplate(w, "partials/game-board.html", gameData); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (h *TemplateHandler) renderProvinceDetails(w http.ResponseWriter, r *http.Request, provinceID string) {
	data := struct {
		ProvinceID string
		Units      []UnitData
		Orders     []OrderData
	}{
		ProvinceID: provinceID,
		Units:      h.getProvinceUnits(provinceID),
		Orders:     h.getProvinceOrders(provinceID),
	}

	w.Header().Set("Content-Type", "text/html")
	if err := h.templates.ExecuteTemplate(w, "partials/province-details.html", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (h *TemplateHandler) renderOrderForm(w http.ResponseWriter, r *http.Request, gameID int) {
	data := struct {
		GameID    int
		Units     []UnitData
		Provinces []ProvinceData
	}{
		GameID:    gameID,
		Units:     h.getPlayerUnits(gameID),
		Provinces: h.getProvinces(),
	}

	w.Header().Set("Content-Type", "text/html")
	if err := h.templates.ExecuteTemplate(w, "partials/order-form.html", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// Data structures for templates

type GameData struct {
	ID            int            `json:"id"`
	Name          string         `json:"name"`
	Status        string         `json:"status"`
	Phase         string         `json:"phase"`
	Year          int            `json:"year"`
	Players       []PlayerData   `json:"players"`
	Units         []UnitData     `json:"units"`
	Provinces     []ProvinceData `json:"provinces"`
	SupplyCenters []string       `json:"supply_centers"`
}

type PlayerData struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Nation   string `json:"nation"`
	Status   string `json:"status"`
	IsOnline bool   `json:"is_online"`
}

type UnitData struct {
	ID        int    `json:"id"`
	Type      string `json:"type"` // "army" or "fleet"
	Province  string `json:"province"`
	Nation    string `json:"nation"`
	CanMove   bool   `json:"can_move"`
	HasOrders bool   `json:"has_orders"`
}

type ProvinceData struct {
	ID             string    `json:"id"`
	Name           string    `json:"name"`
	Type           string    `json:"type"` // "land", "sea", "coast"
	IsSupplyCenter bool      `json:"is_supply_center"`
	Owner          string    `json:"owner,omitempty"`
	Unit           *UnitData `json:"unit,omitempty"`
}

type OrderData struct {
	ID      int    `json:"id"`
	Type    string `json:"type"`
	Unit    string `json:"unit"`
	From    string `json:"from"`
	To      string `json:"to,omitempty"`
	Support string `json:"support,omitempty"`
	Status  string `json:"status"`
}

// Data access methods (placeholder implementations)

func (h *TemplateHandler) getGamesData() []GameData {
	// TODO: Implement actual database queries
	return []GameData{
		{
			ID:     1,
			Name:   "Classic Game 1",
			Status: "active",
			Phase:  "Spring 1901 Movement",
			Year:   1901,
			Players: []PlayerData{
				{ID: 1, Name: "Player 1", Nation: "England", Status: "ready", IsOnline: true},
				{ID: 2, Name: "Player 2", Nation: "France", Status: "waiting", IsOnline: false},
			},
		},
	}
}

func (h *TemplateHandler) getGameData(gameID int) *GameData {
	// TODO: Implement actual database query
	games := h.getGamesData()
	for _, game := range games {
		if game.ID == gameID {
			return &game
		}
	}
	return nil
}

func (h *TemplateHandler) getProvinceUnits(provinceID string) []UnitData {
	// TODO: Implement actual database query
	return []UnitData{}
}

func (h *TemplateHandler) getProvinceOrders(provinceID string) []OrderData {
	// TODO: Implement actual database query
	return []OrderData{}
}

func (h *TemplateHandler) getPlayerUnits(gameID int) []UnitData {
	// TODO: Implement actual database query
	return []UnitData{}
}

func (h *TemplateHandler) getProvinces() []ProvinceData {
	// TODO: Implement actual database query
	return []ProvinceData{}
}
