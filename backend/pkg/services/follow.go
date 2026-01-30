package services

import (
	queries "backend/pkg/db/queries"
	"backend/pkg/models"
	"context"
	"database/sql"
	"errors"
)

type FollowService struct {
	db *sql.DB
}

func NewFollowService(db *sql.DB) *FollowService {
	return &FollowService{db: db}
}

func (s *FollowService) FollowUser(ctx context.Context, req models.FollowRequest) error {
	if req.FollowedID == 0 || req.FollowerID == 0 {
		return errors.New("follower_id and followed_id are required")
	}

	var isUserPrivate bool
	err := queries.UserPrivacy(ctx, s.db, req.FollowedID, &isUserPrivate)
	if err != nil {
		return err
	}

	status := "pending"
	if !isUserPrivate {
		status = "accepted"
	}

	err = queries.CreateFollow(ctx, s.db, req, status)
	if err != nil {
		return err
	}

	return nil
}
