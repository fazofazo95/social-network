package handlers

import (
	"net/http"
	"strconv"

	"backend/pkg/db/queries"
	database "backend/pkg/db/sqlite"
	"backend/pkg/middleware"
	"backend/pkg/responses"
)

func AddReactionHandler(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.UserIDFromContext(r.Context())
	if err != nil {
		responses.SendError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	targetID, _ := strconv.Atoi(r.PathValue("id"))

	likeCount, err := queries.AddReaction(r.Context(), database.DB, targetID, userID)
	if err != nil {
		responses.SendError(w, http.StatusInternalServerError, "failed to add reaction")
		return
	}

	responses.SendSuccess(w, "reaction added successfully", map[string]interface{}{"like_count": likeCount})
}

func RemoveReactionHandler(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.UserIDFromContext(r.Context())
	if err != nil {
		responses.SendError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	targetID, _ := strconv.Atoi(r.PathValue("id"))

	likeCount, err := queries.RemoveReaction(r.Context(), database.DB, targetID, userID)
	if err != nil {
		responses.SendError(w, http.StatusInternalServerError, "failed to remove reaction")
		return
	}

	responses.SendSuccess(w, "reaction removed successfully", map[string]interface{}{"like_count": likeCount})
}
