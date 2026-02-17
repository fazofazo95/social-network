package handlers

import (
	"net/http"

	queries "backend/pkg/db/queries"
	database "backend/pkg/db/sqlite"
	"backend/pkg/middleware"
	"backend/pkg/responses"
)

// DiscoverHandler returns a list of users the current user may want to follow.
// Query params:
// - limit (optional, default 5)
func DiscoverHandler(w http.ResponseWriter, r *http.Request) {
	// ensure auth middleware placed the user id in context
	userID, err := middleware.UserIDFromContext(r.Context())
	if err != nil {
		http.Error(w, "User not authenticated", http.StatusUnauthorized)
		return
	}

	// Use server-side configured limit (ignore client-provided ?limit param).
	const defaultLimit = 5

	users, err := queries.DiscoverUsers(r.Context(), database.DB, userID, defaultLimit)
	if err != nil {
		responses.SendError(w, http.StatusInternalServerError, "Failed to discover users: "+err.Error())
		return
	}

	responses.SendSuccess(w, "Discovered users", users)
}
