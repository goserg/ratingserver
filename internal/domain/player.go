package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

type Player struct {
	ID           uuid.UUID
	Name         string
	RegisteredAt time.Time

	RatingRank   int
	EloRating    int
	GamesPlayed  int
	RatingChange int
}

type PlayerStats struct {
	Player Player
	Wins   int
	Draws  int
	Loses  int
}

type PlayerCardData struct {
	Player  Player
	Results map[uuid.UUID]PlayerStats
}

type Match struct {
	ID      int
	PlayerA Player
	PlayerB Player
	Winner  Player
	Date    time.Time
}

func (m Match) Validate() error {
	if m.Winner.ID != uuid.Nil && m.Winner.ID != m.PlayerA.ID && m.Winner.ID != m.PlayerB.ID {
		return errors.New("winner must be empty or one of the players")
	}
	return nil
}
