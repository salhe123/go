package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"golang.org/x/crypto/bcrypt"
)

// ActionPayload represents the structure of the incoming request payload.
type ActionPayload struct {
	SessionVariables map[string]interface{} `json:"session_variables"`
	Input            signupArgs             `json:"input"`
}

// GraphQLError represents an error response from Hasura.
type GraphQLError struct {
	Message string `json:"message"`
}

// GraphQLRequest is the structure for a GraphQL request.
type GraphQLRequest struct {
	Query     string      `json:"query"`
	Variables interface{} `json:"variables"`
}

// GraphQLResponse represents the response from Hasura.
type GraphQLResponse struct {
	Data   GraphQLData    `json:"data,omitempty"`
	Errors []GraphQLError `json:"errors,omitempty"`
}

// GraphQLData represents the data field in the GraphQL response.
type GraphQLData struct {
	InsertUserOne signupOutput `json:"insert_users_one"`
}

// signupArgs is the structure for the signup input fields.
type signupArgs struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Username string `json:"username"`
}

// signupOutput is the structure for the successful signup response.
type signupOutput struct {
	ID      int    `json:"id"`
	Message string `json:"message"`
}

// SignupHandler handles the signup request.
func SignupHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Read and parse the request body
	reqBody, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "invalid payload", http.StatusBadRequest)
		return
	}

	var actionPayload ActionPayload
	err = json.Unmarshal(reqBody, &actionPayload)
	if err != nil {
		http.Error(w, "invalid payload", http.StatusBadRequest)
		return
	}
	fmt.Println("Received payload:", actionPayload.Input)

	// Call the signup function to handle the business logic
	result, err := signup(actionPayload.Input)
	if err != nil {
		errorObject := GraphQLError{
			Message: err.Error(),
		}
		errorBody, _ := json.Marshal(errorObject)
		w.WriteHeader(http.StatusBadRequest)
		w.Write(errorBody)
		return
	}

	// Respond with the result
	data, _ := json.Marshal(result)
	w.Write(data)
}

// signup processes the signup logic, including password hashing and calling Hasura.
func signup(args signupArgs) (response signupOutput, err error) {
	// Hash the password before saving it
	hashedPassword, err := hashPassword(args.Password)
	if err != nil {
		return
	}
	// Log the variables for debugging purposes
	fmt.Println("Variables to be sent to Hasura:", args.Email, hashedPassword, args.Username)
	// Prepare the variables for the GraphQL mutation
	variables := map[string]interface{}{
		"email":    args.Email,
		"password": hashedPassword,
		"username": args.Username,
	}
	fmt.Println("Variables to be sent to Hasura:", variables)
	// Execute the signup mutation with the prepared variables
	hasuraResponse, err := executeSignup(variables)
	if err != nil {
		return
	}

	// Check if there are any errors from Hasura
	if len(hasuraResponse.Errors) > 0 {
		err = errors.New(hasuraResponse.Errors[0].Message)
		return
	}

	// Return the successful response from Hasura
	response = hasuraResponse.Data.InsertUserOne
	return
}

// executeSignup sends the GraphQL request to Hasura and returns the response.
func executeSignup(variables map[string]interface{}) (response GraphQLResponse, err error) {
	// Define the GraphQL mutation query
	query := `mutation($email: String!, $password: String!, $username: String!) {
    insert_users_one(object: { email: $email, password: $password, username: $username }) {
        id
    }
}`

	// Prepare the request body
	reqBody := GraphQLRequest{
		Query:     query,
		Variables: variables,
	}
	reqBytes, err := json.Marshal(reqBody)
	if err != nil {
		return
	}

	// Send the request to Hasura
	resp, err := http.Post("http://localhost:8080/v1/graphql", "application/json", bytes.NewBuffer(reqBytes))
	if err != nil {
		return
	}
	defer resp.Body.Close()

	// Read the response body
	respBytes, err := io.ReadAll(resp.Body) // Use io.ReadAll instead of ioutil.ReadAll
	if err != nil {
		return
	}
	fmt.Println("Hasura Response:", string(respBytes))
	// Check if the response status is OK
	if resp.StatusCode != http.StatusOK {
		err = fmt.Errorf("failed to execute GraphQL query: %s", string(respBytes))
		return
	}

	// Parse the response
	err = json.Unmarshal(respBytes, &response)
	if err != nil {
		return
	}

	return
}

// hashPassword hashes the password using bcrypt.
func hashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedPassword), nil
}
