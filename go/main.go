package main

import (
	"event_backend/handlers"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}
	router := mux.NewRouter()
	router.HandleFunc("/signup", handlers.SignupHandler).Methods("POST")
	// router.HandleFunc("/api/login", handlers.LoginHandler).Methods("POST") // Uncomment if needed
	fmt.Printf("Server is listening on port %s...\n", port)
	log.Fatal(http.ListenAndServe(":"+port, router))
}
