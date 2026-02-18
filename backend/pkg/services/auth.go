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

// SignUp registers a new user with email, username, and password
func (s *AuthService) SignUp(ctx context.Context, req models.Signup_fields) error {
	log.Println("[INFO] AuthService.SignUp: Starting signup process")

	// Validate input
	if req.Email == "" || req.Username == "" || req.Password == "" {
		log.Println("[WARN] AuthService.SignUp: Missing required fields")
		return errors.New("email, username, and password are required")
	}

	// Execute signup query
	if err := queries.SignUp(ctx, s.db, req); err != nil {
		// Map database errors to service errors
		if err == queries.ErrEmailTaken {
			log.Printf("[WARN] AuthService.SignUp: Email conflict for %s", req.Email)
			return ErrEmailTaken
		}
		if err == queries.ErrUsernameTaken {
			log.Printf("[WARN] AuthService.SignUp: Username conflict for %s", req.Username)
			return ErrUsernameTaken
		}
		log.Printf("[ERROR] AuthService.SignUp: Database error: %v", err)
		return err
	}

	log.Println("[SUCCESS] AuthService.SignUp: User registered successfully")
	return nil
}

// Login authenticates a user and creates a session
func (s *AuthService) Login(ctx context.Context, req models.LoginRequest) (*models.LoginResponse, error) {
	log.Println("[INFO] AuthService.Login: Attempting login")

	// Validate input
	if req.Email == "" {
		log.Println("[WARN] AuthService.Login: Email is missing")
		return nil, errors.New("email is required")
	}
	if req.Password == "" {
		log.Println("[WARN] AuthService.Login: Password is missing")
		return nil, errors.New("password is required")
	}

	// Query user credentials
	input := models.LoginInput{
		Email:    req.Email,
		Password: req.Password,
	}

	userID, err := queries.LogIn(ctx, s.db, input)
	if err != nil {
		// Map database errors to service errors
		if err == queries.ErrInvalidEmail || err == queries.ErrInvalidPassword {
			log.Println("[WARN] AuthService.Login: Invalid credentials provided")
			return nil, ErrInvalidCredentials
		}
		log.Printf("[ERROR] AuthService.Login: Database query failed: %v", err)
		return nil, err
	}

	// Create session
	log.Printf("[INFO] AuthService.Login: Creating session for UserID: %d", userID)
	sessionID, err := queries.CreateSession(ctx, s.db, userID)
	if err != nil {
		log.Printf("[ERROR] AuthService.Login: Session creation failed: %v", err)
		return nil, ErrSessionFailed
	}

	log.Printf("[SUCCESS] AuthService.Login: Login successful for UserID: %d", userID)
	return &models.LoginResponse{
		UserID:    userID,
		SessionID: sessionID,
	}, nil
}

// Logout removes a user's session
func (s *AuthService) Logout(ctx context.Context, sessionID string, userID int) error {
	log.Printf("[INFO] AuthService.Logout: Attempting logout for UserID: %d", userID)

	if err := queries.LogOut(ctx, s.db, sessionID, userID); err != nil {
		log.Printf("[ERROR] AuthService.Logout: LogOut query failed: %v", err)
		return ErrLogoutFailed
	}

	log.Printf("[SUCCESS] AuthService.Logout: Session cleared for UserID: %d", userID)
	return nil
}
