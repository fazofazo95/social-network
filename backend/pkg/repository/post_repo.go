package repository

import (
	"backend/pkg/models"
	"context"
	"database/sql"
	"log"
)

type PostRepository interface {
	// --- Commands (Write Operations) ---
	Create(ctx context.Context, post *models.Post) (int64, error)
	Update(ctx context.Context, postID int, content string) error
	Delete(ctx context.Context, postID int) error
	Restore(ctx context.Context, postID int) error

	// --- Queries (Read Operations) ---
	GetByID(ctx context.Context, postID, viewerID int) (*models.Post, error)
	GetOwnerID(ctx context.Context, postID int) (int, error)

	// --- Feed Queries ---
	GetFeed(ctx context.Context, viewerID, limit, offset int) ([]*models.Post, error)
	GetByUser(ctx context.Context, targetUserID, viewerID, limit, offset int) ([]*models.Post, error)

	// --- Helper ---
	AddPermissions(ctx context.Context, postID int64, userIDs []int) error
}

type sqlitePostRepo struct {
	db *sql.DB
}

func NewPostRepository(db *sql.DB) PostRepository {
	return &sqlitePostRepo{db: db}
}

func (r *sqlitePostRepo) Create(ctx context.Context, post *models.Post) (int64, error) {

	query := "INSERT INTO posts (user_id, content, extra_content, privacy) VALUES (?,?,?,?)"
	res, err := r.db.ExecContext(ctx, query, post.UserID, post.Content, post.Image, post.Privacy)
	if err != nil {
		return 0, err
	}

	postID, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}

	return postID, nil
}

func (r *sqlitePostRepo) AddPermissions(ctx context.Context, postID int64, userIDs []int) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	addPermissionQuery := `INSERT INTO post_permissions (post_id, user_id) VALUES (?,?)`
	for _, uID := range userIDs {
		if _, err := tx.ExecContext(ctx, addPermissionQuery, postID, uID); err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (r *sqlitePostRepo) Update(ctx context.Context, postID int, content string) error {
	query := "UPDATE posts SET content = ? WHERE id = ? AND deleted_at IS NULL"
	_, err := r.db.ExecContext(ctx, query, content, postID)
	if err != nil {
		return err
	}
	return nil
}

func (r *sqlitePostRepo) Delete(ctx context.Context, postID int) error {
	query := "UPDATE posts SET deleted_at = CURRENT_TIMESTAMP WHERE id = ?"
	_, err := r.db.ExecContext(ctx, query, postID)
	if err != nil {
		return err
	}
	return nil
}

func (r *sqlitePostRepo) Restore(ctx context.Context, postID int) error {
	query := "UPDATE posts SET deleted_at = NULL WHERE id = ?"
	_, err := r.db.ExecContext(ctx, query, postID)
	if err != nil {
		return err
	}
	return nil
}

func (r *sqlitePostRepo) GetByID(ctx context.Context, postID int, viewerID int) (*models.Post, error) {
	var post models.Post
	query := `
    SELECT 
        p.id, p.user_id, p.content, COALESCE(p.extra_content, ''), p.created_at,
        u.first_name, u.last_name, COALESCE(u.profile_picture, ''), p.privacy
    FROM active_posts p
    JOIN users u ON p.user_id = u.id
    WHERE p.id = ? 
	AND NOT EXISTS (
		SELECT 1 FROM followers fb
		WHERE fb.status = 'blocked'
		  AND ((fb.follower_id = ? AND fb.followed_id = p.user_id)
			   OR (fb.follower_id = p.user_id AND fb.followed_id = ?))
	)
    AND (
        -- Group posts: viewer must be active member of the group
        (p.group_id IS NOT NULL AND EXISTS (
            SELECT 1 FROM group_members gm
            WHERE gm.group_id = p.group_id AND gm.user_id = ? AND gm.status = 'active'
        ))
        OR
        -- Regular posts: existing privacy rules
        (p.group_id IS NULL AND (
            p.user_id = ? 
            OR p.privacy = 'public'
            OR (p.privacy = 'followers' AND EXISTS (
                SELECT 1 FROM followers WHERE follower_id = ? AND followed_id = p.user_id
            ))
            OR (p.privacy = 'custom' AND EXISTS (
                SELECT 1 FROM post_permissions WHERE post_id = p.id AND user_id = ?
            ))
        ))
    );`

	err := r.db.QueryRowContext(ctx, query, postID, viewerID, viewerID, viewerID, viewerID, viewerID, viewerID).Scan(
		&post.ID, &post.UserID, &post.Content, &post.ExtraContent, &post.CreatedAt,
		&post.AuthorFirstName, &post.AuthorLastName, &post.AuthorProfilePicture, &post.Privacy,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, err
		}
		log.Printf("[ERROR] GetPostByID failed: %v", err)
		return nil, err
	}
	return &post, nil
}

func (r *sqlitePostRepo) GetOwnerID(ctx context.Context, postID int) (int, error) {
	var ownerID int
	query := "SELECT user_id FROM posts WHERE id = ?"
	err := r.db.QueryRowContext(ctx, query, postID).Scan(&ownerID)
	if err != nil {
		return 0, err
	}
	return ownerID, nil
}

func (r *sqlitePostRepo) GetByUser(ctx context.Context, targetUserID int, viewerID int, limit int, offset int) ([]*models.Post, error) {
	query := `
	SELECT 
		p.id, p.user_id, p.content, COALESCE(p.extra_content, ''), p.created_at,
		u.first_name, u.last_name, COALESCE(u.profile_picture, ''), p.privacy,
		p.like_count,
		EXISTS(SELECT 1 FROM reactions r WHERE r.user_id = ? AND r.target_type = 'post' AND r.target_id = p.id) as has_current_user_liked
	FROM active_posts p
	JOIN users u ON p.user_id = u.id
	WHERE p.user_id = ? 
	AND p.group_id IS NULL
	AND NOT EXISTS (
		SELECT 1 FROM followers fb
		WHERE fb.status = 'blocked'
		  AND ((fb.follower_id = ? AND fb.followed_id = p.user_id)
			   OR (fb.follower_id = p.user_id AND fb.followed_id = ?))
	)
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

	rows, err := r.db.QueryContext(ctx, query, viewerID, targetUserID, viewerID, viewerID, viewerID, targetUserID, viewerID, viewerID, targetUserID, limit, offset)
	if err != nil {
		log.Printf("[ERROR] GetUserPosts query failed: %v", err)
		return nil, err
	}
	defer rows.Close()

	posts := make([]*models.Post, 0)
	for rows.Next() {
		post := new(models.Post)
		var hasLikedInt int
		if err := rows.Scan(
			&post.ID, &post.UserID, &post.Content, &post.ExtraContent, &post.CreatedAt,
			&post.AuthorFirstName, &post.AuthorLastName, &post.AuthorProfilePicture, &post.Privacy,
			&post.LikesCount, &hasLikedInt,
		); err != nil {
			log.Printf("[ERROR] GetUserPosts scan failed: %v", err)
			return nil, err
		}
		post.HasCurrentUserLiked = hasLikedInt == 1
		posts = append(posts, post)
	}
	return posts, nil
}

func (r *sqlitePostRepo) GetFeed(ctx context.Context, userID int, limit int, offset int) ([]*models.Post, error) {
	query := `
	SELECT 
		p.id, p.user_id, p.content, COALESCE(p.extra_content, ''), p.created_at,
		u.first_name, u.last_name, COALESCE(u.profile_picture, ''), p.privacy,
		p.like_count,
		EXISTS(SELECT 1 FROM reactions r WHERE r.user_id = ? AND r.target_type = 'post' AND r.target_id = p.id) as has_current_user_liked
	FROM active_posts p
	JOIN users u ON p.user_id = u.id
	WHERE 
		p.group_id IS NULL
		AND NOT EXISTS (
			SELECT 1 FROM followers fb
			WHERE fb.status = 'blocked'
			  AND ((fb.follower_id = ? AND fb.followed_id = p.user_id)
				   OR (fb.follower_id = p.user_id AND fb.followed_id = ?))
		)
		AND (
		p.user_id = ? 
		OR p.privacy = 'public'
		OR (p.privacy = 'followers' AND EXISTS (
			SELECT 1 FROM followers WHERE follower_id = ? AND followed_id = p.user_id AND status = 'accepted'
		))
		OR (p.privacy = 'custom' AND EXISTS (
			SELECT 1 FROM post_permissions WHERE post_id = p.id AND user_id = ?
		))
	)
	ORDER BY p.created_at DESC
	LIMIT ? OFFSET ?;`

	rows, err := r.db.QueryContext(ctx, query, userID, userID, userID, userID, userID, userID, limit, offset)
	if err != nil {
		log.Printf("[ERROR] GetFeedPosts query failed: %v", err)
		return nil, err
	}
	defer rows.Close()

	posts := make([]*models.Post, 0)
	for rows.Next() {
		post := new(models.Post)
		var hasLikedInt int
		err := rows.Scan(
			&post.ID, &post.UserID, &post.Content, &post.ExtraContent, &post.CreatedAt,
			&post.AuthorFirstName, &post.AuthorLastName, &post.AuthorProfilePicture, &post.Privacy,
			&post.LikesCount, &hasLikedInt,
		)
		if err != nil {
			log.Printf("[ERROR] GetFeedPosts scan failed: %v", err)
			return nil, err
		}
		post.HasCurrentUserLiked = hasLikedInt == 1
		posts = append(posts, post)
	}
	return posts, nil
}
