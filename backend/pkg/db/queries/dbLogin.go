package queries

import (
	"context"
	"database/sql"
	"errors"

	"backend/pkg/models"

	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidUsernameOrEmail = errors.New("invalid username or email")
	ErrInvalidPassword        = errors.New("invalid password")
)

func LogIn(ctx context.Context, db *sql.DB, input models.LoginInput) (int, error) {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	var storedHash string
	var userID int
	err = tx.QueryRowContext(ctx, `
		SELECT id, password_hash FROM login_users 
		WHERE username = ? OR email = ?
	`, input.Username, input.Email).Scan(&userID, &storedHash)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, ErrInvalidUsernameOrEmail
		}
		return 0, err
	}

	err = bcrypt.CompareHashAndPassword([]byte(storedHash), []byte(input.Password))
	if err != nil {
		return 0, ErrInvalidPassword
	}

	return userID, tx.Commit()
}

func LogOut(ctx context.Context, db *sql.DB, sessionCookie string, userID int) error{
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	_, err = tx.ExecContext(ctx, `
		DELETE FROM sessions WHERE session_id = ? AND id = ?;
	`, sessionCookie,userID)
	return tx.Commit()
}