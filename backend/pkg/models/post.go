package models

type Post struct {
	ID               int    `json:"id"`
	UserID           int    `json:"user_id"`
	Content          string `json:"content"`
	Image            string `json:"-"`
	Privacy          string `json:"privacy"`
	WhitelistedUsers []int  `json:"whitelisted_users,omitempty"`
	CreatedAt        string `json:"created_at"`
}
