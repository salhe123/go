package handlers

import (
	"encoding/json"
	"net/http"
	"os"

	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
)

type ImageUploadInput struct {
	File   string `json:"file"`
	UserID string `json:"user_id,omitempty"`
}

type ImageUploadOutput struct {
	URL     string `json:"url"`
	Message string `json:"message"`
}

func ImageUploadHandler(w http.ResponseWriter, r *http.Request) {
	var input ImageUploadInput

	// Decode JSON request body
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	// Initialize Cloudinary with credentials from .env
	cld, err := cloudinary.NewFromParams(os.Getenv("CLOUDINARY_CLOUD_NAME"), os.Getenv("CLOUDINARY_API_KEY"), os.Getenv("CLOUDINARY_API_SECRET"))
	if err != nil {
		http.Error(w, "Failed to initialize Cloudinary", http.StatusInternalServerError)
		return
	}

	// Upload the image to Cloudinary
	uploadResult, err := cld.Upload.Upload(r.Context(), input.File, uploader.UploadParams{})
	if err != nil {
		http.Error(w, "Failed to upload image", http.StatusInternalServerError)
		return
	}

	// Prepare response
	response := ImageUploadOutput{
		URL:     uploadResult.SecureURL,
		Message: "Image uploaded successfully",
	}

	// Set response headers and send JSON response
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}
