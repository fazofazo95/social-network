package repository

import (
	"backend/pkg/models"
	"context"
	"database/sql"
	"fmt"
	"log"
)

type FollowRepository interface {
	CreateFollow(ctx context.Context, req models.FollowRequest, status string) error
	DeleteFollow(ctx context.Context, followerID, followedID int) (int64, error)
	RemoveFollower(ctx context.Context, currentUserID, targetFollowerID int) (int64, error)
	AcceptFollow(ctx context.Context, followerID, followedID int) (int64, error)
	RejectFollow(ctx context.Context, followerID, followedID int) (int64, error)
	BlockFollow(ctx context.Context, blockerID, targetID int) (int64, error)
	UnblockFollow(ctx context.Context, blockerID, targetID int) (int64, error)

	// Queries
	DiscoverUsers(ctx context.Context, currentUserID int, limit int) ([]*models.DiscoveredUser, error)
	GetFollowingUsers(ctx context.Context, currentUserID int) ([]*models.FollowListUser, error)
	GetFollowers(ctx context.Context, currentUserID int) ([]*models.FollowListUser, error)
	GetFollowingUsersForViewer(ctx context.Context, targetUserID, viewerID int) ([]*models.FollowListUser, error)
	GetFollowersForViewer(ctx context.Context, targetUserID, viewerID int) ([]*models.FollowListUser, error)
	GetBlockedUsers(ctx context.Context, currentUserID int) ([]*models.FollowListUser, error)
	GetPendingIncomingRequests(ctx context.Context, currentUserID int) ([]*models.FollowListUser, error)

	// Maintenance
	RebuildAllFollowCounts(ctx context.Context) error
}

type sqliteFollowRepo struct {
	db *sql.DB
}

func NewFollowRepository(db *sql.DB) FollowRepository {
	return &sqliteFollowRepo{db: db}
}

// --- Helpers (Private Methods) ---

func (r *sqliteFollowRepo) syncFollowCountsForUser(ctx context.Context, tx *sql.Tx, userID int) error {
	_, err := tx.ExecContext(ctx, `
		UPDATE users
		SET Followers = (SELECT COUNT(*) FROM followers WHERE followed_id = ? AND status = 'accepted'),
			Following = (SELECT COUNT(*) FROM followers WHERE follower_id = ? AND status = 'accepted')
		WHERE id = ?
	`, userID, userID, userID)
	return err
}

func (r *sqliteFollowRepo) syncFollowCountsForPair(ctx context.Context, tx *sql.Tx, userA, userB int) error {
	if err := r.syncFollowCountsForUser(ctx, tx, userA); err != nil {
		log.Printf("[ERROR] syncFollowCountsForPair: failed syncing user %d: %v", userA, err)
		return err
	}
	if err := r.syncFollowCountsForUser(ctx, tx, userB); err != nil {
		log.Printf("[ERROR] syncFollowCountsForPair: failed syncing user %d: %v", userB, err)
		return err
	}
	return nil
}

// --- Implementation ---

func (r *sqliteFollowRepo) RebuildAllFollowCounts(ctx context.Context) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE users
		SET Followers = (SELECT COUNT(*) FROM followers WHERE followed_id = users.id AND status = 'accepted'),
			Following = (SELECT COUNT(*) FROM followers WHERE follower_id = users.id AND status = 'accepted')
	`)
	if err != nil {
		log.Printf("[ERROR] RebuildAllFollowCounts failed: %v", err)
		return err
	}
	return nil
}

func (r *sqliteFollowRepo) CreateFollow(ctx context.Context, req models.FollowRequest, status string) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	query := `INSERT INTO followers (follower_id, followed_id, status) VALUES (?, ?, ?)`
	_, err = tx.ExecContext(ctx, query, req.FollowerID, req.FollowedID, status)
	if err != nil {
		log.Printf("[ERROR] CreateFollow failed: %v", err)
		return err
	}

	err = r.syncFollowCountsForPair(ctx, tx, req.FollowerID, req.FollowedID)
	if err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}

func (r *sqliteFollowRepo) DeleteFollow(ctx context.Context, followerID, followedID int) (int64, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	query := `DELETE FROM followers WHERE follower_id = ? AND followed_id = ?;`
	res, err := tx.ExecContext(ctx, query, followerID, followedID)
	if err != nil {
		log.Printf("[ERROR] DeleteFollow failed: %v", err)
		return 0, err
	}
	err = r.syncFollowCountsForPair(ctx, tx, followerID, followedID)
	if err != nil {
		return 0, err
	}
	rows, _ := res.RowsAffected()

	if err := tx.Commit(); err != nil {
		return 0, err
	}

	return rows, nil
}

func (r *sqliteFollowRepo) RemoveFollower(ctx context.Context, currentUserID, targetFollowerID int) (int64, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	query := `
		DELETE FROM followers
		WHERE follower_id = ? AND followed_id = ? AND status = 'accepted';
	`
	res, err := tx.ExecContext(ctx, query, targetFollowerID, currentUserID)
	if err != nil {
		log.Printf("[ERROR] RemoveFollower failed: %v", err)
		return 0, err
	}
	err = r.syncFollowCountsForPair(ctx, tx, targetFollowerID, currentUserID)
	if err != nil {
		return 0, err
	}

	rows, _ := res.RowsAffected()

	if err := tx.Commit(); err != nil {
		return 0, err
	}

	return rows, nil
}

func (r *sqliteFollowRepo) AcceptFollow(ctx context.Context, followerID, followedID int) (int64, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	query := `UPDATE followers SET status = 'accepted' WHERE follower_id = ? AND followed_id = ? AND status = 'pending';`
	res, err := tx.ExecContext(ctx, query, followerID, followedID)
	if err != nil {
		log.Printf("[ERROR] AcceptFollow failed: %v", err)
		return 0, err
	}

	err = r.syncFollowCountsForPair(ctx, tx, followerID, followedID)
	if err != nil {
		return 0, err
	}

	rows, _ := res.RowsAffected()

	if err := tx.Commit(); err != nil {
		return 0, err
	}

	return rows, nil
}

func (r *sqliteFollowRepo) RejectFollow(ctx context.Context, followerID, followedID int) (int64, error) {
	query := `DELETE FROM followers WHERE follower_id = ? AND followed_id = ? AND status = 'pending';`
	res, err := r.db.ExecContext(ctx, query, followerID, followedID)
	if err != nil {
		log.Printf("[ERROR] RejectFollow failed: %v", err)
		return 0, err
	}
	rows, _ := res.RowsAffected()
	return rows, nil
}

func (r *sqliteFollowRepo) BlockFollow(ctx context.Context, blockerID, targetID int) (int64, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	query := `INSERT INTO followers (follower_id, followed_id, status) VALUES (?, ?, 'blocked')
        ON CONFLICT(follower_id, followed_id) DO UPDATE SET status = excluded.status;`
	res, err := tx.ExecContext(ctx, query, blockerID, targetID)
	if err != nil {
		log.Printf("[ERROR] BlockFollow failed: %v", err)
		return 0, err
	}
	err = r.syncFollowCountsForPair(ctx, tx, blockerID, targetID)
	if err != nil {
		return 0, err
	}

	if err := tx.Commit(); err != nil {
		return 0, err
	}

	rows, _ := res.RowsAffected()
	return rows, nil
}

func (r *sqliteFollowRepo) UnblockFollow(ctx context.Context, blockerID, targetID int) (int64, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	query := `DELETE FROM followers WHERE follower_id = ? AND followed_id = ? AND status = 'blocked';`
	res, err := tx.ExecContext(ctx, query, blockerID, targetID)
	if err != nil {
		log.Printf("[ERROR] UnblockFollow failed: %v", err)
		return 0, err
	}
	err = r.syncFollowCountsForPair(ctx, tx, blockerID, targetID)
	if err != nil {
		return 0, err
	}
	rows, _ := res.RowsAffected()

	if err := tx.Commit(); err != nil {
		return 0, err
	}

	return rows, nil
}

// --- Queries (Returning Pointers) ---

func (r *sqliteFollowRepo) DiscoverUsers(ctx context.Context, currentUserID int, limit int) ([]*models.DiscoveredUser, error) {
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

	rows, err := r.db.QueryContext(ctx, query, currentUserID, currentUserID, currentUserID, currentUserID, currentUserID, limit)
	if err != nil {
		log.Printf("[ERROR] DiscoverUsers query failed: %v", err)
		return nil, err
	}
	defer rows.Close()

	users := make([]*models.DiscoveredUser, 0)
	for rows.Next() {
		u := new(models.DiscoveredUser)
		if err := rows.Scan(&u.ID, &u.FirstName, &u.LastName, &u.ProfilePicture, &u.Status); err != nil {
			log.Printf("[ERROR] DiscoverUsers scan failed: %v", err)
			return nil, err
		}
		users = append(users, u)
	}

	return users, nil
}

func (r *sqliteFollowRepo) GetFollowingUsers(ctx context.Context, currentUserID int) ([]*models.FollowListUser, error) {
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
	rows, err := r.db.QueryContext(ctx, query, currentUserID, currentUserID, currentUserID)
	if err != nil {
		log.Printf("[ERROR] GetFollowingUsers query failed: %v", err)
		return nil, err
	}
	defer rows.Close()

	res := make([]*models.FollowListUser, 0)
	for rows.Next() {
		u := new(models.FollowListUser)
		if err := rows.Scan(&u.ID, &u.FirstName, &u.LastName, &u.ProfilePicture); err != nil {
			log.Printf("[ERROR] GetFollowingUsers scan failed: %v", err)
			return nil, err
		}
		res = append(res, u)
	}
	return res, nil
}

func (r *sqliteFollowRepo) GetFollowers(ctx context.Context, currentUserID int) ([]*models.FollowListUser, error) {
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
	rows, err := r.db.QueryContext(ctx, query, currentUserID, currentUserID, currentUserID)
	if err != nil {
		log.Printf("[ERROR] GetFollowers query failed: %v", err)
		return nil, err
	}
	defer rows.Close()

	res := make([]*models.FollowListUser, 0)
	for rows.Next() {
		u := new(models.FollowListUser)
		if err := rows.Scan(&u.ID, &u.FirstName, &u.LastName, &u.ProfilePicture); err != nil {
			log.Printf("[ERROR] GetFollowers scan failed: %v", err)
			return nil, err
		}
		res = append(res, u)
	}
	return res, nil
}

func (r *sqliteFollowRepo) GetFollowingUsersForViewer(ctx context.Context, targetUserID, viewerID int) ([]*models.FollowListUser, error) {
	return r.fetchFollowListWithViewerStatus(ctx, targetUserID, viewerID, true)
}

func (r *sqliteFollowRepo) GetFollowersForViewer(ctx context.Context, targetUserID, viewerID int) ([]*models.FollowListUser, error) {
	return r.fetchFollowListWithViewerStatus(ctx, targetUserID, viewerID, false)
}

func (r *sqliteFollowRepo) fetchFollowListWithViewerStatus(ctx context.Context, targetUserID, viewerID int, isFollowingList bool) ([]*models.FollowListUser, error) {
	joinCondition := "u.id = t.follower_id AND t.followed_id = ?" // default: followers
	if isFollowingList {
		joinCondition = "u.id = t.followed_id AND t.follower_id = ?" // following
	}

	query := fmt.Sprintf(`
        SELECT u.id, u.first_name, u.last_name, COALESCE(u.profile_picture, ''),
            CASE
                WHEN fv.status = 'blocked' THEN 'Blocked'
                WHEN vf.status = 'blocked' THEN 'You_Are_Blocked'
                WHEN fv.status = 'accepted' THEN 'Following'
                WHEN fv.status = 'pending' THEN 'Pending'
                WHEN vf.status = 'accepted' THEN 'Follow Back'
                ELSE 'Follow'
            END AS current_status
        FROM users u
        JOIN followers t ON %s AND t.status = 'accepted'
        LEFT JOIN followers fv ON fv.follower_id = ? AND fv.followed_id = u.id
        LEFT JOIN followers vf ON vf.follower_id = u.id AND vf.followed_id = ?
        WHERE NOT EXISTS (
            SELECT 1 FROM followers fb
            WHERE fb.status = 'blocked'
              AND ((fb.follower_id = ? AND fb.followed_id = u.id)
                   OR (fb.follower_id = u.id AND fb.followed_id = ?))
        )
        ORDER BY u.first_name, u.last_name
    `, joinCondition)

	rows, err := r.db.QueryContext(ctx, query, targetUserID, viewerID, viewerID, viewerID, viewerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	res := make([]*models.FollowListUser, 0)
	for rows.Next() {
		u := new(models.FollowListUser)
		if err := rows.Scan(&u.ID, &u.FirstName, &u.LastName, &u.ProfilePicture, &u.Status); err != nil {
			return nil, err
		}
		res = append(res, u)
	}
	return res, nil
}

func (r *sqliteFollowRepo) GetBlockedUsers(ctx context.Context, currentUserID int) ([]*models.FollowListUser, error) {
	query := `
        SELECT u.id, u.first_name, u.last_name, COALESCE(u.profile_picture, '')
        FROM users u
        JOIN followers f ON u.id = f.followed_id
        WHERE f.follower_id = ? AND f.status = 'blocked'
        ORDER BY u.first_name, u.last_name
    `
	rows, err := r.db.QueryContext(ctx, query, currentUserID)
	if err != nil {
		log.Printf("[ERROR] GetBlockedUsers query failed: %v", err)
		return nil, err
	}
	defer rows.Close()

	res := make([]*models.FollowListUser, 0)
	for rows.Next() {
		u := new(models.FollowListUser)
		if err := rows.Scan(&u.ID, &u.FirstName, &u.LastName, &u.ProfilePicture); err != nil {
			log.Printf("[ERROR] GetBlockedUsers scan failed: %v", err)
			return nil, err
		}
		res = append(res, u)
	}
	return res, nil
}

func (r *sqliteFollowRepo) GetPendingIncomingRequests(ctx context.Context, currentUserID int) ([]*models.FollowListUser, error) {
	query := `
        SELECT u.id, u.first_name, u.last_name, COALESCE(u.profile_picture, '')
        FROM users u
        JOIN followers f ON u.id = f.follower_id
        WHERE f.followed_id = ? AND f.status = 'pending'
        ORDER BY u.first_name, u.last_name
    `
	rows, err := r.db.QueryContext(ctx, query, currentUserID)
	if err != nil {
		log.Printf("[ERROR] GetPendingIncomingRequests query failed: %v", err)
		return nil, err
	}
	defer rows.Close()

	res := make([]*models.FollowListUser, 0)
	for rows.Next() {
		u := new(models.FollowListUser)
		if err := rows.Scan(&u.ID, &u.FirstName, &u.LastName, &u.ProfilePicture); err != nil {
			log.Printf("[ERROR] GetPendingIncomingRequests scan failed: %v", err)
			return nil, err
		}
		res = append(res, u)
	}
	return res, nil
}
