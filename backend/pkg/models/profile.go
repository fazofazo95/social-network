package models

import (
	"database/sql"
	"time"
)

// UserProfile represents a row from the `users` table
type UserProfile struct {
	ID                 int
	FirstName          string
	LastName           string
	Birthday           sql.NullTime
	RelationshipStatus sql.NullString
	EmployedAt         sql.NullString
	PhoneNumber        sql.NullString
	ProfilePicture     sql.NullString
	Pictures           sql.NullString
	Level              string
}

// UserProfileRequest represents user profile data for creation or update
type UserProfileRequest struct {
	FirstName          string     `json:"first_name"`
	LastName           string     `json:"last_name"`
	Birthday           *time.Time `json:"birthday,omitempty"`
	RelationshipStatus *string    `json:"relationship_status,omitempty"`
	EmployedAt         *string    `json:"employed_at,omitempty"`
	PhoneNumber        *string    `json:"phone_number,omitempty"`
	ProfilePicture     *string    `json:"profile_picture,omitempty"`
	Pictures           *string    `json:"pictures,omitempty"`
	Level              string     `json:"level"`
}

// UserProfileResponse represents user profile data returned to clients
type UserProfileResponse struct {
	ID                 int        `json:"id"`
	FirstName          string     `json:"first_name"`
	LastName           string     `json:"last_name"`
	Birthday           *time.Time `json:"birthday,omitempty"`
	RelationshipStatus *string    `json:"relationship_status,omitempty"`
	EmployedAt         *string    `json:"employed_at,omitempty"`
	PhoneNumber        *string    `json:"phone_number,omitempty"`
	ProfilePicture     *string    `json:"profile_picture,omitempty"`
	Pictures           *string    `json:"pictures,omitempty"`
	Level              string     `json:"level"`
}
