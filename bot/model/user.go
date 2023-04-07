package model

import "time"

type User struct {
	ID        int
	FirstName string
	Username  string
	CreatedAt time.Time
	UpdatedAt time.Time
}
