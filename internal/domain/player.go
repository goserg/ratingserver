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

type Match struct {
	ID      int
	PlayerA *Player
	PlayerB *Player
	Winner  *Player
	Date    time.Time
}
