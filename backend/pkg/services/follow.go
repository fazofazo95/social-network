package services

import (
	queries "backend/pkg/db/queries"
	"backend/pkg/models"
	"context"
	"database/sql"
	"errors"
	"log"
)

type FollowService struct {
	db *sql.DB
}

func NewFollowService(db *sql.DB) *FollowService {
	return &FollowService{db: db}
}

func (s *FollowService) FollowUser(ctx context.Context, req models.FollowRequest) (string, error) {
	log.Println("[INFO] FollowService.FollowUser: Starting process")

	if req.FollowedID == 0 || req.FollowerID == 0 {
		log.Println("[WARN] FollowService.FollowUser: Missing IDs in request")
		return "", errors.New("follower_id and followed_id are required")
	}

	log.Printf("[INFO] FollowService.FollowUser: Checking privacy for UserID: %d", req.FollowedID)
	var isUserPrivate bool
	err := queries.UserPrivacy(ctx, s.db, req.FollowedID, &isUserPrivate)
	if err != nil {
		log.Printf("[ERROR] FollowService.FollowUser: Privacy check failed: %v", err)
		return "", err
	}

	status := "pending"
	if !isUserPrivate {
		log.Printf("[INFO] FollowService.FollowUser: Target User %d is public", req.FollowedID)
		status = "accepted"
	} else {
		log.Printf("[INFO] FollowService.FollowUser: Target User %d is private", req.FollowedID)
	}

	log.Printf("[INFO] FollowService.FollowUser: Creating follow record with status: %s", status)
	err = queries.CreateFollow(ctx, s.db, req, status)
	if err != nil {
		log.Printf("[ERROR] FollowService.FollowUser: Database insert failed: %v", err)
		return "", err
	}

	log.Printf("[SUCCESS] FollowService.FollowUser: Follow relationship established (Status: %s)", status)
	return status, nil
}
