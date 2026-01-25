package models

import "time"

type MessageType string

const (
	MessageTypeText   MessageType = "text"
	MessageTypeImage  MessageType = "image"
	MessageTypeFile   MessageType = "file"
	MessageTypeSystem MessageType = "system"
)

type Message struct {
	ID             int         `json:"id"`
	ConversationID int         `json:"conversation_id"`
	SenderID       int         `json:"sender_id"`
	Type           MessageType `json:"type"`
	Content        string      `json:"content"`
	ReplyToID      *int        `json:"reply_to_id,omitempty"`
	CreatedAt      time.Time   `json:"created_at"`
	UpdatedAt      time.Time   `json:"updated_at"`
	DeletedAt      *time.Time  `json:"deleted_at,omitempty"`

	// Populated
	Sender  *User    `json:"sender,omitempty"`
	ReplyTo *Message `json:"reply_to,omitempty"`
}

type MessageRead struct {
	MessageID int       `json:"message_id"`
	UserID    int       `json:"user_id"`
	ReadAt    time.Time `json:"read_at"`
}
