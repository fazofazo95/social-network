package queries

import (
	"context"
	"database/sql"
)

func AddReaction(ctx context.Context, db *sql.DB, targetID int, userID int) (likeCount int, err error) {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}

	_, err = tx.Exec("INSERT INTO reactions (user_id, target_type, reaction_type, target_id) VALUES (?, ?, ?, ?)", userID, "post", "like", targetID)
	if err != nil {
		tx.Rollback()
		return 0, err
	}

	_, err = tx.Exec("UPDATE posts SET like_count = like_count + 1 WHERE id = ?", targetID)
	if err != nil {
		tx.Rollback()
		return 0, err
	}

	err = tx.Commit()
	if err != nil {
		return 0, err
	}

	// Get the updated like count
	var count int
	err = db.QueryRowContext(ctx, "SELECT like_count FROM posts WHERE id = ?", targetID).Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func RemoveReaction(ctx context.Context, db *sql.DB, targetID int, userID int) (likeCount int, err error) {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}

	_, err = tx.Exec("DELETE FROM reactions WHERE user_id = ? AND target_type = ? AND target_id = ?", userID, "post", targetID)
	if err != nil {
		tx.Rollback()
		return 0, err
	}

	_, err = tx.Exec("UPDATE posts SET like_count = like_count - 1 WHERE id = ?", targetID)
	if err != nil {
		tx.Rollback()
		return 0, err
	}

	err = tx.Commit()
	if err != nil {
		return 0, err
	}

	// Get the updated like count
	var count int
	err = db.QueryRowContext(ctx, "SELECT like_count FROM posts WHERE id = ?", targetID).Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
}
