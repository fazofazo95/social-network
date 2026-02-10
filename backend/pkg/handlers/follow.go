package handlers

import (
	database "backend/pkg/db/sqlite"
	"backend/pkg/models"
	"backend/pkg/responses"
	"backend/pkg/services"
	"encoding/json"
	"net/http"
)

func FollowRequestHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
	w.Header().Set("Access-Control-Allow-Credentials", "true")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	followService := services.NewFollowService(database.DB)

	var followRequest models.FollowRequest
	if err := json.NewDecoder(r.Body).Decode(&followRequest); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	err := followService.FollowUser(r.Context(), followRequest)
	if err != nil {
		http.Error(w, "Failed to process follow request: "+err.Error(), http.StatusInternalServerError)
		return
	}

	responses.SendCreated(w, "follow request created successfully", nil)
}


