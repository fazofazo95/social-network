package models

type SendMessageInput struct {
	MessageType string `json:"message_type"`
	Body        string `json:"body"`
	MediaURL    string `json:"media_url"`
}

type ChatMessage struct {
	ID          int    `json:"id"`
	ChatID      int    `json:"chat_id"`
	SenderID    int    `json:"sender_id"`
	MessageType string `json:"message_type"`
	Body        string `json:"body,omitempty"`
	MediaURL    string `json:"media_url,omitempty"`
	CreatedAt   string `json:"created_at"`
}

type ChatSummary struct {
	ChatID             int    `json:"chat_id"`
	Type               string `json:"type"`
	GroupID            int    `json:"group_id,omitempty"`
	GroupName          string `json:"group_name,omitempty"`
	GroupPicture       string `json:"group_picture,omitempty"`
	OtherUserID        int    `json:"other_user_id,omitempty"`
	OtherUserFirstName string `json:"other_user_first_name,omitempty"`
	OtherUserLastName  string `json:"other_user_last_name,omitempty"`
	OtherUserPicture   string `json:"other_user_picture,omitempty"`
	LastMessageID      int    `json:"last_message_id,omitempty"`
	LastMessageSender  int    `json:"last_message_sender_id,omitempty"`
	LastMessageType    string `json:"last_message_type,omitempty"`
	LastMessagePreview string `json:"last_message_preview,omitempty"`
	LastMessageAt      string `json:"last_message_at,omitempty"`
	Seen               bool   `json:"seen"`
}
