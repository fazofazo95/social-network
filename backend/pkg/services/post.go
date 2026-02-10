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
