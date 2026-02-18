package services

import (
	queries "backend/pkg/db/queries"
	"backend/pkg/models"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
)

type CommentService struct {
	db *sql.DB
}

// NewAuthService creates a new AuthService instance
func NewCommentService(db *sql.DB) *CommentService {
	return &CommentService{db: db}
}

func (s *CommentService) CreateComment(ctx context.Context, req models.Comment) error {
	log.Println("[INFO] CommentService.CreateComment: Starting process")

	if req.Content == "" {
		log.Println("[WARN] CommentService.CreateComment: Empty content provided")
		return errors.New("content are required")
	}

	_, err := queries.CreateComment(ctx, s.db, req)
	if err != nil {
		log.Printf("[ERROR] CommentService.CreateComment: Database error: %v", err)
		return err
	}

	log.Printf("[SUCCESS] CommentService.CreateComment: Comment created for UserID %d on ParentID %d", req.UserID, req.ParentID)
	return nil
}

func (s *CommentService) UpdateComment(ctx context.Context, commentID int, content string) error {
	log.Printf("[INFO] CommentService.UpdateComment: Updating CommentID %d", commentID)

	if content == "" {
		log.Println("[WARN] CommentService.UpdateComment: Empty content provided")
		return errors.New("content is required")
	}

	err := queries.UpdateComment(ctx, s.db, commentID, content)
	if err != nil {
		log.Printf("[ERROR] CommentService.UpdateComment: Database error for CommentID %d: %v", commentID, err)
		return fmt.Errorf("failed to update comment: %v", err)
	}

	log.Printf("[SUCCESS] CommentService.UpdateComment: CommentID %d updated", commentID)
	return nil
}

func (s *CommentService) DeleteComment(ctx context.Context, commentID int) error {
	log.Printf("[INFO] CommentService.DeleteComment: Attempting delete for CommentID %d", commentID)

	err := queries.DeleteComment(ctx, s.db, commentID)
	if err != nil {
		log.Printf("[ERROR] CommentService.DeleteComment: Database error for CommentID %d: %v", commentID, err)
		return err
	}

	log.Printf("[SUCCESS] CommentService.DeleteComment: CommentID %d deleted", commentID)
	return nil
}

func (s *CommentService) RestoreComment(ctx context.Context, commentID int) error {
	log.Printf("[INFO] CommentService.RestoreComment: Attempting restore for CommentID %d", commentID)

	err := queries.RestoreComment(ctx, s.db, commentID)
	if err != nil {
		log.Printf("[ERROR] CommentService.RestoreComment: Database error for CommentID %d: %v", commentID, err)
		return err
	}

	log.Printf("[SUCCESS] CommentService.RestoreComment: CommentID %d restored", commentID)
	return nil
}

func (s *CommentService) GetPostComments(ctx context.Context, postID int, viewerID int) ([]models.Comment, error) {
	log.Printf("[INFO] CommentService.GetPostComments: Fetching comments for PostID %d, ViewerID %d", postID, viewerID)

	comments, err := queries.GetPostComments(ctx, s.db, postID, viewerID)
	if err != nil {
		log.Printf("[ERROR] CommentService.GetPostComments: Database error: %v", err)
		return nil, err
	}

	log.Printf("[SUCCESS] CommentService.GetPostComments: Retrieved %d comments", len(comments))
	return comments, nil
}

func (s *CommentService) GetCommentByID(ctx context.Context, commentID int) (*models.Comment, error) {
	log.Printf("[INFO] CommentService.GetCommentByID: Fetching CommentID %d", commentID)

	comment, err := queries.GetCommentByID(ctx, s.db, commentID)
	if err != nil {
		log.Printf("[ERROR] CommentService.GetCommentByID: Database error for CommentID %d: %v", commentID, err)
		return nil, err
	}

	log.Printf("[SUCCESS] CommentService.GetCommentByID: CommentID %d retrieved", commentID)
	return comment, nil
}
