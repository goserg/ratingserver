package sqlite

import (
	"database/sql"
	"ratingserver/gen/model"
	"ratingserver/gen/table"
	"ratingserver/internal/domain"
	"ratingserver/internal/storage"

	"github.com/google/uuid"
)

type Storage struct {
	db *sql.DB
}

var _ storage.PlayerStorage = (*Storage)(nil)
var _ storage.MatchStorage = (*Storage)(nil)

func New() (*Storage, error) {
	db, err := sql.Open("sqlite3", "file:rating.sqlite?cache=shared")
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)
	err = db.Ping()
	if err != nil {
		return nil, err
	}
	return &Storage{db: db}, nil
}

func (s *Storage) ListPlayers() ([]domain.Player, error) {
	var players []model.Players
	err := table.Players.
		SELECT(table.Players.AllColumns).
		FROM(table.Players).
		Query(s.db, &players)
	if err != nil {
		return nil, err
	}
	return convertPlayers(players), err
}

func (s *Storage) ListMatches() ([]domain.Match, error) {
	var matches []model.Matches
	err := table.Matches.
		SELECT(table.Matches.AllColumns).
		FROM(table.Matches).
		Query(s.db, &matches)
	if err != nil {
		return nil, err
	}
	players, err := s.ListPlayers()
	if err != nil {
		return nil, err
	}
	playerMap := convertPlayersToMap(players)
	domainMatches := convertMatches(matches)
	for i := range domainMatches {
		domainMatches[i].PlayerA = playerMap[domainMatches[i].PlayerA.ID]
		domainMatches[i].PlayerB = playerMap[domainMatches[i].PlayerB.ID]
		if domainMatches[i].Winner != nil {
			domainMatches[i].Winner = playerMap[domainMatches[i].Winner.ID]
		}
	}
	return domainMatches, nil
}

func convertPlayersToMap(players []domain.Player) map[uuid.UUID]*domain.Player {
	m := make(map[uuid.UUID]*domain.Player)
	for i := range players {
		m[players[i].ID] = &players[i]
	}
	return m
}
