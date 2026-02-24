package queries

import (
	"context"
	"database/sql"
	"log"

	"backend/pkg/models"
)

func syncFollowCountsForUser(ctx context.Context, db *sql.DB, userID int) error {
	_, err := db.ExecContext(ctx, `
		UPDATE users
		SET Followers = (SELECT COUNT(*) FROM followers WHERE followed_id = ? AND status = 'accepted'),
			Following = (SELECT COUNT(*) FROM followers WHERE follower_id = ? AND status = 'accepted')
		WHERE id = ?
	`, userID, userID, userID)
	return err
}

func syncFollowCountsForPair(ctx context.Context, db *sql.DB, userA, userB int) {
	if err := syncFollowCountsForUser(ctx, db, userA); err != nil {
		log.Printf("[WARN] syncFollowCountsForPair: failed syncing user %d: %v", userA, err)
	}
	if err := syncFollowCountsForUser(ctx, db, userB); err != nil {
		log.Printf("[WARN] syncFollowCountsForPair: failed syncing user %d: %v", userB, err)
	}
}

// RebuildAllFollowCounts recalculates Followers/Following counters for every user
// from accepted relationships in followers table.
func RebuildAllFollowCounts(ctx context.Context, db *sql.DB) error {
	_, err := db.ExecContext(ctx, `
		UPDATE users
		SET Followers = (SELECT COUNT(*) FROM followers WHERE followed_id = users.id AND status = 'accepted'),
			Following = (SELECT COUNT(*) FROM followers WHERE follower_id = users.id AND status = 'accepted')
	`)
	if err != nil {
		log.Printf("[ERROR] RebuildAllFollowCounts failed: %v", err)
		return err
	}
	log.Printf("[SUCCESS] RebuildAllFollowCounts: counters rebuilt for all users")
	return nil
}

func CreateFollow(ctx context.Context, db *sql.DB, req models.FollowRequest, status string) error {
	log.Printf("[INFO] CreateFollow: Attempting to create follow. Follower: %d, Followed: %d, Status: %s", req.FollowerID, req.FollowedID, status)
	query := `INSERT INTO followers (follower_id, followed_id, status) VALUES (?, ?, ?)`
	_, err := db.ExecContext(ctx, query, req.FollowerID, req.FollowedID, status)
	if err != nil {
		log.Printf("[ERROR] CreateFollow failed: %v", err)
		return err
	}
	syncFollowCountsForPair(ctx, db, req.FollowerID, req.FollowedID)
	log.Printf("[SUCCESS] CreateFollow: Follow relationship created")
	return nil
}

func DeleteFollow(ctx context.Context, db *sql.DB, followerID, followedID int) (int64, error) {
	log.Printf("[INFO] DeleteFollow: Deleting relationship. Follower: %d, Followed: %d", followerID, followedID)
	query := `DELETE FROM followers WHERE follower_id = ? AND followed_id = ?;`
	res, err := db.ExecContext(ctx, query, followerID, followedID)
	if err != nil {
		log.Printf("[ERROR] DeleteFollow failed: %v", err)
		return 0, err
	}
	syncFollowCountsForPair(ctx, db, followerID, followedID)
	rows, _ := res.RowsAffected()
	log.Printf("[INFO] DeleteFollow: Rows affected: %d", rows)
	return rows, nil
}

func AcceptFollow(ctx context.Context, db *sql.DB, followerID, followedID int) (int64, error) {
	log.Printf("[INFO] AcceptFollow: Accepting request. Follower: %d, Followed: %d", followerID, followedID)
	query := `UPDATE followers SET status = 'accepted' WHERE follower_id = ? AND followed_id = ? AND status = 'pending';`
	res, err := db.ExecContext(ctx, query, followerID, followedID)
	if err != nil {
		log.Printf("[ERROR] AcceptFollow failed: %v", err)
		return 0, err
	}
	syncFollowCountsForPair(ctx, db, followerID, followedID)
	rows, _ := res.RowsAffected()
	log.Printf("[INFO] AcceptFollow: Rows affected: %d", rows)
	return rows, nil
}

func BlockFollow(ctx context.Context, db *sql.DB, blockerID, targetID int) (int64, error) {
	log.Printf("[INFO] BlockFollow: Blocker: %d, Target: %d", blockerID, targetID)
	query := `INSERT INTO followers (follower_id, followed_id, status) VALUES (?, ?, 'blocked')
        ON CONFLICT(follower_id, followed_id) DO UPDATE SET status = excluded.status;`
	res, err := db.ExecContext(ctx, query, blockerID, targetID)
	if err != nil {
		log.Printf("[ERROR] BlockFollow failed: %v", err)
		return 0, err
	}
	syncFollowCountsForPair(ctx, db, blockerID, targetID)
	rows, _ := res.RowsAffected()
	log.Printf("[INFO] BlockFollow: Rows affected: %d", rows)
	return rows, nil
}

func UnblockFollow(ctx context.Context, db *sql.DB, blockerID, targetID int) (int64, error) {
	log.Printf("[INFO] UnblockFollow: Blocker: %d, Target: %d", blockerID, targetID)
	query := `DELETE FROM followers WHERE follower_id = ? AND followed_id = ? AND status = 'blocked';`
	res, err := db.ExecContext(ctx, query, blockerID, targetID)
	if err != nil {
		log.Printf("[ERROR] UnblockFollow failed: %v", err)
		return 0, err
	}
	syncFollowCountsForPair(ctx, db, blockerID, targetID)
	rows, _ := res.RowsAffected()
	log.Printf("[INFO] UnblockFollow: Rows affected: %d", rows)
	return rows, nil
}

func GetRelationshipStatus(ctx context.Context, db *sql.DB, currentUserID, targetUserID int) (string, error) {
	log.Printf("[INFO] GetRelationshipStatus: %d -> %d", currentUserID, targetUserID)
	var status string
	query := `SELECT status FROM followers WHERE follower_id = ? AND followed_id = ?`
	err := db.QueryRowContext(ctx, query, currentUserID, targetUserID).Scan(&status)

	if err == nil {
		if status == "accepted" {
			return "Following", nil
		} else if status == "pending" {
			return "Pending", nil
		}
	} else if err != sql.ErrNoRows {
		log.Printf("[ERROR] GetRelationshipStatus (Outgoing) failed: %v", err)
		return "", err
	}

	query = `SELECT status FROM followers WHERE follower_id = ? AND followed_id = ? AND status = 'accepted'`
	err = db.QueryRowContext(ctx, query, targetUserID, currentUserID).Scan(&status)

	if err == nil {
		return "Follow Back", nil
	} else if err != sql.ErrNoRows {
		log.Printf("[ERROR] GetRelationshipStatus (Incoming) failed: %v", err)
		return "", err
	}

	return "Follow", nil
}

func DiscoverUsers(ctx context.Context, db *sql.DB, currentUserID int, limit int) ([]models.DiscoveredUser, error) {
	if limit <= 0 {
		limit = 5
	}
	log.Printf("[INFO] DiscoverUsers: Fetching for user %d, limit %d", currentUserID, limit)

	query := `
    SELECT 
        u.id, u.first_name, u.last_name, COALESCE(u.profile_picture, ''),
        CASE 
            WHEN f.status = 'pending' THEN 'Pending'
            WHEN f.status = 'accepted' THEN 'Following'
            WHEN f_back.status = 'accepted' THEN 'Follow Back'
            ELSE 'Follow'
        END as relationship_status
    FROM users u
    LEFT JOIN followers f ON f.follower_id = ? AND f.followed_id = u.id
    LEFT JOIN followers f_back ON f_back.follower_id = u.id AND f_back.followed_id = ?
    WHERE u.id != ? 
    AND u.id NOT IN (
        SELECT followed_id FROM followers WHERE follower_id = ? AND status IN ('accepted', 'pending', 'blocked')
        UNION
        SELECT follower_id FROM followers WHERE followed_id = ? AND status = 'blocked'
    )
    ORDER BY RANDOM()
    LIMIT ?;
`

	rows, err := db.QueryContext(ctx, query, currentUserID, currentUserID, currentUserID, currentUserID, currentUserID, limit)
	if err != nil {
		log.Printf("[ERROR] DiscoverUsers query failed: %v", err)
		return nil, err
	}
	defer rows.Close()

	var users []models.DiscoveredUser
	for rows.Next() {
		var u models.DiscoveredUser
		if err := rows.Scan(&u.ID, &u.FirstName, &u.LastName, &u.ProfilePicture, &u.Status); err != nil {
			log.Printf("[ERROR] DiscoverUsers scan failed: %v", err)
			return nil, err
		}
		users = append(users, u)
	}

	log.Printf("[SUCCESS] DiscoverUsers: Found %d users", len(users))
	return users, nil
}

func GetFollowingUsers(ctx context.Context, db *sql.DB, currentUserID int) ([]models.FollowListUser, error) {
	log.Printf("[INFO] GetFollowingUsers: Fetching following for %d", currentUserID)
	query := `
        SELECT u.id, u.first_name, u.last_name, COALESCE(u.profile_picture, '')
        FROM users u
        JOIN followers f ON u.id = f.followed_id
        WHERE f.follower_id = ? AND f.status = 'accepted'
		AND NOT EXISTS (
			SELECT 1 FROM followers fb
			WHERE fb.status = 'blocked'
			  AND ((fb.follower_id = ? AND fb.followed_id = u.id)
				   OR (fb.follower_id = u.id AND fb.followed_id = ?))
		)
        ORDER BY u.first_name, u.last_name
    `
	rows, err := db.QueryContext(ctx, query, currentUserID, currentUserID, currentUserID)
	if err != nil {
		log.Printf("[ERROR] GetFollowingUsers query failed: %v", err)
		return nil, err
	}
	defer rows.Close()

	var res []models.FollowListUser
	for rows.Next() {
		var u models.FollowListUser
		if err := rows.Scan(&u.ID, &u.FirstName, &u.LastName, &u.ProfilePicture); err != nil {
			log.Printf("[ERROR] GetFollowingUsers scan failed: %v", err)
			return nil, err
		}
		res = append(res, u)
	}
	return res, nil
}

func GetFollowers(ctx context.Context, db *sql.DB, currentUserID int) ([]models.FollowListUser, error) {
	log.Printf("[INFO] GetFollowers: Fetching followers for %d", currentUserID)
	query := `
        SELECT u.id, u.first_name, u.last_name, COALESCE(u.profile_picture, '')
        FROM users u
        JOIN followers f ON u.id = f.follower_id
        WHERE f.followed_id = ? AND f.status = 'accepted'
		AND NOT EXISTS (
			SELECT 1 FROM followers fb
			WHERE fb.status = 'blocked'
			  AND ((fb.follower_id = ? AND fb.followed_id = u.id)
				   OR (fb.follower_id = u.id AND fb.followed_id = ?))
		)
        ORDER BY u.first_name, u.last_name
    `
	rows, err := db.QueryContext(ctx, query, currentUserID, currentUserID, currentUserID)
	if err != nil {
		log.Printf("[ERROR] GetFollowers query failed: %v", err)
		return nil, err
	}
	defer rows.Close()

	var res []models.FollowListUser
	for rows.Next() {
		var u models.FollowListUser
		if err := rows.Scan(&u.ID, &u.FirstName, &u.LastName, &u.ProfilePicture); err != nil {
			log.Printf("[ERROR] GetFollowers scan failed: %v", err)
			return nil, err
		}
		res = append(res, u)
	}
	return res, nil
}

// GetFollowingUsersForViewer returns users that targetUserID follows (accepted),
// and includes relationship status from viewerID to each listed user.
func GetFollowingUsersForViewer(ctx context.Context, db *sql.DB, targetUserID, viewerID int) ([]models.FollowListUser, error) {
	log.Printf("[INFO] GetFollowingUsersForViewer: target=%d viewer=%d", targetUserID, viewerID)
	query := `
		SELECT
			u.id,
			u.first_name,
			u.last_name,
			COALESCE(u.profile_picture, ''),
			CASE
				WHEN fv.status = 'blocked' THEN 'Blocked'
				WHEN vf.status = 'blocked' THEN 'You_Are_Blocked'
				WHEN fv.status = 'accepted' THEN 'Following'
				WHEN fv.status = 'pending' THEN 'Pending'
				WHEN vf.status = 'accepted' THEN 'Follow Back'
				ELSE 'Follow'
			END AS current_status
		FROM users u
		JOIN followers t ON u.id = t.followed_id AND t.follower_id = ? AND t.status = 'accepted'
		LEFT JOIN followers fv ON fv.follower_id = ? AND fv.followed_id = u.id
		LEFT JOIN followers vf ON vf.follower_id = u.id AND vf.followed_id = ?
		WHERE NOT EXISTS (
			SELECT 1 FROM followers fb
			WHERE fb.status = 'blocked'
			  AND ((fb.follower_id = ? AND fb.followed_id = u.id)
			       OR (fb.follower_id = u.id AND fb.followed_id = ?))
		)
		ORDER BY u.first_name, u.last_name
	`
	rows, err := db.QueryContext(ctx, query, targetUserID, viewerID, viewerID, viewerID, viewerID)
	if err != nil {
		log.Printf("[ERROR] GetFollowingUsersForViewer query failed: %v", err)
		return nil, err
	}
	defer rows.Close()

	var res []models.FollowListUser
	for rows.Next() {
		var u models.FollowListUser
		if err := rows.Scan(&u.ID, &u.FirstName, &u.LastName, &u.ProfilePicture, &u.Status); err != nil {
			log.Printf("[ERROR] GetFollowingUsersForViewer scan failed: %v", err)
			return nil, err
		}
		res = append(res, u)
	}
	return res, nil
}

// GetFollowersForViewer returns users that follow targetUserID (accepted),
// and includes relationship status from viewerID to each listed user.
func GetFollowersForViewer(ctx context.Context, db *sql.DB, targetUserID, viewerID int) ([]models.FollowListUser, error) {
	log.Printf("[INFO] GetFollowersForViewer: target=%d viewer=%d", targetUserID, viewerID)
	query := `
		SELECT
			u.id,
			u.first_name,
			u.last_name,
			COALESCE(u.profile_picture, ''),
			CASE
				WHEN fv.status = 'blocked' THEN 'Blocked'
				WHEN vf.status = 'blocked' THEN 'You_Are_Blocked'
				WHEN fv.status = 'accepted' THEN 'Following'
				WHEN fv.status = 'pending' THEN 'Pending'
				WHEN vf.status = 'accepted' THEN 'Follow Back'
				ELSE 'Follow'
			END AS current_status
		FROM users u
		JOIN followers t ON u.id = t.follower_id AND t.followed_id = ? AND t.status = 'accepted'
		LEFT JOIN followers fv ON fv.follower_id = ? AND fv.followed_id = u.id
		LEFT JOIN followers vf ON vf.follower_id = u.id AND vf.followed_id = ?
		WHERE NOT EXISTS (
			SELECT 1 FROM followers fb
			WHERE fb.status = 'blocked'
			  AND ((fb.follower_id = ? AND fb.followed_id = u.id)
			       OR (fb.follower_id = u.id AND fb.followed_id = ?))
		)
		ORDER BY u.first_name, u.last_name
	`
	rows, err := db.QueryContext(ctx, query, targetUserID, viewerID, viewerID, viewerID, viewerID)
	if err != nil {
		log.Printf("[ERROR] GetFollowersForViewer query failed: %v", err)
		return nil, err
	}
	defer rows.Close()

	var res []models.FollowListUser
	for rows.Next() {
		var u models.FollowListUser
		if err := rows.Scan(&u.ID, &u.FirstName, &u.LastName, &u.ProfilePicture, &u.Status); err != nil {
			log.Printf("[ERROR] GetFollowersForViewer scan failed: %v", err)
			return nil, err
		}
		res = append(res, u)
	}
	return res, nil
}

func GetBlockedUsers(ctx context.Context, db *sql.DB, currentUserID int) ([]models.FollowListUser, error) {
	log.Printf("[INFO] GetBlockedUsers: Fetching blocked for %d", currentUserID)
	query := `
        SELECT u.id, u.first_name, u.last_name, COALESCE(u.profile_picture, '')
        FROM users u
        JOIN followers f ON u.id = f.followed_id
        WHERE f.follower_id = ? AND f.status = 'blocked'
        ORDER BY u.first_name, u.last_name
    `
	rows, err := db.QueryContext(ctx, query, currentUserID)
	if err != nil {
		log.Printf("[ERROR] GetBlockedUsers query failed: %v", err)
		return nil, err
	}
	defer rows.Close()

	var res []models.FollowListUser
	for rows.Next() {
		var u models.FollowListUser
		if err := rows.Scan(&u.ID, &u.FirstName, &u.LastName, &u.ProfilePicture); err != nil {
			log.Printf("[ERROR] GetBlockedUsers scan failed: %v", err)
			return nil, err
		}
		res = append(res, u)
	}
	return res, nil
}

func GetPendingIncomingRequests(ctx context.Context, db *sql.DB, currentUserID int) ([]models.FollowListUser, error) {
	log.Printf("[INFO] GetPendingIncomingRequests: Fetching pending for %d", currentUserID)
	query := `
        SELECT u.id, u.first_name, u.last_name, COALESCE(u.profile_picture, '')
        FROM users u
        JOIN followers f ON u.id = f.follower_id
        WHERE f.followed_id = ? AND f.status = 'pending'
        ORDER BY u.first_name, u.last_name
    `
	rows, err := db.QueryContext(ctx, query, currentUserID)
	if err != nil {
		log.Printf("[ERROR] GetPendingIncomingRequests query failed: %v", err)
		return nil, err
	}
	defer rows.Close()

	var res []models.FollowListUser
	for rows.Next() {
		var u models.FollowListUser
		if err := rows.Scan(&u.ID, &u.FirstName, &u.LastName, &u.ProfilePicture); err != nil {
			log.Printf("[ERROR] GetPendingIncomingRequests scan failed: %v", err)
			return nil, err
		}
		res = append(res, u)
	}
	return res, nil
}
