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

	query := "INSERT INTO posts (user_id, content, extra_content, privacy) VALUES (?,?,?,?)"
	res, err := tx.ExecContext(ctx, query, post.UserID, post.Content, post.Image, post.Privacy)
	if err != nil {
		return err
	}

	if post.Privacy == "custom" && len(post.WhitelistedUsers) > 0 {
		postID, _ := res.LastInsertId()
		addPermissionQuery := `INSERT INTO post_permissions (post_id, user_id) VALUES (?,?)`

		for _, userID := range post.WhitelistedUsers {
			if _, err := tx.ExecContext(ctx, addPermissionQuery, postID, userID); err != nil {
				return err
			}
		}
	}
	return tx.Commit()
}

func GetPostByID(ctx context.Context, db *sql.DB, postID int, viewerID int) (*models.Post, error) {
	var post models.Post
	query := `
    SELECT 
        p.id, p.user_id, p.content, COALESCE(p.extra_content, ''), p.created_at,
        u.first_name, u.last_name, COALESCE(u.profile_picture, ''), p.privacy
    FROM active_posts p
    JOIN users u ON p.user_id = u.id
    WHERE p.id = ? 
    AND (
        p.user_id = ? -- Ιδιοκτήτης
        OR p.privacy = 'public'
        OR (p.privacy = 'followers' AND EXISTS (
            SELECT 1 FROM followers WHERE follower_id = ? AND followed_id = p.user_id
        ))
        OR (p.privacy = 'custom' AND EXISTS (
            SELECT 1 FROM post_permissions WHERE post_id = p.id AND user_id = ?
        ))
    );`

	err := db.QueryRowContext(ctx, query, postID, viewerID, viewerID, viewerID).Scan(
		&post.ID, &post.UserID, &post.Content, &post.ExtraContent, &post.CreatedAt,
		&post.AuthorFirstName, &post.AuthorLastName, &post.AuthorProfilePicture, &post.Privacy,
	)

	if err != nil {
		return nil, err // Θα επιστρέψει sql.ErrNoRows αν δεν επιτρέπεται η πρόσβαση
	}
	return &post, nil
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

func UpdatePost(ctx context.Context, db *sql.DB, postID int, content string) error {
	query := "UPDATE posts SET content = ? WHERE id = ? AND deleted_at IS NULL"
	_, err := db.ExecContext(ctx, query, content, postID)
	return err
}

func DeletePost(ctx context.Context, db *sql.DB, postID int) error {
	query := "UPDATE posts SET deleted_at = CURRENT_TIMESTAMP WHERE id = ?"
	_, err := db.ExecContext(ctx, query, postID)
	return err
}

func RestorePost(ctx context.Context, db *sql.DB, postID int) error {
	query := "UPDATE posts SET deleted_at = NULL WHERE id = ?"
	_, err := db.ExecContext(ctx, query, postID)
	return err
}

func GetUserPosts(ctx context.Context, db *sql.DB, targetUserID int, viewerID int) ([]models.Post, error) {
	query := `
    SELECT 
        p.id, p.user_id, p.content, COALESCE(p.extra_content, ''), p.created_at,
        u.first_name, u.last_name, COALESCE(u.profile_picture, ''), p.privacy
    FROM active_posts p
    JOIN users u ON p.user_id = u.id
    WHERE p.user_id = ? 
    AND (
        p.privacy = 'public'
        OR (p.privacy = 'followers' AND EXISTS (
            SELECT 1 FROM followers WHERE follower_id = ? AND followed_id = ?
        ))
        OR (p.privacy = 'custom' AND EXISTS (
            SELECT 1 FROM post_permissions WHERE post_id = p.id AND user_id = ?
        ))
        OR ? = ?
    )
    ORDER BY p.created_at DESC;`

	rows, err := db.QueryContext(ctx, query, targetUserID, viewerID, targetUserID, viewerID, viewerID, targetUserID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []models.Post
	for rows.Next() {
		var post models.Post
		if err := rows.Scan(
			&post.ID, &post.UserID, &post.Content, &post.ExtraContent, &post.CreatedAt,
			&post.AuthorFirstName, &post.AuthorLastName, &post.AuthorProfilePicture, &post.Privacy,
		); err != nil {
			return nil, err
		}
		posts = append(posts, post)
	}
	return posts, nil
}

func GetFeedPosts(ctx context.Context, db *sql.DB, userID int) ([]models.Post, error) {
	query := `
    SELECT 
        p.id, p.user_id, p.content, COALESCE(p.extra_content, ''), p.created_at,
        u.first_name, u.last_name, COALESCE(u.profile_picture, ''), p.privacy
    FROM active_posts p
    JOIN users u ON p.user_id = u.id
    WHERE 
        p.user_id = ? 
        OR p.privacy = 'public'
        OR (p.privacy = 'followers' AND p.user_id IN (
            SELECT followed_id FROM followers WHERE follower_id = ?
        ))
        OR (p.privacy = 'custom' AND p.id IN (
            SELECT post_id FROM post_permissions WHERE user_id = ?
        ))
    ORDER BY p.created_at DESC;`

	rows, err := db.QueryContext(ctx, query, userID, userID, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []models.Post
	for rows.Next() {
		var post models.Post
		err := rows.Scan(
			&post.ID, &post.UserID, &post.Content, &post.ExtraContent, &post.CreatedAt,
			&post.AuthorFirstName, &post.AuthorLastName, &post.AuthorProfilePicture, &post.Privacy,
		)
		if err != nil {
			return nil, err
		}
		posts = append(posts, post)
	}
	return posts, nil
}
