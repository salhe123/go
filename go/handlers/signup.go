package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"golang.org/x/crypto/bcrypt"
)

type ActionPayload struct {
	SessionVariables map[string]interface{} `json:"session_variables"`
	Input            signupArgs             `json:"input"`
}

type GraphQLError struct {
	Message string `json:"message"`
}

type GraphQLRequest struct {
	Query     string      `json:"query"`
	Variables interface{} `json:"variables"`
}

type GraphQLData struct {
	Signup signupOutput `json:"signup"` // Update to match the mutation name
}

type GraphQLResponse struct {
	Data   GraphQLData    `json:"data,omitempty"`
	Errors []GraphQLError `json:"errors,omitempty"`
}

type signupArgs struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type signupOutput struct {
	ID string `json:"id"` // This corresponds to the `id` returned by the mutation
}

// SignupHandler is the HTTP handler for the signup action
func SignupHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	log.Println("mmmmmmmmmmmm")
	reqBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Invalid payload", http.StatusBadRequest)
		return
	}

	var actionPayload ActionPayload
	err = json.Unmarshal(reqBody, &actionPayload)
	if err != nil {
		http.Error(w, "Invalid payload", http.StatusBadRequest)
		return
	}

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

	data, _ := json.Marshal(result)
	w.Write(data)
}

// signup handles the business logic for user signup
func signup(args signupArgs) (response signupOutput, err error) {
	// Hash the password before saving to the database
	log.Println("signuphandles")
	hashedPassword, err := hashPassword(args.Password)
	if err != nil {
		return signupOutput{}, err
	}

	// Define variables for the GraphQL mutation
	variables := map[string]interface{}{
		"email":    args.Email,
		"password": hashedPassword,
		"username": args.Username,
	}

	// Perform the GraphQL mutation to insert a new user
	hasuraResponse, err := executeSignup(variables)
	if err != nil {
		return signupOutput{}, err
	}

	if len(hasuraResponse.Errors) != 0 {
		err = errors.New(hasuraResponse.Errors[0].Message)
		return signupOutput{}, err
	}

	// Prepare the response with the user ID
	response = signupOutput{
		ID: hasuraResponse.Data.Signup.ID,
	}

	return response, nil
}

// executeSignup executes the GraphQL mutation to insert a new user
func executeSignup(variables map[string]interface{}) (response GraphQLResponse, err error) {
	log.Println("signup data", variables)
	query := `mutation MyMutation($email: String!, $password: String!, $username: String!) {
        signup(input: {email: $email, password: $password, username: $username}) {
            id  
        }
    }`

	reqBody := GraphQLRequest{
		Query:     query,
		Variables: variables,
	}
	reqBytes, err := json.Marshal(reqBody)
	if err != nil {
		return GraphQLResponse{}, err
	}
	log.Println(reqBody)
	// Set the admin secret in the headers
	adminSecret := "ibnuseid27" // Replace with your actual admin secret
	req, err := http.NewRequest("POST", "http://localhost:8080/v1/graphql", bytes.NewBuffer(reqBytes))
	if err != nil {
		return GraphQLResponse{}, err
	}

	// Add the X-Hasura-Admin-Secret header
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Hasura-Admin-Secret", adminSecret)

	// Make the HTTP request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return GraphQLResponse{}, err
	}
	defer resp.Body.Close()

	// Read the response from Hasura
	respBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return GraphQLResponse{}, err
	}

	if resp.StatusCode != http.StatusOK {
		err = fmt.Errorf("failed to execute GraphQL query: %s", string(respBytes))
		return GraphQLResponse{}, err
	}

	// Parse the response body
	err = json.Unmarshal(respBytes, &response)
	if err != nil {
		return GraphQLResponse{}, err
	}

	return response, nil
}

// hashPassword hashes a user's password using bcrypt
func hashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedPassword), nil
}
