package models

type NotificationType string

const (
	NotificationTypeFollowRequest NotificationType = "follow_request"
	NotificationTypeGroupInvite   NotificationType = "group_invite"
	NotificationTypeGroupJoin     NotificationType = "group_join_request"
	NotificationTypeGroupEvent    NotificationType = "group_event_created"
	NotificationStatusPending     string           = "pending"
	NotificationStatusAccepted    string           = "accepted"
	NotificationStatusRejected    string           = "rejected"
	NotificationStatusRead        string           = "read"
)

type Notification struct {
	ID          int              `json:"id"`
	RecipientID int              `json:"recipient_id"`
	ActorID     *int             `json:"actor_id,omitempty"`
	Type        NotificationType `json:"type"`
	Status      string           `json:"status"`
	GroupID     *int             `json:"group_id,omitempty"`
	EventID     *int             `json:"event_id,omitempty"`
	Content     string           `json:"content"`
	Metadata    string           `json:"metadata"`
	Seen        bool             `json:"seen"`
	CreatedAt   string           `json:"created_at"`
	UpdatedAt   string           `json:"updated_at"`
}

type NotificationWithActor struct {
	Notification
	ActorFirstName string `json:"actor_first_name,omitempty"`
	ActorLastName  string `json:"actor_last_name,omitempty"`
	ActorPicture   string `json:"actor_picture,omitempty"`
}

type CreateNotificationInput struct {
	RecipientID int
	ActorID     *int
	Type        NotificationType
	GroupID     *int
	EventID     *int
	Content     string
	Metadata    string
}

type NotificationActionInput struct {
	Action string `json:"action"`
}

type NotificationSSEPayload struct {
	Event        string                `json:"event"`
	Notification NotificationWithActor `json:"notification"`
}
