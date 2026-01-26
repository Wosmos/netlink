package models

import "time"

type MessageType string

const (
	MessageTypeText   MessageType = "text"
	MessageTypeImage  MessageType = "image"
	MessageTypeFile   MessageType = "file"
	MessageTypeSystem MessageType = "system"
	MessageTypeVoice  MessageType = "voice"
)

type Message struct {
	ID             int         `json:"id"`
	ConversationID int         `json:"conversation_id"`
	SenderID       int         `json:"sender_id"`
	Type           MessageType `json:"type"`
	Content        string      `json:"content"`
	ReplyToID      *int        `json:"reply_to_id,omitempty"`
	ForwardedFrom  *int        `json:"forwarded_from,omitempty"` // Original message ID if forwarded
	CreatedAt      time.Time   `json:"created_at"`
	UpdatedAt      time.Time   `json:"updated_at"`
	DeletedAt      *time.Time  `json:"deleted_at,omitempty"`

	// Voice message fields
	VoiceFilePath string    `json:"voice_file_path,omitempty"`
	VoiceDuration float64   `json:"voice_duration,omitempty"`
	VoiceWaveform []float64 `json:"voice_waveform,omitempty"`
	VoiceFileSize int64     `json:"voice_file_size,omitempty"`

	// Populated
	Sender    *User            `json:"sender,omitempty"`
	ReplyTo   *Message         `json:"reply_to,omitempty"`
	Reactions map[string][]int `json:"reactions,omitempty"` // emoji -> [user_ids]
	ReadBy    []int            `json:"read_by,omitempty"`
}

type MessageRead struct {
	MessageID int       `json:"message_id"`
	UserID    int       `json:"user_id"`
	ReadAt    time.Time `json:"read_at"`
}

type MessageReaction struct {
	ID        int       `json:"id"`
	MessageID int       `json:"message_id"`
	UserID    int       `json:"user_id"`
	Emoji     string    `json:"emoji"`
	CreatedAt time.Time `json:"created_at"`
}
