package service

import (
	"ratingserver/internal/domain"
	"ratingserver/internal/elo"
	"ratingserver/internal/storage"
	"sort"
)

type PlayerService struct {
	playerStorage storage.PlayerStorage
	matchStorage  storage.MatchStorage
}

func New(playerStorage storage.PlayerStorage, matchStorage storage.MatchStorage) *PlayerService {
	return &PlayerService{
		playerStorage: playerStorage,
		matchStorage:  matchStorage,
	}
}

func (s *PlayerService) ListPlayers() ([]domain.Player, error) {
	return s.playerStorage.ListPlayers()
}

func (s *PlayerService) GetRatings() ([]domain.Player, error) {
	matches, err := s.matchStorage.ListMatches()
	if err != nil {
		return nil, err
	}
	playerMap := make(map[string]int)
	for _, match := range matches {
		playerRatingA, ok := playerMap[match.PlayerA.ID.String()]
		if !ok {
			playerRatingA = 1000
		}
		playerRatingB, ok := playerMap[match.PlayerB.ID.String()]
		if !ok {
			playerRatingB = 1000
		}
		pointsA, pointsB := calculatePoints(match.PlayerA, match.Winner)
		playerCoefficientA := calculatePlayerCoefficient(calculatePlayerGameCount(), playerRatingA)
		playerCoefficientB := calculatePlayerCoefficient(calculatePlayerGameCount(), playerRatingA)
		playerMap[match.PlayerA.ID.String()] = elo.Calculate(playerRatingA, playerRatingB, playerCoefficientA, pointsA)
		playerMap[match.PlayerB.ID.String()] = elo.Calculate(playerRatingB, playerRatingA, playerCoefficientB, pointsB)
	}
	players, err := s.ListPlayers()
	if err != nil {
		return nil, err
	}
	for i := range players {
		players[i].EloRating = playerMap[players[i].ID.String()]
	}
	sort.SliceStable(players, func(i, j int) bool {
		return players[i].EloRating > players[j].EloRating
	})
	return players, nil
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

func calculatePoints(a *domain.Player, winner *domain.Player) (elo.Points, elo.Points) {
	if winner == nil {
		return elo.Draw, elo.Draw
	}
	if winner.ID == a.ID {
		return elo.Win, elo.Lose
	}
	return elo.Lose, elo.Win
}

func (s *PlayerService) GetMatches() ([]domain.Match, error) {
	matches, err := s.matchStorage.ListMatches()
	if err != nil {
		return nil, err
	}
	return matches, nil
}
