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

// GraphQLData represents the data section in the GraphQL response.
type GraphQLData struct {
	Insert_user_one signupOutput `json:"insert_user_one"`
}

// GraphQLResponse represents the overall response from the GraphQL server.
type GraphQLResponse struct {
	Data   GraphQLData    `json:"data,omitempty"`
	Errors []GraphQLError `json:"errors,omitempty"`
}

// signupArgs defines the input fields required for signup.
type signupArgs struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

// signupOutput defines the output structure of a successful signup.
type signupOutput struct {
	ID string `json:"id"`
}

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

func signup(args signupArgs) (response signupOutput, err error) {
	hashedPassword, err := hashPassword(args.Password)
	if err != nil {
		log.Printf("Failed to hash password: %v", err)
		return response, fmt.Errorf("Failed to hash password: %v", err)
	}

	// Define variables for the GraphQL mutation
	variables := map[string]interface{}{
		"email":    args.Email,
		"username": args.Username,
		"password": hashedPassword,
	}

	// Execute the GraphQL mutation
	hasuraResponse, err := executeSignup(variables)
	if err != nil {
		log.Printf("Failed to execute signup: %v", err)
		return response, fmt.Errorf("Failed to execute signup: %v", err)
	}

	// Handle errors from the GraphQL response
	if len(hasuraResponse.Errors) > 0 {
		log.Printf("Hasura response error: %s", hasuraResponse.Errors[0].Message)
		return response, errors.New(hasuraResponse.Errors[0].Message)
	}

	// Return the signup response
	response = hasuraResponse.Data.Insert_user_one
	return response, nil
}

func executeSignup(variables map[string]interface{}) (response GraphQLResponse, err error) {
	query := `mutation ($username: String!, $email: String!, $password: String!) {
		signup(input: {username: $username, email: $email, password: $password}) {
			id
			message
		}
	}`

	// Prepare request body
	reqBody := GraphQLRequest{
		Query:     query,
		Variables: variables,
	}
	reqBytes, err := json.Marshal(reqBody)
	if err != nil {
		log.Printf("Failed to marshal request: %v", err)
		return response, fmt.Errorf("Failed to marshal request: %v", err)
	}

	// Fetch Hasura environment variables
	hasuraAdminURL := os.Getenv("HASURA_GRAPHQL_ENDPOINT")
	adminSecret := os.Getenv("HASURA_GRAPHQL_ADMIN_SECRET")
	if hasuraAdminURL == "" || adminSecret == "" {
		log.Printf("Missing Hasura Admin URL or Secret")
		return response, errors.New("Missing Hasura Admin URL or Secret")
	}

	// Create HTTP request to Hasura
	req, err := http.NewRequest("POST", hasuraAdminURL, bytes.NewBuffer(reqBytes))
	if err != nil {
		log.Printf("Failed to create request: %v", err)
		return response, fmt.Errorf("Failed to create request: %v", err)
	}

	// Set required headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Hasura-Admin-Secret", adminSecret)

	// Execute the HTTP request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Request failed: %v", err)
		return response, fmt.Errorf("Request failed: %v", err)
	}
	defer resp.Body.Close()

	// Read and process response
	respBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Failed to read response: %v", err)
		return response, fmt.Errorf("Failed to read response: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		log.Printf("GraphQL query failed: %s", string(respBytes))
		return response, fmt.Errorf("GraphQL query failed: %s", string(respBytes))
	}

	// Unmarshal the GraphQL response
	err = json.Unmarshal(respBytes, &response)
	if err != nil {
		log.Printf("Failed to parse response: %v", err)
		return response, fmt.Errorf("Failed to parse response: %v", err)
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
