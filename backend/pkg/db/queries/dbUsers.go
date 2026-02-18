package queries

import (
	"context"
	"database/sql"
	"log"
	"time"

	"backend/pkg/models"
)

func CreateUserProfile(ctx context.Context, db *sql.DB, in models.UserProfileInput) error {
	log.Printf("[INFO] CreateUserProfile: Attempting to create profile for UserID: %d", in.ID)
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		log.Printf("[ERROR] CreateUserProfile: Failed to begin transaction: %v", err)
		return err
	}
	defer tx.Rollback()

	var birthday interface{}
	if in.Birthday != nil {
		birthday = in.Birthday
	} else {
		birthday = nil
	}

	_, err = tx.ExecContext(ctx, `
        INSERT INTO users (
            id, first_name, last_name, birthday_date, relationship_status,
            employed_at, phone_number, profile_picture, pictures, level
        ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?);
    `, in.ID, in.FirstName, in.LastName, birthday, in.RelationshipStatus, in.EmployedAt, in.PhoneNumber, in.ProfilePicture, in.Pictures, in.Level)
	if err != nil {
		log.Printf("[ERROR] CreateUserProfile: Insert failed: %v", err)
		return err
	}

	err = tx.Commit()
	if err != nil {
		log.Printf("[ERROR] CreateUserProfile: Commit failed: %v", err)
		return err
	}
	log.Printf("[SUCCESS] CreateUserProfile: Profile created for UserID: %d", in.ID)
	return nil
}

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

func MockUserProfileInput(id int) models.UserProfileInput {
	log.Printf("[DEBUG] Generating MockUserProfileInput for ID: %d", id)
	now := time.Date(1990, 1, 1, 0, 0, 0, 0, time.UTC)
	rel := "single"
	employed := "Acme Corp"
	phone := "+1234567890"
	profilePic := "profile.jpg"
	pics := "[\"pic1.jpg\",\"pic2.jpg\"]"
	return models.UserProfileInput{
		ID:                 id,
		FirstName:          "Test",
		LastName:           "User",
		Birthday:           &now,
		RelationshipStatus: &rel,
		EmployedAt:         &employed,
		PhoneNumber:        &phone,
		ProfilePicture:     &profilePic,
		Pictures:           &pics,
		Level:              "beginner",
	}
}

func MarkProfileComplete(ctx context.Context, db *sql.DB, userID int) error {
	log.Printf("[INFO] MarkProfileComplete: Setting completed flag for UserID: %d", userID)
	res, err := db.ExecContext(ctx, `UPDATE login_users SET completed = 1 WHERE id = ?;`, userID)
	if err != nil {
		log.Printf("[ERROR] MarkProfileComplete failed: %v", err)
		return err
	}
	rows, _ := res.RowsAffected()
	log.Printf("[SUCCESS] MarkProfileComplete: Rows affected: %d", rows)
	return nil
}

func DeleteStaleIncompleteUsers(ctx context.Context, db *sql.DB, olderThan time.Duration) (int64, error) {
	cutoff := time.Now().Add(-olderThan)
	log.Printf("[INFO] DeleteStaleIncompleteUsers: Deleting users incomplete since before: %v", cutoff)
	res, err := db.ExecContext(ctx, `DELETE FROM login_users WHERE completed = 0 AND created_at < ?;`, cutoff)
	if err != nil {
		log.Printf("[ERROR] DeleteStaleIncompleteUsers failed: %v", err)
		return 0, err
	}
	rows, _ := res.RowsAffected()
	log.Printf("[SUCCESS] DeleteStaleIncompleteUsers: Deleted %d stale users", rows)
	return rows, nil
}

func UserPrivacy(ctx context.Context, db *sql.DB, userID int, isPublic *bool) error {
	log.Printf("[INFO] UserPrivacy: Checking profile_type for UserID: %d", userID)
	query := `
        SELECT profile_type FROM users 
        WHERE id = ?`
	err := db.QueryRowContext(ctx, query, userID).Scan(isPublic)
	if err != nil {
		log.Printf("[ERROR] UserPrivacy failed: %v", err)
		return err
	}
	log.Printf("[INFO] UserPrivacy: UserID: %d, isPublic: %v", userID, *isPublic)
	return nil
}
