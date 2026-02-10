package queries

import (
	"context"
	"database/sql"

	"backend/pkg/models"
)

func CreatePost(ctx context.Context, db *sql.DB, post models.Post) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	query := "INSERT INTO posts (user_id, content, extra_content) VALUES (?,?,?)"
	res, err := tx.ExecContext(ctx, query, post.UserID, post.Content, post.Image)
	if err != nil {
		return err
	}

	if post.Privacy == "custom" && len(post.WhitelistedUsers) >= 1 {

		postID, err := res.LastInsertId()
		if err != nil {
			return err
		}

		addPermissionQuery := `INSERT INTO post_permissions (post_id, user_id) VALUES (?,?)`
		stmt, err := tx.PrepareContext(ctx, addPermissionQuery)
		if err != nil {
			return nil
		}
		defer stmt.Close()

		for _, userID := range post.WhitelistedUsers {
			if _, err := stmt.ExecContext(ctx, postID, userID); err != nil {
				return err
			}
		}
	}

	return tx.Commit()
}

func GetPostByID(ctx context.Context, db *sql.DB, postID int) (models.Post, error) {
	var post models.Post
	query := "SELECT id, user_id, content, extra_content, created_at, privacy FROM posts WHERE id = ?"

	row := db.QueryRowContext(ctx, query, postID)

	err := row.Scan(&post)
	if err != nil {
		return models.Post{}, err
	}

	return post, nil
}

func GetPostsByUserID(ctx context.Context, db *sql.DB, userID int) ([]models.Post, error) {
	query := "SELECT id, user_id, content, extra_content, created_at, privacy FROM posts WHERE user_id = ?"

	rows, err := db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var posts []models.Post

	for rows.Next() {
		var post models.Post
		if err := rows.Scan(
			&post.ID,
			&post.UserID,
			&post.Content,
			&post.Image,
			&post.CreatedAt,
			&post.Privacy,
		); err != nil {
			return nil, err
		}
	}

	return posts, nil
}
