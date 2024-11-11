package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	"golang.org/x/crypto/bcrypt"
)

// ActionPayload structure to handle the input from the Hasura Action
type ActionPayload struct {
	SessionVariables map[string]interface{} `json:"session_variables"`
	Input            signupArgs             `json:"input"`
}

// signupArgs represents the data we expect from the signup action
type signupArgs struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"email"`
	Password  string `json:"password"`
}

// SignupOutput represents the response that will be sent back to Hasura
type SignupOutput struct {
	Id      string `json:"id"`
	Message string `json:"message"`
}

// GraphQLError structure to return error messages
type GraphQLErr struct {
	Message string `json:"message"`
}

// Load environment variables from .env file
func loadEnvVars() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}
}

// hashPassword hashes the password using bcrypt
func hashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedPassword), nil
}

// performHasuraMutation sends a mutation request to Hasura's GraphQL endpoint
func performHasuraMutation(user signupArgs) (SignupOutput, error) {
	// Get GraphQL API URL and Hasura admin secret from environment variables
	hasuraURL := os.Getenv("HASURA_GRAPHQL_ENDPOINT")
	hasuraSecret := os.Getenv("HASURA_GRAPHQL_ADMIN_SECRET")

	// Construct the mutation for Hasura to insert the user
	mutation := `
	mutation($first_name: String!, $last_name: String!, $email: String!, $password: String!) {
	  insert_users_one(object: {
	    first_name: $first_name,
	    last_name: $last_name,
	    email: $email,
	    password: $password
	  }) {
	    id
	    first_name
	    last_name
	    email
	  }
	}
	`

	// Prepare the variables for the mutation
	variables := map[string]interface{}{
		"first_name": user.FirstName,
		"last_name":  user.LastName,
		"email":      user.Email,
		"password":   user.Password, // Already hashed password
	}

	// Prepare the GraphQL request body
	requestBody := map[string]interface{}{
		"query":     mutation,
		"variables": variables,
	}

	// Encode the request body to JSON
	requestBodyJSON, err := json.Marshal(requestBody)
	if err != nil {
		return SignupOutput{}, err
	}

	// Send the request to the Hasura GraphQL endpoint
	req, err := http.NewRequest("POST", hasuraURL, bytes.NewBuffer(requestBodyJSON))
	if err != nil {
		return SignupOutput{}, err
	}

	// Set the authorization header with the Hasura secret
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-hasura-admin-secret", hasuraSecret)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return SignupOutput{}, err
	}
	defer resp.Body.Close()

	// Read the response body
	responseBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return SignupOutput{}, err
	}

	// Check if the response contains an error
	if resp.StatusCode != http.StatusOK {
		return SignupOutput{}, fmt.Errorf("Hasura responded with error: %s", responseBody)
	}

	// Parse the response from Hasura
	var response map[string]interface{}
	err = json.Unmarshal(responseBody, &response)
	if err != nil {
		return SignupOutput{}, err
	}

	// Extract the user ID from the response
	userData := response["data"].(map[string]interface{})["insert_users_one"].(map[string]interface{})
	userID := userData["id"].(string)

	// Return the result
	return SignupOutput{
		Id:      userID,
		Message: "User successfully created",
	}, nil
}

// handler processes the signup request
func SignupHandler(w http.ResponseWriter, r *http.Request) {
	// Set the response header as JSON
	w.Header().Set("Content-Type", "application/json")
	fmt.Println("mmmmmmmmmmmmmmmmmmmmmmm")

	// Read request body
	reqBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "invalid payload", http.StatusBadRequest)
		return
	}

	// Parse the body as Action payload
	var actionPayload ActionPayload
	err = json.Unmarshal(reqBody, &actionPayload)
	if err != nil {
		http.Error(w, "invalid payload", http.StatusBadRequest)
		return
	}

	// Hash the password before sending it to Hasura
	hashedPassword, err := hashPassword(actionPayload.Input.Password)
	if err != nil {
		http.Error(w, "Error hashing password", http.StatusInternalServerError)
		return
	}

	// Replace the plain password with the hashed password
	actionPayload.Input.Password = hashedPassword

	// Send the request params to the Hasura mutation
	result, err := performHasuraMutation(actionPayload.Input)

	// Throw if an error happens
	if err != nil {
		errorObject := GraphQLError{
			Message: err.Error(),
		}
		errorBody, _ := json.Marshal(errorObject)
		w.WriteHeader(http.StatusBadRequest)
		w.Write(errorBody)
		return
	}

	// Write the response as JSON
	data, _ := json.Marshal(result)
	w.Write(data)
}
