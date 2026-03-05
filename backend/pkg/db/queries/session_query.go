package queries

import (
	database "backend/pkg/db/sqlite"
	"context"
	"database/sql"
	"fmt"
)

func AuthenticateSession(ctx context.Context, token string) (int, error) {

	var userID int
	query := `SELECT id FROM sessions WHERE session_id = ? LIMIT 1`

	err := database.DB.QueryRowContext(ctx, query, token).Scan(&userID)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, fmt.Errorf("session not found or expired")
		}
		return 0, err
	}

	return userID, nil
}
