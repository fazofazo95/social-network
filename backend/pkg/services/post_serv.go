package services

import (
	"context"
	"errors"
	"strings"

	"backend/pkg/models"
	"backend/pkg/repository"
)

type PostService interface {
	CreatePost(ctx context.Context, userID int, input *models.Post) (*models.Post, error)
	GetPost(ctx context.Context, postID, viewerID int) (*models.Post, error)
	GetUserFeed(ctx context.Context, userID, limit, offset int) ([]*models.Post, error)
	GetProfilePosts(ctx context.Context, targetUserID, viewerID, limit, offset int) ([]*models.Post, error)
	UpdatePost(ctx context.Context, postID, userID int, content string) error
	DeletePost(ctx context.Context, postID, userID int) error
	RestorePost(ctx context.Context, postID, userID int) error
}

type postService struct {
	repo     repository.PostRepository
}

func NewPostService(r repository.PostRepository) PostService {
	return &postService{
		repo:     r,
	}
}

// --- Create Post ---
func (s *postService) CreatePost(ctx context.Context, userID int, input *models.Post) (*models.Post, error) {
	content := strings.TrimSpace(input.Content)
	if content == "" && input.Image == "" {
		return nil, errors.New("post cant be empty")
	}

	post := &models.Post{
		UserID:  userID,
		Content: content,
		Image:   input.Image,
		Privacy: input.Privacy,
		WhitelistedUsers: input.WhitelistedUsers,
	}

	// 3. Execution & Transaction Handling
	id, err := s.repo.Create(ctx, post)
	if err != nil {
		return nil, err
	}

	if input.Privacy == "custom" && len(input.WhitelistedUsers) > 0 {
		if err := s.repo.AddPermissions(ctx, id, input.WhitelistedUsers); err != nil {
			return nil, err
		}
	}

	return s.repo.GetByID(ctx, int(id), userID)
}

// --- Get Single Post ---
func (s *postService) GetPost(ctx context.Context, postID, viewerID int) (*models.Post, error) {
	post, err := s.repo.GetByID(ctx, postID, viewerID)
	if err != nil {
		return nil, err 
	}
	return post, nil
}

// --- Feeds ---
func (s *postService) GetUserFeed(ctx context.Context, userID, limit, offset int) ([]*models.Post, error) {
	if limit <= 0 { limit = 10 }
	return s.repo.GetFeed(ctx, userID, limit, offset)
}

func (s *postService) GetProfilePosts(ctx context.Context, targetUserID, viewerID, limit, offset int) ([]*models.Post, error) {
	if limit <= 0 { limit = 10 }
	return s.repo.GetByUser(ctx, targetUserID, viewerID, limit, offset)
}

// --- Mutations with Ownership Check ---
func (s *postService) UpdatePost(ctx context.Context, postID, userID int, content string) error {
	ownerID, err := s.repo.GetOwnerID(ctx, postID)
	if err != nil {
		return err
	}
	if ownerID != userID {
		return errors.New("dont have permissions")
	}
	
	return s.repo.Update(ctx, postID, strings.TrimSpace(content))
}

func (s *postService) DeletePost(ctx context.Context, postID, userID int) error {
	ownerID, err := s.repo.GetOwnerID(ctx, postID)
	if err != nil {
		return err
	}
	if ownerID != userID {
		return errors.New("dont have permissions")
	}
	
	return s.repo.Delete(ctx, postID)
}

func (s *postService) RestorePost(ctx context.Context, postID, userID int) error {
	ownerID, err := s.repo.GetOwnerID(ctx, postID)
	if err != nil {
		return err
	}
	if ownerID != userID {
		return errors.New("dont have permissions")
	}
	
	return s.repo.Restore(ctx, postID)
}