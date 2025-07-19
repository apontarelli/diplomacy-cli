package handlers

import (
	"net/http"
	"strings"
	"time"

	"diplomacy-cli/backend/internal/api/middleware"
	"diplomacy-cli/backend/internal/storage"
)

// AuthHandler handles authentication-related HTTP requests
type AuthHandler struct {
	authService *storage.AuthService
}

// NewAuthHandler creates a new authentication handler
func NewAuthHandler(authService *storage.AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

// LoginRequest represents the login request payload
type LoginRequest struct {
	Username string `json:"username,omitempty"`
	Email    string `json:"email,omitempty"`
	Password string `json:"password"`
}

// LoginResponse represents the login response payload
type LoginResponse struct {
	Token     string    `json:"token"`
	User      UserInfo  `json:"user"`
	ExpiresAt time.Time `json:"expires_at"`
	TokenType string    `json:"token_type"`
}

// RegisterRequest represents the registration request payload
type RegisterRequest struct {
	Username    string `json:"username"`
	Email       string `json:"email"`
	Password    string `json:"password"`
	DisplayName string `json:"display_name"`
}

// RegisterResponse represents the registration response payload
type RegisterResponse struct {
	Token     string    `json:"token"`
	User      UserInfo  `json:"user"`
	ExpiresAt time.Time `json:"expires_at"`
	TokenType string    `json:"token_type"`
}

// UserInfo represents user information in responses
type UserInfo struct {
	ID          int64     `json:"id"`
	Username    string    `json:"username"`
	Email       string    `json:"email"`
	DisplayName string    `json:"display_name"`
	CreatedAt   time.Time `json:"created_at"`
}

// RefreshRequest represents the token refresh request payload
type RefreshRequest struct {
	Token string `json:"token"`
}

// RefreshResponse represents the token refresh response payload
type RefreshResponse struct {
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
	TokenType string    `json:"token_type"`
}

// TokenInfoResponse represents token information response
type TokenInfoResponse struct {
	Valid     bool      `json:"valid"`
	UserID    int64     `json:"user_id,omitempty"`
	Username  string    `json:"username,omitempty"`
	Email     string    `json:"email,omitempty"`
	ExpiresAt time.Time `json:"expires_at,omitempty"`
	IssuedAt  time.Time `json:"issued_at,omitempty"`
}

// Login handles POST /api/auth/login
func (ah *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := middleware.ReadJSON(r, &req); err != nil {
		apiErr := middleware.NewValidationError("Invalid request body", "INVALID_JSON", nil)
		middleware.WriteError(w, apiErr)
		return
	}

	// Validate request
	if req.Password == "" {
		apiErr := middleware.NewValidationError("Password is required", "MISSING_PASSWORD", nil)
		middleware.WriteError(w, apiErr)
		return
	}

	if req.Username == "" && req.Email == "" {
		apiErr := middleware.NewValidationError("Username or email is required", "MISSING_CREDENTIALS", nil)
		middleware.WriteError(w, apiErr)
		return
	}

	// Authenticate user
	var token string
	var user *storage.User
	var err error

	if req.Email != "" {
		token, user, err = ah.authService.LoginByEmail(strings.TrimSpace(req.Email), req.Password)
	} else {
		token, user, err = ah.authService.Login(strings.TrimSpace(req.Username), req.Password)
	}

	if err != nil {
		apiErr := middleware.NewAuthError("Invalid credentials", "INVALID_CREDENTIALS")
		middleware.WriteError(w, apiErr)
		return
	}

	// Get token info for expiration
	tokenInfo := ah.authService.GetTokenInfo(token)

	// Create response
	response := LoginResponse{
		Token: token,
		User: UserInfo{
			ID:          user.ID,
			Username:    user.Username,
			Email:       user.Email,
			DisplayName: user.DisplayName,
			CreatedAt:   user.CreatedAt,
		},
		ExpiresAt: tokenInfo.ExpiresAt,
		TokenType: "Bearer",
	}

	middleware.WriteJSON(w, http.StatusOK, response)
}

// Register handles POST /api/auth/register
func (ah *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := middleware.ReadJSON(r, &req); err != nil {
		apiErr := middleware.NewValidationError("Invalid request body", "INVALID_JSON", nil)
		middleware.WriteError(w, apiErr)
		return
	}

	// Validate request
	req.Username = strings.TrimSpace(req.Username)
	req.Email = strings.TrimSpace(req.Email)
	req.DisplayName = strings.TrimSpace(req.DisplayName)

	if req.Username == "" {
		apiErr := middleware.NewValidationError("Username is required", "MISSING_USERNAME", nil)
		middleware.WriteError(w, apiErr)
		return
	}

	if req.Email == "" {
		apiErr := middleware.NewValidationError("Email is required", "MISSING_EMAIL", nil)
		middleware.WriteError(w, apiErr)
		return
	}

	if req.Password == "" {
		apiErr := middleware.NewValidationError("Password is required", "MISSING_PASSWORD", nil)
		middleware.WriteError(w, apiErr)
		return
	}

	if len(req.Password) < 6 {
		apiErr := middleware.NewValidationError("Password must be at least 6 characters", "PASSWORD_TOO_SHORT", nil)
		middleware.WriteError(w, apiErr)
		return
	}

	if req.DisplayName == "" {
		req.DisplayName = req.Username
	}

	// Register user
	token, user, err := ah.authService.Register(req.Username, req.Email, req.Password, req.DisplayName)
	if err != nil {
		if strings.Contains(err.Error(), "already exists") {
			apiErr := middleware.NewValidationError("User already exists", "USER_EXISTS", nil)
			middleware.WriteError(w, apiErr)
			return
		}
		middleware.WriteInternalError(w, err)
		return
	}

	// Get token info for expiration
	tokenInfo := ah.authService.GetTokenInfo(token)

	// Create response
	response := RegisterResponse{
		Token: token,
		User: UserInfo{
			ID:          user.ID,
			Username:    user.Username,
			Email:       user.Email,
			DisplayName: user.DisplayName,
			CreatedAt:   user.CreatedAt,
		},
		ExpiresAt: tokenInfo.ExpiresAt,
		TokenType: "Bearer",
	}

	middleware.WriteJSON(w, http.StatusCreated, response)
}

// RefreshToken handles POST /api/auth/refresh
func (ah *AuthHandler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	var req RefreshRequest
	if err := middleware.ReadJSON(r, &req); err != nil {
		apiErr := middleware.NewValidationError("Invalid request body", "INVALID_JSON", nil)
		middleware.WriteError(w, apiErr)
		return
	}

	// Validate request
	if req.Token == "" {
		apiErr := middleware.NewValidationError("Token is required", "MISSING_TOKEN", nil)
		middleware.WriteError(w, apiErr)
		return
	}

	// Refresh token
	newToken, err := ah.authService.RefreshToken(req.Token)
	if err != nil {
		apiErr := middleware.NewAuthError("Invalid or expired token", "INVALID_TOKEN")
		middleware.WriteError(w, apiErr)
		return
	}

	// Get token info for expiration
	tokenInfo := ah.authService.GetTokenInfo(newToken)

	// Create response
	response := RefreshResponse{
		Token:     newToken,
		ExpiresAt: tokenInfo.ExpiresAt,
		TokenType: "Bearer",
	}

	middleware.WriteJSON(w, http.StatusOK, response)
}

// GetTokenInfo handles GET /api/auth/token-info
func (ah *AuthHandler) GetTokenInfo(w http.ResponseWriter, r *http.Request) {
	// Extract token from Authorization header
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		apiErr := middleware.NewAuthError("Authorization header required", "MISSING_AUTH_HEADER")
		middleware.WriteError(w, apiErr)
		return
	}

	// Check for Bearer token format
	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || parts[0] != "Bearer" {
		apiErr := middleware.NewAuthError("Invalid authorization header format", "INVALID_AUTH_FORMAT")
		middleware.WriteError(w, apiErr)
		return
	}

	token := parts[1]

	// Get token info
	tokenInfo := ah.authService.GetTokenInfo(token)

	// Create response
	response := TokenInfoResponse{
		Valid:     tokenInfo.Valid,
		UserID:    tokenInfo.UserID,
		Username:  tokenInfo.Username,
		Email:     tokenInfo.Email,
		ExpiresAt: tokenInfo.ExpiresAt,
		IssuedAt:  tokenInfo.IssuedAt,
	}

	middleware.WriteJSON(w, http.StatusOK, response)
}

// Logout handles POST /api/auth/logout
func (ah *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	// Extract token from Authorization header
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		apiErr := middleware.NewAuthError("Authorization header required", "MISSING_AUTH_HEADER")
		middleware.WriteError(w, apiErr)
		return
	}

	// Check for Bearer token format
	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || parts[0] != "Bearer" {
		apiErr := middleware.NewAuthError("Invalid authorization header format", "INVALID_AUTH_FORMAT")
		middleware.WriteError(w, apiErr)
		return
	}

	token := parts[1]

	// Revoke token
	err := ah.authService.RevokeToken(token)
	if err != nil {
		apiErr := middleware.NewAuthError("Invalid token", "INVALID_TOKEN")
		middleware.WriteError(w, apiErr)
		return
	}

	// Return success response
	response := map[string]string{
		"message": "Successfully logged out",
	}

	middleware.WriteJSON(w, http.StatusOK, response)
}

// GetProfile handles GET /api/auth/profile (requires authentication)
func (ah *AuthHandler) GetProfile(w http.ResponseWriter, r *http.Request) {
	// Get user from context (set by RequireAuth middleware)
	user := middleware.MustGetUserFromContext(r.Context())

	// Create response
	response := UserInfo{
		ID:          user.ID,
		Username:    user.Username,
		Email:       user.Email,
		DisplayName: user.DisplayName,
		CreatedAt:   user.CreatedAt,
	}

	middleware.WriteJSON(w, http.StatusOK, response)
}
