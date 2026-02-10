package queries

import (
	"context"
	"database/sql"
	"log"

	"backend/pkg/models"
)

// GetFollowedUsersPosts retrieves the most recent posts from users that the current user follows
func GetFollowedUsersPosts(ctx context.Context, db *sql.DB, currentUserID int, limit int) ([]models.Post, error) {
	if limit <= 0 {
		limit = 5
	}

	query := `
		SELECT 
			p.id,
			p.user_id,
			p.content,
			COALESCE(p.extra_content, '') as extra_content,
			p.created_at,
			u.first_name,
			u.last_name,
			COALESCE(u.profile_picture, '') as profile_picture
		FROM posts p
		INNER JOIN users u ON p.user_id = u.id
		WHERE p.user_id IN (
			SELECT followed_id 
			FROM followers 
			WHERE follower_id = ? AND status = 'accepted'
		)
		ORDER BY p.created_at DESC
		LIMIT ?
	`

	// log query start
	// Note: avoid importing log at top-level if unused elsewhere — use Printf directly
	log.Printf("GetFollowedUsersPosts: executing query for user %d limit %d", currentUserID, limit)
	rows, err := db.QueryContext(ctx, query, currentUserID, limit)
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
		)
		if err != nil {
			return nil, err
		}
		posts = append(posts, post)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	log.Printf("GetFollowedUsersPosts: fetched %d posts for user %d", len(posts), currentUserID)

	return posts, nil
}
