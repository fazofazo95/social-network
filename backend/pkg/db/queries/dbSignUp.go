package queries

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"strings"

	"backend/pkg/models"

	"golang.org/x/crypto/bcrypt"
)

var (
	ErrEmailTaken    = errors.New("email already in use")
	ErrUsernameTaken = errors.New("username already in use")
)

func SignUp(ctx context.Context, db *sql.DB, input models.Signup_fields) error {
	log.Printf("[INFO] SignUp: Starting signup process for Email: %s, Username: %s", input.Email, input.Username)

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		log.Printf("[ERROR] SignUp: Failed to begin transaction: %v", err)
		return err
	}
	defer tx.Rollback()

	log.Printf("[INFO] SignUp: Hashing password for %s", input.Email)
	hash, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("[ERROR] SignUp: Password hashing failed: %v", err)
		return err
	}

	loginUserQuery := `
        INSERT INTO login_users (email, username, password_hash)
        VALUES (?, ?, ?);
    `
	log.Printf("[INFO] SignUp: Inserting into login_users")
	_, err = tx.ExecContext(ctx, loginUserQuery, input.Email, input.Username, string(hash))
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
	log.Printf("[INFO] SignUp: Generated UserID: %d", userID)

	userQuery := `
        INSERT INTO users (
            id, first_name, last_name, birthday_date, relationship_status,
            employed_at, phone_number, profile_picture, pictures, nickname, about_me
        ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);
    `
	log.Printf("[INFO] SignUp: Inserting profile data into users table")
	_, err = tx.ExecContext(ctx, userQuery, userID, input.FirstName, input.LastName, input.Birthday, nil, nil, nil, input.Avatar, nil, input.Nickname, input.AboutMe)
	if err != nil {
		log.Printf("[ERROR] SignUp: users table insertion failed: %v", err)
		return err
	}

	// Ensure a default user_settings row exists for this user so frontend can rely on defaults.
	if err := InsertDefaultUserSettingsTx(ctx, tx, userID); err != nil {
		log.Printf("[ERROR] SignUp: failed to insert default user_settings for user %d: %v", userID, err)
		return err
	}

	err = tx.Commit()
	if err != nil {
		log.Printf("[ERROR] SignUp: Transaction commit failed: %v", err)
		return err
	}

	log.Printf("[SUCCESS] SignUp: User %d registered successfully", userID)
	return nil
}

func mapSignupError(err error) error {
	msg := err.Error()
	lowMsg := strings.ToLower(msg)

	if strings.Contains(lowMsg, "email") {
		log.Printf("[WARN] SignUp: Email conflict detected")
		return ErrEmailTaken
	}
	if strings.Contains(lowMsg, "username") {
		log.Printf("[WARN] SignUp: Username conflict detected")
		return ErrUsernameTaken
	}

	log.Printf("[ERROR] SignUp: Unmapped database error: %s", msg)
	return err
}
