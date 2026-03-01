package services

import (
	"context"
	"database/sql"
	"errors"
	"log"

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


// UpdateProfile updates an existing user profile
func (s *UserService) UpdateProfile(ctx context.Context, userID int, req models.UserProfileRequest) error {
	log.Printf("[INFO] UserService.UpdateProfile: Starting for UserID: %d", userID)

	// Verify user profile exists first
	_, err := queries.GetUserByID(ctx, s.db, userID)
	if err == sql.ErrNoRows {
		log.Printf("[WARN] UserService.UpdateProfile: Profile not found for UserID: %d", userID)
		return ErrUserProfileNotFound
	}
	if err != nil {
		log.Printf("[ERROR] UserService.UpdateProfile: Database check failed for UserID %d: %v", userID, err)
		return err
	}

	// Validate required fields
	if req.FirstName == "" || req.LastName == "" || req.Level == "" {
		log.Println("[WARN] UserService.UpdateProfile: Validation failed - missing required fields")
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
		log.Printf("[ERROR] UserService.UpdateProfile: Database update failed for UserID %d: %v", userID, err)
		return ErrUpdateFailed
	}

	log.Printf("[SUCCESS] UserService.UpdateProfile: Profile updated for UserID: %d", userID)
	return nil
}

// GetProfile retrieves a user profile by ID
func (s *UserService) GetProfile(ctx context.Context, userID int) (*models.UserProfileResponse, error) {
	log.Printf("[INFO] UserService.GetProfile: Fetching Profile for UserID: %d", userID)

	profile, err := queries.GetUserByID(ctx, s.db, userID)
	if err == sql.ErrNoRows {
		log.Printf("[WARN] UserService.GetProfile: Profile not found for UserID: %d", userID)
		return nil, ErrUserProfileNotFound
	}
	if err != nil {
		log.Printf("[ERROR] UserService.GetProfile: Database query failed for UserID %d: %v", userID, err)
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

	log.Printf("[SUCCESS] UserService.GetProfile: Profile retrieved for UserID: %d", userID)
	return response, nil
}

func (s *UserService) DiscoveredUser(ctx context.Context, currentUserID int, limit int) ([]models.DiscoveredUser, error) {
	log.Printf("[INFO] UserService.DiscoveredUser: Finding users for CurrentUserID: %d (Limit: %d)", currentUserID, limit)

	if limit <= 0 {
		limit = 5
	}

	users, err := queries.DiscoverUsers(ctx, s.db, currentUserID, limit)
	if err != nil {
		log.Printf("[ERROR] UserService.DiscoveredUser: Query failed for UserID %d: %v", currentUserID, err)
		return nil, err
	}

	log.Printf("[SUCCESS] UserService.DiscoveredUser: Found %d users for CurrentUserID: %d", len(users), currentUserID)
	return users, nil
}
