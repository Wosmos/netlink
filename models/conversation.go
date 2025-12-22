package models

import "time"

type ConversationType string

const (
	ConversationTypeDirect ConversationType = "direct"
	ConversationTypeGroup  ConversationType = "group"
)

type Conversation struct {
	ID        int              `json:"id"`
	Type      ConversationType `json:"type"`
	Name      string           `json:"name"`   // For groups
	Avatar    string           `json:"avatar"` // For groups
	CreatedBy int              `json:"created_by"`
	CreatedAt time.Time        `json:"created_at"`
	UpdatedAt time.Time        `json:"updated_at"`

	// Populated fields
	Members     []ConversationMember `json:"members,omitempty"`
	LastMessage *Message             `json:"last_message,omitempty"`
	UnreadCount int                  `json:"unread_count,omitempty"`
}

type ConversationMember struct {
	ConversationID int       `json:"conversation_id"`
	UserID         int       `json:"user_id"`
	Role           string    `json:"role"` // admin, member
	JoinedAt       time.Time `json:"joined_at"`
	LastReadAt     time.Time `json:"last_read_at"`

	// Populated
	User *User `json:"user,omitempty"`
}
