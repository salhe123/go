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
		port = "5000"
	}
	router := mux.NewRouter()
	router.HandleFunc("/signup", handlers.SignupHandler).Methods("POST")
	router.HandleFunc("/login", handlers.LoginHandler).Methods("POST")
	router.HandleFunc("/uploadImages", handlers.UploadBase64Image).Methods("POST")
	router.HandleFunc("/acceptPayment", handlers.PaymentsHandler).Methods("POST")
	router.HandleFunc("/welcome_email", handlers.WelcomeEmailHandler)
	fmt.Printf("Server is listening on port %s...\n", port)
	log.Fatal(http.ListenAndServe(":"+port, router))
}
