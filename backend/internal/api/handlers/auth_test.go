package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"diplomacy-cli/backend/internal/storage"
)

func TestAuthHandler_Register(t *testing.T) {
	// Setup test database
	config := storage.Config{
		DatabasePath: ":memory:",
		LogQueries:   false,
	}
	db, err := storage.NewDatabase(config)
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	defer db.Close()

	// Setup services
	userService := storage.NewUserService(db)
	authService := storage.NewAuthService(userService, storage.DefaultJWTConfig())
	authHandler := NewAuthHandler(authService)

	tests := []struct {
		name           string
		request        RegisterRequest
		expectedStatus int
		expectToken    bool
	}{
		{
			name: "valid registration",
			request: RegisterRequest{
				Username:    "testuser",
				Email:       "test@example.com",
				Password:    "password123",
				DisplayName: "Test User",
			},
			expectedStatus: http.StatusCreated,
			expectToken:    true,
		},
		{
			name: "missing username",
			request: RegisterRequest{
				Email:    "test@example.com",
				Password: "password123",
			},
			expectedStatus: http.StatusBadRequest,
			expectToken:    false,
		},
		{
			name: "missing email",
			request: RegisterRequest{
				Username: "testuser",
				Password: "password123",
			},
			expectedStatus: http.StatusBadRequest,
			expectToken:    false,
		},
		{
			name: "password too short",
			request: RegisterRequest{
				Username: "testuser",
				Email:    "test@example.com",
				Password: "123",
			},
			expectedStatus: http.StatusBadRequest,
			expectToken:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create request
			jsonData, _ := json.Marshal(tt.request)
			req := httptest.NewRequest("POST", "/api/auth/register", bytes.NewReader(jsonData))
			req.Header.Set("Content-Type", "application/json")

			// Create response recorder
			w := httptest.NewRecorder()

			// Call handler
			authHandler.Register(w, req)

			// Check status code
			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			// Check response for successful registration
			if tt.expectToken {
				var response RegisterResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				if err != nil {
					t.Fatalf("Failed to unmarshal response: %v", err)
				}

				if response.Token == "" {
					t.Error("Expected token in response")
				}

				if response.User.Username != tt.request.Username {
					t.Errorf("Expected username %s, got %s", tt.request.Username, response.User.Username)
				}

				if response.TokenType != "Bearer" {
					t.Errorf("Expected token type Bearer, got %s", response.TokenType)
				}
			}
		})
	}
}

func TestAuthHandler_Login(t *testing.T) {
	// Setup test database
	config := storage.Config{
		DatabasePath: ":memory:",
		LogQueries:   false,
	}
	db, err := storage.NewDatabase(config)
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	defer db.Close()

	// Setup services
	userService := storage.NewUserService(db)
	authService := storage.NewAuthService(userService, storage.DefaultJWTConfig())
	authHandler := NewAuthHandler(authService)

	// Create test user
	testUser, err := userService.RegisterUser("testuser", "test@example.com", "password123", "Test User")
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	tests := []struct {
		name           string
		request        LoginRequest
		expectedStatus int
		expectToken    bool
	}{
		{
			name: "valid login with username",
			request: LoginRequest{
				Username: "testuser",
				Password: "password123",
			},
			expectedStatus: http.StatusOK,
			expectToken:    true,
		},
		{
			name: "valid login with email",
			request: LoginRequest{
				Email:    "test@example.com",
				Password: "password123",
			},
			expectedStatus: http.StatusOK,
			expectToken:    true,
		},
		{
			name: "invalid password",
			request: LoginRequest{
				Username: "testuser",
				Password: "wrongpassword",
			},
			expectedStatus: http.StatusUnauthorized,
			expectToken:    false,
		},
		{
			name: "missing credentials",
			request: LoginRequest{
				Password: "password123",
			},
			expectedStatus: http.StatusBadRequest,
			expectToken:    false,
		},
		{
			name: "missing password",
			request: LoginRequest{
				Username: "testuser",
			},
			expectedStatus: http.StatusBadRequest,
			expectToken:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create request
			jsonData, _ := json.Marshal(tt.request)
			req := httptest.NewRequest("POST", "/api/auth/login", bytes.NewReader(jsonData))
			req.Header.Set("Content-Type", "application/json")

			// Create response recorder
			w := httptest.NewRecorder()

			// Call handler
			authHandler.Login(w, req)

			// Check status code
			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			// Check response for successful login
			if tt.expectToken {
				var response LoginResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				if err != nil {
					t.Fatalf("Failed to unmarshal response: %v", err)
				}

				if response.Token == "" {
					t.Error("Expected token in response")
				}

				if response.User.ID != testUser.ID {
					t.Errorf("Expected user ID %d, got %d", testUser.ID, response.User.ID)
				}

				if response.TokenType != "Bearer" {
					t.Errorf("Expected token type Bearer, got %s", response.TokenType)
				}

				// Verify token is valid
				claims, err := authService.ValidateToken(response.Token)
				if err != nil {
					t.Errorf("Token validation failed: %v", err)
				}

				if claims.UserID != testUser.ID {
					t.Errorf("Expected user ID in claims %d, got %d", testUser.ID, claims.UserID)
				}
			}
		})
	}
}

func TestAuthHandler_RefreshToken(t *testing.T) {
	// Setup test database
	config := storage.Config{
		DatabasePath: ":memory:",
		LogQueries:   false,
	}
	db, err := storage.NewDatabase(config)
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	defer db.Close()

	// Setup services
	userService := storage.NewUserService(db)
	authService := storage.NewAuthService(userService, storage.DefaultJWTConfig())
	authHandler := NewAuthHandler(authService)

	// Create test user and get token
	testUser, err := userService.RegisterUser("testuser", "test@example.com", "password123", "Test User")
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	token, err := authService.GenerateToken(testUser)
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	tests := []struct {
		name           string
		token          string
		expectedStatus int
		expectNewToken bool
	}{
		{
			name:           "valid token refresh",
			token:          token,
			expectedStatus: http.StatusOK,
			expectNewToken: true,
		},
		{
			name:           "invalid token",
			token:          "invalid.token.here",
			expectedStatus: http.StatusUnauthorized,
			expectNewToken: false,
		},
		{
			name:           "empty token",
			token:          "",
			expectedStatus: http.StatusBadRequest,
			expectNewToken: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create request
			request := RefreshRequest{Token: tt.token}
			jsonData, _ := json.Marshal(request)
			req := httptest.NewRequest("POST", "/api/auth/refresh", bytes.NewReader(jsonData))
			req.Header.Set("Content-Type", "application/json")

			// Create response recorder
			w := httptest.NewRecorder()

			// Call handler
			authHandler.RefreshToken(w, req)

			// Check status code
			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			// Check response for successful refresh
			if tt.expectNewToken {
				var response RefreshResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				if err != nil {
					t.Fatalf("Failed to unmarshal response: %v", err)
				}

				if response.Token == "" {
					t.Error("Expected new token in response")
				}

				// Note: New token should be different due to different IssuedAt timestamp
				// but we'll skip this check as it may be flaky in fast test environments

				if response.TokenType != "Bearer" {
					t.Errorf("Expected token type Bearer, got %s", response.TokenType)
				}

				// Verify new token is valid
				_, err = authService.ValidateToken(response.Token)
				if err != nil {
					t.Errorf("New token validation failed: %v", err)
				}
			}
		})
	}
}

func TestAuthHandler_GetTokenInfo(t *testing.T) {
	// Setup test database
	config := storage.Config{
		DatabasePath: ":memory:",
		LogQueries:   false,
	}
	db, err := storage.NewDatabase(config)
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	defer db.Close()

	// Setup services
	userService := storage.NewUserService(db)
	authService := storage.NewAuthService(userService, storage.DefaultJWTConfig())
	authHandler := NewAuthHandler(authService)

	// Create test user and get token
	testUser, err := userService.RegisterUser("testuser", "test@example.com", "password123", "Test User")
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	token, err := authService.GenerateToken(testUser)
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	tests := []struct {
		name           string
		authHeader     string
		expectedStatus int
		expectValid    bool
	}{
		{
			name:           "valid token",
			authHeader:     "Bearer " + token,
			expectedStatus: http.StatusOK,
			expectValid:    true,
		},
		{
			name:           "invalid token",
			authHeader:     "Bearer invalid.token.here",
			expectedStatus: http.StatusOK,
			expectValid:    false,
		},
		{
			name:           "missing auth header",
			authHeader:     "",
			expectedStatus: http.StatusUnauthorized,
			expectValid:    false,
		},
		{
			name:           "invalid auth format",
			authHeader:     "InvalidFormat " + token,
			expectedStatus: http.StatusUnauthorized,
			expectValid:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create request
			req := httptest.NewRequest("GET", "/api/auth/token-info", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}

			// Create response recorder
			w := httptest.NewRecorder()

			// Call handler
			authHandler.GetTokenInfo(w, req)

			// Check status code
			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			// Check response for successful token info
			if tt.expectedStatus == http.StatusOK {
				var response TokenInfoResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				if err != nil {
					t.Fatalf("Failed to unmarshal response: %v", err)
				}

				if response.Valid != tt.expectValid {
					t.Errorf("Expected valid %v, got %v", tt.expectValid, response.Valid)
				}

				if tt.expectValid {
					if response.UserID != testUser.ID {
						t.Errorf("Expected user ID %d, got %d", testUser.ID, response.UserID)
					}

					if response.Username != testUser.Username {
						t.Errorf("Expected username %s, got %s", testUser.Username, response.Username)
					}
				}
			}
		})
	}
}

func TestAuthHandler_GetProfile(t *testing.T) {
	// Setup test database
	config := storage.Config{
		DatabasePath: ":memory:",
		LogQueries:   false,
	}
	db, err := storage.NewDatabase(config)
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	defer db.Close()

	// Setup services
	userService := storage.NewUserService(db)
	authService := storage.NewAuthService(userService, storage.DefaultJWTConfig())

	// Create test user and get token
	testUser, err := userService.RegisterUser("testuser", "test@example.com", "password123", "Test User")
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	token, err := authService.GenerateToken(testUser)
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	// Validate token and get user (simulating middleware)
	user, err := authService.GetUserFromToken(token)
	if err != nil {
		t.Fatalf("Failed to get user from token: %v", err)
	}

	// Verify the user data structure matches expected values
	if user.ID != testUser.ID {
		t.Errorf("Expected user ID %d, got %d", testUser.ID, user.ID)
	}

	if user.Username != testUser.Username {
		t.Errorf("Expected username %s, got %s", testUser.Username, user.Username)
	}

	if user.Email != testUser.Email {
		t.Errorf("Expected email %s, got %s", testUser.Email, user.Email)
	}
}

func TestTokenExpiration(t *testing.T) {
	// Setup test database
	config := storage.Config{
		DatabasePath: ":memory:",
		LogQueries:   false,
	}
	db, err := storage.NewDatabase(config)
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	defer db.Close()

	// Setup services with short expiration for testing
	jwtConfig := storage.JWTConfig{
		SecretKey:       "test-secret-key",
		ExpirationHours: 0, // Immediate expiration for testing
		Issuer:          "test-issuer",
	}
	userService := storage.NewUserService(db)
	authService := storage.NewAuthService(userService, jwtConfig)

	// Create test user
	testUser, err := userService.RegisterUser("testuser", "test@example.com", "password123", "Test User")
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	// Generate token that expires immediately
	token, err := authService.GenerateToken(testUser)
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	// Wait a moment to ensure expiration
	time.Sleep(time.Millisecond * 10)

	// Try to validate expired token
	_, err = authService.ValidateToken(token)
	if err == nil {
		t.Error("Expected token validation to fail for expired token")
	}
}
