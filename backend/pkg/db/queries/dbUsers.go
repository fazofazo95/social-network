package queries

import (
	"context"
	"database/sql"
	"log"
	"strings"

	"backend/pkg/models"
)

func UpdateUserProfile(ctx context.Context, db *sql.DB, in models.UserProfileInput) error {
	log.Printf("[INFO] UpdateUserProfile: Updating profile for UserID: %d", in.ID)
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		log.Printf("[ERROR] UpdateUserProfile: Failed to begin transaction: %v", err)
		return err
	}
	defer tx.Rollback()

	var birthday interface{}
	if in.Birthday != nil {
		birthday = in.Birthday
	} else {
		birthday = nil
	}

	res, err := tx.ExecContext(ctx, `
        UPDATE users SET
            first_name = ?,
            last_name = ?,
            birthday_date = ?,
            relationship_status = ?,
            employed_at = ?,
            phone_number = ?,
            profile_picture = ?,
            pictures = ?,
            level = ?
        WHERE id = ?;
    `, in.FirstName, in.LastName, birthday, in.RelationshipStatus, in.EmployedAt, in.PhoneNumber, in.ProfilePicture, in.Pictures, in.Level, in.ID)
	if err != nil {
		log.Printf("[ERROR] UpdateUserProfile: Update failed: %v", err)
		return err
	}

	rows, _ := res.RowsAffected()
	log.Printf("[INFO] UpdateUserProfile: Rows affected: %d", rows)

	err = tx.Commit()
	if err != nil {
		log.Printf("[ERROR] UpdateUserProfile: Commit failed: %v", err)
		return err
	}
	log.Printf("[SUCCESS] UpdateUserProfile: Profile updated for UserID: %d", in.ID)
	return nil
}

func GetUserByID(ctx context.Context, db *sql.DB, id int) (models.UserProfile, error) {
	log.Printf("[INFO] GetUserByID: Fetching profile for UserID: %d", id)
	var u models.UserProfile
	row := db.QueryRowContext(ctx, `
        SELECT id, first_name, last_name, birthday_date, relationship_status,
               employed_at, phone_number, profile_picture, pictures, level
        FROM users WHERE id = ?
    `, id)

	var birthday sql.NullTime
	err := row.Scan(&u.ID, &u.FirstName, &u.LastName, &birthday, &u.RelationshipStatus, &u.EmployedAt, &u.PhoneNumber, &u.ProfilePicture, &u.Pictures, &u.Level)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Printf("[WARN] GetUserByID: No user found with ID: %d", id)
		} else {
			log.Printf("[ERROR] GetUserByID: Scan failed: %v", err)
		}
		return models.UserProfile{}, err
	}
	u.Birthday = birthday
	return u, nil
}

func UserPrivacy(ctx context.Context, db *sql.DB, userID int, isPublic *bool) error {
	query := `
        SELECT profile_type FROM users 
        WHERE id = ?`
	err := db.QueryRowContext(ctx, query, userID).Scan(isPublic)
	if err != nil {
		return err
	}
	return nil
}

// GetUserVisibilitySettings returns the user's visibility settings and profile_type
func GetUserVisibilitySettings(ctx context.Context, db *sql.DB, userID int) (map[string]interface{}, error) {
	log.Printf("[INFO] GetUserVisibilitySettings: Fetching settings for user %d", userID)
	query := `
		SELECT u.profile_type,
			   COALESCE(us.email_vis, 0) as email_vis,
			   COALESCE(us.birthday_date_vis, 1) as birthday_date_vis,
			   COALESCE(us.relationship_status_vis, 1) as relationship_status_vis,
			   COALESCE(us.employed_at_vis, 1) as employed_at_vis,
			   COALESCE(us.phone_number_vis, 0) as phone_number_vis,
			   COALESCE(us.about_me_vis, 1) as about_me_vis,
			   COALESCE(us.nickname_vis, 1) as nickname_vis,
			   COALESCE(us.follow_vis, 0) as follow_vis
		FROM users u
		LEFT JOIN user_settings us ON u.id = us.id
		WHERE u.id = ?
	`

	var profileTypeRaw interface{}
	var emailVis, birthdayVis, relVis, employedVis, phoneVis int
	var aboutVis, nickVis, followVis int
	row := db.QueryRowContext(ctx, query, userID)
	if err := row.Scan(&profileTypeRaw, &emailVis, &birthdayVis, &relVis, &employedVis, &phoneVis, &aboutVis, &nickVis, &followVis); err != nil {
		if err == sql.ErrNoRows {
			log.Printf("[WARN] GetUserVisibilitySettings: no row for user %d", userID)
		}
		return nil, err
	}

	// Map integer flags to strings
	mapVis := func(v int) string {
		if v == 1 {
			return "visible"
		}
		return "hidden"
	}

	// profile_type can be stored as integer or boolean depending on how it was inserted.
	pt := "public"
	switch v := profileTypeRaw.(type) {
	case int64:
		if v == 1 {
			pt = "private"
		}
	case int:
		if v == 1 {
			pt = "private"
		}
	case bool:
		if v {
			pt = "private"
		}
	case nil:
		// default public
	default:
		// try to handle numeric-like strings
		// ignore and default to public
	}

	res := map[string]interface{}{
		"email_vis":               mapVis(emailVis),
		"birthday_date_vis":       mapVis(birthdayVis),
		"relationship_status_vis": mapVis(relVis),
		"employed_at_vis":         mapVis(employedVis),
		"phone_number_vis":        mapVis(phoneVis),
		"about_me_vis":            mapVis(aboutVis),
		"nickname_vis":            mapVis(nickVis),
		"follow_vis":              mapVis(followVis),
		"profile_type":            pt,
	}

	return res, nil
}

// UpdateUserVisibilitySettings updates the visibility settings and optionally profile_type for a user.
// If a pointer argument is nil, that field is not changed.
func UpdateUserVisibilitySettings(ctx context.Context, db *sql.DB, userID int,
	emailVis, birthdayVis, relVis, employedVis, phoneVis, aboutVis, nickVis, followVis *int, profileType *int) error {
	log.Printf("[INFO] UpdateUserVisibilitySettings: user=%d", userID)

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Fetch current values (use defaults if missing)
	var curEmail, curBirthday, curRel, curEmployed, curPhone, curAbout, curNick, curFollow int
	row := tx.QueryRowContext(ctx, `SELECT COALESCE(email_vis,0), COALESCE(birthday_date_vis,1), COALESCE(relationship_status_vis,1), COALESCE(employed_at_vis,1), COALESCE(phone_number_vis,0), COALESCE(about_me_vis,1), COALESCE(nickname_vis,1), COALESCE(follow_vis,0) FROM user_settings WHERE id = ?`, userID)
	if err := row.Scan(&curEmail, &curBirthday, &curRel, &curEmployed, &curPhone, &curAbout, &curNick, &curFollow); err != nil {
		if err == sql.ErrNoRows {
			// use defaults
			curEmail, curBirthday, curRel, curEmployed, curPhone, curAbout, curNick, curFollow = 0, 1, 1, 1, 0, 1, 1, 0
		} else {
			return err
		}
	}

	// Determine final values
	if emailVis != nil {
		curEmail = *emailVis
	}
	if birthdayVis != nil {
		curBirthday = *birthdayVis
	}
	if relVis != nil {
		curRel = *relVis
	}
	if employedVis != nil {
		curEmployed = *employedVis
	}
	if phoneVis != nil {
		curPhone = *phoneVis
	}
	if aboutVis != nil {
		curAbout = *aboutVis
	}
	if nickVis != nil {
		curNick = *nickVis
	}
	if followVis != nil {
		curFollow = *followVis
	}

	// Upsert into user_settings
	_, err = tx.ExecContext(ctx, `INSERT INTO user_settings (id, email_vis, birthday_date_vis, relationship_status_vis, employed_at_vis, phone_number_vis, about_me_vis, nickname_vis, follow_vis) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET email_vis = excluded.email_vis, birthday_date_vis = excluded.birthday_date_vis, relationship_status_vis = excluded.relationship_status_vis, employed_at_vis = excluded.employed_at_vis, phone_number_vis = excluded.phone_number_vis, about_me_vis = excluded.about_me_vis, nickname_vis = excluded.nickname_vis, follow_vis = excluded.follow_vis;`, userID, curEmail, curBirthday, curRel, curEmployed, curPhone, curAbout, curNick, curFollow)
	if err != nil {
		return err
	}

	// Update profile_type if provided
	if profileType != nil {
		_, err = tx.ExecContext(ctx, `UPDATE users SET profile_type = ? WHERE id = ?;`, *profileType, userID)
		if err != nil {
			return err
		}
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	log.Printf("[SUCCESS] UpdateUserVisibilitySettings: updated user=%d", userID)
	return nil
}

// InsertDefaultUserSettings inserts a default user_settings row for userID using the DB connection.
// It is safe to call if a row already exists.
func InsertDefaultUserSettings(ctx context.Context, db *sql.DB, userID int) error {
	_, err := db.ExecContext(ctx, `INSERT INTO user_settings (id, email_vis, birthday_date_vis, relationship_status_vis, employed_at_vis, phone_number_vis, about_me_vis, nickname_vis, follow_vis) VALUES (?, 0, 1, 1, 1, 0, 1, 1, 0) ON CONFLICT(id) DO NOTHING;`, userID)
	return err
}

// GetUserContentSettings returns editable profile fields used by settings content manager.
func GetUserContentSettings(ctx context.Context, db *sql.DB, userID int) (map[string]interface{}, error) {
	log.Printf("[INFO] GetUserContentSettings: Fetching content settings for user %d", userID)

	query := `
		SELECT
			COALESCE(first_name, ''),
			COALESCE(last_name, ''),
			COALESCE(date(birthday_date), ''),
			COALESCE(relationship_status, ''),
			COALESCE(employed_at, ''),
			COALESCE(location, ''),
			COALESCE(phone_number, ''),
			COALESCE(nickname, ''),
			COALESCE(about_me, '')
		FROM users
		WHERE id = ?
	`

	var firstName, lastName, birthdayDate, relationshipStatus, employedAt, location, phoneNumber, nickname, aboutMe string
	row := db.QueryRowContext(ctx, query, userID)
	if err := row.Scan(&firstName, &lastName, &birthdayDate, &relationshipStatus, &employedAt, &location, &phoneNumber, &nickname, &aboutMe); err != nil {
		if err == sql.ErrNoRows {
			log.Printf("[WARN] GetUserContentSettings: no row for user %d", userID)
		}
		return nil, err
	}

	return map[string]interface{}{
		"first_name":          firstName,
		"last_name":           lastName,
		"birthday_date":       birthdayDate,
		"relationship_status": relationshipStatus,
		"employed_at":         employedAt,
		"location":            location,
		"phone_number":        phoneNumber,
		"nickname":            nickname,
		"about_me":            aboutMe,
	}, nil
}

// UpdateUserContentSettings updates editable profile fields for a user.
// Only non-nil fields are updated.
func UpdateUserContentSettings(ctx context.Context, db *sql.DB, userID int,
	firstName, lastName, birthdayDate, relationshipStatus, employedAt, location, phoneNumber, nickname, aboutMe *string) error {
	setClauses := make([]string, 0, 8)
	args := make([]interface{}, 0, 9)

	if firstName != nil {
		setClauses = append(setClauses, "first_name = ?")
		args = append(args, *firstName)
	}
	if lastName != nil {
		setClauses = append(setClauses, "last_name = ?")
		args = append(args, *lastName)
	}

	addNullableField := func(column string, value *string) {
		if value == nil {
			return
		}
		if strings.TrimSpace(*value) == "" {
			setClauses = append(setClauses, column+" = NULL")
			return
		}
		setClauses = append(setClauses, column+" = ?")
		args = append(args, *value)
	}

	addNullableField("birthday_date", birthdayDate)
	addNullableField("relationship_status", relationshipStatus)
	addNullableField("employed_at", employedAt)
	addNullableField("location", location)
	addNullableField("phone_number", phoneNumber)
	addNullableField("nickname", nickname)
	addNullableField("about_me", aboutMe)

	if len(setClauses) == 0 {
		return nil
	}

	query := "UPDATE users SET " + strings.Join(setClauses, ", ") + " WHERE id = ?"
	args = append(args, userID)

	res, err := db.ExecContext(ctx, query, args...)
	if err != nil {
		return err
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return sql.ErrNoRows
	}

	return nil
}

func normalizeProfileType(raw interface{}) int {
	switch v := raw.(type) {
	case int64:
		if v == 1 {
			return 1
		}
		return 0
	case int:
		if v == 1 {
			return 1
		}
		return 0
	case bool:
		if v {
			return 1
		}
		return 0
	default:
		return 0
	}
}

func hasRelationshipStatus(ctx context.Context, db *sql.DB, followerID, followedID int, status string) (bool, error) {
	var one int
	err := db.QueryRowContext(ctx,
		`SELECT 1 FROM followers WHERE follower_id = ? AND followed_id = ? AND status = ? LIMIT 1`,
		followerID, followedID, status,
	).Scan(&one)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

func relationshipStatusForProfileView(ctx context.Context, db *sql.DB, viewerID, targetID int) (string, error) {
	var outgoing string
	err := db.QueryRowContext(ctx, `SELECT status FROM followers WHERE follower_id = ? AND followed_id = ?`, viewerID, targetID).Scan(&outgoing)
	if err != nil && err != sql.ErrNoRows {
		return "", err
	}

	if err == nil {
		switch outgoing {
		case "accepted":
			return "Following", nil
		case "pending":
			return "Pending", nil
		case "blocked":
			return "Blocked", nil
		}
	}

	followsBack, err := hasRelationshipStatus(ctx, db, targetID, viewerID, "accepted")
	if err != nil {
		return "", err
	}
	if followsBack {
		return "Follow Back", nil
	}

	return "Follow", nil
}

// GetUserProfileView builds the profile payload according to relationship, privacy and visibility rules.
func GetUserProfileView(ctx context.Context, db *sql.DB, viewerID, targetID int) (map[string]interface{}, error) {
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

	var id int
	var email, firstName, lastName, profilePicture, coverImage, birthdayDate, relationshipStatus, employedAt, location, phoneNumber, nickname, aboutMe string
	var followersCount, followingCount int
	var profileTypeRaw interface{}
	var emailVis, birthdayVis, relationshipVis, employedVis, phoneVis, nicknameVis, aboutVis int
	var followVis int

	if err := db.QueryRowContext(ctx, query, targetID).Scan(
		&id,
		&email,
		&firstName,
		&lastName,
		&profilePicture,
		&coverImage,
		&birthdayDate,
		&relationshipStatus,
		&employedAt,
		&location,
		&phoneNumber,
		&nickname,
		&aboutMe,
		&followersCount,
		&followingCount,
		&profileTypeRaw,
		&emailVis,
		&birthdayVis,
		&relationshipVis,
		&employedVis,
		&phoneVis,
		&nicknameVis,
		&aboutVis,
		&followVis,
	); err != nil {
		return nil, err
	}

	res := map[string]interface{}{
		"id":              id,
		"first_name":      firstName,
		"last_name":       lastName,
		"profile_picture": profilePicture,
		"cover_image":     coverImage,
		"followers":       followersCount,
		"following":       followingCount,
	}
	mapVis := func(v int) string {
		if v == 1 {
			return "visible"
		}
		return "hidden"
	}

	// Own profile: skip relationship checks.
	if viewerID == targetID {
		res["own_profile"] = true
		res["follow_vis"] = "visible"
		res["location"] = location
		if emailVis == 1 {
			res["email"] = email
		}
		if birthdayVis == 1 {
			res["birthday_date"] = birthdayDate
		}
		if relationshipVis == 1 {
			res["relationship_status"] = relationshipStatus
		}
		if employedVis == 1 {
			res["employed_at"] = employedAt
		}
		if phoneVis == 1 {
			res["phone_number"] = phoneNumber
		}
		if nicknameVis == 1 {
			res["nickname"] = nickname
		}
		if aboutVis == 1 {
			res["about_me"] = aboutMe
		}
		return res, nil
	}

	res["own_profile"] = false

	// If viewer has blocked target -> return blocked status and no profile data.
	viewerBlockedTarget, err := hasRelationshipStatus(ctx, db, viewerID, targetID, "blocked")
	if err != nil {
		return nil, err
	}
	if viewerBlockedTarget {
		return map[string]interface{}{
			"id":             id,
			"own_profile":    false,
			"current_status": "Blocked",
		}, nil
	}

	// If target has blocked viewer -> return blocked status and no profile data.
	blockedByTarget, err := hasRelationshipStatus(ctx, db, targetID, viewerID, "blocked")
	if err != nil {
		return nil, err
	}
	if blockedByTarget {
		return map[string]interface{}{
			"id":             id,
			"own_profile":    false,
			"current_status": "You_Are_Blocked",
		}, nil
	}

	currentStatus, err := relationshipStatusForProfileView(ctx, db, viewerID, targetID)
	if err != nil {
		return nil, err
	}
	res["current_status"] = currentStatus

	isPrivate := normalizeProfileType(profileTypeRaw) == 1
	canViewFull := !isPrivate || currentStatus == "Following"
	if !canViewFull {
		res["follow_vis"] = "hidden"
		return res, nil
	}

	res["follow_vis"] = mapVis(followVis)

	if emailVis == 1 {
		res["email"] = email
	}

	if birthdayVis == 1 {
		res["birthday_date"] = birthdayDate
	}
	if relationshipVis == 1 {
		res["relationship_status"] = relationshipStatus
	}
	if employedVis == 1 {
		res["employed_at"] = employedAt
		res["location"] = location
	}
	if phoneVis == 1 {
		res["phone_number"] = phoneNumber
	}
	if nicknameVis == 1 {
		res["nickname"] = nickname
	}
	if aboutVis == 1 {
		res["about_me"] = aboutMe
	}

	return res, nil
}
