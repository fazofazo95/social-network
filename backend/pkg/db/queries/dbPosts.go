package queries

import (
	"context"
	"database/sql"
	"log"

	"backend/pkg/models"
)

func GetFollowedUsersPosts(ctx context.Context, db *sql.DB, currentUserID int, limit int) ([]models.Post, error) {
	if limit <= 0 {
		limit = 5
	}

	log.Printf("[INFO] GetFollowedUsersPosts: Fetching posts for UserID: %d, Limit: %d", currentUserID, limit)

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

	rows, err := db.QueryContext(ctx, query, currentUserID, limit)
	if err != nil {
		log.Printf("[ERROR] GetFollowedUsersPosts query failed: %v", err)
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
			log.Printf("[ERROR] GetFollowedUsersPosts scan failed: %v", err)
			return nil, err
		}
		posts = append(posts, post)
	}

	if err = rows.Err(); err != nil {
		log.Printf("[ERROR] GetFollowedUsersPosts rows error: %v", err)
		return nil, err
	}

	log.Printf("[SUCCESS] GetFollowedUsersPosts: Found %d posts for UserID: %d", len(posts), currentUserID)

	return posts, nil
}
