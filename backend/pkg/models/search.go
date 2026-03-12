package models

type SearchUserItem struct {
	ID             int    `json:"id"`
	FirstName      string `json:"first_name"`
	LastName       string `json:"last_name"`
	Username       string `json:"username"`
	ProfilePicture string `json:"profile_picture"`
	CurrentStatus  string `json:"current_status"`
}

type SearchGroupItem struct {
	ID            int    `json:"id"`
	Name          string `json:"name"`
	Description   string `json:"description"`
	GroupPicture  string `json:"group_picture"`
	GroupMembers  int    `json:"group_members"`
	Visibility    string `json:"visibility"`
	JoinMode      string `json:"join_mode"`
	CurrentStatus string `json:"current_status"`
}
