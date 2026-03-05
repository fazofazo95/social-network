package services

import (
	"context"
	"errors"
	"log"

	"backend/pkg/models"
	"backend/pkg/repository"

	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidCredentials = errors.New("invalid username, email, or password")
	ErrEmailTaken         = errors.New("email already in use")
	ErrUsernameTaken      = errors.New("username already in use")
	ErrSessionFailed      = errors.New("failed to create session")
	ErrLogoutFailed       = errors.New("failed to logout")
)

type AuthService interface {
	SignUp(ctx context.Context, req models.SignupFields) error
	Login(ctx context.Context, req models.LoginRequest) (*models.LoginResponse, error)
	Logout(ctx context.Context, sessionID string, userID int) error
}

type authService struct {
	repo repository.AuthRepository
}

func NewAuthService(r repository.AuthRepository) AuthService {
	return &authService{repo: r}
}

// SignUp registers a new user with email, username, and password
func (s *authService) SignUp(ctx context.Context, req models.SignupFields) error {
	log.Println("[INFO] AuthService.SignUp: Starting signup process")

	// Validate input
	if req.Email == "" || req.Username == "" || req.Password == "" {
		log.Println("[WARN] AuthService.SignUp: Missing required fields")
		return errors.New("email, username, and password are required")
	}

	log.Printf("[INFO] SignUp: Hashing password for %s", req.Email)
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("[ERROR] SignUp: Password hashing failed: %v", err)
		return err
	}

	req.Password = string(hash)

	// Execute signup query
	if err := s.repo.SignUp(ctx, req); err != nil {
		// Map database errors to service errors
		if err == repository.ErrEmailTaken {
			log.Printf("[WARN] AuthService.SignUp: Email conflict for %s", req.Email)
			return ErrEmailTaken
		}
		if err == repository.ErrUsernameTaken {
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
func (s *authService) Login(ctx context.Context, req models.LoginRequest) (*models.LoginResponse, error) {
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
	input := models.LoginRequest{
		Email:    req.Email,
		Password: req.Password,
	}

	userID, err := s.repo.LogIn(ctx, input)
	if err != nil {
		// Map database errors to service errors
		if err == repository.ErrInvalidEmail || err == repository.ErrInvalidPassword {
			log.Println("[WARN] AuthService.Login: Invalid credentials provided")
			return nil, ErrInvalidCredentials
		}
		log.Printf("[ERROR] AuthService.Login: Database query failed: %v", err)
		return nil, err
	}

	// Create session
	log.Printf("[INFO] AuthService.Login: Creating session for UserID: %d", userID)
	sessionID, err := s.repo.CreateSession(ctx, userID)
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
func (s *authService) Logout(ctx context.Context, sessionID string, userID int) error {
	log.Printf("[INFO] AuthService.Logout: Attempting logout for UserID: %d", userID)

	if err := s.repo.LogOut(ctx, sessionID, userID); err != nil {
		log.Printf("[ERROR] AuthService.Logout: LogOut query failed: %v", err)
		return ErrLogoutFailed
	}

	log.Printf("[SUCCESS] AuthService.Logout: Session cleared for UserID: %d", userID)
	return nil
}
