package queries

import (
	"backend/pkg/models"
	"context"
	"database/sql"
)

func CreateFollow(ctx context.Context, db *sql.DB, req models.FollowRequest, status string) error {
	query := `INSERT INTO follows (follower_id, followed_id, status) VALUES (?, ?, ?)`
	_, err := db.ExecContext(ctx, query, req.FollowerID, req.FollowedID, status)
	if err != nil {
		return err
	}
	return nil
}
