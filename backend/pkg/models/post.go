package models

import "time"

// Post is a merged model preserving fields from both branches.
type Post struct {
    ID                   int       `json:"id"`
    UserID               int       `json:"user_id"`
    Content              string    `json:"content"`
    ExtraContent         string    `json:"extra_content,omitempty"`
    Image                string    `json:"image,omitempty"`
    Privacy              string    `json:"privacy,omitempty"`
    WhitelistedUsers     []int     `json:"whitelisted_users,omitempty"`
    CreatedAt            time.Time `json:"created_at_time,omitempty"`
    CreatedAtRaw         string    `json:"created_at,omitempty"`
    // Author info
    AuthorFirstName      string `json:"author_first_name,omitempty"`
    AuthorLastName       string `json:"author_last_name,omitempty"`
    AuthorProfilePicture string `json:"author_profile_picture,omitempty"`
}

type FeedResponse struct {
    Posts           []Post           `json:"posts"`
    DiscoveredUsers []DiscoveredUser `json:"discovered_users"`
}
