package main

import (
	"event_backend/handlers" // Adjust based on your module name
	"log"
	"net/http"

	"github.com/gorilla/mux" // Import the mux package here
	"github.com/joho/godotenv"
)

func main() {
	// Load the .env file
	err := godotenv.Load(".env") // Load .env from the current directory
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	router := mux.NewRouter() // Initialize the router using mux

	// Initialize your handlers
	router.HandleFunc("/api/signup", handlers.SignupHandler).Methods("POST")
	// http.HandleFunc("/api/login", handlers.LoginHandler)

	// Start the server on port 5000
	log.Println("Starting server on port 5000...")
	err = http.ListenAndServe(":5000", router) // Use the router here
	if err != nil {
		log.Fatalf("Could not start server: %s", err)
	}
}
