package handlers

import (
	"backend/pkg/middleware"
	"backend/pkg/responses"
	"backend/pkg/services"
	"net/http"
	"strconv"
)

type RelationsHandler struct {
	Service services.FollowService
}

func NewRelationsHandler(s services.FollowService) *RelationsHandler {
	return &RelationsHandler{Service: s}
}

func (h *RelationsHandler) RegisterRoutes(mux *http.ServeMux) {
	auth := middleware.WithAuth

	mux.Handle("GET /api/users/following", middleware.Chain(h.FollowingHandler, auth))
	mux.Handle("GET /api/users/followers", middleware.Chain(h.FollowersHandler, auth))
	mux.Handle("GET /api/users/{id}/following", middleware.Chain(h.FollowingByUserHandler, auth))
	mux.Handle("GET /api/users/{id}/followers", middleware.Chain(h.FollowersByUserHandler, auth))
	mux.Handle("GET /api/users/blocked", middleware.Chain(h.BlockedHandler, auth))
	mux.Handle("GET /api/users/pending", middleware.Chain(h.PendingRequestsHandler, auth))
}

func (h *RelationsHandler) FollowingHandler(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.UserIDFromContext(r.Context())
	if err != nil {
		responses.SendError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	users, err := h.Service.GetFollowingUsers(r.Context(), userID, userID)
	if err != nil {
		responses.SendError(w, http.StatusInternalServerError, "failed to get following list: "+err.Error())
		return
	}

	responses.SendSuccess(w, "following list", users)
}

func (h *RelationsHandler) FollowersHandler(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.UserIDFromContext(r.Context())
	if err != nil {
		responses.SendError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	users, err := h.Service.GetFollowers(r.Context(), userID)
	if err != nil {
		responses.SendError(w, http.StatusInternalServerError, "failed to get followers list: "+err.Error())
		return
	}

	responses.SendSuccess(w, "followers list", users)
}

func (h *RelationsHandler) FollowingByUserHandler(w http.ResponseWriter, r *http.Request) {
	viewerID, err := middleware.UserIDFromContext(r.Context())
	if err != nil {
		responses.SendError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	targetID, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || targetID <= 0 {
		responses.SendError(w, http.StatusBadRequest, "invalid target id")
		return
	}

	users, err := h.Service.GetFollowingUsers(r.Context(), targetID, viewerID)
	if err != nil {
		responses.SendError(w, http.StatusInternalServerError, "failed to get following list: "+err.Error())
		return
	}

	responses.SendSuccess(w, "following list", users)
}

func (h *RelationsHandler) FollowersByUserHandler(w http.ResponseWriter, r *http.Request) {
	viewerID, err := middleware.UserIDFromContext(r.Context())
	if err != nil {
		responses.SendError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	targetID, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || targetID <= 0 {
		responses.SendError(w, http.StatusBadRequest, "invalid target id")
		return
	}

	users, err := h.Service.GetFollowersByUser(r.Context(), viewerID, targetID)
	if err != nil {
		responses.SendError(w, http.StatusInternalServerError, "failed to get followers list: "+err.Error())
		return
	}

	responses.SendSuccess(w, "followers list", users)
}

func (h *RelationsHandler) BlockedHandler(w http.ResponseWriter, r *http.Request) {

	userID, err := middleware.UserIDFromContext(r.Context())
	if err != nil {
		responses.SendError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	users, err := h.Service.GetBlockedUsers(r.Context(), userID)
	if err != nil {
		responses.SendError(w, http.StatusInternalServerError, "failed to get blocked users: "+err.Error())
		return
	}

	responses.SendSuccess(w, "blocked users list", users)
}

func (h *RelationsHandler) PendingRequestsHandler(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.UserIDFromContext(r.Context())
	if err != nil {
		responses.SendError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	users, err := h.Service.GetPendingRequests(r.Context(), userID)
	if err != nil {
		responses.SendError(w, http.StatusInternalServerError, "failed to get pending requests: "+err.Error())
		return
	}

	responses.SendSuccess(w, "pending requests", users)
}
