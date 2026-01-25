package models

import "time"

type User struct {
	ID                int        `json:"id"`
	Email             string     `json:"email"`
	Phone             string     `json:"phone,omitempty"`
	Name              string     `json:"name"`
	Avatar            string     `json:"avatar,omitempty"`
	PasswordHash      string     `json:"-"`
	IsVerified        bool       `json:"is_verified"`
	VerificationToken *string    `json:"-"`
	ResetToken        *string    `json:"-"`
	ResetTokenExpires *time.Time `json:"-"`
	LastSeenAt        *time.Time `json:"last_seen_at,omitempty"`
	CreatedAt         time.Time  `json:"created_at"`
}

// IsOnline returns true if user was seen in last 2 minutes
func (u *User) IsOnline() bool {
	if u.LastSeenAt == nil {
		return false
	}
	return time.Since(*u.LastSeenAt) < 2*time.Minute
}
