package service

import (
	"encoding/json"
	"errors"
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
		playerCoefficientB := calculatePlayerCoefficient(calculatePlayerGameCount(), playerRatingB)
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

const exportVersion = 1

type export struct {
	Version int
	Players []domain.Player
	Matches []domain.Match
}

func (s *PlayerService) Export() ([]byte, error) {
	players, err := s.GetRatings()
	if err != nil {
		return nil, err
	}
	matches, err := s.GetMatches()
	if err != nil {
		return nil, err
	}
	exportData := export{
		Version: exportVersion,
		Players: players,
		Matches: matches,
	}
	data, err := json.Marshal(exportData)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (s *PlayerService) Import(data []byte) error {
	var importData export
	err := json.Unmarshal(data, &importData)
	if err != nil {
		return err
	}
	if importData.Version != exportVersion {
		return errors.New("invalid export file version")
	}
	err = s.playerStorage.ImportPlayers(importData.Players)
	if err != nil {
		return err
	}
	err = s.matchStorage.ImportMatches(importData.Matches)
	if err != nil {
		return err
	}
	return nil
}

func (s *PlayerService) CreateMatch(match domain.Match) error {
	err := s.matchStorage.Create(match)
	if err != nil {
		return err
	}
	return nil
}
