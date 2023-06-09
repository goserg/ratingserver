package users

import (
	"github.com/google/uuid"
	"time"
)

type User struct {
	ID           uuid.UUID
	Name         string
	Roles        []string // TODO role type
	RegisteredAt time.Time
}

type Secret struct {
	PasswordHash []byte
	Salt         []byte
}
