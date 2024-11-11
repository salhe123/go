package handlers

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/dgrijalva/jwt-go"
	"golang.org/x/crypto/bcrypt"
)

// JWT Signing key (should be securely stored, e.g., in environment variables)
var mySigningKey = []byte("your_secret_key") // Use a secret key

// LoginArgs struct to parse the request payload
type LoginArgs struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// LoginOutput struct for the response
type LoginOutput struct {
	ID    string `json:"id"`
	Token string `json:"token"`
}

// GraphQLError struct for handling error from Hasura
type GraphQLError struct {
	Message string `json:"message"`
}

// Login handler to authenticate user
func LoginHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Read the request body
	reqBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "invalid payload", http.StatusBadRequest)
		return
	}

	// Parse the body as action payload
	var loginArgs LoginArgs
	err = json.Unmarshal(reqBody, &loginArgs)
	if err != nil {
		http.Error(w, "invalid payload", http.StatusBadRequest)
		return
	}

	// Get Hasura endpoint and admin secret from environment variables
	hasuraEndpoint := os.Getenv("HASURA_GRAPHQL_ENDPOINT")
	hasuraAdminSecret := os.Getenv("HASURA_GRAPHQL_ADMIN_SECRET")

	// Create the GraphQL query to fetch user data by email
	query := `query($email: String!) {
		users(where: {email: {_eq: $email}}) {
			id
			password
		}
	}`

	// Create a request payload
	var requestBody = map[string]interface{}{
		"query": query,
		"variables": map[string]interface{}{
			"email": loginArgs.Email,
		},
	}

	// Convert request body to JSON
	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		http.Error(w, "Error encoding request", http.StatusInternalServerError)
		return
	}

	// Make the HTTP request to Hasura
	req, err := http.NewRequest("POST", hasuraEndpoint, bytes.NewBuffer(jsonBody))
	if err != nil {
		http.Error(w, "Error creating request", http.StatusInternalServerError)
		return
	}

	// Set headers for Hasura
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-hasura-admin-secret", hasuraAdminSecret)

	// Send the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		http.Error(w, "Error sending request to Hasura", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	// Parse the response from Hasura
	var response map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		http.Error(w, "Error decoding response", http.StatusInternalServerError)
		return
	}
	log.Printf("Hasura Response: %+v", response)

	// Check if the user exists (ensure "users" field exists)
	data, ok := response["data"].(map[string]interface{})
	if !ok {
		http.Error(w, "Unexpected response structure", http.StatusInternalServerError)
		return
	}

	users, ok := data["users"].([]interface{})
	if !ok || len(users) == 0 {
		http.Error(w, "Invalid email or password", http.StatusUnauthorized)
		return
	}

	// Retrieve the user info from the response
	user := users[0].(map[string]interface{})
	hashedPassword := user["password"].(string)
	userID := user["id"].(string)

	// Compare the password with the stored hashed password
	err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(loginArgs.Password))
	if err != nil {
		http.Error(w, "Invalid email or password", http.StatusUnauthorized)
		return
	}

	// Generate JWT token for the user
	token, err := generateJWT(userID) // Pass the actual userID here
	if err != nil {
		http.Error(w, "Error generating token", http.StatusInternalServerError)
		return
	}

	// Prepare the response
	loginOutput := LoginOutput{
		ID:    userID,
		Token: token,
	}

	// Send the response
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(loginOutput)
}

func generateJWT(userID string) (string, error) {
	// Create JWT claims
	claims := jwt.MapClaims{
		"sub": userID,
		"exp": time.Now().Add(time.Hour * 24).Unix(),
		"https://hasura.io/jwt/claims": map[string]interface{}{
			"x-hasura-user-id":       userID,
			"x-hasura-default-role":  "user",
			"x-hasura-allowed-roles": []string{"user"},
		},
	}

	// Generate JWT token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(mySigningKey)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}
