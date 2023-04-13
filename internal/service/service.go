package service

import (
	"encoding/json"
	"errors"
	"ratingserver/internal/domain"
	"ratingserver/internal/elo"
	"ratingserver/internal/storage"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
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
	matches = calculateMatches(matches)
	playerRatings := make(map[string]int)
	playerGamesPlayed := make(map[string]int)
	for i := range matches {
		playerRatings[matches[i].PlayerA.ID.String()] = matches[i].PlayerA.EloRating
		playerRatings[matches[i].PlayerB.ID.String()] = matches[i].PlayerB.EloRating
		playerGamesPlayed[matches[i].PlayerA.ID.String()] = matches[i].PlayerA.GamesPlayed
		playerGamesPlayed[matches[i].PlayerB.ID.String()] = matches[i].PlayerB.GamesPlayed
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
	for i := range players {
		players[i].RatingRank = i + 1
	}
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
	matches = calculateMatches(matches)
	reverse(matches)
	return matches, nil
}

func calculateMatches(matches []domain.Match) []domain.Match {
	players := make(map[uuid.UUID]domain.Player)
	for i := range matches {
		playerA, ok := players[matches[i].PlayerA.ID]
		if !ok {
			playerA = matches[i].PlayerA
			playerA.EloRating = 1000
		}
		playerB, ok := players[matches[i].PlayerB.ID]
		if !ok {
			playerB = matches[i].PlayerB
			playerB.EloRating = 1000
		}
		pointsA, pointsB := calculatePoints(playerA, matches[i].Winner)
		playerCoefficientA := calculatePlayerCoefficient(playerA.GamesPlayed, playerA.EloRating)
		playerCoefficientB := calculatePlayerCoefficient(playerB.GamesPlayed, playerB.EloRating)

		newRatingA := elo.Calculate(playerA.EloRating, playerB.EloRating, playerCoefficientA, pointsA)
		newRatingB := elo.Calculate(playerB.EloRating, playerA.EloRating, playerCoefficientB, pointsB)

		playerA.RatingChange = newRatingA - playerA.EloRating
		playerA.EloRating = newRatingA
		playerB.RatingChange = newRatingB - playerB.EloRating
		playerB.EloRating = newRatingB

		playerA.GamesPlayed++
		playerB.GamesPlayed++

		matches[i].PlayerA = playerA
		matches[i].PlayerB = playerB
		players[playerA.ID] = playerA
		players[playerB.ID] = playerB
	}
	return matches
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

func (s *PlayerService) CreateMatch(match domain.Match) (domain.Match, error) {
	return s.matchStorage.Create(match)
}

func (s *PlayerService) Get(id uuid.UUID) (domain.Player, error) {
	return s.playerStorage.Get(id)
}

func (s *PlayerService) GetPlayerData(id uuid.UUID) (domain.PlayerCardData, error) {
	var data domain.PlayerCardData

	matches, err := s.matchStorage.ListMatches()
	if err != nil {
		return domain.PlayerCardData{}, err
	}
	results := make(map[uuid.UUID]domain.PlayerStats)
	players, err := s.GetRatings()
	if err != nil {
		return domain.PlayerCardData{}, err
	}
	for _, player := range players {
		if player.ID == id {
			data.Player = player
			continue
		}
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
			this = &matches[i].PlayerB
			other = &matches[i].PlayerA
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
	data.Results = results
	return data, nil
}

func (s *PlayerService) GetByName(name string) (domain.Player, error) {
	rating, err := s.GetRatings()
	if err != nil {
		return domain.Player{}, err
	}
	for i := range rating {
		if strings.EqualFold(rating[i].Name, name) {
			return rating[i], nil
		}
	}
	return domain.Player{}, errors.New("not found")
}

func (s *PlayerService) CreatePlayer(name string) (domain.Player, error) {
	newPlayer := domain.Player{
		ID:           uuid.New(),
		Name:         name,
		RegisteredAt: time.Now(),
	}
	player, err := s.playerStorage.Add(newPlayer)
	if err != nil {
		return domain.Player{}, err
	}
	return player, nil
}
