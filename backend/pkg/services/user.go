package services

import (
	"context"
	"database/sql"
	"errors"
	"time"

	queries "backend/pkg/db/queries"
	"backend/pkg/models"
)

var (
	ErrUserNotFound        = errors.New("user not found")
	ErrUserProfileNotFound = errors.New("user profile not found")
	ErrUpdateFailed        = errors.New("failed to update user profile")
	ErrCreateProfileFailed = errors.New("failed to create user profile")
)

// UserService handles user-related business logic
type UserService struct {
	db *sql.DB
}

// NewUserService creates a new UserService instance
func NewUserService(db *sql.DB) *UserService {
	return &UserService{db: db}
}

// CreateProfile creates a new user profile
func (s *UserService) CreateProfile(ctx context.Context, userID int, req models.UserProfileRequest) error {
	// Validate required fields
	if req.FirstName == "" || req.LastName == "" || req.Level == "" {
		return errors.New("first name, last name, and level are required")
	}

	// Create profile input for database query
	input := models.UserProfileInput{
		ID:                 userID,
		FirstName:          req.FirstName,
		LastName:           req.LastName,
		Birthday:           req.Birthday,
		RelationshipStatus: req.RelationshipStatus,
		EmployedAt:         req.EmployedAt,
		PhoneNumber:        req.PhoneNumber,
		ProfilePicture:     req.ProfilePicture,
		Pictures:           req.Pictures,
		Level:              req.Level,
	}

	// Execute create profile query
	if err := queries.CreateUserProfile(ctx, s.db, input); err != nil {
		return ErrCreateProfileFailed
	}

	return nil
}

// UpdateProfile updates an existing user profile
func (s *UserService) UpdateProfile(ctx context.Context, userID int, req models.UserProfileRequest) error {
	// Verify user profile exists first
	_, err := queries.GetUserByID(ctx, s.db, userID)
	if err == sql.ErrNoRows {
		return ErrUserProfileNotFound
	}
	if err != nil {
		return err
	}

	// Validate required fields
	if req.FirstName == "" || req.LastName == "" || req.Level == "" {
		return errors.New("first name, last name, and level are required")
	}

	// Create profile input for database query
	input := models.UserProfileInput{
		ID:                 userID,
		FirstName:          req.FirstName,
		LastName:           req.LastName,
		Birthday:           req.Birthday,
		RelationshipStatus: req.RelationshipStatus,
		EmployedAt:         req.EmployedAt,
		PhoneNumber:        req.PhoneNumber,
		ProfilePicture:     req.ProfilePicture,
		Pictures:           req.Pictures,
		Level:              req.Level,
	}

	// Execute update profile query
	if err := queries.UpdateUserProfile(ctx, s.db, input); err != nil {
		return ErrUpdateFailed
	}

	return nil
}

// GetProfile retrieves a user profile by ID
func (s *UserService) GetProfile(ctx context.Context, userID int) (*models.UserProfileResponse, error) {
	profile, err := queries.GetUserByID(ctx, s.db, userID)
	if err == sql.ErrNoRows {
		return nil, ErrUserProfileNotFound
	}
	if err != nil {
		return nil, err
	}

	// Convert database profile to response
	response := &models.UserProfileResponse{
		ID:        profile.ID,
		FirstName: profile.FirstName,
		LastName:  profile.LastName,
		Level:     profile.Level,
	}

	// Handle nullable fields
	if profile.Birthday.Valid {
		response.Birthday = &profile.Birthday.Time
	}
	if profile.RelationshipStatus.Valid {
		response.RelationshipStatus = &profile.RelationshipStatus.String
	}
	if profile.EmployedAt.Valid {
		response.EmployedAt = &profile.EmployedAt.String
	}
	if profile.PhoneNumber.Valid {
		response.PhoneNumber = &profile.PhoneNumber.String
	}
	if profile.ProfilePicture.Valid {
		response.ProfilePicture = &profile.ProfilePicture.String
	}
	if profile.Pictures.Valid {
		response.Pictures = &profile.Pictures.String
	}

	return response, nil
}

// MarkProfileComplete marks a user's profile as complete
func (s *UserService) MarkProfileComplete(ctx context.Context, userID int) error {
	// Verify user profile exists first
	_, err := queries.GetUserByID(ctx, s.db, userID)
	if err == sql.ErrNoRows {
		return ErrUserProfileNotFound
	}
	if err != nil {
		return err
	}

	if err := queries.MarkProfileComplete(ctx, s.db, userID); err != nil {
		return errors.New("failed to mark profile as complete")
	}

	return nil
}

// CleanupStaleProfiles deletes incomplete user profiles older than the specified duration
func (s *UserService) CleanupStaleProfiles(ctx context.Context, olderThan time.Duration) (int, error) {
	deleted, err := queries.DeleteStaleIncompleteUsers(ctx, s.db, olderThan)
	if err != nil {
		return 0, errors.New("failed to cleanup stale profiles")
	}
	return int(deleted), nil
}
