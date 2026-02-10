package models

import "time"

type Post struct {
	ID           int       `json:"id"`
	UserID       int       `json:"user_id"`
	Content      string    `json:"content"`
	ExtraContent string    `json:"extra_content,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
	// Author info
	AuthorFirstName      string `json:"author_first_name"`
	AuthorLastName       string `json:"author_last_name"`
	AuthorProfilePicture string `json:"author_profile_picture,omitempty"`
}

type FeedResponse struct {
	Posts           []Post           `json:"posts"`
	DiscoveredUsers []DiscoveredUser `json:"discovered_users"`
}
