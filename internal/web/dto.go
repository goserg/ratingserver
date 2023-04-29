package web

import (
	"errors"
	"ratingserver/internal/domain"
	"time"

	"github.com/google/uuid"
)

type createMatch struct {
	PlayerA uuid.UUID `json:"playerAId"`
	PlayerB uuid.UUID `json:"playerBId"`
	Winner  uuid.UUID `json:"winnerId"`
}

var ErrWrongWinner = errors.New("ID победителя не совпадает с ID учасников")
var ErrMissingPlayer = errors.New("оба игрока должны присутствовать")

func (c createMatch) Validate() error {
	if c.PlayerA.ID() == 0 || c.PlayerB.ID() == 0 {
		return ErrMissingPlayer
	}
	if c.Winner.ID() != 0 && (c.Winner.ID() != c.PlayerA.ID() && c.Winner.ID() != c.PlayerB.ID()) {
		return ErrWrongWinner
	}
	return nil
}

func (c createMatch) convertToDomainMatch() domain.Match {
	return domain.Match{
		PlayerA: domain.Player{ID: c.PlayerA},
		PlayerB: domain.Player{ID: c.PlayerB},
		Winner:  domain.Player{ID: c.Winner},
		Date:    time.Now(),
	}
}
