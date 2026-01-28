package models

import "time"

// LoginInput represents login credentials for database queries
type LoginInput struct {
	Email    string
	Password string
}

// SignUpInput represents signup data for database queries
type SignUpInput struct {
	Username string
	Email    string
	Password string
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
