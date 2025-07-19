package storage

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// JWTConfig holds JWT configuration
type JWTConfig struct {
	SecretKey       string
	ExpirationHours int
	Issuer          string
}

// DefaultJWTConfig returns a default JWT configuration
func DefaultJWTConfig() JWTConfig {
	return JWTConfig{
		SecretKey:       "your-secret-key-change-this-in-production",
		ExpirationHours: 24, // 24 hours
		Issuer:          "diplomacy-cli",
	}
}

// Claims represents the JWT claims
type Claims struct {
	UserID      int64  `json:"user_id"`
	Username    string `json:"username"`
	Email       string `json:"email"`
	DisplayName string `json:"display_name"`
	jwt.RegisteredClaims
}

// AuthService provides authentication operations
type AuthService struct {
	userService *UserService
	jwtConfig   JWTConfig
}

// NewAuthService creates a new authentication service
func NewAuthService(userService *UserService, jwtConfig JWTConfig) *AuthService {
	return &AuthService{
		userService: userService,
		jwtConfig:   jwtConfig,
	}
}

// Login authenticates a user and returns a JWT token
func (as *AuthService) Login(username, password string) (string, *User, error) {
	// Authenticate user
	user, err := as.userService.Authenticate(username, password)
	if err != nil {
		return "", nil, fmt.Errorf("login failed: %w", err)
	}

	// Generate JWT token
	token, err := as.GenerateToken(user)
	if err != nil {
		return "", nil, fmt.Errorf("failed to generate token: %w", err)
	}

	return token, user, nil
}

// LoginByEmail authenticates a user by email and returns a JWT token
func (as *AuthService) LoginByEmail(email, password string) (string, *User, error) {
	// Authenticate user by email
	user, err := as.userService.AuthenticateByEmail(email, password)
	if err != nil {
		return "", nil, fmt.Errorf("login failed: %w", err)
	}

	// Generate JWT token
	token, err := as.GenerateToken(user)
	if err != nil {
		return "", nil, fmt.Errorf("failed to generate token: %w", err)
	}

	return token, user, nil
}

// GenerateToken creates a JWT token for a user
func (as *AuthService) GenerateToken(user *User) (string, error) {
	// Create claims
	claims := Claims{
		UserID:      user.ID,
		Username:    user.Username,
		Email:       user.Email,
		DisplayName: user.DisplayName,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(as.jwtConfig.ExpirationHours) * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    as.jwtConfig.Issuer,
			Subject:   fmt.Sprintf("%d", user.ID),
		},
	}

	// Create token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Sign token
	tokenString, err := token.SignedString([]byte(as.jwtConfig.SecretKey))
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return tokenString, nil
}

// ValidateToken validates a JWT token and returns the claims
func (as *AuthService) ValidateToken(tokenString string) (*Claims, error) {
	// Parse token
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// Validate signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(as.jwtConfig.SecretKey), nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	// Validate token and extract claims
	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("invalid token")
}

// RefreshToken generates a new token for a user (if the current token is valid)
func (as *AuthService) RefreshToken(tokenString string) (string, error) {
	// Validate current token
	claims, err := as.ValidateToken(tokenString)
	if err != nil {
		return "", fmt.Errorf("cannot refresh invalid token: %w", err)
	}

	// Get user from database to ensure they still exist
	user, err := as.userService.GetUserByID(claims.UserID)
	if err != nil {
		return "", fmt.Errorf("user not found: %w", err)
	}

	// Generate new token
	newToken, err := as.GenerateToken(user)
	if err != nil {
		return "", fmt.Errorf("failed to generate new token: %w", err)
	}

	return newToken, nil
}

// GetUserFromToken extracts user information from a valid JWT token
func (as *AuthService) GetUserFromToken(tokenString string) (*User, error) {
	// Validate token
	claims, err := as.ValidateToken(tokenString)
	if err != nil {
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	// Get user from database
	user, err := as.userService.GetUserByID(claims.UserID)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	return user, nil
}

// RevokeToken marks a token as revoked (for future implementation with token blacklist)
func (as *AuthService) RevokeToken(tokenString string) error {
	// Validate token first
	_, err := as.ValidateToken(tokenString)
	if err != nil {
		return fmt.Errorf("cannot revoke invalid token: %w", err)
	}

	// TODO: Implement token blacklist in database
	// For now, we just validate that the token is valid
	// In a production system, you would store revoked tokens in a blacklist table

	return nil
}

// ChangePassword changes a user's password and optionally revokes all existing tokens
func (as *AuthService) ChangePassword(userID int64, oldPassword, newPassword string) error {
	// Get user
	user, err := as.userService.GetUserByID(userID)
	if err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	// Verify old password
	_, err = as.userService.Authenticate(user.Username, oldPassword)
	if err != nil {
		return fmt.Errorf("invalid old password: %w", err)
	}

	// Update password
	err = as.userService.UpdateUserPassword(userID, newPassword)
	if err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	// TODO: In a production system, you might want to revoke all existing tokens
	// when a password is changed for security reasons

	return nil
}

// Register creates a new user account and returns a JWT token
func (as *AuthService) Register(username, email, password, displayName string) (string, *User, error) {
	// Check if user already exists
	exists, err := as.userService.UserExists(username, email)
	if err != nil {
		return "", nil, fmt.Errorf("failed to check user existence: %w", err)
	}
	if exists {
		return "", nil, fmt.Errorf("user already exists")
	}

	// Register user
	user, err := as.userService.RegisterUser(username, email, password, displayName)
	if err != nil {
		return "", nil, fmt.Errorf("failed to register user: %w", err)
	}

	// Generate JWT token
	token, err := as.GenerateToken(user)
	if err != nil {
		return "", nil, fmt.Errorf("failed to generate token: %w", err)
	}

	return token, user, nil
}

// TokenInfo represents information about a JWT token
type TokenInfo struct {
	Valid     bool      `json:"valid"`
	UserID    int64     `json:"user_id,omitempty"`
	Username  string    `json:"username,omitempty"`
	Email     string    `json:"email,omitempty"`
	ExpiresAt time.Time `json:"expires_at,omitempty"`
	IssuedAt  time.Time `json:"issued_at,omitempty"`
}

// GetTokenInfo returns information about a JWT token without requiring it to be valid
func (as *AuthService) GetTokenInfo(tokenString string) TokenInfo {
	claims, err := as.ValidateToken(tokenString)
	if err != nil {
		return TokenInfo{Valid: false}
	}

	return TokenInfo{
		Valid:     true,
		UserID:    claims.UserID,
		Username:  claims.Username,
		Email:     claims.Email,
		ExpiresAt: claims.ExpiresAt.Time,
		IssuedAt:  claims.IssuedAt.Time,
	}
}
