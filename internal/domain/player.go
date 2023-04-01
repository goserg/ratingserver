package domain

import (
	"time"

	"github.com/google/uuid"
)

type Player struct {
	ID           uuid.UUID
	Name         string
	RegisteredAt time.Time
	EloRating    int
}

type Rating struct {
	Player *Player
	Value  int
}
