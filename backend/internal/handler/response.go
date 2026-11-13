package handler

import (
	"encoding/json"
	"net/http"

	"streamforge-backend/internal/models"
)

// WriteJSON serializes data into JSON and sends HTTP response with specified status code
func WriteJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(models.SuccessResponse(data))
}

// WriteError sends a standardized error JSON response
func WriteError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(models.ErrorResponse(message))
}
