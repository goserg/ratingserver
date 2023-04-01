package service

import (
	"database/sql"
	"ratingserver/gen/model"
	"ratingserver/gen/table"
	"ratingserver/internal/domain"

	"github.com/google/uuid"
)

type PlayerService struct {
	db *sql.DB
}

func New(db *sql.DB) *PlayerService {
	return &PlayerService{
		db: db,
	}
}

func (s *PlayerService) List() ([]domain.Player, error) {
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

func convertPlayers(players []model.Players) []domain.Player {
	converted := make([]domain.Player, 0, len(players))
	for _, player := range players {
		id, err := uuid.Parse(player.ID)
		if err != nil {
			return nil
		}
		converted = append(converted, domain.Player{
			ID:           id,
			Name:         player.Name,
			RegisteredAt: player.CreatedAt,
			EloRating:    777,
		})
	}
	return converted
}
