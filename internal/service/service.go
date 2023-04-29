package service

import (
	"encoding/json"
	"errors"
	glicko "github.com/zelenin/go-glicko2"
	"ratingserver/internal/cache/mem"
	"ratingserver/internal/domain"
	"ratingserver/internal/elo"
	"ratingserver/internal/normalize"
	"ratingserver/internal/storage"
	"sort"
	"time"

	"github.com/google/uuid"
)

type PlayerService struct {
	playerStorage storage.PlayerStorage
	matchStorage  storage.MatchStorage
	cache         *mem.Cache
}

func New(playerStorage storage.PlayerStorage, matchStorage storage.MatchStorage, cache *mem.Cache) (*PlayerService, error) {
	p := PlayerService{
		playerStorage: playerStorage,
		matchStorage:  matchStorage,
		cache:         cache,
	}
	return &p, p.updateCache()
}

func (s *PlayerService) updateCache() error {
	players, err := s.getRatings()
	if err != nil {
		return err
	}
	glicko2, err := s.getGlicko2()
	if err != nil {
		return err
	}
	for i := range players {
		players[i].Glicko2Rating = glicko2[players[i].ID].Glicko2Rating
	}
	s.cache.Update(players)
	return nil
}

func (s *PlayerService) getGlicko2() (map[uuid.UUID]domain.Player, error) {
	matches, err := s.matchStorage.ListMatches()
	if err != nil {
		return nil, err
	}
	ps, err := s.playerStorage.ListPlayers()
	if err != nil {
		return nil, err
	}
	players := make(map[uuid.UUID]*glicko.Player)
	for i := range ps {
		players[ps[i].ID] = glicko.NewDefaultPlayer()
	}
	period := glicko.NewRatingPeriod()
	for _, player := range players {
		period.AddPlayer(player)
	}
	start := matches[0].Date
	for i := range matches {
		if matches[i].Date.After(start.Add(time.Hour * 24)) {
			start = matches[i].Date
			period.Calculate()
			period = glicko.NewRatingPeriod()
			for _, player := range players {
				period.AddPlayer(player)
			}
		}
		w := glicko.MATCH_RESULT_DRAW
		pA := matches[i].PlayerA
		pB := matches[i].PlayerB
		switch matches[i].Winner.ID {
		case pA.ID:
			w = glicko.MATCH_RESULT_WIN
		case pB.ID:
			w = glicko.MATCH_RESULT_LOSS
		}
		period.AddMatch(players[pA.ID], players[pB.ID], w)
	}
	period.Calculate()
	for i := range ps {
		ps[i].Glicko2Rating.Rating = players[ps[i].ID].Rating().R()
		ps[i].Glicko2Rating.RatingDeviation = players[ps[i].ID].Rating().Rd()
		ps[i].Glicko2Rating.Sigma = players[ps[i].ID].Rating().Sigma()
		ps[i].Glicko2Rating.Interval.Min, ps[i].Glicko2Rating.Interval.Max = players[ps[i].ID].Rating().ConfidenceInterval()
	}
	playersMap := make(map[uuid.UUID]domain.Player)
	for i := range ps {
		playersMap[ps[i].ID] = ps[i]
	}
	return playersMap, nil
}

func (s *PlayerService) getRatings() ([]domain.Player, error) {
	matches, err := s.matchStorage.ListMatches()
	if err != nil {
		return nil, err
	}
	players, err := s.playerStorage.ListPlayers()
	if err != nil {
		return nil, err
	}
	playersMap := make(map[uuid.UUID]*domain.Player)
	for _, player := range players {
		p := player
		playersMap[player.ID] = &p
	}
	matches = calculateMatches(matches)
	for i := range matches {
		playersMap[matches[i].PlayerA.ID].EloRating = matches[i].PlayerA.EloRating
		playersMap[matches[i].PlayerB.ID].EloRating = matches[i].PlayerB.EloRating
		playersMap[matches[i].PlayerA.ID].GamesPlayed = matches[i].PlayerA.GamesPlayed
		playersMap[matches[i].PlayerB.ID].GamesPlayed = matches[i].PlayerB.GamesPlayed
		playersMap[matches[i].PlayerA.ID].RatingChange = matches[i].PlayerA.RatingChange
		playersMap[matches[i].PlayerB.ID].RatingChange = matches[i].PlayerB.RatingChange
	}
	for i := range players {
		players[i] = *playersMap[players[i].ID]
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
		calculateMatch(&matches[i], players)
	}
	return matches
}

func calculateMatch(match *domain.Match, players map[uuid.UUID]domain.Player) {
	playerA, ok := players[match.PlayerA.ID]
	if !ok {
		playerA = match.PlayerA
		playerA.EloRating = 1000
	}
	playerB, ok := players[match.PlayerB.ID]
	if !ok {
		playerB = match.PlayerB
		playerB.EloRating = 1000
	}
	pointsA, pointsB := calculatePoints(playerA, match.Winner)
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

	match.PlayerA = playerA
	match.PlayerB = playerB
	players[playerA.ID] = playerA
	players[playerB.ID] = playerB
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
	players, err := s.getRatings()
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

func (s *PlayerService) CreateMatch(match domain.Match) (m domain.Match, err error) {
	defer func() {
		if err != nil {
			return
		}
		err = s.updateCache()
	}()

	if match.PlayerA.ID == match.PlayerB.ID {
		return domain.Match{}, errors.New("должно участвовать два разных игрока")
	}
	return s.matchStorage.Create(match)
}

func (s *PlayerService) Get(playerID uuid.UUID) (domain.Player, error) {
	rating, err := s.getRatings()
	if err != nil {
		return domain.Player{}, err
	}
	for i := range rating {
		if rating[i].ID == playerID {
			return rating[i], nil
		}
	}
	return domain.Player{}, errors.New("not found")
}

func (s *PlayerService) GetRatings() []domain.Player {
	return s.cache.GetRatings()
}

func (s *PlayerService) GetPlayerData(id uuid.UUID) (domain.PlayerCardData, error) {
	var data domain.PlayerCardData

	matches, err := s.matchStorage.ListMatches()
	if err != nil {
		return domain.PlayerCardData{}, err
	}
	results := make(map[uuid.UUID]domain.PlayerStats)
	players, err := s.getRatings()
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
	player, ok := s.cache.GetPlayerByName(name)
	if !ok {
		return domain.Player{}, errors.New("not found")
	}
	return player, nil
}

func (s *PlayerService) CreatePlayer(name string) (player domain.Player, err error) {
	defer func() {
		if err != nil {
			return
		}
		err = s.updateCache()
	}()

	players, err := s.playerStorage.ListPlayers()
	if err != nil {
		return domain.Player{}, err
	}
	normName := normalize.Name(name)
	for _, player := range players {
		if normalize.Name(player.Name) == normName {
			return domain.Player{}, errors.New("player " + player.Name + " already exists")
		}
	}
	newPlayer := domain.Player{
		ID:           uuid.New(),
		Name:         name,
		RegisteredAt: time.Now(),
	}
	player, err = s.playerStorage.Add(newPlayer)
	if err != nil {
		return domain.Player{}, err
	}
	return player, nil
}
