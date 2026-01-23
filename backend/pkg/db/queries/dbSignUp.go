package queries

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

var (
	ErrEmailTaken    = errors.New("email already in use")
	ErrUsernameTaken = errors.New("username already in use")
)

type SignUpInput struct {
	Email    string
	Username string
	Password string
}

func SignUp(ctx context.Context, db *sql.DB, input SignUpInput) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	hash, err := bcrypt.GenerateFromPassword(
		[]byte(input.Password),
		bcrypt.DefaultCost,
	)
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx, `
		INSERT INTO login_users (email, username, password_hash)
		VALUES (?, ?, ?);
	`, input.Email, input.Username, string(hash))
	if err != nil {
		return mapSignupError(err)
	}

	return tx.Commit()
}

func mapSignupError(err error) error {
	msg := err.Error()

	switch {
	case strings.Contains(msg, "login_users.email"):
		return ErrEmailTaken
	case strings.Contains(msg, "login_users.username"):
		return ErrUsernameTaken
	default:
		return err
	}
}
