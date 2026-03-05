package handlers

import (
	"backend/pkg/middleware"
	"backend/pkg/models"
	"backend/pkg/responses"
	"backend/pkg/services"
	"net/http"
	"strconv"
)

type FollowHandler struct {
	Service services.FollowService
}

func NewFollowHandler(s services.FollowService) *FollowHandler {
	return &FollowHandler{
		Service: s,
	}
}

func (h *FollowHandler) RegisterRoutes(mux *http.ServeMux) {
	auth := middleware.WithAuth

	mux.Handle("POST /api/users/{id}/follow", middleware.Chain(h.FollowUserHandler, auth))
	mux.Handle("DELETE /api/users/{id}/unfollow", middleware.Chain(h.UnfollowUserHandler, auth))
	mux.Handle("DELETE /api/users/{id}/remove-follower", middleware.Chain(h.RemoveFollowerHandler, auth))
	mux.Handle("POST /api/users/{id}/follow/accept", middleware.Chain(h.AcceptFollowHandler, auth))
	mux.Handle("DELETE /api/users/{id}/follow/reject", middleware.Chain(h.RejectFollowHandler, auth))
	mux.Handle("POST /api/users/{id}/block", middleware.Chain(h.BlockUserHandler, auth))
	mux.Handle("DELETE /api/users/{id}/unblock", middleware.Chain(h.UnblockUserHandler, auth))
}

func (h *FollowHandler) FollowUserHandler(w http.ResponseWriter, r *http.Request) {
	followerID, err := middleware.UserIDFromContext(r.Context())
	if err != nil {
		responses.SendError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	targetIDstr := r.PathValue("id")
	targetID, err := strconv.Atoi(targetIDstr)
	if err != nil || targetID <= 0 {
		responses.SendError(w, http.StatusBadRequest, "invalid target user id")
		return
	}

	req := models.FollowRequest{
		FollowerID: followerID,
		FollowedID: targetID,
	}

	status, err := h.Service.FollowUser(r.Context(), req)
	if err != nil {
		responses.SendError(w, http.StatusInternalServerError, "Failed to process follow request")
		return
	}

	resp := map[string]interface{}{
		"follower_id": req.FollowerID,
		"followed_id": req.FollowedID,
		"status":      status,
	}

	responses.SendCreated(w, "follow request created successfully", resp)
}

func (h *FollowHandler) UnfollowUserHandler(w http.ResponseWriter, r *http.Request) {
	followerID, err := middleware.UserIDFromContext(r.Context())
	if err != nil {
		responses.SendError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	targetIDstr := r.PathValue("id")
	targetID, err := strconv.Atoi(targetIDstr)
	if err != nil || targetID <= 0 {
		responses.SendError(w, http.StatusBadRequest, "invalid target user id")
		return
	}

	deleted, err := h.Service.DeleteFollow(r.Context(), followerID, targetID)
	if err != nil {
		responses.SendError(w, http.StatusInternalServerError, "failed to unfollow: "+err.Error())
		return
	}

	if deleted == 0 {
		responses.SendError(w, http.StatusNotFound, "no follow relationship found")
		return
	}

	responses.SendSuccess(w, "unfollowed successfully", map[string]interface{}{"follower_id": followerID, "followed_id": targetID})
}

func (h *FollowHandler) RemoveFollowerHandler(w http.ResponseWriter, r *http.Request) {
	currentUserID, err := middleware.UserIDFromContext(r.Context())
	if err != nil {
		responses.SendError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	targetFollowerIDstr := r.PathValue("id")
	targetFollowerID, err := strconv.Atoi(targetFollowerIDstr)
	if err != nil || targetFollowerID <= 0 {
		responses.SendError(w, http.StatusBadRequest, "invalid target follower id")
		return
	}

	deleted, err := h.Service.RemoveFollower(r.Context(), currentUserID, targetFollowerID)
	if err != nil {
		responses.SendError(w, http.StatusInternalServerError, "failed to remove follower: "+err.Error())
		return
	}

	if deleted == 0 {
		responses.SendError(w, http.StatusNotFound, "no accepted follower relationship found")
		return
	}

	responses.SendSuccess(w, "follower removed successfully", map[string]interface{}{"user_id": currentUserID, "removed_follower_id": targetFollowerID})
}

func (h *FollowHandler) AcceptFollowHandler(w http.ResponseWriter, r *http.Request) {
	followedID, err := middleware.UserIDFromContext(r.Context())
	if err != nil {
		responses.SendError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	followerIDstr := r.PathValue("id")
	followerID, err := strconv.Atoi(followerIDstr)
	if err != nil || followerID <= 0 {
		responses.SendError(w, http.StatusBadRequest, "invalid follower user id")
		return
	}

	updated, err := h.Service.AcceptFollow(r.Context(), followerID, followedID)
	if err != nil {
		responses.SendError(w, http.StatusInternalServerError, "failed to accept follow: "+err.Error())
		return
	}

	if updated == 0 {
		responses.SendError(w, http.StatusNotFound, "no pending follow request found")
		return
	}

	responses.SendSuccess(w, "follow request accepted", map[string]interface{}{"follower_id": followerID, "followed_id": followedID})
}

func (h *FollowHandler) RejectFollowHandler(w http.ResponseWriter, r *http.Request) {

	followedID, err := middleware.UserIDFromContext(r.Context())
	if err != nil {
		responses.SendError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	followerIDstr := r.PathValue("id")
	followerID, err := strconv.Atoi(followerIDstr)
	if err != nil || followerID <= 0 {
		responses.SendError(w, http.StatusBadRequest, "invalid follower user id")
		return
	}

	deleted, err := h.Service.RejectFollow(r.Context(), followerID, followedID)
	if err != nil {
		responses.SendError(w, http.StatusInternalServerError, "failed to reject follow: "+err.Error())
		return
	}

	if deleted == 0 {
		responses.SendError(w, http.StatusNotFound, "no pending follow request found")
		return
	}

	responses.SendSuccess(w, "follow request rejected", map[string]interface{}{"follower_id": followerID, "followed_id": followedID})
}

func (h *FollowHandler) BlockUserHandler(w http.ResponseWriter, r *http.Request) {
	blockerID, err := middleware.UserIDFromContext(r.Context())
	if err != nil {
		responses.SendError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	targetIDstr := r.PathValue("id")
	targetID, err := strconv.Atoi(targetIDstr)
	if err != nil || targetID <= 0 {
		responses.SendError(w, http.StatusBadRequest, "invalid target user id")
		return
	}

	updated, err := h.Service.BlockFollow(r.Context(), blockerID, targetID)
	if err != nil {
		responses.SendError(w, http.StatusInternalServerError, "failed to block user: "+err.Error())
		return
	}

	responses.SendSuccess(w, "user blocked", map[string]interface{}{"blocker_id": blockerID, "blocked_id": targetID, "rows": updated})
}

func (h *FollowHandler) UnblockUserHandler(w http.ResponseWriter, r *http.Request) {
	blockerID, err := middleware.UserIDFromContext(r.Context())
	if err != nil {
		responses.SendError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	targetIDstr := r.PathValue("id")
	targetID, err := strconv.Atoi(targetIDstr)
	if err != nil || targetID <= 0 {
		responses.SendError(w, http.StatusBadRequest, "invalid target user id")
		return
	}

	deleted, err := h.Service.UnblockFollow(r.Context(), blockerID, targetID)
	if err != nil {
		responses.SendError(w, http.StatusInternalServerError, "failed to unblock user: "+err.Error())
		return
	}
	if deleted == 0 {
		responses.SendError(w, http.StatusNotFound, "no blocked relationship found")
		return
	}

	responses.SendSuccess(w, "user unblocked", map[string]interface{}{"blocker_id": blockerID, "unblocked_id": targetID})
}
