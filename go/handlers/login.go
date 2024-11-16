package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv" // Import for int to string conversion
	"time"

	"github.com/dgrijalva/jwt-go"
	"golang.org/x/crypto/bcrypt"
)

type JWTClaims struct {
	ID   string `json:"id"`
	Role string `json:"role"` // Add role field for Hasura
	jwt.StandardClaims
}

// Secret key for signing JWT
var jwtSecret = []byte("your_secret_key_here")

// Struct to capture the payload for login
type loginArgs struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// The LoginHandler processes the login request
func LoginHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Read the request body
	reqBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "invalid payload", http.StatusBadRequest)
		return
	}

	// Parse the JSON payload into a struct (simplified input structure)
	var actionPayload struct {
		Input loginArgs `json:"input"` // Directly access the input fields
	}

	// Unmarshal the request body into the actionPayload
	err = json.Unmarshal(reqBody, &actionPayload)
	if err != nil {
		http.Error(w, "invalid payload", http.StatusBadRequest)
		return
	}

	// Get the login input (email and password)
	loginPayload := actionPayload.Input

	// Call the login function to process the login
	result, err := login(loginPayload)
	if err != nil {
		errorObject := GraphQLError{
			Message: err.Error(),
		}
		errorBody, _ := json.Marshal(errorObject)
		w.WriteHeader(http.StatusUnauthorized)
		w.Write(errorBody)
		return
	}

	// Convert `int` ID to `string`
	idAsString := strconv.Itoa(result.ID)

	// Generate JWT token
	token, err := generateJWT(idAsString, result.Role)
	if err != nil {
		http.Error(w, "failed to generate token", http.StatusInternalServerError)
		return
	}

	// Respond with user info and JWT token
	response := struct {
		ID    int    `json:"id"`
		Token string `json:"token"`
		Role  string `json:"role"`
	}{
		ID:    result.ID,
		Token: token,
		Role:  result.Role, // Return the role
	}

	// Send the response back to Hasura
	data, _ := json.Marshal(response)
	w.Write(data)
}

// The login function validates the credentials
func login(args loginArgs) (response userOutput, err error) {
	// Check credentials via Hasura GraphQL query
	hasuraResponse, err := executeLogin(map[string]interface{}{
		"email": args.Email,
	})
	if err != nil {
		return
	}

	// Handle errors from Hasura response
	if len(hasuraResponse.Errors) != 0 {
		err = errors.New(hasuraResponse.Errors[0].Message)
		return
	}
	if len(hasuraResponse.Data.User) == 0 {
		err = errors.New("invalid credentials")
		return
	}

	// Check password hash for validity
	user := hasuraResponse.Data.User[0] // Assuming we only need the first match
	isValid := checkPasswordHash(args.Password, user.Password)
	if !isValid {
		err = errors.New("invalid credentials")
		return
	}

	// Return the user response
	response = user
	return
}

// Check password hash against stored hash
func checkPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// Generate JWT token
func generateJWT(userID, role string) (string, error) {
	claims := JWTClaims{
		ID:   userID,
		Role: role, // Include role in the token
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(24 * time.Hour).Unix(),
			Issuer:    "your_app", // Adjust the issuer name as needed
		},
	}

	// Sign and return the JWT token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtSecret)
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

// Execute GraphQL query for login
func executeLogin(variables map[string]interface{}) (response GraphQLResponse, err error) {
	query := `query ($email: String!) {
        users(where: {email: {_eq: $email}}) {
            id
            password
            role
        }
    }`

	// Construct GraphQL request
	reqBody := GraphQLRequest{
		Query:     query,
		Variables: variables,
	}
	reqBytes, err := json.Marshal(reqBody)
	if err != nil {
		return
	}

	// Send the request to Hasura GraphQL endpoint
	resp, err := http.Post("http://localhost:8080/v1/graphql", "application/json", bytes.NewBuffer(reqBytes))
	if err != nil {
		return
	}
	defer resp.Body.Close()

	// Read the response
	respBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}

	// Handle non-OK responses from Hasura
	if resp.StatusCode != http.StatusOK {
		err = fmt.Errorf("failed to execute GraphQL query: %s", string(respBytes))
		return
	}

	// Parse the response JSON
	err = json.Unmarshal(respBytes, &response)
	if err != nil {
		return
	}

	return
}
