package queries

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
)

func GenerateSession(ctx context.Context, db *sql.DB, userID int) (string, error) {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return "", err
	}
	defer tx.Rollback()

	// Generate a random session ID
	sessionIDBytes := make([]byte, 32)
	_, err = rand.Read(sessionIDBytes)
	if err != nil {
		return "", err
	}
	sessionID := hex.EncodeToString(sessionIDBytes)

	// Insert the session
	_, err = tx.ExecContext(ctx, `
		INSERT INTO sessions (id, session_id) VALUES (?, ?)
	`, userID, sessionID)
	if err != nil {
		return "", err
	}

	return sessionID, tx.Commit()
}

func SessionExists(ctx context.Context, db *sql.DB, sessionCookie string) (int, error) {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	var userID int
	err = tx.QueryRowContext(ctx, `
		SELECT id FROM sessions
		WHERE session_id = ?
	`, sessionCookie).Scan(&userID)
	if err != nil {
		return 0, err
	}

	return userID, tx.Commit()
}
