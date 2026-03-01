package services

import (
	"context"
	"errors"
	"strings"

	"backend/pkg/models"
	"backend/pkg/repository"
)

type CommentService interface {
	CreateComment(ctx context.Context, userID int, input models.Comment) (*models.Comment, error)
	GetPostComments(ctx context.Context, postID, viewerID int) ([]*models.Comment, error)
	UpdateComment(ctx context.Context, commentID, userID int, content string) error
	DeleteComment(ctx context.Context, commentID, userID int) error
	RestoreComment(ctx context.Context, commentID, userID int) error
}

type commentService struct {
	repo     repository.CommentRepository
	postRepo repository.PostRepository
}

func NewCommentService(r repository.CommentRepository, pr repository.PostRepository) CommentService {
	return &commentService{
		repo:     r,
		postRepo: pr,
	}
}

// --- Create Comment ---
func (s *commentService) CreateComment(ctx context.Context, userID int, input models.Comment) (*models.Comment, error) {
	content := strings.TrimSpace(input.Content)
	if content == "" {
		return nil, errors.New("the comment cannot be empty")
	}

	post, err := s.postRepo.GetByID(ctx, input.ParentID, userID)
	if err != nil {
		return nil, errors.New("You do not have permission to comment on this post, or the post does not exist.")
	}

	comment := models.Comment{
		UserID:     userID,
		ParentType: "post", 
		ParentID:   post.ID,
		Content:    content,
		ExtraContent:      input.ExtraContent,
	}

	id, err := s.repo.CreateComment(ctx, comment)
	if err != nil {
		return nil, err
	}

	return s.repo.GetCommentByID(ctx, id)
}

// --- Get Comments ---
func (s *commentService) GetPostComments(ctx context.Context, postID, viewerID int) ([]*models.Comment, error) {
	_, err := s.postRepo.GetByID(ctx, postID, viewerID)
	if err != nil {
		return nil, errors.New("access denied")
	}

	return s.repo.GetPostComments(ctx, postID, viewerID)
}

// --- Mutations ---
func (s *commentService) UpdateComment(ctx context.Context, commentID, userID int, content string) error {
	ownerID, err := s.repo.GetCommentOwnerID(ctx, commentID)
	if err != nil {
		return err
	}
	if ownerID != userID {
		return errors.New("dont have permission to edit")
	}

	newContent := strings.TrimSpace(content)
	if newContent == "" {
		return errors.New("the content cannot be empty")
	}

	return s.repo.UpdateComment(ctx, commentID, newContent)
}

func (s *commentService) DeleteComment(ctx context.Context, commentID, userID int) error {
	ownerID, err := s.repo.GetCommentOwnerID(ctx, commentID)
	if err != nil {
		return err
	}

	if ownerID != userID {
		return errors.New("dont have permission to delete")
	}

	return s.repo.DeleteComment(ctx, commentID)
}

func (s *commentService) RestoreComment(ctx context.Context, commentID, userID int) error {
	ownerID, err := s.repo.GetCommentOwnerID(ctx, commentID)
	if err != nil {
		return err
	}

	if ownerID != userID {
		return errors.New("dont have permission to delete")
	}

	return s.repo.RestoreComment(ctx, commentID)
}