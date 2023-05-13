package service

import (
	"errors"
	"sort"
	"time"

	"github.com/goserg/ratingserver/internal/cache/mem"
	"github.com/goserg/ratingserver/internal/domain"
	"github.com/goserg/ratingserver/internal/elo"
	"github.com/goserg/ratingserver/internal/normalize"
	"github.com/goserg/ratingserver/internal/storage"

	glicko "github.com/zelenin/go-glicko2"

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
	if len(matches) == 0 {
		return make(map[uuid.UUID]domain.Player), nil
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
		pA := matches[i].PlayerA
		pB := matches[i].PlayerB
		period.AddMatch(players[pA.ID], players[pB.ID], findResult(matches[i].Winner.ID, pA.ID, pB.ID))
	}
	period.Calculate()
	for i := range ps {
		fillGlickoStats(&ps[i], players[ps[i].ID])
	}
	playersMap := make(map[uuid.UUID]domain.Player)
	for i := range ps {
		playersMap[ps[i].ID] = ps[i]
	}
	return playersMap, nil
}

func findResult(winnerID, pAID, pBID uuid.UUID) glicko.MatchResult {
	w := glicko.MATCH_RESULT_DRAW
	switch winnerID {
	case pAID:
		w = glicko.MATCH_RESULT_WIN
	case pBID:
		w = glicko.MATCH_RESULT_LOSS
	}
	return w
}

func fillGlickoStats(player *domain.Player, gPlayer *glicko.Player) {
	player.Glicko2Rating.Rating = gPlayer.Rating().R()
	player.Glicko2Rating.RatingDeviation = gPlayer.Rating().Rd()
	player.Glicko2Rating.Sigma = gPlayer.Rating().Sigma()
	player.Glicko2Rating.Interval.Min, player.Glicko2Rating.Interval.Max = gPlayer.Rating().ConfidenceInterval()
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
		if matches[i].PlayerA.ID != id && matches[i].PlayerB.ID != id {
			continue
		}
		other, r := s.calculateResult(id, &matches[i], results)
		results[other] = r
	}
	data.Results = results
	return data, nil
}

func (s *PlayerService) calculateResult(id uuid.UUID, match *domain.Match, results map[uuid.UUID]domain.PlayerStats) (otherID uuid.UUID, result domain.PlayerStats) {
	var this, other *domain.Player
	if match.PlayerA.ID == id {
		this = &match.PlayerA
		other = &match.PlayerB
	} else {
		this = &match.PlayerB
		other = &match.PlayerA
	}
	r := results[other.ID]
	switch {
	case this.ID == match.Winner.ID:
		r.Wins++
	case other.ID == match.Winner.ID:
		r.Loses++
	default:
		r.Draws++
	}
	return other.ID, r
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
