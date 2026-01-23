package responses

import (
	"encoding/json"
	"net/http"
)

// Response represents a standard JSON response
type Response struct {
	Status  string      `json:"status"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

// SendJSON sends a JSON response with the given status code and data
func SendJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// SendSuccess sends a success JSON response
func SendSuccess(w http.ResponseWriter, message string, data interface{}) {
	resp := Response{
		Status:  "success",
		Message: message,
		Data:    data,
	}
	SendJSON(w, http.StatusOK, resp)
}

// SendError sends an error JSON response
func SendError(w http.ResponseWriter, status int, message string) {
	resp := Response{
		Status:  "error",
		Message: message,
	}
	SendJSON(w, status, resp)
}

// SendCreated sends a created JSON response
func SendCreated(w http.ResponseWriter, message string, data interface{}) {
	resp := Response{
		Status:  "success",
		Message: message,
		Data:    data,
	}
	SendJSON(w, http.StatusCreated, resp)
}