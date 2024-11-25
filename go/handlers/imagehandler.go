package handlers

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
)

var cld *cloudinary.Cloudinary

func init() {
	var err error

	// Load Cloudinary parameters from environment variables
	cloudinaryName := os.Getenv("CLOUDINARY_NAME")
	cloudinaryAPIKey := os.Getenv("CLOUDINARY_API_KEY")
	cloudinaryAPISecret := os.Getenv("CLOUDINARY_API_SECRET")

	cld, err = cloudinary.NewFromParams(cloudinaryName, cloudinaryAPIKey, cloudinaryAPISecret)
	if err != nil {
		log.Fatalf("Failed to create Cloudinary instance: %v", err)
	}
}

type UploadImagesRequest struct {
	Input struct {
		Base64Strs []string `json:"base64_strs"` // Array of base64-encoded strings
	} `json:"input"`
}

type ImageUploadResponse struct {
	Urls []string `json:"urls"` // List of uploaded image URLs
}

func UploadBase64Images(w http.ResponseWriter, r *http.Request) {
	var request UploadImagesRequest

	// Decode incoming JSON payload
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&request); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	// Validate input
	if len(request.Input.Base64Strs) == 0 {
		http.Error(w, "No images provided", http.StatusBadRequest)
		return
	}

	var uploadedURLs []string
	for _, base64Str := range request.Input.Base64Strs {
		// Remove potential prefix
		if strings.HasPrefix(base64Str, "data:image/") {
			base64Str = base64Str[strings.IndexByte(base64Str, ',')+1:]
		}

		// Decode the image
		decodedImage, err := base64.StdEncoding.DecodeString(base64Str)
		if err != nil {
			log.Printf("Failed to decode base64 string: %v", err)
			http.Error(w, "Failed to decode base64 string", http.StatusInternalServerError)
			return
		}

		// Upload the image to Cloudinary
		uploadResp, err := cld.Upload.Upload(r.Context(), bytes.NewReader(decodedImage), uploader.UploadParams{Folder: "images"})
		if err != nil {
			log.Printf("Error while uploading to Cloudinary: %v", err)
			http.Error(w, "Failed to upload image to Cloudinary", http.StatusInternalServerError)
			return
		}

		// Append the uploaded URL to the list
		uploadedURLs = append(uploadedURLs, uploadResp.SecureURL)
	}

	// Return the list of uploaded URLs
	response := ImageUploadResponse{Urls: uploadedURLs}
	responseData, _ := json.Marshal(response)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(responseData)
}
