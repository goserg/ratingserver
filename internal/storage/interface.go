package storage

import (
	"github.com/google/uuid"
	"ratingserver/internal/domain"
)

type PlayerStorage interface {
	ListPlayers() ([]domain.Player, error)
	Get(uuid uuid.UUID) (domain.Player, error)

	ImportPlayers([]domain.Player) error
}

type MatchStorage interface {
	ListMatches() ([]domain.Match, error)
	Create(domain.Match) error

	ImportMatches([]domain.Match) error
}
