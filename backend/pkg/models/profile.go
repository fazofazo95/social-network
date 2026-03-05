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

// UserProfileInput is used to create or update a user's profile in the database
// Required fields: ID, FirstName, LastName, Level
// Other fields are optional and may be nil to leave NULL in the DB
type UserProfileInput struct {
	ID                 int
	FirstName          string
	LastName           string
	Birthday           *time.Time
	RelationshipStatus *string
	EmployedAt         *string
	PhoneNumber        *string
	ProfilePicture     *string
	Pictures           *string
	Level              string
}

type SignupFields struct {
	Email     string
	Password  string
	FirstName string
	LastName  string
	Username  string
	Birthday  string
	Avatar    string
	Nickname  string
	AboutMe   string
}

type RawVisibilityData struct {
	ProfileType bool
	EmailVis    int
	BirthdayVis int
	RelVis      int
	EmployedVis int
	PhoneVis    int
	AboutVis    int
	NickVis     int
	FollowVis   int
}

type VisibilitySettings struct {
	EmailVis              string `json:"email_vis"`
	BirthdayDateVis       string `json:"birthday_date_vis"`
	RelationshipStatusVis string `json:"relationship_status_vis"`
	EmployedAtVis         string `json:"employed_at_vis"`
	PhoneNumberVis        string `json:"phone_number_vis"`
	AboutMeVis            string `json:"about_me_vis"`
	NicknameVis           string `json:"nickname_vis"`
	FollowVis             string `json:"follow_vis"`
	ProfileType           string `json:"profile_type"`
}

type UpdateVisibilityRequest struct {
	ProfileType           *string `json:"profile_type"`
	EmailVis              *string `json:"email_vis"`
	BirthdayVis           *string `json:"birthday_date_vis"`
	RelationshipStatusVis *string `json:"relationship_status_vis"`
	EmployedAtVis         *string `json:"employed_at_vis"`
	PhoneNumberVis        *string `json:"phone_number_vis"`
	AboutMeVis            *string `json:"about_me_vis"`
	NicknameVis           *string `json:"nickname_vis"`
	FollowVis             *string `json:"follow_vis"`
}

type RawProfileData struct {
	ID                 int
	Email              string
	FirstName          string
	LastName           string
	ProfilePicture     string
	CoverImage         string
	BirthdayDate       string
	RelationshipStatus string
	EmployedAt         string
	Location           string
	PhoneNumber        string
	Nickname           string
	AboutMe            string
	FollowersCount     int
	FollowingCount     int
	ProfileType        bool

	// Visibility Flags
	EmailVis        int
	BirthdayVis     int
	RelationshipVis int
	EmployedVis     int
	PhoneVis        int
	NicknameVis     int
	AboutVis        int
	FollowVis       int
}

type UserProfileDTO struct {
	ID             int    `json:"id"`
	FirstName      string `json:"first_name"`
	LastName       string `json:"last_name"`
	ProfilePicture string `json:"profile_picture"`
	CoverImage     string `json:"cover_image"`
	Followers      int    `json:"followers"`
	Following      int    `json:"following"`
	OwnProfile     bool   `json:"own_profile"`
	CurrentStatus  string `json:"current_status,omitempty"` // Following, Blocked, etc
	FollowVis      string `json:"follow_vis,omitempty"`     // visible/hidden

	Email              string `json:"email,omitempty"`
	BirthdayDate       string `json:"birthday_date,omitempty"`
	RelationshipStatus string `json:"relationship_status,omitempty"`
	EmployedAt         string `json:"employed_at,omitempty"`
	Location           string `json:"location,omitempty"`
	PhoneNumber        string `json:"phone_number,omitempty"`
	Nickname           string `json:"nickname,omitempty"`
	AboutMe            string `json:"about_me,omitempty"`
}
