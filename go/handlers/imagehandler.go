package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
	"github.com/joho/godotenv"
)

type ActionPayloads struct {
	SessionVariables map[string]interface{} `json:"session_variables"`
	Input            image_uploadArgs       `json:"input"`
}

type GraphQLEr struct {
	Message string `json:"message"`
}

type image_uploadArgs struct {
	Images           []string `json:"images"`
	FeaturedImageURL string   `json:"featured_image_url"`
}

// Event represents the event with the images
type Event struct {
	ID               string   `json:"id"`
	Image            []string `json:"image"`
	FeaturedImageURL string   `json:"featured_image_url"`
}

// Function to upload image to Cloudinary
func uploadToCloudinary(base64Image string) (string, error) {
	// Load the environment variables from the .env file
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
		return "", err
	}

	// Get Cloudinary credentials from environment variables
	cloudinaryURL := os.Getenv("CLOUDINARY_URL")
	if cloudinaryURL == "" {
		return "", fmt.Errorf("CLOUDINARY_URL is not set in environment variables")
	}

	cld, err := cloudinary.NewFromURL(cloudinaryURL)
	if err != nil {
		return "", err
	}

	// Upload base64 image to Cloudinary
	resp, err := cld.Upload.Upload(context.Background(), base64Image, uploader.UploadParams{ResourceType: "image"})
	if err != nil {
		return "", err
	}

	// Return the URL of the uploaded image
	return resp.SecureURL, nil
}

// image_handler is the handler function that processes the image upload request
func image_handler(w http.ResponseWriter, r *http.Request) {
	// Set response header as JSON
	w.Header().Set("Content-Type", "application/json")

	// Read request body
	reqBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "invalid payload", http.StatusBadRequest)
		return
	}

	// Parse the body as action payload
	var actionPayload ActionPayloads
	err = json.Unmarshal(reqBody, &actionPayload)
	if err != nil {
		http.Error(w, "invalid payload", http.StatusBadRequest)
		return
	}

	// Send the request params to the Action's generated handler function
	result, err := image_upload(actionPayload.Input)

	// Throw error if an error happens
	if err != nil {
		errorObject := GraphQLEr{
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

// image_upload function that processes the image upload
func image_upload(args image_uploadArgs) (response Event, err error) {
	var uploadedImages []string

	// Loop through each image, upload to Cloudinary, and store the URLs
	for _, base64Image := range args.Images {
		imageURL, uploadErr := uploadToCloudinary(base64Image)
		if uploadErr != nil {
			return Event{}, fmt.Errorf("error uploading image: %v", uploadErr)
		}
		uploadedImages = append(uploadedImages, imageURL)
	}

	// Set the featured image URL (the first image in the uploaded list)
	featuredImageURL := ""
	if len(uploadedImages) > 0 {
		featuredImageURL = uploadedImages[0]
	}

	// Now, update the database (here you would have your database logic to insert this data)

	// Example response - you can customize this based on your database response
	response = Event{
		Image:            uploadedImages,
		FeaturedImageURL: featuredImageURL,
	}

	// Return the updated event with URLs
	return response, nil
}
