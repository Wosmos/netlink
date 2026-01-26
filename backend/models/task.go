package models

import "time"

type Task struct {
	ID             int       `json:"id"`
	UserID         int       `json:"user_id"`
	ConversationID *int      `json:"conversation_id"` // Optional - null for personal tasks
	Text           string    `json:"text"`
	Completed      bool      `json:"completed"`
	CreatedAt      time.Time `json:"created_at"`
}
