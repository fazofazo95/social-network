package queries

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
)

func CreateSession(ctx context.Context, db *sql.DB, userID int) (string, error) {
	sessionIDBytes := make([]byte, 32)
	if _, err := rand.Read(sessionIDBytes); err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %w", err)
	}
	sessionID := hex.EncodeToString(sessionIDBytes)

	query := `INSERT INTO sessions (id, session_id)
        VALUES (?, ?)
        ON CONFLICT(id) DO UPDATE SET
            session_id = excluded.session_id;`
	_, err := db.ExecContext(ctx, query, userID, sessionID)
	if err != nil {
		return "", fmt.Errorf("failed to insert session: %w", err)
	}

	return sessionID, nil
}

func AuthenticateSession(ctx context.Context, db *sql.DB, token string) (int, error) {
	var userID int

	query := `SELECT id FROM sessions WHERE session_id = ? LIMIT 1`

	err := db.QueryRowContext(ctx, query, token).Scan(&userID)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, fmt.Errorf("session not found or expired")
		}
		return 0, err
	}

	return userID, nil
}
