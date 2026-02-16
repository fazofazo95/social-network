package queries

import (
	"context"
	"database/sql"
	"fmt"

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

func GetPostByPostID(ctx context.Context, db *sql.DB, postID int) (models.Post, error) {
	var post models.Post
	query := `
        SELECT 
            p.id,
            p.user_id,
            p.content,
            COALESCE(p.extra_content, ''),
            p.created_at,
            u.first_name,
            u.last_name,
            COALESCE(u.profile_picture, ''),
			p.privacy
        FROM posts p
        INNER JOIN users u ON p.user_id = u.id
        WHERE p.id = ?;
    `

	row := db.QueryRowContext(ctx, query, postID)

	err := row.Scan(
		&post.ID,
		&post.UserID,
		&post.Content,
		&post.ExtraContent,
		&post.CreatedAt,
		&post.AuthorFirstName,
		&post.AuthorLastName,
		&post.AuthorProfilePicture,
		&post.Privacy,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return models.Post{}, nil
		}
		return models.Post{}, err
	}

	return post, nil
}

func GetPostsByUserID(ctx context.Context, db *sql.DB, userID int) ([]models.Post, error) {
	query := `
        SELECT 
            p.id,
            p.user_id,
            p.content,
            COALESCE(p.extra_content, ''),
            p.created_at,
            u.first_name,
            u.last_name,
            COALESCE(u.profile_picture, ''),
			p.privacy
        FROM posts p
        INNER JOIN users u ON p.user_id = u.id
        WHERE p.user_id = ?
        ORDER BY p.created_at DESC
    `

	rows, err := db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []models.Post

	for rows.Next() {
		var post models.Post

		err := rows.Scan(
			&post.ID,
			&post.UserID,
			&post.Content,
			&post.ExtraContent,
			&post.CreatedAt,
			&post.AuthorFirstName,
			&post.AuthorLastName,
			&post.AuthorProfilePicture,
			&post.Privacy,
		)
		if err != nil {
			return nil, err
		}

		posts = append(posts, post)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return posts, nil
}

func GetPostOwnerID(ctx context.Context, db *sql.DB, postID int) (int, error) {
	var ownerID int
	query := "SELECT user_id FROM posts WHERE id = ?"
	err := db.QueryRowContext(ctx, query, postID).Scan(&ownerID)
	if err != nil {
		return 0, err
	}
	return ownerID, nil
}

func UpdatePost(ctx context.Context, db *sql.DB, req models.UpdateData) error {
	query := "UPDATE posts SET content = ? WHERE id = ?"
	_, err := db.ExecContext(ctx, query, req.Content, req.ParentID)
	return err
}

func GetResourceOwnerID(ctx context.Context, db *sql.DB, resourceType string, resourceID int) (int, error) {
	var ownerID int
	query := fmt.Sprintf("SELECT user_id FROM %s WHERE id = ?", resourceType)
	err := db.QueryRowContext(ctx, query, resourceID).Scan(&ownerID)
	if err != nil {
		return 0, err
	}
	return ownerID, nil
}
