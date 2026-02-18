package handlers

import (
	"log"
	"net/http"
	"strconv"
	"strings"

	"backend/pkg/db/queries"
	database "backend/pkg/db/sqlite"
	"backend/pkg/middleware"
	"backend/pkg/models"
	"backend/pkg/responses"
	"backend/pkg/services"
)

func FollowUserHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("[INFO] FollowUserHandler: Received follow request")

	followerID, err := middleware.UserIDFromContext(r.Context())
	if err != nil {
		log.Printf("[ERROR] FollowUserHandler: Unauthorized: %v", err)
		responses.SendError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 4 {
		log.Printf("[ERROR] FollowUserHandler: Invalid path: %s", r.URL.Path)
		responses.SendError(w, http.StatusBadRequest, "invalid path")
		return
	}
	targetStr := parts[3]
	targetID, err := strconv.Atoi(targetStr)
	if err != nil || targetID <= 0 {
		log.Printf("[ERROR] FollowUserHandler: Invalid target ID: %s", targetStr)
		responses.SendError(w, http.StatusBadRequest, "invalid target id")
		return
	}

	log.Printf("[INFO] FollowUserHandler: User %d attempting to follow User %d", followerID, targetID)
	followService := services.NewFollowService(database.DB)

	req := models.FollowRequest{
		FollowerID: followerID,
		FollowedID: targetID,
	}

	status, err := followService.FollowUser(r.Context(), req)
	if err != nil {
		log.Printf("[ERROR] FollowUserHandler: Service error: %v", err)
		responses.SendError(w, http.StatusInternalServerError, "Failed to process follow request: "+err.Error())
		return
	}

	log.Printf("[SUCCESS] FollowUserHandler: Relationship created with status: %s", status)
	resp := map[string]interface{}{
		"follower_id": req.FollowerID,
		"followed_id": req.FollowedID,
		"status":      status,
	}

	responses.SendCreated(w, "follow request created successfully", resp)
}

func UnfollowUserHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("[INFO] UnfollowUserHandler: Received unfollow request")

	followerID, err := middleware.UserIDFromContext(r.Context())
	if err != nil {
		log.Printf("[ERROR] UnfollowUserHandler: Unauthorized: %v", err)
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

	log.Printf("[INFO] UnfollowUserHandler: User %d unfollowing User %d", followerID, targetID)
	deleted, err := queries.DeleteFollow(r.Context(), database.DB, followerID, targetID)
	if err != nil {
		log.Printf("[ERROR] UnfollowUserHandler: Query error: %v", err)
		responses.SendError(w, http.StatusInternalServerError, "failed to unfollow: "+err.Error())
		return
	}

	if deleted == 0 {
		log.Printf("[WARN] UnfollowUserHandler: No relationship found to delete")
		responses.SendError(w, http.StatusNotFound, "no follow relationship found")
		return
	}

	log.Printf("[SUCCESS] UnfollowUserHandler: Relationship deleted")
	responses.SendSuccess(w, "unfollowed successfully", map[string]interface{}{"follower_id": followerID, "followed_id": targetID})
}

func AcceptFollowHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("[INFO] AcceptFollowHandler: Received accept request")

	followedID, err := middleware.UserIDFromContext(r.Context())
	if err != nil {
		log.Printf("[ERROR] AcceptFollowHandler: Unauthorized: %v", err)
		responses.SendError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 5 {
		responses.SendError(w, http.StatusBadRequest, "invalid path")
		return
	}
	followerStr := parts[3]
	followerID, err := strconv.Atoi(followerStr)
	if err != nil || followerID <= 0 {
		responses.SendError(w, http.StatusBadRequest, "invalid follower id")
		return
	}

	log.Printf("[INFO] AcceptFollowHandler: User %d accepting request from User %d", followedID, followerID)
	updated, err := queries.AcceptFollow(r.Context(), database.DB, followerID, followedID)
	if err != nil {
		log.Printf("[ERROR] AcceptFollowHandler: Query error: %v", err)
		responses.SendError(w, http.StatusInternalServerError, "failed to accept follow: "+err.Error())
		return
	}

	if updated == 0 {
		log.Printf("[WARN] AcceptFollowHandler: No pending request found")
		responses.SendError(w, http.StatusNotFound, "no pending follow request found")
		return
	}

	log.Printf("[SUCCESS] AcceptFollowHandler: Request accepted")
	responses.SendSuccess(w, "follow request accepted", map[string]interface{}{"follower_id": followerID, "followed_id": followedID})
}

func BlockUserHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("[INFO] BlockUserHandler: Received block request")

	blockerID, err := middleware.UserIDFromContext(r.Context())
	if err != nil {
		log.Printf("[ERROR] BlockUserHandler: Unauthorized: %v", err)
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

	log.Printf("[INFO] BlockUserHandler: User %d blocking User %d", blockerID, targetID)
	updated, err := queries.BlockFollow(r.Context(), database.DB, blockerID, targetID)
	if err != nil {
		log.Printf("[ERROR] BlockUserHandler: Query error: %v", err)
		responses.SendError(w, http.StatusInternalServerError, "failed to block user: "+err.Error())
		return
	}

	log.Printf("[SUCCESS] BlockUserHandler: User blocked. Rows affected: %d", updated)
	responses.SendSuccess(w, "user blocked", map[string]interface{}{"blocker_id": blockerID, "blocked_id": targetID, "rows": updated})
}

func UnblockUserHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("[INFO] UnblockUserHandler: Received unblock request")

	blockerID, err := middleware.UserIDFromContext(r.Context())
	if err != nil {
		log.Printf("[ERROR] UnblockUserHandler: Unauthorized: %v", err)
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

	log.Printf("[INFO] UnblockUserHandler: User %d unblocking User %d", blockerID, targetID)
	deleted, err := queries.UnblockFollow(r.Context(), database.DB, blockerID, targetID)
	if err != nil {
		log.Printf("[ERROR] UnblockUserHandler: Query error: %v", err)
		responses.SendError(w, http.StatusInternalServerError, "failed to unblock user: "+err.Error())
		return
	}
	if deleted == 0 {
		log.Printf("[WARN] UnblockUserHandler: No blocked relationship found")
		responses.SendError(w, http.StatusNotFound, "no blocked relationship found")
		return
	}

	log.Printf("[SUCCESS] UnblockUserHandler: User unblocked")
	responses.SendSuccess(w, "user unblocked", map[string]interface{}{"blocker_id": blockerID, "unblocked_id": targetID})
}

func FollowingHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("[INFO] FollowingHandler: Fetching following list")

	userID, err := middleware.UserIDFromContext(r.Context())
	if err != nil {
		log.Printf("[ERROR] FollowingHandler: Unauthorized: %v", err)
		responses.SendError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	users, err := queries.GetFollowingUsers(r.Context(), database.DB, userID)
	if err != nil {
		log.Printf("[ERROR] FollowingHandler: Query error: %v", err)
		responses.SendError(w, http.StatusInternalServerError, "failed to get following list: "+err.Error())
		return
	}

	log.Printf("[SUCCESS] FollowingHandler: Found %d users", len(users))
	responses.SendSuccess(w, "following list", users)
}

func FollowersHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("[INFO] FollowersHandler: Fetching followers list")

	userID, err := middleware.UserIDFromContext(r.Context())
	if err != nil {
		log.Printf("[ERROR] FollowersHandler: Unauthorized: %v", err)
		responses.SendError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	users, err := queries.GetFollowers(r.Context(), database.DB, userID)
	if err != nil {
		log.Printf("[ERROR] FollowersHandler: Query error: %v", err)
		responses.SendError(w, http.StatusInternalServerError, "failed to get followers list: "+err.Error())
		return
	}

	log.Printf("[SUCCESS] FollowersHandler: Found %d users", len(users))
	responses.SendSuccess(w, "followers list", users)
}

func BlockedHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("[INFO] BlockedHandler: Fetching blocked list")

	userID, err := middleware.UserIDFromContext(r.Context())
	if err != nil {
		log.Printf("[ERROR] BlockedHandler: Unauthorized: %v", err)
		responses.SendError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	users, err := queries.GetBlockedUsers(r.Context(), database.DB, userID)
	if err != nil {
		log.Printf("[ERROR] BlockedHandler: Query error: %v", err)
		responses.SendError(w, http.StatusInternalServerError, "failed to get blocked users: "+err.Error())
		return
	}

	log.Printf("[SUCCESS] BlockedHandler: Found %d users", len(users))
	responses.SendSuccess(w, "blocked users list", users)
}

func PendingRequestsHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("[INFO] PendingRequestsHandler: Fetching pending requests")

	userID, err := middleware.UserIDFromContext(r.Context())
	if err != nil {
		log.Printf("[ERROR] PendingRequestsHandler: Unauthorized: %v", err)
		responses.SendError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	users, err := queries.GetPendingIncomingRequests(r.Context(), database.DB, userID)
	if err != nil {
		log.Printf("[ERROR] PendingRequestsHandler: Query error: %v", err)
		responses.SendError(w, http.StatusInternalServerError, "failed to get pending requests: "+err.Error())
		return
	}

	log.Printf("[SUCCESS] PendingRequestsHandler: Found %d requests", len(users))
	responses.SendSuccess(w, "pending requests", users)
}
