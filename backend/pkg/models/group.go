package models

type CreateGroupInput struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Visibility  string `json:"visibility"`
	JoinMode    string `json:"join_mode"`
	Picture     string `json:"picture"`
}

type GroupResponse struct {
	ID           int    `json:"id"`
	Name         string `json:"name"`
	Description  string `json:"description,omitempty"`
	OwnerID      int    `json:"owner_id"`
	Visibility   string `json:"visibility"`
	JoinMode     string `json:"join_mode"`
	GroupPicture string `json:"group_picture,omitempty"`
	GroupMembers int    `json:"group_members"`
	CreatedAt    string `json:"created_at"`
}

type GroupMemberListItem struct {
	ID             int    `json:"id"`
	FirstName      string `json:"first_name"`
	LastName       string `json:"last_name"`
	ProfilePicture string `json:"profile_picture"`
	Role           string `json:"group_status"`
}

type GroupPendingItem struct {
	ID             int    `json:"id"`
	FirstName      string `json:"first_name"`
	LastName       string `json:"last_name"`
	ProfilePicture string `json:"profile_picture"`
	Type           string `json:"type"`
}

type GroupDiscoverItem struct {
	ID           int    `json:"id"`
	Name         string `json:"name"`
	Description  string `json:"description"`
	GroupPicture string `json:"group_picture"`
	GroupMembers int    `json:"group_members"`
	OwnerFirst   string `json:"owner_first_name"`
	OwnerLast    string `json:"owner_last_name"`
	Type         string `json:"type"`
}
