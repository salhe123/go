package handlers

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"os"

	"golang.org/x/crypto/bcrypt"
)

// Define the user input data structure matching Hasura's SignupInput
type SignupInput struct {
	Email     string `json:"email"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Password  string `json:"password"`
}

// GraphQL request structure
type GraphQLRequest struct {
	Query     string      `json:"query"`
	Variables interface{} `json:"variables"`
}

// Define the expected structure for Hasura's SignupOutput
type SignupOutput struct {
	ID      string `json:"id"`
	Message string `json:"message"`
}

// Define a structure to match Hasura's GraphQL response
type GraphQLResponse struct {
	Data struct {
		Signup SignupOutput `json:"signup"`
	} `json:"data"`
}

// SignupHandler to handle the signup mutation
func SignupHandler(w http.ResponseWriter, r *http.Request) {
	// Parse the request JSON into SignupInput struct
	var input SignupInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}
	log.Printf("Received Signup request: %+v\n", input)

	// Hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Failed to hash password", http.StatusInternalServerError)
		return
	}
	input.Password = string(hashedPassword)

	// GraphQL mutation with variables
	mutation := `
        mutation Signup($email: String!, $first_name: String!, $last_name: String!, $password: String!) {
            signup(arg1: {email: $email, first_name: $first_name, last_name: $last_name, password: $password}) {
                id
                message
            }
        }
    `

	// Prepare GraphQL request payload
	requestPayload := GraphQLRequest{
		Query: mutation,
		Variables: map[string]interface{}{
			"email":      input.Email,
			"first_name": input.FirstName,
			"last_name":  input.LastName,
			"password":   input.Password,
		},
	}

	jsonData, err := json.Marshal(requestPayload)
	if err != nil {
		http.Error(w, "Failed to serialize payload", http.StatusInternalServerError)
		return
	}

	// Use environment variables for sensitive information
	endpoint := os.Getenv("HASURA_GRAPHQL_ENDPOINT")
	adminSecret := os.Getenv("HASURA_GRAPHQL_ADMIN_SECRET")

	// Send the HTTP request to Hasura
	req, err := http.NewRequest("POST", endpoint, bytes.NewBuffer(jsonData))
	if err != nil {
		http.Error(w, "Failed to create request", http.StatusInternalServerError)
		return
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-hasura-admin-secret", adminSecret)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		http.Error(w, "Request to Hasura failed", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		http.Error(w, "Non-200 response from Hasura", http.StatusInternalServerError)
		return
	}

	// Parse the GraphQL response and log it
	var result GraphQLResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		http.Error(w, "Failed to decode response", http.StatusInternalServerError)
		return
	}

	// Log the Hasura response for debugging
	log.Printf("Hasura response: %+v\n", result)

	// Check if the response is empty
	if result.Data.Signup.ID == "" {
		http.Error(w, "Signup mutation failed", http.StatusInternalServerError)
		return
	}

	// Send the signup response back to the client
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(result.Data.Signup)
}
