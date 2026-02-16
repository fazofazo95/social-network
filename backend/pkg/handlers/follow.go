package handlers

import (
	database "backend/pkg/db/sqlite"
	"backend/pkg/models"
	"backend/pkg/responses"
	"backend/pkg/services"
	"encoding/json"
	"net/http"
)

func FollowUserHandler(w http.ResponseWriter, r *http.Request) {

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

func UnfollowUserHandler(w http.ResponseWriter, r *http.Request) {}
