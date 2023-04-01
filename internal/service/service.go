package service

import (
	"database/sql"
	"fmt"
	"ratingserver/gen/model"
	"ratingserver/gen/table"
	"ratingserver/internal/domain"
	"ratingserver/internal/elo"
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

func (s *PlayerService) GetRatings() ([]domain.Rating, error) {
	var matches []model.Matches
	err := table.Matches.
		SELECT(table.Matches.AllColumns).
		FROM(table.Matches).
		Query(s.db, &matches)
	if err != nil {
		return nil, err
	}
	playerMap := make(map[string]int)
	for _, match := range matches {
		playerRatingA, ok := playerMap[match.PlayerA]
		if !ok {
			playerRatingA = 1000
		}
		playerRatingB, ok := playerMap[match.PlayerB]
		if !ok {
			playerRatingB = 1000
		}
		pointsA, pointsB := calculatePoints(match.PlayerA, match.Winner)
		playerCoefficientA := calculatePlayerCoefficient(calculatePlayerGameCount(), playerRatingA)
		playerCoefficientB := calculatePlayerCoefficient(calculatePlayerGameCount(), playerRatingA)
		playerMap[match.PlayerA] = elo.Calculate(playerRatingA, playerRatingB, playerCoefficientA, pointsA)
		playerMap[match.PlayerB] = elo.Calculate(playerRatingB, playerRatingA, playerCoefficientB, pointsB)
	}
	players, err := s.List()
	if err != nil {
		return nil, err
	}
	for i := range players {
		players[i].EloRating = playerMap[players[i].ID.String()]
	}
	fmt.Println(players)
	return nil, nil
}

func calculatePlayerGameCount() int {
	// TODO not implemented
	return 0
}

func calculatePlayerCoefficient(n int, rating int) int {
	if n <= 30 {
		return 40
	}
	if rating >= 2400 {
		return 10
	}
	return 20
}

func calculatePoints(a string, winner *string) (elo.Points, elo.Points) {
	if winner == nil {
		return elo.Draw, elo.Draw
	}
	if *winner == a {
		return elo.Win, elo.Lose
	}
	return elo.Lose, elo.Win
}
