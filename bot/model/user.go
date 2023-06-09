package model

import "time"

type EventType string

const (
	NewMatch EventType = "new_match"
)

type UserRole int

const (
	RoleAdmin     UserRole = 1
	RoleModerator UserRole = 2
	RoleUser      UserRole = 3
)

type User struct {
	ID        int
	FirstName string
	Username  string
	CreatedAt time.Time
	UpdatedAt time.Time

	Role UserRole

	Subscriptions []EventType
}
