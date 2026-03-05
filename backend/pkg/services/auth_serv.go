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
	if req.Email == "" || req.Username == "" || req.Password == "" {
		return errors.New("email, username, and password are required")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("[ERROR] SignUp: Password hashing failed: %v", err)
		return err
	}

	req.Password = string(hash)

	if err := s.repo.SignUp(ctx, req); err != nil {
		if err == repository.ErrEmailTaken {
			return ErrEmailTaken
		}
		if err == repository.ErrUsernameTaken {
			return ErrUsernameTaken
		}
		log.Printf("[ERROR] AuthService.SignUp: Database error: %v", err)
		return err
	}

	return nil
}

// Login authenticates a user and creates a session
func (s *authService) Login(ctx context.Context, req models.LoginRequest) (*models.LoginResponse, error) {
	if req.Email == "" {
		return nil, errors.New("email is required")
	}
	if req.Password == "" {
		return nil, errors.New("password is required")
	}

	input := models.LoginRequest{
		Email:    req.Email,
		Password: req.Password,
	}

	userID, err := s.repo.LogIn(ctx, input)
	if err != nil {
		if err == repository.ErrInvalidEmail || err == repository.ErrInvalidPassword {
			return nil, ErrInvalidCredentials
		}
		log.Printf("[ERROR] AuthService.Login: Database query failed: %v", err)
		return nil, err
	}

	sessionID, err := s.repo.CreateSession(ctx, userID)
	if err != nil {
		log.Printf("[ERROR] AuthService.Login: Session creation failed: %v", err)
		return nil, ErrSessionFailed
	}

	return &models.LoginResponse{
		UserID:    userID,
		SessionID: sessionID,
	}, nil
}

// Logout removes a user's session
func (s *authService) Logout(ctx context.Context, sessionID string, userID int) error {
	if err := s.repo.LogOut(ctx, sessionID, userID); err != nil {
		log.Printf("[ERROR] AuthService.Logout: LogOut query failed: %v", err)
		return ErrLogoutFailed
	}

	return nil
}
