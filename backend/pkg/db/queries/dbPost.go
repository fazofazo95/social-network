package queries

import (
	"context"
	"database/sql"
	"log"

	"backend/pkg/models"
)

func CreatePost(ctx context.Context, db *sql.DB, post models.Post) error {
	log.Printf("[INFO] CreatePost: UserID: %d, Privacy: %s", post.UserID, post.Privacy)
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		log.Printf("[ERROR] CreatePost transaction begin failed: %v", err)
		return err
	}
	defer tx.Rollback()

	query := "INSERT INTO posts (user_id, content, extra_content, privacy) VALUES (?,?,?,?)"
	res, err := tx.ExecContext(ctx, query, post.UserID, post.Content, post.Image, post.Privacy)
	if err != nil {
		log.Printf("[ERROR] CreatePost insert failed: %v", err)
		return err
	}

	if post.Privacy == "custom" && len(post.WhitelistedUsers) > 0 {
		postID, _ := res.LastInsertId()
		log.Printf("[INFO] CreatePost: Adding custom permissions for PostID: %d", postID)
		addPermissionQuery := `INSERT INTO post_permissions (post_id, user_id) VALUES (?,?)`

		for _, userID := range post.WhitelistedUsers {
			if _, err := tx.ExecContext(ctx, addPermissionQuery, postID, userID); err != nil {
				log.Printf("[ERROR] CreatePost permission insert failed for UserID %d: %v", userID, err)
				return err
			}
		}
	}

	err = tx.Commit()
	if err != nil {
		log.Printf("[ERROR] CreatePost commit failed: %v", err)
		return err
	}
	log.Printf("[SUCCESS] CreatePost: Post created successfully")
	return nil
}

func GetPostByID(ctx context.Context, db *sql.DB, postID int, viewerID int) (*models.Post, error) {
	log.Printf("[INFO] GetPostByID: PostID: %d, ViewerID: %d", postID, viewerID)
	var post models.Post
	query := `
    SELECT 
        p.id, p.user_id, p.content, COALESCE(p.extra_content, ''), p.created_at,
        u.first_name, u.last_name, COALESCE(u.profile_picture, ''), p.privacy
    FROM active_posts p
    JOIN users u ON p.user_id = u.id
    WHERE p.id = ? 
    AND (
        p.user_id = ? 
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
		if err == sql.ErrNoRows {
			log.Printf("[WARN] GetPostByID: Post not found or access denied. PostID: %d, ViewerID: %d", postID, viewerID)
		} else {
			log.Printf("[ERROR] GetPostByID failed: %v", err)
		}
		return nil, err
	}
	return &post, nil
}

func GetPostsByUserID(ctx context.Context, db *sql.DB, userID int) ([]models.Post, error) {
	log.Printf("[INFO] GetPostsByUserID: Fetching all posts for UserID: %d", userID)
	query := `
        SELECT 
            p.id, p.user_id, p.content, COALESCE(p.extra_content, ''), p.created_at,
            u.first_name, u.last_name, COALESCE(u.profile_picture, ''), p.privacy
        FROM posts p
        INNER JOIN users u ON p.user_id = u.id
        WHERE p.user_id = ?
        ORDER BY p.created_at DESC
    `

	rows, err := db.QueryContext(ctx, query, userID)
	if err != nil {
		log.Printf("[ERROR] GetPostsByUserID query failed: %v", err)
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
			log.Printf("[ERROR] GetPostsByUserID scan failed: %v", err)
			return nil, err
		}
		posts = append(posts, post)
	}

	log.Printf("[SUCCESS] GetPostsByUserID: Found %d posts", len(posts))
	return posts, nil
}

func GetPostOwnerID(ctx context.Context, db *sql.DB, postID int) (int, error) {
	log.Printf("[INFO] GetPostOwnerID: PostID: %d", postID)
	var ownerID int
	query := "SELECT user_id FROM posts WHERE id = ?"
	err := db.QueryRowContext(ctx, query, postID).Scan(&ownerID)
	if err != nil {
		log.Printf("[ERROR] GetPostOwnerID failed: %v", err)
		return 0, err
	}
	return ownerID, nil
}

func UpdatePost(ctx context.Context, db *sql.DB, postID int, content string) error {
	log.Printf("[INFO] UpdatePost: PostID: %d", postID)
	query := "UPDATE posts SET content = ? WHERE id = ? AND deleted_at IS NULL"
	res, err := db.ExecContext(ctx, query, content, postID)
	if err != nil {
		log.Printf("[ERROR] UpdatePost failed: %v", err)
		return err
	}
	rows, _ := res.RowsAffected()
	log.Printf("[INFO] UpdatePost: Rows affected: %d", rows)
	return nil
}

func DeletePost(ctx context.Context, db *sql.DB, postID int) error {
	log.Printf("[INFO] DeletePost: PostID: %d", postID)
	query := "UPDATE posts SET deleted_at = CURRENT_TIMESTAMP WHERE id = ?"
	_, err := db.ExecContext(ctx, query, postID)
	if err != nil {
		log.Printf("[ERROR] DeletePost failed: %v", err)
	}
	return err
}

func RestorePost(ctx context.Context, db *sql.DB, postID int) error {
	log.Printf("[INFO] RestorePost: PostID: %d", postID)
	query := "UPDATE posts SET deleted_at = NULL WHERE id = ?"
	_, err := db.ExecContext(ctx, query, postID)
	if err != nil {
		log.Printf("[ERROR] RestorePost failed: %v", err)
	}
	return err
}

func GetUserPosts(ctx context.Context, db *sql.DB, targetUserID int, viewerID int, limit int, offset int) ([]models.Post, error) {
	log.Printf("[INFO] GetUserPosts: TargetUserID: %d, ViewerID: %d, Limit: %d, Offset: %d", targetUserID, viewerID, limit, offset)
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
            SELECT 1 FROM followers WHERE follower_id = ? AND followed_id = ? AND status = 'accepted'
        ))
        OR (p.privacy = 'custom' AND EXISTS (
            SELECT 1 FROM post_permissions WHERE post_id = p.id AND user_id = ?
        ))
        OR ? = ?
    )
    ORDER BY p.created_at DESC
    LIMIT ? OFFSET ?;`

	rows, err := db.QueryContext(ctx, query, targetUserID, viewerID, targetUserID, viewerID, viewerID, targetUserID, limit, offset)
	if err != nil {
		log.Printf("[ERROR] GetUserPosts query failed: %v", err)
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
			log.Printf("[ERROR] GetUserPosts scan failed: %v", err)
			return nil, err
		}
		posts = append(posts, post)
	}
	log.Printf("[SUCCESS] GetUserPosts: Found %d posts", len(posts))
	return posts, nil
}

func GetFeedPosts(ctx context.Context, db *sql.DB, userID int, limit int, offset int) ([]models.Post, error) {
	log.Printf("[INFO] GetFeedPosts: UserID: %d, Limit: %d, Offset: %d", userID, limit, offset)
	query := `
    SELECT 
        p.id, p.user_id, p.content, COALESCE(p.extra_content, ''), p.created_at,
        u.first_name, u.last_name, COALESCE(u.profile_picture, ''), p.privacy
    FROM active_posts p
    JOIN users u ON p.user_id = u.id
    WHERE 
        p.user_id = ? 
        OR p.privacy = 'public'
        OR (p.privacy = 'followers' AND EXISTS (
            SELECT 1 FROM followers WHERE follower_id = ? AND followed_id = p.user_id AND status = 'accepted'
        ))
        OR (p.privacy = 'custom' AND EXISTS (
            SELECT 1 FROM post_permissions WHERE post_id = p.id AND user_id = ?
        ))
    ORDER BY p.created_at DESC
    LIMIT ? OFFSET ?;`

	rows, err := db.QueryContext(ctx, query, userID, userID, userID, limit, offset)
	if err != nil {
		log.Printf("[ERROR] GetFeedPosts query failed: %v", err)
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
			log.Printf("[ERROR] GetFeedPosts scan failed: %v", err)
			return nil, err
		}
		posts = append(posts, post)
	}
	log.Printf("[SUCCESS] GetFeedPosts: Found %d posts", len(posts))
	return posts, nil
}
