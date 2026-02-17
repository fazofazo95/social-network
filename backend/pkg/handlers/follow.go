package handlers

import (
	"backend/pkg/db/queries"
	database "backend/pkg/db/sqlite"
	"backend/pkg/middleware"
	"backend/pkg/models"
	"backend/pkg/responses"
	"backend/pkg/services"
	"net/http"
	"strconv"
	"strings"
)

// FollowUserHandler handles POST /api/users/{id}/follow
// The authenticated user (from context) is the follower. The {id} in the URL
// is the target user to follow. The handler decides whether the follow is
// 'accepted' or 'pending' based on the target user's `profile_type`.
func FollowUserHandler(w http.ResponseWriter, r *http.Request) {
	// get authenticated user id from context
	followerID, err := middleware.UserIDFromContext(r.Context())
	if err != nil {
		responses.SendError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	// extract target id from URL path: /api/users/{id}/follow
	parts := strings.Split(r.URL.Path, "/")
	// parts: ["", "api", "users", "{id}", "follow"]
	if len(parts) < 4 {
		responses.SendError(w, http.StatusBadRequest, "invalid path")
		return
	}
	targetStr := parts[3]
	targetID, err := strconv.Atoi(targetStr)
	if err != nil || targetID <= 0 {
		responses.SendError(w, http.StatusBadRequest, "invalid target id")
		return
	}

	followService := services.NewFollowService(database.DB)

	req := models.FollowRequest{
		FollowerID: followerID,
		FollowedID: targetID,
	}

	status, err := followService.FollowUser(r.Context(), req)
	if err != nil {
		responses.SendError(w, http.StatusInternalServerError, "Failed to process follow request: "+err.Error())
		return
	}

	// return the created follow object so frontend can update UI immediately
	resp := map[string]interface{}{
		"follower_id": req.FollowerID,
		"followed_id": req.FollowedID,
		"status":      status,
	}

	responses.SendCreated(w, "follow request created successfully", resp)
}

// UnfollowUserHandler handles DELETE /api/users/{id}/unfollow
// The authenticated user is the follower; the {id} in the path is the target to unfollow.
func UnfollowUserHandler(w http.ResponseWriter, r *http.Request) {
	followerID, err := middleware.UserIDFromContext(r.Context())
	if err != nil {
		responses.SendError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	// parse target id from path
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 4 {
		responses.SendError(w, http.StatusBadRequest, "invalid path")
		return
	}
	targetStr := parts[3]
	targetID, err := strconv.Atoi(targetStr)
	if err != nil || targetID <= 0 {
		responses.SendError(w, http.StatusBadRequest, "invalid target id")
		return
	}

	// perform delete
	deleted, err := queries.DeleteFollow(r.Context(), database.DB, followerID, targetID)
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

// AcceptFollowHandler handles POST /api/users/{id}/follow/accept
// Authenticated user is the followed user (the one who received the follow request).
// The {id} in the path is the follower id whose pending request should be accepted.
func AcceptFollowHandler(w http.ResponseWriter, r *http.Request) {
	followedID, err := middleware.UserIDFromContext(r.Context())
	if err != nil {
		responses.SendError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 5 {
		responses.SendError(w, http.StatusBadRequest, "invalid path")
		return
	}
	// expected: ["", "api", "users", "{id}", "follow", "accept"] or similar
	followerStr := parts[3]
	followerID, err := strconv.Atoi(followerStr)
	if err != nil || followerID <= 0 {
		responses.SendError(w, http.StatusBadRequest, "invalid follower id")
		return
	}

	// perform accept (pending -> accepted)
	updated, err := queries.AcceptFollow(r.Context(), database.DB, followerID, followedID)
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

// BlockUserHandler handles POST /api/users/{id}/block
// Authenticated user is the blocker; {id} is the user to block.
func BlockUserHandler(w http.ResponseWriter, r *http.Request) {
	blockerID, err := middleware.UserIDFromContext(r.Context())
	if err != nil {
		responses.SendError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 4 {
		responses.SendError(w, http.StatusBadRequest, "invalid path")
		return
	}
	targetStr := parts[3]
	targetID, err := strconv.Atoi(targetStr)
	if err != nil || targetID <= 0 {
		responses.SendError(w, http.StatusBadRequest, "invalid target id")
		return
	}

	updated, err := queries.BlockFollow(r.Context(), database.DB, blockerID, targetID)
	if err != nil {
		responses.SendError(w, http.StatusInternalServerError, "failed to block user: "+err.Error())
		return
	}

	// Return the blocked pair and whether an existing row was updated/inserted
	responses.SendSuccess(w, "user blocked", map[string]interface{}{"blocker_id": blockerID, "blocked_id": targetID, "rows": updated})
}

// UnblockUserHandler handles DELETE /api/users/{id}/unblock
// Authenticated user is the blocker; removes the blocked row if present.
func UnblockUserHandler(w http.ResponseWriter, r *http.Request) {
	blockerID, err := middleware.UserIDFromContext(r.Context())
	if err != nil {
		responses.SendError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 4 {
		responses.SendError(w, http.StatusBadRequest, "invalid path")
		return
	}
	targetStr := parts[3]
	targetID, err := strconv.Atoi(targetStr)
	if err != nil || targetID <= 0 {
		responses.SendError(w, http.StatusBadRequest, "invalid target id")
		return
	}

	deleted, err := queries.UnblockFollow(r.Context(), database.DB, blockerID, targetID)
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

// FollowingHandler returns list of users the authenticated user is following (status = 'accepted')
func FollowingHandler(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.UserIDFromContext(r.Context())
	if err != nil {
		responses.SendError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	users, err := queries.GetFollowingUsers(r.Context(), database.DB, userID)
	if err != nil {
		responses.SendError(w, http.StatusInternalServerError, "failed to get following list: "+err.Error())
		return
	}

	// Return the list (may be empty)
	responses.SendSuccess(w, "following list", users)
}

// FollowersHandler returns list of users who follow the authenticated user (status = 'accepted')
func FollowersHandler(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.UserIDFromContext(r.Context())
	if err != nil {
		responses.SendError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	users, err := queries.GetFollowers(r.Context(), database.DB, userID)
	if err != nil {
		responses.SendError(w, http.StatusInternalServerError, "failed to get followers list: "+err.Error())
		return
	}

	responses.SendSuccess(w, "followers list", users)
}

// BlockedHandler returns list of users the authenticated user has blocked
func BlockedHandler(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.UserIDFromContext(r.Context())
	if err != nil {
		responses.SendError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	users, err := queries.GetBlockedUsers(r.Context(), database.DB, userID)
	if err != nil {
		responses.SendError(w, http.StatusInternalServerError, "failed to get blocked users: "+err.Error())
		return
	}

	responses.SendSuccess(w, "blocked users list", users)
}

// PendingRequestsHandler returns list of users who have sent a pending follow request to the authenticated user
func PendingRequestsHandler(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.UserIDFromContext(r.Context())
	if err != nil {
		responses.SendError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	users, err := queries.GetPendingIncomingRequests(r.Context(), database.DB, userID)
	if err != nil {
		responses.SendError(w, http.StatusInternalServerError, "failed to get pending requests: "+err.Error())
		return
	}

	responses.SendSuccess(w, "pending requests", users)
}
