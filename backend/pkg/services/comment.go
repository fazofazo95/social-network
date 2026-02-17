package services

import (
	queries "backend/pkg/db/queries"
	"backend/pkg/models"
	"context"
	"database/sql"
	"errors"
	"fmt"
)

type CommentService struct {
	db *sql.DB
}

// NewAuthService creates a new AuthService instance
func NewCommentService(db *sql.DB) *CommentService {
	return &CommentService{db: db}
}

func (s *CommentService) CreateComment(ctx context.Context, req models.Comment) error {
	if req.Content == "" {
		return errors.New("content are required")
	}

	_, err := queries.CreateComment(ctx, s.db, req)
	if err != nil {
		return err
	}

	return nil
}

func (s *CommentService) UpdateComment(ctx context.Context, commentID int, content string) error {
	if content == "" {
		return errors.New("content is required")
	}

	err := queries.UpdateComment(ctx, s.db, commentID, content)
	if err != nil {
		return fmt.Errorf("failed to update comment: %v", err)
	}

	return nil
}

func (s *CommentService) DeleteComment(ctx context.Context, commentID int) error {
	return queries.DeleteComment(ctx, s.db, commentID)
}

func (s *CommentService) RestoreComment(ctx context.Context, commentID int) error {
	return queries.RestoreComment(ctx, s.db, commentID)
}

func (s *CommentService) GetPostComments(ctx context.Context, postID int, viewerID int) ([]models.Comment, error) {
	return queries.GetPostComments(ctx, s.db, postID, viewerID)
}

func (s *CommentService) GetCommentByID(ctx context.Context, commentID int) (*models.Comment, error) {
	return queries.GetCommentByID(ctx, s.db, commentID)
}
