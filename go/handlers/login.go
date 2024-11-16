package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/dgrijalva/jwt-go"
	"golang.org/x/crypto/bcrypt"
)

type JWTClaims struct {
	ID   string `json:"id"`
	Role string `json:"role"`
	jwt.StandardClaims
}

var jwtSecret = []byte("your_secret_key_here")

type loginArgs struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	reqBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Println("Error reading request body:", err)
		http.Error(w, "invalid payload", http.StatusBadRequest)
		return
	}

	var actionPayload struct {
		Action struct {
			Name string `json:"name"`
		} `json:"action"`
		Input struct {
			Email    string `json:"email"`
			Password string `json:"password"`
		} `json:"input"`
	}

	err = json.Unmarshal(reqBody, &actionPayload)
	if err != nil {
		log.Println("Error unmarshaling request body:", err)
		http.Error(w, "invalid payload", http.StatusBadRequest)
		return
	}

	// Extract email and password from the input field
	email := actionPayload.Input.Email
	password := actionPayload.Input.Password
	log.Printf("Parsed login payload: email=%s, password=%s\n", email, password)

	if email == "" || password == "" {
		log.Println("Missing email or password")
		http.Error(w, "email and password are required", http.StatusBadRequest)
		return
	}
	result, err := login(loginArgs{Email: email, Password: password})
	if err != nil {
		log.Println("Login failed:", err)
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}
	token, err := generateJWT(strconv.Itoa(result.ID), result.Role)
	if err != nil {
		log.Println("Error generating token:", err)
		http.Error(w, "failed to generate token", http.StatusInternalServerError)
		return
	}
	response := struct {
		ID    int    `json:"id"`
		Token string `json:"token"`
		Role  string `json:"role"`
	}{
		ID:    result.ID,
		Token: token,
		Role:  result.Role,
	}

	data, _ := json.Marshal(response)
	w.Write(data)
}

func login(args loginArgs) (response userOutput, err error) {
	hasuraResponse, err := executeLogin(map[string]interface{}{
		"email": args.Email,
	})
	if err != nil {
		return
	}
	log.Println("hasura response", hasuraResponse)

	if len(hasuraResponse.Errors) != 0 {
		err = errors.New(hasuraResponse.Errors[0].Message)
		return
	}
	if len(hasuraResponse.Data.Users) == 0 {
		err = errors.New("invalid credentials")
		return
	}

	user := hasuraResponse.Data.Users[0]
	isValid := checkPasswordHash(args.Password, user.Password)
	if !isValid {
		err = errors.New("invalid credentials")
		return
	}
	response = user
	return
}

func checkPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func generateJWT(ID string, role string) (string, error) {
	claims := jwt.MapClaims{
		"sub":   ID,
		"name":  "seya",
		"admin": role == "admin",
		"iat":   time.Now().Unix(),
		"exp":   time.Now().Add(time.Hour * 24).Unix(),
		"iss":   "Event",
		// Hasura-specific claims
		"https://hasura.io/jwt/claims": map[string]interface{}{
			"x-hasura-default-role":  role,
			"x-hasura-allowed-roles": []string{role, "admin"},
			"x-hasura-user-id":       ID,
			"x-hasura-org-id":        "456",
			"x-hasura-custom":        "custom-value",
		},
	}

	// Create the JWT token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	secretKey := []byte("your_secret_key")
	tokenString, err := token.SignedString(secretKey)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func executeLogin(variables map[string]interface{}) (response GraphQLResponse, err error) {
	query := `query ($email: String!) {
        users(where: {email: {_eq: $email}}) {
            id
            password
            role
        }
    }`

	reqBody := GraphQLRequest{
		Query:     query,
		Variables: variables,
	}
	reqBytes, err := json.Marshal(reqBody)
	if err != nil {
		return
	}

	resp, err := http.Post("http://localhost:8080/v1/graphql", "application/json", bytes.NewBuffer(reqBytes))
	if err != nil {
		return
	}
	defer resp.Body.Close()

	respBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}

	if resp.StatusCode != http.StatusOK {
		err = fmt.Errorf("failed to execute GraphQL query: %s", string(respBytes))
		return
	}

	err = json.Unmarshal(respBytes, &response)
	if err != nil {
		return
	}

	return
}
