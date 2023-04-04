package web

import (
	"errors"
	"ratingserver/internal/domain"
	"time"

	"github.com/google/uuid"
)

type createMatch struct {
	PlayerA uuid.UUID `json:"player_a_id"`
	PlayerB uuid.UUID `json:"player_b_id"`
	Winner  uuid.UUID `json:"winner_id"`
}

func (c createMatch) Validate() error {
	if c.PlayerA.ID() == 0 || c.PlayerB.ID() == 0 {
		return errors.New("оба игрока должны присутствовать")
	}
	if c.Winner.ID() != 0 && (c.Winner.ID() != c.PlayerA.ID() && c.Winner.ID() != c.PlayerB.ID()) {
		return errors.New("ID победителя не совпадает с ID учасников")
	}
	return nil
}

func (c createMatch) convertToDomainMatch() (domain.Match, error) {
	return domain.Match{
		PlayerA: domain.Player{ID: c.PlayerA},
		PlayerB: domain.Player{ID: c.PlayerB},
		Winner:  domain.Player{ID: c.Winner},
		Date:    time.Now(),
	}, nil
}
