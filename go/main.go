package main

import (
	"event_backend/handlers" // Adjust based on your module name
	"log"
	"net/http"

	"github.com/joho/godotenv"
)

func main() {
	// Load the .env file
	err := godotenv.Load(".env") // Load .env from the current directory
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	// Initialize your handlers
	http.HandleFunc("/api/signup", handlers.SignupHandler)
	http.HandleFunc("/api/login", handlers.LoginHandler)

	// Start the server on port 7070
	log.Println("Starting server on port 7070...")
	err = http.ListenAndServe(":7070", nil)
	if err != nil {
		log.Fatalf("Could not start server: %s", err)
	}
}
