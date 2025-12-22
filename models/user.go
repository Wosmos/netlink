package models

import "time"

type User struct {
	ID                int
	Email             string
	PasswordHash      string
	IsVerified        bool
	VerificationToken *string
	ResetToken        *string
	ResetTokenExpires *time.Time
	CreatedAt         time.Time
}
