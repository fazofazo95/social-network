package queries

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	"log"
)

func CreateSession(ctx context.Context, db *sql.DB, userID int) (string, error) {
	log.Printf("[INFO] CreateSession: Generating session for UserID: %d", userID)

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
	_, err := db.ExecContext(ctx, query, userID, sessionID)
	if err != nil {
		log.Printf("[ERROR] CreateSession: Database insert/update failed for UserID %d: %v", userID, err)
		return "", fmt.Errorf("failed to insert session: %w", err)
	}

	log.Printf("[SUCCESS] CreateSession: Session created/updated for UserID: %d", userID)
	return sessionID, nil
}

func AuthenticateSession(ctx context.Context, db *sql.DB, token string) (int, error) {
	log.Printf("[INFO] AuthenticateSession: Validating token")

	var userID int
	query := `SELECT id FROM sessions WHERE session_id = ? LIMIT 1`

	err := db.QueryRowContext(ctx, query, token).Scan(&userID)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Printf("[WARN] AuthenticateSession: No session found for token")
			return 0, fmt.Errorf("session not found or expired")
		}
		log.Printf("[ERROR] AuthenticateSession: Database query failed: %v", err)
		return 0, err
	}

	log.Printf("[SUCCESS] AuthenticateSession: Token valid for UserID: %d", userID)
	return userID, nil
}
