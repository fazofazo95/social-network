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

// DeleteFollow removes any follow relationship where follower_id and followed_id match.
// It returns the number of rows deleted.
func DeleteFollow(ctx context.Context, db *sql.DB, followerID, followedID int) (int64, error) {
	query := `DELETE FROM followers WHERE follower_id = ? AND followed_id = ?;`
	res, err := db.ExecContext(ctx, query, followerID, followedID)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

// AcceptFollow sets a pending follow to accepted. It returns number of rows updated.
func AcceptFollow(ctx context.Context, db *sql.DB, followerID, followedID int) (int64, error) {
	query := `UPDATE followers SET status = 'accepted' WHERE follower_id = ? AND followed_id = ? AND status = 'pending';`
	res, err := db.ExecContext(ctx, query, followerID, followedID)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

// BlockFollow sets the relationship to 'blocked'. It will update an existing row
// or create one if none exists. Returns rows affected (update counts as 1, insert may be 1).
func BlockFollow(ctx context.Context, db *sql.DB, blockerID, targetID int) (int64, error) {
	// Use SQLite upsert to set status = 'blocked'
	query := `INSERT INTO followers (follower_id, followed_id, status) VALUES (?, ?, 'blocked')
		ON CONFLICT(follower_id, followed_id) DO UPDATE SET status = excluded.status;`
	res, err := db.ExecContext(ctx, query, blockerID, targetID)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

// UnblockFollow deletes a blocked relationship (only if status = 'blocked').
func UnblockFollow(ctx context.Context, db *sql.DB, blockerID, targetID int) (int64, error) {
	query := `DELETE FROM followers WHERE follower_id = ? AND followed_id = ? AND status = 'blocked';`
	res, err := db.ExecContext(ctx, query, blockerID, targetID)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
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
	// Fetch a larger candidate pool and filter out users who have blocked the current user
	// so we can replace them until we reach the requested limit.
	fetchSize := min(limit*5, 100)

	query := `
        SELECT id, first_name, last_name, COALESCE(profile_picture, '')
        FROM users
        WHERE id != ? 
        AND id NOT IN (
            SELECT followed_id FROM followers WHERE follower_id = ?
            UNION
            SELECT follower_id FROM followers WHERE followed_id = ?
        )
        ORDER BY RANDOM()
        LIMIT ?;
    `

	log.Printf("DiscoverUsers: executing discover query for user %d fetchSize %d (will filter blocked)", currentUserID, fetchSize)
	rows, err := db.QueryContext(ctx, query, currentUserID, currentUserID, currentUserID, fetchSize)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []models.DiscoveredUser
	for rows.Next() {
		var u models.DiscoveredUser
		if err := rows.Scan(&u.ID, &u.FirstName, &u.LastName, &u.ProfilePicture, &u.Status); err != nil {
			return nil, err
		}
		users = append(users, u)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	log.Printf("DiscoverUsers: returning %d users for user %d (requested %d)", len(users), currentUserID, limit)
	return users, nil
}

// GetFollowingUsers returns users that the current user is following with status 'accepted'
func GetFollowingUsers(ctx context.Context, db *sql.DB, currentUserID int) ([]models.FollowListUser, error) {
	query := `
		SELECT u.id, u.first_name, u.last_name, COALESCE(u.profile_picture, '')
		FROM users u
		JOIN followers f ON u.id = f.followed_id
		WHERE f.follower_id = ? AND f.status = 'accepted'
		ORDER BY u.first_name, u.last_name
	`
	rows, err := db.QueryContext(ctx, query, currentUserID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var res []models.FollowListUser
	for rows.Next() {
		var u models.FollowListUser
		if err := rows.Scan(&u.ID, &u.FirstName, &u.LastName, &u.ProfilePicture); err != nil {
			return nil, err
		}
		res = append(res, u)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return res, nil
}

// GetFollowers returns users that follow the current user with status 'accepted'
func GetFollowers(ctx context.Context, db *sql.DB, currentUserID int) ([]models.FollowListUser, error) {
	query := `
		SELECT u.id, u.first_name, u.last_name, COALESCE(u.profile_picture, '')
		FROM users u
		JOIN followers f ON u.id = f.follower_id
		WHERE f.followed_id = ? AND f.status = 'accepted'
		ORDER BY u.first_name, u.last_name
	`
	rows, err := db.QueryContext(ctx, query, currentUserID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var res []models.FollowListUser
	for rows.Next() {
		var u models.FollowListUser
		if err := rows.Scan(&u.ID, &u.FirstName, &u.LastName, &u.ProfilePicture); err != nil {
			return nil, err
		}
		res = append(res, u)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return res, nil
}

// GetBlockedUsers returns users that the current user has blocked (follower_id = currentUserID, status = 'blocked')
func GetBlockedUsers(ctx context.Context, db *sql.DB, currentUserID int) ([]models.FollowListUser, error) {
	query := `
		SELECT u.id, u.first_name, u.last_name, COALESCE(u.profile_picture, '')
		FROM users u
		JOIN followers f ON u.id = f.followed_id
		WHERE f.follower_id = ? AND f.status = 'blocked'
		ORDER BY u.first_name, u.last_name
	`
	rows, err := db.QueryContext(ctx, query, currentUserID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var res []models.FollowListUser
	for rows.Next() {
		var u models.FollowListUser
		if err := rows.Scan(&u.ID, &u.FirstName, &u.LastName, &u.ProfilePicture); err != nil {
			return nil, err
		}
		res = append(res, u)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return res, nil
}

// GetPendingIncomingRequests returns users who sent a pending follow request to the current user
func GetPendingIncomingRequests(ctx context.Context, db *sql.DB, currentUserID int) ([]models.FollowListUser, error) {
	query := `
		SELECT u.id, u.first_name, u.last_name, COALESCE(u.profile_picture, '')
		FROM users u
		JOIN followers f ON u.id = f.follower_id
		WHERE f.followed_id = ? AND f.status = 'pending'
		ORDER BY u.first_name, u.last_name
	`
	rows, err := db.QueryContext(ctx, query, currentUserID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var res []models.FollowListUser
	for rows.Next() {
		var u models.FollowListUser
		if err := rows.Scan(&u.ID, &u.FirstName, &u.LastName, &u.ProfilePicture); err != nil {
			return nil, err
		}
		res = append(res, u)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return res, nil
}
