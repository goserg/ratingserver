package service

import (
	"database/sql"
	"ratingserver/gen/model"
	"ratingserver/gen/table"
	"ratingserver/internal/domain"
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
