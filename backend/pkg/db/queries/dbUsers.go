package queries

import (
	"context"
	"database/sql"
	"time"

	"backend/pkg/models"
)

// CreateUserProfile inserts a new row into users. The caller must ensure a corresponding login_users row exists
// (users.id is a FK to login_users.id). Required fields: ID, FirstName, LastName, Level.
func CreateUserProfile(ctx context.Context, db *sql.DB, in models.UserProfileInput) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
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
		return err
	}

	return tx.Commit()
}

// UpdateUserProfile updates all profile columns for the given ID. Pass nil for optional fields to set them to NULL.
func UpdateUserProfile(ctx context.Context, db *sql.DB, in models.UserProfileInput) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
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
		return err
	}

	return tx.Commit()
}

// GetUserByID loads a user profile by id. If no row exists, sql.ErrNoRows is returned.
func GetUserByID(ctx context.Context, db *sql.DB, id int) (models.UserProfile, error) {
	var u models.UserProfile
	row := db.QueryRowContext(ctx, `
        SELECT id, first_name, last_name, birthday_date, relationship_status,
               employed_at, phone_number, profile_picture, pictures, level
        FROM users WHERE id = ?
    `, id)

	var birthday sql.NullTime
	err := row.Scan(&u.ID, &u.FirstName, &u.LastName, &birthday, &u.RelationshipStatus, &u.EmployedAt, &u.PhoneNumber, &u.ProfilePicture, &u.Pictures, &u.Level)
	if err != nil {
		return models.UserProfile{}, err
	}
	u.Birthday = birthday
	return u, nil
}

// Helper to create a mock UserProfileInput filled with example data for manual checks.
func MockUserProfileInput(id int) models.UserProfileInput {
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

// MarkProfileComplete marks the login_users.completed flag for the given user id.
func MarkProfileComplete(ctx context.Context, db *sql.DB, userID int) error {
	_, err := db.ExecContext(ctx, `UPDATE login_users SET completed = 1 WHERE id = ?;`, userID)
	return err
}

// DeleteStaleIncompleteUsers deletes login_users rows (and cascades) where completed = 0
// and created_at is older than the provided duration (olderThan). It returns the number of deleted rows.
func DeleteStaleIncompleteUsers(ctx context.Context, db *sql.DB, olderThan time.Duration) (int64, error) {
	cutoff := time.Now().Add(-olderThan)
	res, err := db.ExecContext(ctx, `DELETE FROM login_users WHERE completed = 0 AND created_at < ?;`, cutoff)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}
