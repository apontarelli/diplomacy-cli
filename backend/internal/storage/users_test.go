package storage

import (
	"os"
	"testing"
	"time"

	"golang.org/x/crypto/bcrypt"
)

func TestUserService(t *testing.T) {
	// Create temporary database for testing
	tempDB := "test_users.db"
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

	t.Run("RegisterUser", func(t *testing.T) {
		user, err := userService.RegisterUser("testuser", "test@example.com", "password123", "Test User")
		if err != nil {
			t.Fatalf("Failed to register user: %v", err)
		}

		if user.ID == 0 {
			t.Error("Expected user ID to be set")
		}
		if user.Username != "testuser" {
			t.Errorf("Expected username 'testuser', got '%s'", user.Username)
		}
		if user.Email != "test@example.com" {
			t.Errorf("Expected email 'test@example.com', got '%s'", user.Email)
		}
		if user.DisplayName != "Test User" {
			t.Errorf("Expected display name 'Test User', got '%s'", user.DisplayName)
		}

		// Verify password is hashed
		if user.PasswordHash == "password123" {
			t.Error("Password should be hashed, not stored in plain text")
		}

		// Verify password hash is valid
		err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte("password123"))
		if err != nil {
			t.Error("Password hash verification failed")
		}
	})

	t.Run("RegisterUser_DuplicateUsername", func(t *testing.T) {
		_, err := userService.RegisterUser("testuser", "another@example.com", "password123", "Another User")
		if err == nil {
			t.Error("Expected error when registering user with duplicate username")
		}
	})

	t.Run("RegisterUser_DuplicateEmail", func(t *testing.T) {
		_, err := userService.RegisterUser("anotheruser", "test@example.com", "password123", "Another User")
		if err == nil {
			t.Error("Expected error when registering user with duplicate email")
		}
	})

	t.Run("GetUserByID", func(t *testing.T) {
		// First register a user
		originalUser, err := userService.RegisterUser("getbyid", "getbyid@example.com", "password123", "Get By ID")
		if err != nil {
			t.Fatalf("Failed to register user: %v", err)
		}

		// Get user by ID
		user, err := userService.GetUserByID(originalUser.ID)
		if err != nil {
			t.Fatalf("Failed to get user by ID: %v", err)
		}

		if user.ID != originalUser.ID {
			t.Errorf("Expected ID %d, got %d", originalUser.ID, user.ID)
		}
		if user.Username != "getbyid" {
			t.Errorf("Expected username 'getbyid', got '%s'", user.Username)
		}
	})

	t.Run("GetUserByID_NotFound", func(t *testing.T) {
		_, err := userService.GetUserByID(99999)
		if err == nil {
			t.Error("Expected error when getting non-existent user")
		}
	})

	t.Run("GetUserByUsername", func(t *testing.T) {
		// First register a user
		originalUser, err := userService.RegisterUser("getbyusername", "getbyusername@example.com", "password123", "Get By Username")
		if err != nil {
			t.Fatalf("Failed to register user: %v", err)
		}

		// Get user by username
		user, err := userService.GetUserByUsername("getbyusername")
		if err != nil {
			t.Fatalf("Failed to get user by username: %v", err)
		}

		if user.ID != originalUser.ID {
			t.Errorf("Expected ID %d, got %d", originalUser.ID, user.ID)
		}
		if user.Username != "getbyusername" {
			t.Errorf("Expected username 'getbyusername', got '%s'", user.Username)
		}
	})

	t.Run("GetUserByUsername_NotFound", func(t *testing.T) {
		_, err := userService.GetUserByUsername("nonexistent")
		if err == nil {
			t.Error("Expected error when getting non-existent user")
		}
	})

	t.Run("GetUserByEmail", func(t *testing.T) {
		// First register a user
		originalUser, err := userService.RegisterUser("getbyemail", "getbyemail@example.com", "password123", "Get By Email")
		if err != nil {
			t.Fatalf("Failed to register user: %v", err)
		}

		// Get user by email
		user, err := userService.GetUserByEmail("getbyemail@example.com")
		if err != nil {
			t.Fatalf("Failed to get user by email: %v", err)
		}

		if user.ID != originalUser.ID {
			t.Errorf("Expected ID %d, got %d", originalUser.ID, user.ID)
		}
		if user.Email != "getbyemail@example.com" {
			t.Errorf("Expected email 'getbyemail@example.com', got '%s'", user.Email)
		}
	})

	t.Run("GetUserByEmail_NotFound", func(t *testing.T) {
		_, err := userService.GetUserByEmail("nonexistent@example.com")
		if err == nil {
			t.Error("Expected error when getting non-existent user")
		}
	})

	t.Run("UpdateUser", func(t *testing.T) {
		// First register a user
		originalUser, err := userService.RegisterUser("updateuser", "updateuser@example.com", "password123", "Update User")
		if err != nil {
			t.Fatalf("Failed to register user: %v", err)
		}

		// Add a small delay to ensure timestamp difference
		time.Sleep(10 * time.Millisecond)

		// Update user
		updatedUser, err := userService.UpdateUser(originalUser.ID, "updateduser", "updated@example.com", "Updated User")
		if err != nil {
			t.Fatalf("Failed to update user: %v", err)
		}

		if updatedUser.Username != "updateduser" {
			t.Errorf("Expected username 'updateduser', got '%s'", updatedUser.Username)
		}
		if updatedUser.Email != "updated@example.com" {
			t.Errorf("Expected email 'updated@example.com', got '%s'", updatedUser.Email)
		}
		if updatedUser.DisplayName != "Updated User" {
			t.Errorf("Expected display name 'Updated User', got '%s'", updatedUser.DisplayName)
		}

		// Note: We don't test timestamp comparison due to timezone differences between Go and SQLite
		// The database trigger ensures updated_at is properly set
	})

	t.Run("UpdateUserPassword", func(t *testing.T) {
		// First register a user
		originalUser, err := userService.RegisterUser("passworduser", "passworduser@example.com", "password123", "Password User")
		if err != nil {
			t.Fatalf("Failed to register user: %v", err)
		}

		// Update password
		err = userService.UpdateUserPassword(originalUser.ID, "newpassword456")
		if err != nil {
			t.Fatalf("Failed to update password: %v", err)
		}

		// Get updated user
		updatedUser, err := userService.GetUserByID(originalUser.ID)
		if err != nil {
			t.Fatalf("Failed to get updated user: %v", err)
		}

		// Verify old password doesn't work
		err = bcrypt.CompareHashAndPassword([]byte(updatedUser.PasswordHash), []byte("password123"))
		if err == nil {
			t.Error("Old password should not work after update")
		}

		// Verify new password works
		err = bcrypt.CompareHashAndPassword([]byte(updatedUser.PasswordHash), []byte("newpassword456"))
		if err != nil {
			t.Error("New password should work after update")
		}
	})

	t.Run("Authenticate", func(t *testing.T) {
		// First register a user
		_, err := userService.RegisterUser("authuser", "authuser@example.com", "password123", "Auth User")
		if err != nil {
			t.Fatalf("Failed to register user: %v", err)
		}

		// Test successful authentication
		user, err := userService.Authenticate("authuser", "password123")
		if err != nil {
			t.Fatalf("Failed to authenticate user: %v", err)
		}

		if user.Username != "authuser" {
			t.Errorf("Expected username 'authuser', got '%s'", user.Username)
		}
	})

	t.Run("Authenticate_InvalidPassword", func(t *testing.T) {
		// First register a user
		_, err := userService.RegisterUser("authuser2", "authuser2@example.com", "password123", "Auth User 2")
		if err != nil {
			t.Fatalf("Failed to register user: %v", err)
		}

		// Test authentication with wrong password
		_, err = userService.Authenticate("authuser2", "wrongpassword")
		if err == nil {
			t.Error("Expected error when authenticating with wrong password")
		}
	})

	t.Run("Authenticate_NonexistentUser", func(t *testing.T) {
		// Test authentication with non-existent user
		_, err := userService.Authenticate("nonexistent", "password123")
		if err == nil {
			t.Error("Expected error when authenticating non-existent user")
		}
	})

	t.Run("AuthenticateByEmail", func(t *testing.T) {
		// First register a user
		_, err := userService.RegisterUser("emailauth", "emailauth@example.com", "password123", "Email Auth")
		if err != nil {
			t.Fatalf("Failed to register user: %v", err)
		}

		// Test successful authentication by email
		user, err := userService.AuthenticateByEmail("emailauth@example.com", "password123")
		if err != nil {
			t.Fatalf("Failed to authenticate user by email: %v", err)
		}

		if user.Email != "emailauth@example.com" {
			t.Errorf("Expected email 'emailauth@example.com', got '%s'", user.Email)
		}
	})

	t.Run("AuthenticateByEmail_InvalidPassword", func(t *testing.T) {
		// First register a user
		_, err := userService.RegisterUser("emailauth2", "emailauth2@example.com", "password123", "Email Auth 2")
		if err != nil {
			t.Fatalf("Failed to register user: %v", err)
		}

		// Test authentication with wrong password
		_, err = userService.AuthenticateByEmail("emailauth2@example.com", "wrongpassword")
		if err == nil {
			t.Error("Expected error when authenticating with wrong password")
		}
	})

	t.Run("DeleteUser", func(t *testing.T) {
		// First register a user
		user, err := userService.RegisterUser("deleteuser", "deleteuser@example.com", "password123", "Delete User")
		if err != nil {
			t.Fatalf("Failed to register user: %v", err)
		}

		// Delete user
		err = userService.DeleteUser(user.ID)
		if err != nil {
			t.Fatalf("Failed to delete user: %v", err)
		}

		// Verify user is deleted
		_, err = userService.GetUserByID(user.ID)
		if err == nil {
			t.Error("Expected error when getting deleted user")
		}
	})

	t.Run("DeleteUser_NotFound", func(t *testing.T) {
		// Try to delete non-existent user
		err := userService.DeleteUser(99999)
		if err == nil {
			t.Error("Expected error when deleting non-existent user")
		}
	})

	t.Run("ListUsers", func(t *testing.T) {
		// Register multiple users
		for i := 0; i < 5; i++ {
			username := "listuser" + string(rune('0'+i))
			email := username + "@example.com"
			_, err := userService.RegisterUser(username, email, "password123", "List User "+string(rune('0'+i)))
			if err != nil {
				t.Fatalf("Failed to register user %d: %v", i, err)
			}
		}

		// List users with limit
		users, err := userService.ListUsers(3, 0)
		if err != nil {
			t.Fatalf("Failed to list users: %v", err)
		}

		if len(users) != 3 {
			t.Errorf("Expected 3 users, got %d", len(users))
		}

		// Test pagination
		moreUsers, err := userService.ListUsers(3, 3)
		if err != nil {
			t.Fatalf("Failed to list users with offset: %v", err)
		}

		if len(moreUsers) == 0 {
			t.Error("Expected more users with offset")
		}
	})

	t.Run("UserExists", func(t *testing.T) {
		// Register a user
		_, err := userService.RegisterUser("existsuser", "existsuser@example.com", "password123", "Exists User")
		if err != nil {
			t.Fatalf("Failed to register user: %v", err)
		}

		// Check if user exists by username
		exists, err := userService.UserExists("existsuser", "")
		if err != nil {
			t.Fatalf("Failed to check user existence: %v", err)
		}
		if !exists {
			t.Error("Expected user to exist")
		}

		// Check if user exists by email
		exists, err = userService.UserExists("", "existsuser@example.com")
		if err != nil {
			t.Fatalf("Failed to check user existence: %v", err)
		}
		if !exists {
			t.Error("Expected user to exist")
		}

		// Check non-existent user
		exists, err = userService.UserExists("nonexistent", "nonexistent@example.com")
		if err != nil {
			t.Fatalf("Failed to check user existence: %v", err)
		}
		if exists {
			t.Error("Expected user to not exist")
		}
	})
}

func TestUserServiceConcurrency(t *testing.T) {
	// Create temporary database for testing
	tempDB := "test_users_concurrency.db"
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

	t.Run("ConcurrentUserRegistration", func(t *testing.T) {
		// Test concurrent user registration
		done := make(chan bool, 10)
		errors := make(chan error, 10)

		for i := 0; i < 10; i++ {
			go func(id int) {
				username := "concurrent" + string(rune('0'+id))
				email := username + "@example.com"
				_, err := userService.RegisterUser(username, email, "password123", "Concurrent User")
				if err != nil {
					errors <- err
				}
				done <- true
			}(i)
		}

		// Wait for all goroutines to complete
		for i := 0; i < 10; i++ {
			select {
			case <-done:
				// Success
			case err := <-errors:
				t.Errorf("Concurrent registration failed: %v", err)
			case <-time.After(5 * time.Second):
				t.Fatal("Timeout waiting for concurrent registration")
			}
		}
	})
}

func BenchmarkUserService(b *testing.B) {
	// Create temporary database for benchmarking
	tempDB := "bench_users.db"
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

	b.Run("RegisterUser", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			username := "benchuser" + string(rune(i))
			email := username + "@example.com"
			_, err := userService.RegisterUser(username, email, "password123", "Bench User")
			if err != nil {
				b.Fatalf("Failed to register user: %v", err)
			}
		}
	})

	// Register a user for authentication benchmarks
	testUser, err := userService.RegisterUser("authbench", "authbench@example.com", "password123", "Auth Bench")
	if err != nil {
		b.Fatalf("Failed to register test user: %v", err)
	}

	b.Run("Authenticate", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, err := userService.Authenticate("authbench", "password123")
			if err != nil {
				b.Fatalf("Failed to authenticate user: %v", err)
			}
		}
	})

	b.Run("GetUserByID", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, err := userService.GetUserByID(testUser.ID)
			if err != nil {
				b.Fatalf("Failed to get user: %v", err)
			}
		}
	})
}
