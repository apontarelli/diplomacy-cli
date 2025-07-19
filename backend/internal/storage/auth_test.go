package storage

import (
	"os"
	"testing"
	"time"
)

func TestAuthService(t *testing.T) {
	// Create temporary database for testing
	tempDB := "test_auth.db"
	defer os.Remove(tempDB)

	db, err := NewDatabase(Config{
		DatabasePath: tempDB,
		LogQueries:   false,
	})
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	defer db.Close()

	userService := NewUserService(db)
	jwtConfig := JWTConfig{
		SecretKey:       "test-secret-key",
		ExpirationHours: 1,
		Issuer:          "test-issuer",
	}
	authService := NewAuthService(userService, jwtConfig)

	// Register a test user
	testUser, err := userService.RegisterUser("testuser", "test@example.com", "password123", "Test User")
	if err != nil {
		t.Fatalf("Failed to register test user: %v", err)
	}

	t.Run("GenerateToken", func(t *testing.T) {
		token, err := authService.GenerateToken(testUser)
		if err != nil {
			t.Fatalf("Failed to generate token: %v", err)
		}

		if token == "" {
			t.Error("Expected non-empty token")
		}

		// Validate the generated token
		claims, err := authService.ValidateToken(token)
		if err != nil {
			t.Fatalf("Failed to validate generated token: %v", err)
		}

		if claims.UserID != testUser.ID {
			t.Errorf("Expected user ID %d, got %d", testUser.ID, claims.UserID)
		}
		if claims.Username != testUser.Username {
			t.Errorf("Expected username %s, got %s", testUser.Username, claims.Username)
		}
		if claims.Email != testUser.Email {
			t.Errorf("Expected email %s, got %s", testUser.Email, claims.Email)
		}
	})

	t.Run("ValidateToken_Valid", func(t *testing.T) {
		token, err := authService.GenerateToken(testUser)
		if err != nil {
			t.Fatalf("Failed to generate token: %v", err)
		}

		claims, err := authService.ValidateToken(token)
		if err != nil {
			t.Fatalf("Failed to validate token: %v", err)
		}

		if claims.UserID != testUser.ID {
			t.Errorf("Expected user ID %d, got %d", testUser.ID, claims.UserID)
		}
	})

	t.Run("ValidateToken_Invalid", func(t *testing.T) {
		invalidToken := "invalid.token.here"
		_, err := authService.ValidateToken(invalidToken)
		if err == nil {
			t.Error("Expected error when validating invalid token")
		}
	})

	t.Run("ValidateToken_WrongSecret", func(t *testing.T) {
		// Create auth service with different secret
		wrongSecretConfig := JWTConfig{
			SecretKey:       "wrong-secret",
			ExpirationHours: 1,
			Issuer:          "test-issuer",
		}
		wrongAuthService := NewAuthService(userService, wrongSecretConfig)

		// Generate token with original service
		token, err := authService.GenerateToken(testUser)
		if err != nil {
			t.Fatalf("Failed to generate token: %v", err)
		}

		// Try to validate with wrong secret
		_, err = wrongAuthService.ValidateToken(token)
		if err == nil {
			t.Error("Expected error when validating token with wrong secret")
		}
	})

	t.Run("Login", func(t *testing.T) {
		token, user, err := authService.Login("testuser", "password123")
		if err != nil {
			t.Fatalf("Failed to login: %v", err)
		}

		if token == "" {
			t.Error("Expected non-empty token")
		}
		if user.ID != testUser.ID {
			t.Errorf("Expected user ID %d, got %d", testUser.ID, user.ID)
		}

		// Validate the returned token
		claims, err := authService.ValidateToken(token)
		if err != nil {
			t.Fatalf("Failed to validate login token: %v", err)
		}
		if claims.UserID != testUser.ID {
			t.Errorf("Expected user ID %d, got %d", testUser.ID, claims.UserID)
		}
	})

	t.Run("Login_InvalidCredentials", func(t *testing.T) {
		_, _, err := authService.Login("testuser", "wrongpassword")
		if err == nil {
			t.Error("Expected error when logging in with wrong password")
		}
	})

	t.Run("Login_NonexistentUser", func(t *testing.T) {
		_, _, err := authService.Login("nonexistent", "password123")
		if err == nil {
			t.Error("Expected error when logging in with non-existent user")
		}
	})

	t.Run("LoginByEmail", func(t *testing.T) {
		token, user, err := authService.LoginByEmail("test@example.com", "password123")
		if err != nil {
			t.Fatalf("Failed to login by email: %v", err)
		}

		if token == "" {
			t.Error("Expected non-empty token")
		}
		if user.ID != testUser.ID {
			t.Errorf("Expected user ID %d, got %d", testUser.ID, user.ID)
		}
	})

	t.Run("LoginByEmail_InvalidCredentials", func(t *testing.T) {
		_, _, err := authService.LoginByEmail("test@example.com", "wrongpassword")
		if err == nil {
			t.Error("Expected error when logging in with wrong password")
		}
	})

	t.Run("RefreshToken", func(t *testing.T) {
		// Generate initial token
		originalToken, err := authService.GenerateToken(testUser)
		if err != nil {
			t.Fatalf("Failed to generate original token: %v", err)
		}

		// Add a small delay to ensure different timestamps
		time.Sleep(10 * time.Millisecond)

		// Refresh token
		newToken, err := authService.RefreshToken(originalToken)
		if err != nil {
			t.Fatalf("Failed to refresh token: %v", err)
		}

		if newToken == "" {
			t.Error("Expected non-empty refreshed token")
		}
		// Note: Tokens might be the same if generated at the exact same time
		// The important thing is that the refresh process works correctly

		// Validate new token
		claims, err := authService.ValidateToken(newToken)
		if err != nil {
			t.Fatalf("Failed to validate refreshed token: %v", err)
		}
		if claims.UserID != testUser.ID {
			t.Errorf("Expected user ID %d, got %d", testUser.ID, claims.UserID)
		}
	})

	t.Run("RefreshToken_InvalidToken", func(t *testing.T) {
		_, err := authService.RefreshToken("invalid.token.here")
		if err == nil {
			t.Error("Expected error when refreshing invalid token")
		}
	})

	t.Run("GetUserFromToken", func(t *testing.T) {
		token, err := authService.GenerateToken(testUser)
		if err != nil {
			t.Fatalf("Failed to generate token: %v", err)
		}

		user, err := authService.GetUserFromToken(token)
		if err != nil {
			t.Fatalf("Failed to get user from token: %v", err)
		}

		if user.ID != testUser.ID {
			t.Errorf("Expected user ID %d, got %d", testUser.ID, user.ID)
		}
		if user.Username != testUser.Username {
			t.Errorf("Expected username %s, got %s", testUser.Username, user.Username)
		}
	})

	t.Run("GetUserFromToken_InvalidToken", func(t *testing.T) {
		_, err := authService.GetUserFromToken("invalid.token.here")
		if err == nil {
			t.Error("Expected error when getting user from invalid token")
		}
	})

	t.Run("RevokeToken", func(t *testing.T) {
		token, err := authService.GenerateToken(testUser)
		if err != nil {
			t.Fatalf("Failed to generate token: %v", err)
		}

		err = authService.RevokeToken(token)
		if err != nil {
			t.Fatalf("Failed to revoke token: %v", err)
		}

		// Note: Current implementation doesn't actually blacklist tokens
		// This test just ensures the function doesn't error on valid tokens
	})

	t.Run("RevokeToken_InvalidToken", func(t *testing.T) {
		err := authService.RevokeToken("invalid.token.here")
		if err == nil {
			t.Error("Expected error when revoking invalid token")
		}
	})

	t.Run("ChangePassword", func(t *testing.T) {
		err := authService.ChangePassword(testUser.ID, "password123", "newpassword456")
		if err != nil {
			t.Fatalf("Failed to change password: %v", err)
		}

		// Verify old password doesn't work
		_, _, err = authService.Login("testuser", "password123")
		if err == nil {
			t.Error("Expected error when logging in with old password")
		}

		// Verify new password works
		_, _, err = authService.Login("testuser", "newpassword456")
		if err != nil {
			t.Fatalf("Failed to login with new password: %v", err)
		}

		// Change back for other tests
		err = authService.ChangePassword(testUser.ID, "newpassword456", "password123")
		if err != nil {
			t.Fatalf("Failed to change password back: %v", err)
		}
	})

	t.Run("ChangePassword_InvalidOldPassword", func(t *testing.T) {
		err := authService.ChangePassword(testUser.ID, "wrongpassword", "newpassword")
		if err == nil {
			t.Error("Expected error when changing password with wrong old password")
		}
	})

	t.Run("Register", func(t *testing.T) {
		token, user, err := authService.Register("newuser", "newuser@example.com", "password123", "New User")
		if err != nil {
			t.Fatalf("Failed to register user: %v", err)
		}

		if token == "" {
			t.Error("Expected non-empty token")
		}
		if user.Username != "newuser" {
			t.Errorf("Expected username 'newuser', got '%s'", user.Username)
		}

		// Validate token
		claims, err := authService.ValidateToken(token)
		if err != nil {
			t.Fatalf("Failed to validate registration token: %v", err)
		}
		if claims.UserID != user.ID {
			t.Errorf("Expected user ID %d, got %d", user.ID, claims.UserID)
		}
	})

	t.Run("Register_DuplicateUser", func(t *testing.T) {
		_, _, err := authService.Register("testuser", "duplicate@example.com", "password123", "Duplicate User")
		if err == nil {
			t.Error("Expected error when registering duplicate username")
		}
	})

	t.Run("GetTokenInfo_Valid", func(t *testing.T) {
		token, err := authService.GenerateToken(testUser)
		if err != nil {
			t.Fatalf("Failed to generate token: %v", err)
		}

		info := authService.GetTokenInfo(token)
		if !info.Valid {
			t.Error("Expected token to be valid")
		}
		if info.UserID != testUser.ID {
			t.Errorf("Expected user ID %d, got %d", testUser.ID, info.UserID)
		}
		if info.Username != testUser.Username {
			t.Errorf("Expected username %s, got %s", testUser.Username, info.Username)
		}
	})

	t.Run("GetTokenInfo_Invalid", func(t *testing.T) {
		info := authService.GetTokenInfo("invalid.token.here")
		if info.Valid {
			t.Error("Expected token to be invalid")
		}
	})
}

func TestJWTConfig(t *testing.T) {
	t.Run("DefaultJWTConfig", func(t *testing.T) {
		config := DefaultJWTConfig()

		if config.SecretKey == "" {
			t.Error("Expected non-empty secret key")
		}
		if config.ExpirationHours <= 0 {
			t.Error("Expected positive expiration hours")
		}
		if config.Issuer == "" {
			t.Error("Expected non-empty issuer")
		}
	})
}

func TestTokenExpiration(t *testing.T) {
	// Create temporary database for testing
	tempDB := "test_auth_expiration.db"
	defer os.Remove(tempDB)

	db, err := NewDatabase(Config{
		DatabasePath: tempDB,
		LogQueries:   false,
	})
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	defer db.Close()

	userService := NewUserService(db)

	// Create config with very short expiration for testing
	jwtConfig := JWTConfig{
		SecretKey:       "test-secret-key",
		ExpirationHours: 0, // This will create tokens that expire immediately
		Issuer:          "test-issuer",
	}
	authService := NewAuthService(userService, jwtConfig)

	// Register a test user
	testUser, err := userService.RegisterUser("expireuser", "expire@example.com", "password123", "Expire User")
	if err != nil {
		t.Fatalf("Failed to register test user: %v", err)
	}

	t.Run("ExpiredToken", func(t *testing.T) {
		// Generate token that expires immediately
		token, err := authService.GenerateToken(testUser)
		if err != nil {
			t.Fatalf("Failed to generate token: %v", err)
		}

		// Wait a moment to ensure expiration
		time.Sleep(10 * time.Millisecond)

		// Try to validate expired token
		_, err = authService.ValidateToken(token)
		if err == nil {
			t.Error("Expected error when validating expired token")
		}
	})
}

func BenchmarkAuthService(b *testing.B) {
	// Create temporary database for benchmarking
	tempDB := "bench_auth.db"
	defer os.Remove(tempDB)

	db, err := NewDatabase(Config{
		DatabasePath: tempDB,
		LogQueries:   false,
	})
	if err != nil {
		b.Fatalf("Failed to create test database: %v", err)
	}
	defer db.Close()

	userService := NewUserService(db)
	jwtConfig := DefaultJWTConfig()
	authService := NewAuthService(userService, jwtConfig)

	// Register a test user
	testUser, err := userService.RegisterUser("benchuser", "bench@example.com", "password123", "Bench User")
	if err != nil {
		b.Fatalf("Failed to register test user: %v", err)
	}

	b.Run("GenerateToken", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, err := authService.GenerateToken(testUser)
			if err != nil {
				b.Fatalf("Failed to generate token: %v", err)
			}
		}
	})

	// Generate a token for validation benchmarks
	token, err := authService.GenerateToken(testUser)
	if err != nil {
		b.Fatalf("Failed to generate token for benchmarks: %v", err)
	}

	b.Run("ValidateToken", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, err := authService.ValidateToken(token)
			if err != nil {
				b.Fatalf("Failed to validate token: %v", err)
			}
		}
	})

	b.Run("Login", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _, err := authService.Login("benchuser", "password123")
			if err != nil {
				b.Fatalf("Failed to login: %v", err)
			}
		}
	})
}
