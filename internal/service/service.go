package service

import (
	"encoding/json"
	"errors"
	"github.com/google/uuid"
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
	playerRatings := make(map[string]int)
	playerGamesPlayed := make(map[string]int)
	for _, match := range matches {
		playerRatingA, ok := playerRatings[match.PlayerA.ID.String()]
		if !ok {
			playerRatingA = 1000
		}
		playerRatingB, ok := playerRatings[match.PlayerB.ID.String()]
		if !ok {
			playerRatingB = 1000
		}
		pointsA, pointsB := calculatePoints(match.PlayerA, match.Winner)
		playerCoefficientA := calculatePlayerCoefficient(playerGamesPlayed[match.PlayerA.ID.String()], playerRatingA)
		playerCoefficientB := calculatePlayerCoefficient(playerGamesPlayed[match.PlayerB.ID.String()], playerRatingB)
		playerRatings[match.PlayerA.ID.String()] = elo.Calculate(playerRatingA, playerRatingB, playerCoefficientA, pointsA)
		playerRatings[match.PlayerB.ID.String()] = elo.Calculate(playerRatingB, playerRatingA, playerCoefficientB, pointsB)

		playerGamesPlayed[match.PlayerA.ID.String()]++
		playerGamesPlayed[match.PlayerB.ID.String()]++
	}
	players, err := s.ListPlayers()
	if err != nil {
		return nil, err
	}
	for i := range players {
		players[i].EloRating = playerRatings[players[i].ID.String()]
		players[i].GamesPlayed = playerGamesPlayed[players[i].ID.String()]
	}
	sort.SliceStable(players, func(i, j int) bool {
		return players[i].EloRating > players[j].EloRating
	})
	return players, nil
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

func calculatePoints(a domain.Player, winner domain.Player) (elo.Points, elo.Points) {
	if winner.ID == uuid.Nil {
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
	playerRatings := make(map[string]int)
	playerGamesPlayed := make(map[string]int)
	for i := range matches {
		playerRatingA, ok := playerRatings[matches[i].PlayerA.ID.String()]
		if !ok {
			playerRatingA = 1000
		}
		playerRatingB, ok := playerRatings[matches[i].PlayerB.ID.String()]
		if !ok {
			playerRatingB = 1000
		}
		pointsA, pointsB := calculatePoints(matches[i].PlayerA, matches[i].Winner)
		playerCoefficientA := calculatePlayerCoefficient(playerGamesPlayed[matches[i].PlayerA.ID.String()], playerRatingA)
		playerCoefficientB := calculatePlayerCoefficient(playerGamesPlayed[matches[i].PlayerB.ID.String()], playerRatingB)

		newRatingA := elo.Calculate(playerRatingA, playerRatingB, playerCoefficientA, pointsA)
		matches[i].PlayerA.RatingChange = newRatingA - playerRatingA
		playerRatings[matches[i].PlayerA.ID.String()] = newRatingA
		newRatingB := elo.Calculate(playerRatingB, playerRatingA, playerCoefficientB, pointsB)
		matches[i].PlayerB.RatingChange = newRatingB - playerRatingB
		playerRatings[matches[i].PlayerB.ID.String()] = newRatingB

		playerGamesPlayed[matches[i].PlayerA.ID.String()]++
		playerGamesPlayed[matches[i].PlayerB.ID.String()]++
	}
	reverse(matches)
	return matches, nil
}

func reverse(m []domain.Match) {
	for i, j := 0, len(m)-1; i < j; i, j = i+1, j-1 {
		m[i], m[j] = m[j], m[i]
	}
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

func (s *PlayerService) Get(id uuid.UUID) (domain.Player, error) {
	return s.playerStorage.Get(id)
}

func (s *PlayerService) GetPlayerGames(id uuid.UUID) (map[uuid.UUID]domain.PlayerStats, error) {
	matches, err := s.matchStorage.ListMatches()
	if err != nil {
		return nil, err
	}
	results := make(map[uuid.UUID]domain.PlayerStats)
	players, err := s.GetRatings()
	if err != nil {
		return nil, err
	}
	for _, player := range players {
		results[player.ID] = domain.PlayerStats{Player: player}
	}
	for i := range matches {
		var this, other *domain.Player
		if matches[i].PlayerA.ID != id && matches[i].PlayerB.ID != id {
			continue
		}
		if matches[i].PlayerA.ID == id {
			this = &matches[i].PlayerA
			other = &matches[i].PlayerB
		} else {
			this = &matches[i].PlayerA
			other = &matches[i].PlayerB
		}
		r := results[other.ID]
		switch {
		case this.ID == matches[i].Winner.ID:
			r.Wins++
		case other.ID == matches[i].Winner.ID:
			r.Loses++
		default:
			r.Draws++
		}
		results[other.ID] = r
	}
	return results, nil
}
