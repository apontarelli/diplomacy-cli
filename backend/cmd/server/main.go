package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"diplomacy-cli/backend/internal/api"
	"diplomacy-cli/backend/internal/storage"
)

func main() {
	// Initialize database
	config := storage.Config{
		DatabasePath: "diplomacy.db",
		LogQueries:   false,
	}
	db, err := storage.NewDatabase(config)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	// Initialize services
	userService := storage.NewUserService(db)
	authService := storage.NewAuthService(userService, storage.DefaultJWTConfig())

	// Initialize API server
	server := api.NewServer(db, authService)

	// Configure HTTP server
	httpServer := &http.Server{
		Addr:         ":8080",
		Handler:      server.Handler(),
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in goroutine
	go func() {
		log.Printf("Starting server on %s", httpServer.Addr)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := httpServer.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
}
