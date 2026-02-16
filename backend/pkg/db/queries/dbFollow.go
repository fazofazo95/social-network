package queries

import (
	"context"
	"database/sql"
	"log"

	"backend/pkg/models"
)

func CreateFollow(ctx context.Context, db *sql.DB, req models.FollowRequest, status string) error {
	query := `INSERT INTO followers (follower_id, followed_id, status) VALUES (?, ?, ?)`
	_, err := db.ExecContext(ctx, query, req.FollowerID, req.FollowedID, status)
	if err != nil {
		return err
	}
	return nil
}

// GetRelationshipStatus checks the relationship between current user and target user
// Returns: "Following", "Pending", "Follow Back", or "Follow"
func GetRelationshipStatus(ctx context.Context, db *sql.DB, currentUserID, targetUserID int) (string, error) {
	log.Printf("GetRelationshipStatus: checking relationship from %d -> %d", currentUserID, targetUserID)
	// First check if current user follows target user
	var status string
	query := `SELECT status FROM followers WHERE follower_id = ? AND followed_id = ?`
	log.Printf("GetRelationshipStatus: exec query current->target")
	err := db.QueryRowContext(ctx, query, currentUserID, targetUserID).Scan(&status)

	if err == nil {
		// Found a relationship from current user to target user
		if status == "accepted" {
			return "Following", nil
		} else if status == "pending" {
			return "Pending", nil
		}
	} else if err != sql.ErrNoRows {
		// An actual error occurred (not just no rows found)
		return "", err
	}

	// Check if target user follows current user (follow back scenario)
	query = `SELECT status FROM followers WHERE follower_id = ? AND followed_id = ? AND status = 'accepted'`
	log.Printf("GetRelationshipStatus: exec query target->current")
	err = db.QueryRowContext(ctx, query, targetUserID, currentUserID).Scan(&status)

	if err == nil {
		// Target user follows current user
		return "Follow Back", nil
	} else if err != sql.ErrNoRows {
		// An actual error occurred
		return "", err
	}

	// No relationship exists
	return "Follow", nil
}

// DiscoverUsers finds up to 5 random users that the current user can follow
// Excludes users with existing relationships and users who have blocked the current user
func DiscoverUsers(ctx context.Context, db *sql.DB, currentUserID int, limit int) ([]models.DiscoveredUser, error) {
	if limit <= 0 {
		limit = 5
	}


	query := `
		SELECT
			u.id,
			u.first_name,
			u.last_name,
			COALESCE(u.profile_picture, '') as profile_picture,
			CASE
				WHEN EXISTS(SELECT 1 FROM followers f WHERE f.follower_id = ? AND f.followed_id = u.id AND f.status = 'accepted') THEN 'Following'
				WHEN EXISTS(SELECT 1 FROM followers f WHERE f.follower_id = ? AND f.followed_id = u.id AND f.status = 'pending') THEN 'Pending'
				WHEN EXISTS(SELECT 1 FROM followers f WHERE f.follower_id = u.id AND f.followed_id = ? AND f.status = 'accepted') THEN 'Follow Back'
				ELSE 'Follow'
			END as status
		FROM users u
		WHERE u.id != ?
		AND u.id NOT IN (
			SELECT followed_id FROM followers WHERE follower_id = ?
		)
		AND u.id NOT IN (
			SELECT follower_id FROM followers WHERE followed_id = ? AND status = 'blocked'
		)
		ORDER BY RANDOM()
		LIMIT ?
	`

	log.Printf("DiscoverUsers: executing discover query for user %d limit %d (SQL computed status)", currentUserID, limit)
	rows, err := db.QueryContext(ctx, query, currentUserID, currentUserID, currentUserID, currentUserID, currentUserID, currentUserID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []models.DiscoveredUser
	for rows.Next() {
		var user models.DiscoveredUser
		err := rows.Scan(&user.ID, &user.FirstName, &user.LastName, &user.ProfilePicture, &user.Status)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	log.Printf("DiscoverUsers: found %d users for user %d", len(users), currentUserID)

	return users, nil
}
