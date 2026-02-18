package handlers

import (
	"log"
	"net/http"

	queries "backend/pkg/db/queries"
	database "backend/pkg/db/sqlite"
	"backend/pkg/middleware"
	"backend/pkg/responses"
)

func DiscoverHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("[INFO] DiscoverHandler: Received request")

	userID, err := middleware.UserIDFromContext(r.Context())
	if err != nil {
		log.Printf("[ERROR] DiscoverHandler: Authentication failed: %v", err)
		http.Error(w, "User not authenticated", http.StatusUnauthorized)
		return
	}
	log.Printf("[INFO] DiscoverHandler: Fetching recommendations for UserID: %d", userID)

	const defaultLimit = 5

	users, err := queries.DiscoverUsers(r.Context(), database.DB, userID, defaultLimit)
	if err != nil {
		log.Printf("[ERROR] DiscoverHandler: Query failed for UserID %d: %v", userID, err)
		responses.SendError(w, http.StatusInternalServerError, "Failed to discover users: "+err.Error())
		return
	}

	log.Printf("[SUCCESS] DiscoverHandler: Found %d users for UserID: %d", len(users), userID)
	responses.SendSuccess(w, "Discovered users", users)
}
