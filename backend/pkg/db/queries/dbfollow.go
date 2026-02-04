package queries

import (
	"context"
	"database/sql"

	"backend/pkg/models"
)

func CreateFollow(ctx context.Context, db *sql.DB, req models.FollowRequest, status string) error {
	query := `INSERT INTO followers (follower_id, followed_id, status) VALUES (?, ?, ?)`
	_, err := db.ExecContext(ctx, query, req.FollowerID, req.FollowedID, status)
	if err != nil {
		return err
	}
	return nil
}
