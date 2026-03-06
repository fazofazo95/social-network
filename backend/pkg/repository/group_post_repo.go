package repository

import (
	"backend/pkg/models"
	"context"
	"database/sql"
	"errors"
	"log"
)

// CreateGroupPost creates a post within a group. The actor must be an active member.
func (r *sqliteGroupRepo) CreateGroupPost(ctx context.Context, actorID, groupID int, content, image string) (*models.Post, error) {
	// Verify actor is an active group member
	var role string
	err := r.db.QueryRowContext(ctx, `
		SELECT role FROM group_members
		WHERE group_id = ? AND user_id = ? AND status = 'active'
	`, groupID, actorID).Scan(&role)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotActiveGroupMember
		}
		return nil, err
	}

	res, err := r.db.ExecContext(ctx, `
		INSERT INTO posts (user_id, group_id, content, extra_content, privacy)
		VALUES (?, ?, ?, ?, 'group')
	`, actorID, groupID, content, image)
	if err != nil {
		return nil, err
	}

	postID, err := res.LastInsertId()
	if err != nil {
		return nil, err
	}

	// Fetch the created post with author info
	var post models.Post
	err = r.db.QueryRowContext(ctx, `
		SELECT 
			p.id, p.user_id, p.group_id, p.content, COALESCE(p.extra_content, ''), p.created_at,
			u.first_name, u.last_name, COALESCE(u.profile_picture, '')
		FROM posts p
		JOIN users u ON p.user_id = u.id
		WHERE p.id = ?
	`, postID).Scan(
		&post.ID, &post.UserID, &post.GroupID, &post.Content, &post.ExtraContent, &post.CreatedAt,
		&post.AuthorFirstName, &post.AuthorLastName, &post.AuthorProfilePicture,
	)
	if err != nil {
		return nil, err
	}

	return &post, nil
}

// GetGroupPosts returns posts for a group. The viewer must be an active member.
func (r *sqliteGroupRepo) GetGroupPosts(ctx context.Context, viewerID, groupID, limit, offset int) ([]*models.Post, error) {
	// Verify viewer is an active group member
	var role string
	err := r.db.QueryRowContext(ctx, `
		SELECT role FROM group_members
		WHERE group_id = ? AND user_id = ? AND status = 'active'
	`, groupID, viewerID).Scan(&role)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotActiveGroupMember
		}
		return nil, err
	}

	rows, err := r.db.QueryContext(ctx, `
		SELECT 
			p.id, p.user_id, p.group_id, p.content, COALESCE(p.extra_content, ''), p.created_at,
			u.first_name, u.last_name, COALESCE(u.profile_picture, ''),
			p.like_count,
			EXISTS(SELECT 1 FROM reactions r WHERE r.user_id = ? AND r.target_type = 'post' AND r.target_id = p.id) as has_current_user_liked
		FROM active_posts p
		JOIN users u ON p.user_id = u.id
		WHERE p.group_id = ?
		ORDER BY p.created_at DESC
		LIMIT ? OFFSET ?
	`, viewerID, groupID, limit, offset)
	if err != nil {
		log.Printf("[ERROR] GetGroupPosts query failed: %v", err)
		return nil, err
	}
	defer rows.Close()

	posts := make([]*models.Post, 0)
	for rows.Next() {
		post := new(models.Post)
		var hasLikedInt int
		if err := rows.Scan(
			&post.ID, &post.UserID, &post.GroupID, &post.Content, &post.ExtraContent, &post.CreatedAt,
			&post.AuthorFirstName, &post.AuthorLastName, &post.AuthorProfilePicture,
			&post.LikesCount, &hasLikedInt,
		); err != nil {
			log.Printf("[ERROR] GetGroupPosts scan failed: %v", err)
			return nil, err
		}
		post.HasCurrentUserLiked = hasLikedInt == 1
		posts = append(posts, post)
	}

	return posts, nil
}

// DeleteGroupPost deletes a group post. Allowed for the post author, group owner, or group moderator.
func (r *sqliteGroupRepo) DeleteGroupPost(ctx context.Context, actorID, groupID, postID int) error {
	// Verify the post belongs to this group and is not already deleted
	var postOwnerID int
	err := r.db.QueryRowContext(ctx, `
		SELECT user_id FROM active_posts
		WHERE id = ? AND group_id = ?
	`, postID, groupID).Scan(&postOwnerID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errors.New("group post not found")
		}
		return err
	}

	// If the actor is the post author, allow deletion
	if postOwnerID == actorID {
		_, err := r.db.ExecContext(ctx, `
			UPDATE posts SET deleted_at = CURRENT_TIMESTAMP WHERE id = ?
		`, postID)
		return err
	}

	// Otherwise, check if the actor is group owner or moderator
	var actorRole string
	err = r.db.QueryRowContext(ctx, `
		SELECT role FROM group_members
		WHERE group_id = ? AND user_id = ? AND status = 'active'
	`, groupID, actorID).Scan(&actorRole)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrNotActiveGroupMember
		}
		return err
	}

	if actorRole != "owner" && actorRole != "moderator" {
		return errors.New("dont have permissions")
	}

	_, err = r.db.ExecContext(ctx, `
		UPDATE posts SET deleted_at = CURRENT_TIMESTAMP WHERE id = ?
	`, postID)
	return err
}
