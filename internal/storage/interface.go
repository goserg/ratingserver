package storage

import "ratingserver/internal/domain"

type PlayerStorage interface {
	ListPlayers() ([]domain.Player, error)

	ImportPlayers([]domain.Player) error
}

type MatchStorage interface {
	ListMatches() ([]domain.Match, error)

	ImportMatches([]domain.Match) error
}
