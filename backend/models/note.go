package models

import "time"

type Note struct {
	ID             int       `json:"id"`
	UserID         int       `json:"user_id"`
	ConversationID *int      `json:"conversation_id"` // Optional - null for personal notes
	Title          string    `json:"title"`
	Content        string    `json:"content"`
	Color          string    `json:"color"`
	Pinned         bool      `json:"is_pinned"` // Changed to match frontend
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}
