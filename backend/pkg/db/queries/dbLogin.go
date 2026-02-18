package queries

import (
	"context"
	"database/sql"
	"errors"
	"log"

	"backend/pkg/models"

	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidEmail    = errors.New("invalid email")
	ErrInvalidPassword = errors.New("invalid password")
)

func LogIn(ctx context.Context, db *sql.DB, input models.LoginInput) (int, error) {
	log.Printf("[INFO] LogIn: Attempt for email: %s", input.Email)

	var storedHash string
	var userID int

	query := `
        SELECT id, password_hash FROM login_users 
        WHERE email = ?`

	err := db.QueryRowContext(ctx, query, input.Email).Scan(&userID, &storedHash)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Printf("[WARN] LogIn: Email not found: %s", input.Email)
			return 0, ErrInvalidEmail
		}
		log.Printf("[ERROR] LogIn: Database error for email %s: %v", input.Email, err)
		return 0, err
	}

	err = bcrypt.CompareHashAndPassword([]byte(storedHash), []byte(input.Password))
	if err != nil {
		log.Printf("[WARN] LogIn: Invalid password for email: %s", input.Email)
		return 0, ErrInvalidPassword
	}

	log.Printf("[SUCCESS] LogIn: User authenticated. UserID: %d", userID)
	return userID, nil
}

func LogOut(ctx context.Context, db *sql.DB, sessionCookie string, userID int) error {
	log.Printf("[INFO] LogOut: Attempting to delete session for UserID: %d", userID)

	query := `DELETE FROM sessions WHERE session_id = ? AND id = ?;`
	res, err := db.ExecContext(ctx, query, sessionCookie, userID)
	if err != nil {
		log.Printf("[ERROR] LogOut: Database error for UserID %d: %v", userID, err)
		return err
	}

	rows, _ := res.RowsAffected()
	log.Printf("[SUCCESS] LogOut: Session deleted for UserID: %d. Rows affected: %d", userID, rows)
	return nil
}
