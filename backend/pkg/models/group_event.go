package models

type GroupEventCreateInput struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	EventDay    string `json:"event_day"`
	EventTime   string `json:"event_time"`
}

type GroupEventResponseInput struct {
	ReactionType string `json:"reaction_type"`
}

type GroupEvent struct {
	ID          int    `json:"id"`
	GroupID     int    `json:"group_id"`
	CreatorID   int    `json:"creator_id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	EventDay    string `json:"event_day"`
	EventTime   string `json:"event_time"`
	CreatedAt   string `json:"created_at"`
	Going       int    `json:"going"`
	NotGoing    int    `json:"not_going"`
	Invited     int    `json:"invited"`
}
