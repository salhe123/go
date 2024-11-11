package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

type PaymentRequest struct {
	Amount   int64  `json:"amount"`
	Currency string `json:"currency"`
	Email    string `json:"email"`
}

type ChapaResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
	Data    struct {
		PaymentUrl string `json:"payment_url"`
	} `json:"data"`
}

// Exported function
func PaymentHandler(w http.ResponseWriter, r *http.Request) {
	// Set content type
	w.Header().Set("Content-Type", "application/json")

	// Decode the payment request
	var paymentRequest PaymentRequest
	err := json.NewDecoder(r.Body).Decode(&paymentRequest)
	if err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	// Create JSON payload for Chapa API
	payload, err := json.Marshal(map[string]interface{}{
		"amount":   paymentRequest.Amount,
		"currency": paymentRequest.Currency,
		"email":    paymentRequest.Email,
	})
	if err != nil {
		http.Error(w, "Error creating payment payload", http.StatusInternalServerError)
		return
	}

	// Send payment request to Chapa API
	req, err := http.NewRequest("POST", "https://api.chapa.dev/transaction/initialize", bytes.NewBuffer(payload))
	if err != nil {
		http.Error(w, "Error creating request", http.StatusInternalServerError)
		return
	}

	// Set required headers
	req.Header.Set("Authorization", "Bearer your_chapa_secret_key") // Replace with your Chapa secret key
	req.Header.Set("Content-Type", "application/json")

	// Perform the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		http.Error(w, "Error sending payment request", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	// Read the response body
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, "Error reading response", http.StatusInternalServerError)
		return
	}

	// Check the response status
	if resp.StatusCode != http.StatusOK {
		http.Error(w, fmt.Sprintf("Chapa API error: %s", string(body)), resp.StatusCode)
		return
	}

	// Unmarshal the response into ChapaResponse
	var chapaResponse ChapaResponse
	if err := json.Unmarshal(body, &chapaResponse); err != nil {
		http.Error(w, "Error parsing Chapa response", http.StatusInternalServerError)
		return
	}

	// Send response back to the client
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(chapaResponse)
}
