package services

import (
	"context"
	"database/sql"
	"errors"

	queries "backend/pkg/db/queries"
	"backend/pkg/models"
)

// AuthService handles authentication business logic
type PostService struct {
	db *sql.DB
}

// NewAuthService creates a new AuthService instance
func NewPostService(db *sql.DB) *PostService {
	return &PostService{db: db}
}

func (s *PostService) CreatePost(ctx context.Context, req models.Post) error {
	if req.Content == "" {
		return errors.New("content are required")
	}

	err := queries.CreatePost(ctx, s.db, req)
	if err != nil {
		return err
	}

	return nil
}

func (s *PostService) UpdatePost(ctx context.Context, userID int, req models.UpdateData) error {
	if req.Content == "" {
		return errors.New("content are required")
	}

	ownerID, err := queries.GetPostOwnerID(ctx, s.db, req.ParentID)
	if err != nil {
		return err
	}

	req.UserID = ownerID

	if userID != req.UserID {
		return errors.New("user does not own this post")
	}

	err = queries.UpdatePost(ctx, s.db, req)
	if err != nil {
		return err
	}

	return nil
}

func (s *PostService) IsOwner(ctx context.Context, userID int, resourceID int) (bool, error) {
	post := "posts"
	ownerID, err := queries.GetResourceOwnerID(ctx, s.db, post, resourceID)
	if err != nil {
		return false, err
	}

	return ownerID == userID, nil
}
