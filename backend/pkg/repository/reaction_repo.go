package repository

import (
	"context"
	"database/sql"
)

type ReactionRepository interface {
	AddReaction(ctx context.Context, targetID int, userID int) (int, error)
	RemoveReaction(ctx context.Context, targetID int, userID int) (int, error)
}

type sqliteReactionRepo struct {
	db *sql.DB
}

func NewReactionRepository(db *sql.DB) ReactionRepository {
	return &sqliteReactionRepo{
		db: db,
	}
}

func (r *sqliteReactionRepo) AddReaction(ctx context.Context, targetID int, userID int) (int, error) {
	tx, err := r.db.BeginTx(ctx, nil)
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

	var count int
	err = r.db.QueryRowContext(ctx, "SELECT like_count FROM posts WHERE id = ?", targetID).Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (r *sqliteReactionRepo) RemoveReaction(ctx context.Context, targetID int, userID int) (int, error) {
	tx, err := r.db.BeginTx(ctx, nil)
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
	err = r.db.QueryRowContext(ctx, "SELECT like_count FROM posts WHERE id = ?", targetID).Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
}
