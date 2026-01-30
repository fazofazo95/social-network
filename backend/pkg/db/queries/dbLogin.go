package queries

import (
	"context"
	"database/sql"
	"errors"

	"backend/pkg/models"

	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidEmail    = errors.New("invalid email")
	ErrInvalidPassword = errors.New("invalid password")
)

func LogIn(ctx context.Context, db *sql.DB, input models.LoginInput) (int, error) {

	var storedHash string
	var userID int

	query := `
		SELECT id, password_hash FROM login_users 
		WHERE email = ?`

	err := db.QueryRowContext(ctx, query, input.Email).Scan(&userID, &storedHash)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, ErrInvalidEmail
		}
		return 0, err
	}

	err = bcrypt.CompareHashAndPassword([]byte(storedHash), []byte(input.Password))
	if err != nil {
		return 0, ErrInvalidPassword
	}

	return userID, nil
}

func LogOut(ctx context.Context, db *sql.DB, sessionCookie string, userID int) error {
	query := `DELETE FROM sessions WHERE session_id = ? AND id = ?;`
	_, err := db.ExecContext(ctx, query, sessionCookie, userID)
	if err != nil {
		return err
	}
	return nil
}
