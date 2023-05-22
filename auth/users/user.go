package users

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID           uuid.UUID
	Name         string
	Email        string
	Roles        []string // TODO role type
	RegisteredAt time.Time
}

type Secret struct {
	PasswordHash []byte
	Salt         []byte
}
