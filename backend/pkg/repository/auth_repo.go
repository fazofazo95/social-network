package repository

import (
	"backend/pkg/models"
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

type AuthRepository interface {
	SignUp(ctx context.Context, input models.SignupFields) error
	LogIn(ctx context.Context, input models.LoginRequest) (int, error)
	LogOut(ctx context.Context, sessionCookie string, userID int) error
	CreateSession(ctx context.Context, userID int) (string, error)
}

type sqliteAuthRepo struct {
	db *sql.DB
}

func NewAuthRepository(db *sql.DB) AuthRepository {
	return &sqliteAuthRepo{db: db}
}

var (
	ErrEmailTaken      = errors.New("email already in use")
	ErrUsernameTaken   = errors.New("username already in use")
	ErrInvalidEmail    = errors.New("invalid email")
	ErrInvalidPassword = errors.New("invalid password")
)

func (r *sqliteAuthRepo) SignUp(ctx context.Context, input models.SignupFields) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		log.Printf("[ERROR] SignUp: Failed to begin transaction: %v", err)
		return err
	}
	defer tx.Rollback()

	loginUserQuery := `
        INSERT INTO login_users (email, username, password_hash)
        VALUES (?, ?, ?);
    `
	_, err = tx.ExecContext(ctx, loginUserQuery, input.Email, input.Username, input.Password)
	if err != nil {
		log.Printf("[ERROR] SignUp: login_users insertion failed: %v", err)
		return mapSignupError(err)
	}

	var userID int
	userIdQuery := `SELECT id FROM login_users WHERE email = ?`
	err = tx.QueryRowContext(ctx, userIdQuery, input.Email).Scan(&userID)
	if err != nil {
		log.Printf("[ERROR] SignUp: Failed to retrieve new UserID: %v", err)
		return err
	}

	userQuery := `
        INSERT INTO users (
            id, first_name, last_name, birthday_date, relationship_status,
            employed_at, phone_number, profile_picture, pictures, nickname, about_me
        ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);
    `
	_, err = tx.ExecContext(ctx, userQuery, userID, input.FirstName, input.LastName, input.Birthday, nil, nil, nil, input.Avatar, nil, input.Nickname, input.AboutMe)
	if err != nil {
		log.Printf("[ERROR] SignUp: users table insertion failed: %v", err)
		return err
	}

	_, err = tx.ExecContext(ctx, `INSERT INTO user_settings (id, email_vis, birthday_date_vis, relationship_status_vis, 
	employed_at_vis, phone_number_vis, about_me_vis, nickname_vis, follow_vis) 
	VALUES (?, 0, 1, 1, 1, 0, 1, 1, 0) ON CONFLICT(id) DO NOTHING;`, userID)
	if err != nil {
		log.Printf("[ERROR] SignUp: failed to insert default user_settings for user %d: %v", userID, err)
		return err
	}

	err = tx.Commit()
	if err != nil {
		log.Printf("[ERROR] SignUp: Transaction commit failed: %v", err)
		return err
	}

	return nil
}

func (r *sqliteAuthRepo) LogIn(ctx context.Context, input models.LoginRequest) (int, error) {
	var storedHash string
	var userID int

	query := `
        SELECT id, password_hash FROM login_users 
        WHERE email = ?`

	err := r.db.QueryRowContext(ctx, query, input.Email).Scan(&userID, &storedHash)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, ErrInvalidEmail
		}
		log.Printf("[ERROR] LogIn: Database error for email %s: %v", input.Email, err)
		return 0, err
	}

	err = bcrypt.CompareHashAndPassword([]byte(storedHash), []byte(input.Password))
	if err != nil {
		return 0, ErrInvalidPassword
	}

	return userID, nil
}

func (r *sqliteAuthRepo) LogOut(ctx context.Context, sessionCookie string, userID int) error {
	query := `DELETE FROM sessions WHERE session_id = ? AND id = ?;`
	_, err := r.db.ExecContext(ctx, query, sessionCookie, userID)
	if err != nil {
		log.Printf("[ERROR] LogOut: Database error for UserID %d: %v", userID, err)
		return err
	}

	return nil
}

func (r *sqliteAuthRepo) CreateSession(ctx context.Context, userID int) (string, error) {
	sessionIDBytes := make([]byte, 32)
	if _, err := rand.Read(sessionIDBytes); err != nil {
		log.Printf("[ERROR] CreateSession: Random bytes generation failed: %v", err)
		return "", fmt.Errorf("failed to generate random bytes: %w", err)
	}
	sessionID := hex.EncodeToString(sessionIDBytes)

	query := `INSERT INTO sessions (id, session_id)
        VALUES (?, ?)
        ON CONFLICT(id) DO UPDATE SET
            session_id = excluded.session_id;`
	_, err := r.db.ExecContext(ctx, query, userID, sessionID)
	if err != nil {
		log.Printf("[ERROR] CreateSession: Database insert/update failed for UserID %d: %v", userID, err)
		return "", fmt.Errorf("failed to insert session: %w", err)
	}

	return sessionID, nil
}

func mapSignupError(err error) error {
	msg := err.Error()
	lowMsg := strings.ToLower(msg)

	if strings.Contains(lowMsg, "email") {
		return ErrEmailTaken
	}
	if strings.Contains(lowMsg, "username") {
		return ErrUsernameTaken
	}

	log.Printf("[ERROR] SignUp: Unmapped database error: %s", msg)
	return err
}
