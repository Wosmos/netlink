package models

import "time"

type Task struct {
	ID        int
	UserID    int
	Text      string
	Completed bool
	CreatedAt time.Time
}
