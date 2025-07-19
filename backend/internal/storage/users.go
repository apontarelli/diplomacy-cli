package storage

import (
	"database/sql"
	"fmt"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// UserService provides user management operations
type UserService struct {
	db *Database
}

// NewUserService creates a new user service
func NewUserService(db *Database) *UserService {
	return &UserService{db: db}
}

// RegisterUser creates a new user with hashed password
func (us *UserService) RegisterUser(username, email, password, displayName string) (*User, error) {
	// Hash the password using bcrypt
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Create user in database
	query := `
		INSERT INTO users (username, email, password_hash, display_name, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?)
	`

	now := time.Now()
	result, err := us.db.Exec(query, username, email, string(hashedPassword), displayName, now, now)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("failed to get user ID: %w", err)
	}

	return &User{
		ID:           id,
		Username:     username,
		Email:        email,
		PasswordHash: string(hashedPassword),
		DisplayName:  displayName,
		CreatedAt:    now,
		UpdatedAt:    now,
	}, nil
}

// GetUserByID retrieves a user by their ID
func (us *UserService) GetUserByID(id int64) (*User, error) {
	query := `
		SELECT id, username, email, password_hash, display_name, created_at, updated_at
		FROM users
		WHERE id = ?
	`

	user := &User{}
	err := us.db.QueryRow(query, id).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.PasswordHash,
		&user.DisplayName,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return user, nil
}

// GetUserByUsername retrieves a user by their username
func (us *UserService) GetUserByUsername(username string) (*User, error) {
	query := `
		SELECT id, username, email, password_hash, display_name, created_at, updated_at
		FROM users
		WHERE username = ?
	`

	user := &User{}
	err := us.db.QueryRow(query, username).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.PasswordHash,
		&user.DisplayName,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return user, nil
}

// GetUserByEmail retrieves a user by their email
func (us *UserService) GetUserByEmail(email string) (*User, error) {
	query := `
		SELECT id, username, email, password_hash, display_name, created_at, updated_at
		FROM users
		WHERE email = ?
	`

	user := &User{}
	err := us.db.QueryRow(query, email).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.PasswordHash,
		&user.DisplayName,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return user, nil
}

// UpdateUser updates user information
func (us *UserService) UpdateUser(id int64, username, email, displayName string) (*User, error) {
	query := `
		UPDATE users 
		SET username = ?, email = ?, display_name = ?, updated_at = ?
		WHERE id = ?
	`

	now := time.Now()
	_, err := us.db.Exec(query, username, email, displayName, now, id)
	if err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	return us.GetUserByID(id)
}

// UpdateUserPassword updates a user's password
func (us *UserService) UpdateUserPassword(id int64, newPassword string) error {
	// Hash the new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	query := `
		UPDATE users 
		SET password_hash = ?, updated_at = ?
		WHERE id = ?
	`

	_, err = us.db.Exec(query, string(hashedPassword), time.Now(), id)
	if err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	return nil
}

// DeleteUser deletes a user by ID
func (us *UserService) DeleteUser(id int64) error {
	query := `DELETE FROM users WHERE id = ?`

	result, err := us.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("user not found")
	}

	return nil
}

// Authenticate verifies user credentials
func (us *UserService) Authenticate(username, password string) (*User, error) {
	user, err := us.GetUserByUsername(username)
	if err != nil {
		return nil, fmt.Errorf("authentication failed: %w", err)
	}

	// Compare the provided password with the stored hash
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		return nil, fmt.Errorf("authentication failed: invalid credentials")
	}

	return user, nil
}

// AuthenticateByEmail verifies user credentials using email
func (us *UserService) AuthenticateByEmail(email, password string) (*User, error) {
	user, err := us.GetUserByEmail(email)
	if err != nil {
		return nil, fmt.Errorf("authentication failed: %w", err)
	}

	// Compare the provided password with the stored hash
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		return nil, fmt.Errorf("authentication failed: invalid credentials")
	}

	return user, nil
}

// ListUsers retrieves all users (for admin purposes)
func (us *UserService) ListUsers(limit, offset int) ([]*User, error) {
	query := `
		SELECT id, username, email, password_hash, display_name, created_at, updated_at
		FROM users
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`

	rows, err := us.db.Query(query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list users: %w", err)
	}
	defer rows.Close()

	var users []*User
	for rows.Next() {
		user := &User{}
		err := rows.Scan(
			&user.ID,
			&user.Username,
			&user.Email,
			&user.PasswordHash,
			&user.DisplayName,
			&user.CreatedAt,
			&user.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan user: %w", err)
		}
		users = append(users, user)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating users: %w", err)
	}

	return users, nil
}

// UserExists checks if a user exists by username or email
func (us *UserService) UserExists(username, email string) (bool, error) {
	query := `
		SELECT COUNT(*) 
		FROM users 
		WHERE username = ? OR email = ?
	`

	var count int
	err := us.db.QueryRow(query, username, email).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to check user existence: %w", err)
	}

	return count > 0, nil
}
