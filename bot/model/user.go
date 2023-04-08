package model

import "time"

type EventType string

const (
	NewMatch EventType = "new_match"
)

type User struct {
	ID        int
	FirstName string
	Username  string
	CreatedAt time.Time
	UpdatedAt time.Time

	Subscriptions []EventType
}
