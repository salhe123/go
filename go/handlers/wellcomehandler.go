package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/smtp"
	"os"
	"time"

	"github.com/joho/godotenv"
)

// SMTPConfig holds the mail server credentials
type SMTPConfig struct {
	Host     string
	Port     string
	Username string
	Password string
	From     string
}

func init() {
	// Load environment variables from the .env file
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}
}

func GetSMTPConfig() *SMTPConfig {
	// Gmail SMTP server details
	return &SMTPConfig{
		Host:     "smtp.gmail.com",
		Port:     "587",                           // TLS port for Gmail
		Username: os.Getenv("GMAIL_USERNAME"),     // Set in the environment
		Password: os.Getenv("GMAIL_APP_PASSWORD"), // Set the app-specific password
		From:     os.Getenv("FROM_EMAIL"),         // Your Gmail address
	}
}

// sendWelcomeEmail sends the welcome email to the user
func sendWelcomeEmail(toEmail, username string) error {
	smtpConfig := GetSMTPConfig()

	// Construct the email message
	subject := "Welcome to Event Management"
	body := fmt.Sprintf("Hello %s,\n\nThank you for joining our event management community! We're excited to have you on board and look forward to helping you create and manage your events with ease.\nIf you have any questions or need assistance getting started, feel free to reach out. We're here to help!\n\nBest regards,\nThe Event Management Team", username)
	message := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\n\r\n%s", smtpConfig.From, toEmail, subject, body)

	// Set up authentication information
	auth := smtp.PlainAuth("", smtpConfig.Username, smtpConfig.Password, smtpConfig.Host)

	// Send the email
	err := smtp.SendMail(
		smtpConfig.Host+":"+smtpConfig.Port,
		auth,
		smtpConfig.From,
		[]string{toEmail},
		[]byte(message),
	)

	if err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	return nil
}

// User represents the structure of the incoming user data for the welcome email
type User struct {
	Event struct {
		Data struct {
			New struct {
				Email    string `json:"email"`
				Username string `json:"username"`
			} `json:"new"`
		} `json:"data"`
	} `json:"event"`
}

// WelcomeEmailHandler handles the incoming request to send a welcome email
func WelcomeEmailHandler(w http.ResponseWriter, r *http.Request) {
	var user User
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		log.Printf("Error decoding request body: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	email := user.Event.Data.New.Email
	log.Printf("Sending welcome email to: %s", email)

	username := user.Event.Data.New.Username
	err = sendWelcomeEmail(email, username)
	if err != nil {
		log.Printf("Failed to send email to %s: %v", email, err)
		http.Error(w, "Failed to send email", http.StatusInternalServerError)
		return
	}

	// Return a success message with the email sent time
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Welcome email sent successfully to %s at %s", email, fmt.Sprintf("%v", time.Now()))
}
