package models

type CreateGroupInput struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Visibility  string `json:"visibility"`
	JoinMode    string `json:"join_mode"`
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
