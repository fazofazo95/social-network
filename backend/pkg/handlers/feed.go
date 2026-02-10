package handlers

import (
	queries "backend/pkg/db/queries"
	database "backend/pkg/db/sqlite"
	"backend/pkg/middleware"
	"backend/pkg/models"
	"backend/pkg/responses"
	"log"
	"net/http"
)

func FeedHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
	w.Header().Set("Access-Control-Allow-Credentials", "true")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get user ID from context (set by auth middleware)
	userID, err := middleware.UserIDFromContext(r.Context())
	if err != nil {
		http.Error(w, "User not authenticated", http.StatusUnauthorized)
		return
	}

	log.Printf("FeedHandler: fetching posts for user %d", userID)
	// Get posts from followed users
	posts, err := queries.GetFollowedUsersPosts(r.Context(), database.DB, userID, 5)
	if err != nil {
		http.Error(w, "Failed to fetch posts: "+err.Error(), http.StatusInternalServerError)
		return
	}

	log.Printf("FeedHandler: fetching discovered users for user %d", userID)
	// Get discovered users
	discoveredUsers, err := queries.DiscoverUsers(r.Context(), database.DB, userID, 5)
	if err != nil {
		http.Error(w, "Failed to discover users: "+err.Error(), http.StatusInternalServerError)
		return
	}

	feedResponse := models.FeedResponse{
		Posts:           posts,
		DiscoveredUsers: discoveredUsers,
	}

	responses.SendSuccess(w, "Feed retrieved successfully", feedResponse)
}
