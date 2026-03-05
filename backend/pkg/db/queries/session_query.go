package queries

import (
	database "backend/pkg/db/sqlite"
	"context"
	"database/sql"
	"fmt"
	"log"
)

func AuthenticateSession(ctx context.Context, token string) (int, error) {
	log.Printf("[INFO] AuthenticateSession: Validating token")

	var userID int
	query := `SELECT id FROM sessions WHERE session_id = ? LIMIT 1`

	err := database.DB.QueryRowContext(ctx, query, token).Scan(&userID)
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
