package models

type FollowRequest struct {
	FollowerID int    `json:"follower_id"`
	FollowedID int    `json:"followed_id"`
	Status     string `json:"status"`
}

type DiscoveredUser struct {
	ID             int    `json:"id"`
	FirstName      string `json:"first_name"`
	LastName       string `json:"last_name"`
	ProfilePicture string `json:"profile_picture"`
	Status         string `json:"status"` // Follow, Follow Back, Following, Pending
}
