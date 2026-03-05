package handlers

import (
	"net/http"
	"strconv"

	"backend/pkg/middleware"
	"backend/pkg/responses"
	"backend/pkg/services"
)

type ReactionHandler struct {
	reactionService services.ReactionService
}

func NewReactionHandler(s services.ReactionService) *ReactionHandler {
	return &ReactionHandler{reactionService: s}
}

func (h *ReactionHandler) RegisterRoutes(mux *http.ServeMux) {
	auth := middleware.WithAuth

	mux.Handle("POST /api/posts/{id}/reactions", middleware.Chain(h.AddReactionHandler, auth))
	mux.Handle("DELETE /api/posts/{id}/reactions", middleware.Chain(h.RemoveReactionHandler, auth))
}

func (h *ReactionHandler) AddReactionHandler(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.UserIDFromContext(r.Context())
	if err != nil {
		responses.SendError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	targetID, _ := strconv.Atoi(r.PathValue("id"))

	likeCount, err := h.reactionService.AddReaction(r.Context(), userID, targetID)
	if err != nil {
		responses.SendError(w, http.StatusInternalServerError, "failed to add reaction")
		return
	}

	responses.SendSuccess(w, "reaction added successfully", map[string]interface{}{"like_count": likeCount})
}

func (h *ReactionHandler) RemoveReactionHandler(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.UserIDFromContext(r.Context())
	if err != nil {
		responses.SendError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	targetID, _ := strconv.Atoi(r.PathValue("id"))

	likeCount, err := h.reactionService.RemoveReaction(r.Context(), userID, targetID)
	if err != nil {
		responses.SendError(w, http.StatusInternalServerError, "failed to remove reaction")
		return
	}

	responses.SendSuccess(w, "reaction removed successfully", map[string]interface{}{"like_count": likeCount})
}
