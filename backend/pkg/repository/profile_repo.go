package repository

import (
	"backend/pkg/models"
	"context"
	"database/sql"
	"fmt"
	"log"
)

type ProfileRepository interface {
	GetRawUserProfile(ctx context.Context, viewerID, targetID int) (*models.RawProfileData, error)
	GetRelation(ctx context.Context, fromID, toID int) (string, error)
	UpdateUserMedia(ctx context.Context, targetID int, imageURL, coverImageURL string) error
	UserPrivacy(ctx context.Context, userID int) (bool, error)
	DiscoverUsers(ctx context.Context, currentUserID int, limit int) ([]*models.DiscoveredUser, error)
	GetVisibilityRaw(ctx context.Context, userID int) (*models.RawVisibilityData, error)
	UpdateUserVisibilitySettings(ctx context.Context, userID int,
		emailVis, birthdayVis, relVis, employedVis, phoneVis, aboutVis, nickVis, followVis, profileType *int) error
	GetContentRaw(ctx context.Context, userID int) (*models.RawProfileData, error)
	UpdateProfileContent(ctx context.Context, userID int, req models.UserProfileRequest, birthdayStr *string) error
}

type sqliteProfileRepo struct {
	db *sql.DB
}

func NewProfileRepository(db *sql.DB) ProfileRepository {
	return &sqliteProfileRepo{db: db}
}

func (r *sqliteProfileRepo) GetRawUserProfile(ctx context.Context, viewerID, targetID int) (*models.RawProfileData, error) {
	rawProfileData := new(models.RawProfileData)
	query := `
		SELECT
			u.id,
			COALESCE(lu.email, ''),
			COALESCE(u.first_name, ''),
			COALESCE(u.last_name, ''),
			COALESCE(u.profile_picture, ''),
			COALESCE(u.cover_image, ''),
			COALESCE(date(u.birthday_date), ''),
			COALESCE(u.relationship_status, ''),
			COALESCE(u.employed_at, ''),
			COALESCE(u.location, ''),
			COALESCE(u.phone_number, ''),
			COALESCE(u.nickname, ''),
			COALESCE(u.about_me, ''),
			COALESCE(u.Followers, 0),
			COALESCE(u.Following, 0),
			u.profile_type,
			COALESCE(us.email_vis, 0),
			COALESCE(us.birthday_date_vis, 1),
			COALESCE(us.relationship_status_vis, 1),
			COALESCE(us.employed_at_vis, 1),
			COALESCE(us.phone_number_vis, 0),
			COALESCE(us.nickname_vis, 1),
			COALESCE(us.about_me_vis, 1),
			COALESCE(us.follow_vis, 0)
		FROM users u
		LEFT JOIN login_users lu ON lu.id = u.id
		LEFT JOIN user_settings us ON us.id = u.id
		WHERE u.id = ?
	`

	if err := r.db.QueryRowContext(ctx, query, targetID).Scan(
		&rawProfileData.ID,
		&rawProfileData.Email,
		&rawProfileData.FirstName,
		&rawProfileData.LastName,
		&rawProfileData.ProfilePicture,
		&rawProfileData.CoverImage,
		&rawProfileData.BirthdayDate,
		&rawProfileData.RelationshipStatus,
		&rawProfileData.EmployedAt,
		&rawProfileData.Location,
		&rawProfileData.PhoneNumber,
		&rawProfileData.Nickname,
		&rawProfileData.AboutMe,
		&rawProfileData.FollowersCount,
		&rawProfileData.FollowingCount,
		&rawProfileData.ProfileType,
		&rawProfileData.EmailVis,
		&rawProfileData.BirthdayVis,
		&rawProfileData.RelationshipVis,
		&rawProfileData.EmployedVis,
		&rawProfileData.PhoneVis,
		&rawProfileData.NicknameVis,
		&rawProfileData.AboutVis,
		&rawProfileData.FollowVis,
	); err != nil {
		return nil, err
	}

	return rawProfileData, nil
}

func (r *sqliteProfileRepo) GetRelation(ctx context.Context, fromID, toID int) (string, error) {
	var status string
	query := `SELECT status FROM followers WHERE follower_id = ? AND followed_id = ?`
	err := r.db.QueryRowContext(ctx, query, fromID, toID).Scan(&status)
	if err == sql.ErrNoRows {
		return "none", nil
	}
	return status, err
}

func (r *sqliteProfileRepo) UpdateUserMedia(ctx context.Context, targetID int, imageURL, coverImageURL string) error {

	query := "UPDATE users SET "
	args := make([]interface{}, 0, 3)
	if imageURL != "" {
		query += "profile_picture = ?"
		args = append(args, imageURL)
	}
	if coverImageURL != "" {
		if len(args) > 0 {
			query += ", "
		}
		query += "cover_image = ?"
		args = append(args, coverImageURL)
	}
	query += " WHERE id = ?"
	args = append(args, targetID)

	res, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return err
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return fmt.Errorf("user not found")
	}
	return nil
}

func (r *sqliteProfileRepo) UserPrivacy(ctx context.Context, userID int) (bool, error) {
	var isPublic bool
	query := `
        SELECT profile_type FROM users 
        WHERE id = ?`
	err := r.db.QueryRowContext(ctx, query, userID).Scan(&isPublic)
	if err != nil {
		return false, err
	}
	return isPublic, nil
}

func (r *sqliteProfileRepo) DiscoverUsers(ctx context.Context, currentUserID int, limit int) ([]*models.DiscoveredUser, error) {
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

	log.Printf("[SUCCESS] DiscoverUsers: Found %d users", len(users))
	return users, nil
}

func (r *sqliteProfileRepo) GetVisibilityRaw(ctx context.Context, userID int) (*models.RawVisibilityData, error) {
	query := `
        SELECT u.profile_type,
               COALESCE(us.email_vis, 0),
               COALESCE(us.birthday_date_vis, 1),
               COALESCE(us.relationship_status_vis, 1),
               COALESCE(us.employed_at_vis, 1),
               COALESCE(us.phone_number_vis, 0),
               COALESCE(us.about_me_vis, 1),
               COALESCE(us.nickname_vis, 1),
               COALESCE(us.follow_vis, 0)
        FROM users u
        LEFT JOIN user_settings us ON u.id = us.id
        WHERE u.id = ?`

	var v models.RawVisibilityData
	err := r.db.QueryRowContext(ctx, query, userID).Scan(
		&v.ProfileType, &v.EmailVis, &v.BirthdayVis, &v.RelVis,
		&v.EmployedVis, &v.PhoneVis, &v.AboutVis, &v.NickVis, &v.FollowVis,
	)
	return &v, err
}

func (r *sqliteProfileRepo) UpdateUserVisibilitySettings(ctx context.Context, userID int,
	emailVis, birthdayVis, relVis, employedVis, phoneVis, aboutVis, nickVis, followVis, profileType *int) error {

	log.Printf("[INFO] UpdateUserVisibilitySettings: user=%d", userID)

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	querySettings := `
        INSERT INTO user_settings (
            id, email_vis, birthday_date_vis, relationship_status_vis, 
            employed_at_vis, phone_number_vis, about_me_vis, nickname_vis, follow_vis
        ) VALUES (?, 
            COALESCE(?, 0), COALESCE(?, 1), COALESCE(?, 1), 
            COALESCE(?, 1), COALESCE(?, 0), COALESCE(?, 1), 
            COALESCE(?, 1), COALESCE(?, 0)
        )
        ON CONFLICT(id) DO UPDATE SET 
            email_vis = COALESCE(?, email_vis),
            birthday_date_vis = COALESCE(?, birthday_date_vis),
            relationship_status_vis = COALESCE(?, relationship_status_vis),
            employed_at_vis = COALESCE(?, employed_at_vis),
            phone_number_vis = COALESCE(?, phone_number_vis),
            about_me_vis = COALESCE(?, about_me_vis),
            nickname_vis = COALESCE(?, nickname_vis),
            follow_vis = COALESCE(?, follow_vis);`

	_, err = tx.ExecContext(ctx, querySettings,
		userID,
		emailVis, birthdayVis, relVis, employedVis, phoneVis, aboutVis, nickVis, followVis, // για το VALUES
		emailVis, birthdayVis, relVis, employedVis, phoneVis, aboutVis, nickVis, followVis, // για το UPDATE
	)
	if err != nil {
		return err
	}

	if profileType != nil {
		_, err = tx.ExecContext(ctx, `UPDATE users SET profile_type = ? WHERE id = ?;`, *profileType, userID)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (r *sqliteProfileRepo) GetContentRaw(ctx context.Context, userID int) (*models.RawProfileData, error) {
	query := `
        SELECT 
            id, first_name, last_name, 
            COALESCE(date(birthday_date), ''), 
            COALESCE(relationship_status, ''), 
            COALESCE(employed_at, ''), 
            COALESCE(location, ''), 
            COALESCE(phone_number, ''), 
            COALESCE(nickname, ''), 
            COALESCE(about_me, '')
        FROM users WHERE id = ?`

	var p models.RawProfileData
	err := r.db.QueryRowContext(ctx, query, userID).Scan(
		&p.ID, &p.FirstName, &p.LastName, &p.BirthdayDate,
		&p.RelationshipStatus, &p.EmployedAt, &p.Location,
		&p.PhoneNumber, &p.Nickname, &p.AboutMe,
	)
	return &p, err
}

func (r *sqliteProfileRepo) UpdateProfileContent(ctx context.Context, userID int, req models.UserProfileRequest, birthdayStr *string) error {
	query := `
        UPDATE users SET
            first_name = CASE WHEN ? = '' THEN first_name ELSE ? END,
            last_name  = CASE WHEN ? = '' THEN last_name ELSE ? END,
            birthday_date       = COALESCE(?, birthday_date),
            relationship_status = COALESCE(?, relationship_status),
            employed_at         = COALESCE(?, employed_at),
            phone_number        = COALESCE(?, phone_number),
            profile_picture     = COALESCE(?, profile_picture),
            level               = CASE WHEN ? = '' THEN level ELSE ? END
        WHERE id = ?`

	_, err := r.db.ExecContext(ctx, query,
		req.FirstName, req.FirstName,
		req.LastName, req.LastName,
		birthdayStr,
		req.RelationshipStatus,
		req.EmployedAt,
		req.PhoneNumber,
		req.ProfilePicture,
		req.Level, req.Level,
		userID,
	)
	return err
}
