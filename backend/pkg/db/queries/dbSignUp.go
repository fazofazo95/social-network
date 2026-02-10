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

// SignUp inserts a new user into the database.
func SignUp(ctx context.Context, db *sql.DB, input models.Signup_fields) error {
	log.Printf("SignUp: start signup for email=%s username=%s", input.Email, input.Username)
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	log.Printf("SignUp: tx begun for email=%s", input.Email)
	defer tx.Rollback()

	log.Printf("SignUp: generating bcrypt hash for email=%s", input.Email)
	hash, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	loginUserQuery := `
        INSERT INTO login_users (email, username, password_hash)
        VALUES (?, ?, ?);
    `
	log.Printf("SignUp: inserting into login_users email=%s", input.Email)
	_, err = tx.ExecContext(ctx, loginUserQuery, input.Email, input.Username, string(hash))
	if err != nil {
		log.Printf("SignUp: insert login_users error: %v", err)
		return mapSignupError(err)
	}

	var userID int
	userIdQuery := `SELECT id FROM login_users WHERE email = ?`
	log.Printf("SignUp: querying id for email=%s", input.Email)
	err = tx.QueryRowContext(ctx, userIdQuery, input.Email).Scan(&userID)
	if err != nil {
		log.Printf("SignUp: query id error: %v", err)
		return err
	}

	userQuery := `
        INSERT INTO users (id, first_name, last_name, birthday_date, profile_picture, nickname, about_me)
        VALUES (?, ?, ?, ?, ?, ?, ?);
    `
	log.Printf("SignUp: inserting into users id=%d", userID)
	_, err = tx.ExecContext(ctx, userQuery, userID, input.FirstName, input.LastName, input.Birthday, input.Avatar, input.Nickname, input.AboutMe)
	if err != nil {
		log.Printf("SignUp: insert users error: %v", err)
		return err
	}

	log.Printf("SignUp: committing tx for id=%d", userID)
	return tx.Commit()
}

func mapSignupError(err error) error {
	msg := err.Error()

	lowMsg := strings.ToLower(msg)

	switch {
	case strings.Contains(lowMsg, "email"):
		return ErrEmailTaken
	case strings.Contains(lowMsg, "username"):
		return ErrUsernameTaken
	default:
		return err
	}
}
