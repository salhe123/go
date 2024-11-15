package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"golang.org/x/crypto/bcrypt"
)

// ActionPayload represents the payload structure received in the signup request.
type ActionPayload struct {
	SessionVariables map[string]interface{} `json:"session_variables"`
	Input            signupArgs             `json:"input"`
}

// GraphQLError represents the structure of an error returned by the Hasura GraphQL API.
type GraphQLError struct {
	Message string `json:"message"`
}

// GraphQLRequest represents the structure for sending a GraphQL request.
type GraphQLRequest struct {
	Query     string      `json:"query"`
	Variables interface{} `json:"variables"`
}

// GraphQLResponse represents the overall response from the GraphQL server.
type GraphQLResponse struct {
	Data   interface{}    `json:"data,omitempty"`
	Errors []GraphQLError `json:"errors,omitempty"`
}

// signupArgs defines the input fields required for signup.
type signupArgs struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Username string `json:"username"`
}

// signupOutput defines the output structure of a successful signup.
type signupOutput struct {
	ID      string `json:"id"`
	Message string `json:"message"`
}

// SignupHandler is the HTTP handler for processing signup requests.
func SignupHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Read request body
	reqBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Printf("Error reading request body: %v", err)
		http.Error(w, "Invalid payload", http.StatusBadRequest)
		return
	}

	// Parse the request payload
	var actionPayload ActionPayload
	err = json.Unmarshal(reqBody, &actionPayload)
	if err != nil {
		log.Printf("Error unmarshaling request body: %v", err)
		http.Error(w, "Invalid payload", http.StatusBadRequest)
		return
	}

	// Perform signup
	result, err := signup(actionPayload.Input)
	if err != nil {
		log.Printf("Signup failed: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Respond with the result
	responseData, _ := json.Marshal(result)
	w.Write(responseData)
}

// signup handles the signup logic, including password hashing and GraphQL execution.
func signup(args signupArgs) (signupOutput, error) {
	// Hash the password before saving
	hashedPassword, err := hashPassword(args.Password)
	if err != nil {
		log.Printf("Failed to hash password: %v", err)
		return signupOutput{}, fmt.Errorf("Failed to hash password: %v", err)
	}

	// Define variables for the GraphQL mutation
	variables := map[string]interface{}{
		"input": signupArgs{
			Email:    args.Email,
			Password: hashedPassword,
			Username: args.Username,
		},
	}

	// Execute the GraphQL mutation
	hasuraResponse, err := executeSignup(variables)
	if err != nil {
		log.Printf("Failed to execute signup: %v", err)
		return signupOutput{}, fmt.Errorf("Failed to execute signup: %v", err)
	}

	// Handle errors from the GraphQL response
	if len(hasuraResponse.Errors) > 0 {
		log.Printf("Hasura response error: %s", hasuraResponse.Errors[0].Message)
		return signupOutput{}, errors.New(hasuraResponse.Errors[0].Message)
	}

	// Return the signup response
	if data, ok := hasuraResponse.Data.(*signupOutput); ok {
		return *data, nil
	}
	return signupOutput{}, fmt.Errorf("Unexpected response format")
}

// executeSignup sends the GraphQL request to Hasura and returns the response.
func executeSignup(variables map[string]interface{}) (GraphQLResponse, error) {
	query := `mutation MyMutation( $email: String!, $password: String!,$username: String!) {
  signup(input: { email: $email, password: $password, username: $username }) {
    id
    message
  }
}
`

	// Log the variables before sending the request
	log.Printf("GraphQL Variables: %v", variables)

	// Prepare request body
	reqBody := GraphQLRequest{
		Query:     query,
		Variables: variables,
	}
	reqBytes, err := json.Marshal(reqBody)
	if err != nil {
		log.Printf("Failed to marshal request: %v", err)
		return GraphQLResponse{}, fmt.Errorf("Failed to marshal request: %v", err)
	}

	// Fetch Hasura environment variables
	hasuraAdminURL := os.Getenv("HASURA_GRAPHQL_ENDPOINT")
	adminSecret := os.Getenv("HASURA_GRAPHQL_ADMIN_SECRET")
	if hasuraAdminURL == "" || adminSecret == "" {
		log.Printf("Missing Hasura Admin URL or Secret")
		return GraphQLResponse{}, errors.New("Missing Hasura Admin URL or Secret")
	}

	// Create HTTP request to Hasura
	req, err := http.NewRequest("POST", hasuraAdminURL, bytes.NewBuffer(reqBytes))
	if err != nil {
		log.Printf("Failed to create request: %v", err)
		return GraphQLResponse{}, fmt.Errorf("Failed to create request: %v", err)
	}

	// Set required headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Hasura-Admin-Secret", adminSecret)

	// Execute the HTTP request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Request failed: %v", err)
		return GraphQLResponse{}, fmt.Errorf("Request failed: %v", err)
	}
	defer resp.Body.Close()

	// Read and process response
	respBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Failed to read response: %v", err)
		return GraphQLResponse{}, fmt.Errorf("Failed to read response: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		log.Printf("GraphQL query failed: %s", string(respBytes))
		return GraphQLResponse{}, fmt.Errorf("GraphQL query failed: %s", string(respBytes))
	}

	// Unmarshal the GraphQL response
	var response GraphQLResponse
	err = json.Unmarshal(respBytes, &response)
	if err != nil {
		log.Printf("Failed to parse response: %v", err)
		return GraphQLResponse{}, fmt.Errorf("Failed to parse response: %v", err)
	}

	return response, nil
}

// hashPassword hashes the given password using bcrypt.
func hashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedPassword), nil
}
