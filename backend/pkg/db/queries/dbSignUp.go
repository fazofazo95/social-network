package queries

import (
	"context"
	"database/sql"
	"errors"
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
    tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
    
    hash, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
    if err != nil {
        return err
    }

    loginUserQuery := `
        INSERT INTO login_users (email, username, password_hash)
        VALUES (?, ?, ?);
    `
    _, err = tx.ExecContext(ctx, loginUserQuery, input.Email, input.Username, string(hash))
    if err != nil {
        return mapSignupError(err)
    }

    var userID int
    userIdQuery := `SELECT id FROM login_users WHERE email = ?`
    err = tx.QueryRowContext(ctx, userIdQuery, input.Email).Scan(&userID)

    userQuery := `
        INSERT INTO users (id, first_name, last_name, birthday_date, profile_picture, nickname, about_me)
        VALUES (?, ?, ?, ?, ?, ?, ?);
    `
    _, err = tx.ExecContext(ctx, userQuery, userID, input.FirstName, input.LastName, input.Birthday, input.Avatar, input.Nickname, input.AboutMe)
    if err != nil {
        return err
    }

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