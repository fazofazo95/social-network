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

func (s *PostService) UpdatePost(ctx context.Context, postID int, content string) error {
	if content == "" {
		return errors.New("content is required")
	}

	return queries.UpdatePost(ctx, s.db, postID, content)
}

func (s *PostService) DeletePost(ctx context.Context, postID int) error {
	return queries.DeletePost(ctx, s.db, postID)
}

func (s *PostService) RestorePost(ctx context.Context, postID int) error {
	return queries.RestorePost(ctx, s.db, postID)
}

func (s *PostService) GetPostByID(ctx context.Context, userID int, postID int) (*models.Post, error) {
	post, err := queries.GetPostByID(ctx, s.db, postID, userID)
	if err != nil {
		return nil, err
	}

	return post, nil
}

func (s *PostService) GetUserPosts(ctx context.Context, targetUserID int, viewerID int) ([]models.Post, error) {
	posts, err := queries.GetUserPosts(ctx, s.db, targetUserID, viewerID)
	if err != nil {
		return nil, err
	}

	return posts, nil
}
