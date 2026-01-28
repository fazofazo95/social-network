package services

import (
	"context"
	"database/sql"
	"errors"

	queries "backend/pkg/db/queries"
)

var (
	ErrInvalidCredentials = errors.New("invalid username, email, or password")
	ErrEmailTaken         = errors.New("email already in use")
	ErrUsernameTaken      = errors.New("username already in use")
	ErrSessionFailed      = errors.New("failed to create session")
	ErrLogoutFailed       = errors.New("failed to logout")
)

// AuthService handles authentication business logic
type AuthService struct {
	db *sql.DB
}

// NewAuthService creates a new AuthService instance
func NewAuthService(db *sql.DB) *AuthService {
	return &AuthService{db: db}
}

// SignUpRequest represents the input for user signup
type SignUpRequest struct {
	Username string
	Email    string
	Password string
}

// LoginRequest represents the input for user login
type LoginRequest struct {
	Username string
	Email    string
	Password string
}

// LoginResponse represents the output of a successful login
type LoginResponse struct {
	UserID    int
	SessionID string
}

// SignUp registers a new user with email, username, and password
func (s *AuthService) SignUp(ctx context.Context, req SignUpRequest) error {
	// Validate input
	if req.Email == "" || req.Username == "" || req.Password == "" {
		return errors.New("email, username, and password are required")
	}

	// Create signup input for database query
	input := queries.SignUpInput{
		Email:    req.Email,
		Username: req.Username,
		Password: req.Password,
	}

	// Execute signup query
	if err := queries.SignUp(ctx, s.db, input); err != nil {
		// Map database errors to service errors
		if err == queries.ErrEmailTaken {
			return ErrEmailTaken
		}
		if err == queries.ErrUsernameTaken {
			return ErrUsernameTaken
		}
		return err
	}

	return nil
}

// Login authenticates a user and creates a session
func (s *AuthService) Login(ctx context.Context, req LoginRequest) (*LoginResponse, error) {
	// Validate input
	if req.Email == "" && req.Username == "" {
		return nil, errors.New("email or username is required")
	}
	if req.Password == "" {
		return nil, errors.New("password is required")
	}

	// Query user credentials
	input := queries.LogInInput{
		Email:    req.Email,
		Username: req.Username,
		Password: req.Password,
	}

	userID, err := queries.LogIn(ctx, s.db, input)
	if err != nil {
		// Map database errors to service errors
		if err == queries.ErrInvalidUsernameOrEmail || err == queries.ErrInvalidPassword {
			return nil, ErrInvalidCredentials
		}
		return nil, err
	}

	// Create session
	sessionID, err := queries.GenerateSession(ctx, s.db, userID)
	if err != nil {
		return nil, ErrSessionFailed
	}

	return &LoginResponse{
		UserID:    userID,
		SessionID: sessionID,
	}, nil
}

// Logout removes a user's session
func (s *AuthService) Logout(ctx context.Context, sessionID string, userID int) error {
	if err := queries.LogOut(ctx, s.db, sessionID, userID); err != nil {
		return ErrLogoutFailed
	}
	return nil
}
