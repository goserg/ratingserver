package sqlite

import (
	"database/sql"
	"ratingserver/gen/model"
	"ratingserver/gen/table"
	"ratingserver/internal/domain"
	"ratingserver/internal/storage"

	"github.com/sirupsen/logrus"

	"github.com/go-jet/jet/v2/sqlite"

	"github.com/google/uuid"
)

type Storage struct {
	db  *sql.DB
	log *logrus.Entry
}

var _ storage.PlayerStorage = (*Storage)(nil)
var _ storage.MatchStorage = (*Storage)(nil)

func New(l *logrus.Logger) (*Storage, error) {
	log := l.WithFields(map[string]interface{}{
		"from": "sqlite",
	})
	db, err := sql.Open("sqlite3", "file:rating.sqlite?cache=shared")
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)
	err = db.Ping()
	if err != nil {
		return nil, err
	}
	log.Info("db connected")
	return &Storage{
		db:  db,
		log: log,
	}, nil
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
	return convertPlayersToDomain(players), err
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
	domainMatches, err := convertMatchesToDomain(matches)
	if err != nil {
		return nil, err
	}
	for i := range domainMatches {
		domainMatches[i].PlayerA = playerMap[domainMatches[i].PlayerA.ID]
		domainMatches[i].PlayerB = playerMap[domainMatches[i].PlayerB.ID]
		if domainMatches[i].Winner.ID != uuid.Nil {
			domainMatches[i].Winner = playerMap[domainMatches[i].Winner.ID]
		}
	}
	return domainMatches, nil
}

func convertPlayersToMap(players []domain.Player) map[uuid.UUID]domain.Player {
	m := make(map[uuid.UUID]domain.Player)
	for i := range players {
		m[players[i].ID] = players[i]
	}
	return m
}

func (s *Storage) ImportPlayers(players []domain.Player) error {
	var mPlayers []model.Players
	for i := range players {
		mPlayers = append(mPlayers, convertPlayerFromDomain(players[i]))
	}
	_, err := table.Players.INSERT(table.Players.AllColumns).MODELS(mPlayers).Exec(s.db)
	if err != nil {
		return err
	}
	return nil
}

func (s *Storage) ImportMatches(matches []domain.Match) error {
	var mMatches []model.Matches
	for i := range matches {
		mMatches = append(mMatches, convertMatchesFromDomain(matches[i]))
	}
	_, err := table.Matches.INSERT(table.Matches.AllColumns).MODELS(mMatches).Exec(s.db)
	if err != nil {
		return err
	}
	return nil
}

func convertMatchesFromDomain(match domain.Match) model.Matches {
	var m model.Matches
	m.ID = int32(match.ID)
	m.PlayerA = match.PlayerA.ID.String()
	m.PlayerB = match.PlayerB.ID.String()
	if match.Winner.ID.String() != "" {
		id := match.Winner.ID.String()
		m.Winner = &id
	}
	m.CreatedAt = match.Date
	return m
}

func (s *Storage) Create(match domain.Match) (domain.Match, error) {
	dMatch := convertMatchesFromDomain(match)
	err := table.Matches.
		INSERT(
			table.Matches.PlayerA,
			table.Matches.PlayerB,
			table.Matches.Winner,
			table.Matches.CreatedAt,
		).
		MODEL(dMatch).
		RETURNING(table.Matches.AllColumns).
		Query(s.db, &dMatch)
	if err != nil {
		return domain.Match{}, err
	}
	return convertMatchToDomain(dMatch)
}

func (s *Storage) Get(uuid uuid.UUID) (domain.Player, error) {
	var p model.Players
	err := table.Players.
		SELECT(table.Players.AllColumns).
		FROM(table.Players).
		WHERE(table.Players.ID.EQ(sqlite.String(uuid.String()))).
		Query(s.db, &p)
	if err != nil {
		return domain.Player{}, err
	}
	return convertPlayerToDomain(p)
}
