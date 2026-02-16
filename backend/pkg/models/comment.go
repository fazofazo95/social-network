package models

type Comment struct {
	ID           int    `json:"id"`
	UserID       int    `json:"user_id"`
	ParentType   string `json:"parent_type"`
	ParentID     int    `json:"parent_id"`
	Content      string `json:"content"`
	ExtraContent string `json:"extra_content,omitempty"`
	CreatedAt    string `json:"created_at"`

	AuthorFirstName      string `json:"author_first_name"`
	AuthorLastName       string `json:"author_last_name"`
	AuthorProfilePicture string `json:"author_profile_picture,omitempty"`
}
