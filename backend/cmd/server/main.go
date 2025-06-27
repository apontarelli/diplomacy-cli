package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"diplomacy-backend/internal/api"
	"diplomacy-backend/internal/game"
	"diplomacy-backend/internal/storage"
)

func main() {
	dbPath := getDBPath()
	
	db, err := storage.NewDB(dbPath)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	// Load game data
	dataPath := getDataPath()
	loader := game.NewDataLoader(dataPath)
	gameData, err := loader.LoadClassicGameData()
	if err != nil {
		log.Fatalf("Failed to load game data: %v", err)
	}

	server := api.NewServer(db, gameData)
	mux := server.SetupRoutes()

	port := getPort()
	fmt.Printf("Starting server on port %s\n", port)
	fmt.Printf("Database: %s\n", dbPath)
	fmt.Printf("Game data: %s\n", dataPath)
	
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}

func getDBPath() string {
	if dbPath := os.Getenv("DB_PATH"); dbPath != "" {
		return dbPath
	}
	
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "./diplomacy.db"
	}
	
	return filepath.Join(homeDir, ".diplomacy", "diplomacy.db")
}

func getPort() string {
	if port := os.Getenv("PORT"); port != "" {
		return port
	}
	return "8080"
}

func getDataPath() string {
	if dataPath := os.Getenv("DATA_PATH"); dataPath != "" {
		return dataPath
	}
	
	// Default to the Python CLI data directory
	return "../src/diplomacy_cli/data/classic"
}