package main

import (
	"event_backend/handlers" // Adjust based on your module name
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
)

func main() {
	// Load the .env file
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	// Get the port from environment variables, with a default fallback
	port := os.Getenv("PORT")
	if port == "" {
		port = "5000"
	}

	router := mux.NewRouter() // Initialize the router using mux

	// Initialize your handlers
	router.HandleFunc("/signup", handlers.SignupHandler).Methods("POST")
	// router.HandleFunc("/api/login", handlers.LoginHandler).Methods("POST") // Uncomment if needed

	// Start the server
	fmt.Printf("Server is listening on port %s...\n", port)
	log.Fatal(http.ListenAndServe(":"+port, router))
}
