package services

import (
	"context"
	"database/sql"
	"errors"
	"log"

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
	log.Println("[INFO] PostService.CreatePost: Starting process")

	if req.Content == "" {
		log.Println("[WARN] PostService.CreatePost: Empty content provided")
		return errors.New("content are required")
	}

	err := queries.CreatePost(ctx, s.db, req)
	if err != nil {
		log.Printf("[ERROR] PostService.CreatePost: Database error: %v", err)
		return err
	}

	log.Printf("[SUCCESS] PostService.CreatePost: Post created for UserID: %d", req.UserID)
	return nil
}

func (s *PostService) UpdatePost(ctx context.Context, postID int, content string) error {
	log.Printf("[INFO] PostService.UpdatePost: Updating PostID %d", postID)

	if content == "" {
		log.Println("[WARN] PostService.UpdatePost: Empty content provided")
		return errors.New("content is required")
	}

	err := queries.UpdatePost(ctx, s.db, postID, content)
	if err != nil {
		log.Printf("[ERROR] PostService.UpdatePost: Database error for PostID %d: %v", postID, err)
		return err
	}

	log.Printf("[SUCCESS] PostService.UpdatePost: PostID %d updated", postID)
	return nil
}

func (s *PostService) DeletePost(ctx context.Context, postID int) error {
	log.Printf("[INFO] PostService.DeletePost: Deleting PostID %d", postID)

	err := queries.DeletePost(ctx, s.db, postID)
	if err != nil {
		log.Printf("[ERROR] PostService.DeletePost: Database error for PostID %d: %v", postID, err)
		return err
	}

	log.Printf("[SUCCESS] PostService.DeletePost: PostID %d deleted", postID)
	return nil
}

func (s *PostService) RestorePost(ctx context.Context, postID int) error {
	log.Printf("[INFO] PostService.RestorePost: Restoring PostID %d", postID)

	err := queries.RestorePost(ctx, s.db, postID)
	if err != nil {
		log.Printf("[ERROR] PostService.RestorePost: Database error for PostID %d: %v", postID, err)
		return err
	}

	log.Printf("[SUCCESS] PostService.RestorePost: PostID %d restored", postID)
	return nil
}

func (s *PostService) GetPostByID(ctx context.Context, userID int, postID int) (*models.Post, error) {
	log.Printf("[INFO] PostService.GetPostByID: Fetching PostID %d for UserID %d", postID, userID)

	post, err := queries.GetPostByID(ctx, s.db, postID, userID)
	if err != nil {
		log.Printf("[ERROR] PostService.GetPostByID: Database error for PostID %d: %v", postID, err)
		return nil, err
	}

	log.Printf("[SUCCESS] PostService.GetPostByID: PostID %d retrieved", postID)
	return post, nil
}

func (s *PostService) GetUserPosts(ctx context.Context, targetUserID int, viewerID int, limit int, offset int) ([]models.Post, error) {
	log.Printf("[INFO] PostService.GetUserPosts: Fetching posts for TargetUserID %d (ViewerID %d)", targetUserID, viewerID)

	posts, err := queries.GetUserPosts(ctx, s.db, targetUserID, viewerID, limit, offset)
	if err != nil {
		log.Printf("[ERROR] PostService.GetUserPosts: Database error: %v", err)
		return nil, err
	}

	log.Printf("[SUCCESS] PostService.GetUserPosts: Retrieved %d posts", len(posts))
	return posts, nil
}

func (s *PostService) GetFeedPosts(ctx context.Context, userID int, limit int, offset int) ([]models.Post, error) {
	log.Printf("[INFO] PostService.GetFeedPosts: Loading feed for UserID %d (Limit: %d, Offset: %d)", userID, limit, offset)

	posts, err := queries.GetFeedPosts(ctx, s.db, userID, limit, offset)
	if err != nil {
		log.Printf("[ERROR] PostService.GetFeedPosts: Database error: %v", err)
		return nil, err
	}

	log.Printf("[SUCCESS] PostService.GetFeedPosts: Retrieved %d feed posts", len(posts))
	return posts, nil
}
