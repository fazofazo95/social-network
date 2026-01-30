package models

type FollowRequest struct {
	FollowerID int    `json:"follower_id"`
	FollowedID int    `json:"followed_id"`
	Status     string `json:"status"`
}
