package sqlite

import (
	"database/sql"
	"ratingserver/bot/botstorage"
	dbmodel "ratingserver/bot/gen/model"
	"ratingserver/bot/gen/table"
	"ratingserver/bot/model"
	"strings"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/go-jet/jet/v2/sqlite"
)

type Storage struct {
	db  *sql.DB
	log *logrus.Entry
}

var _ botstorage.BotStorage = (*Storage)(nil)

func New(l *logrus.Logger) (*Storage, error) {
	log := l.WithFields(map[string]interface{}{
		"from": "bot-storage",
	})
	db, err := sql.Open("sqlite3", "file:bot.sqlite?cache=shared")
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)
	err = db.Ping()
	if err != nil {
		return nil, err
	}
	log.Info("bot storage connected")
	return &Storage{
		db: db,
	}, nil
}

func (s *Storage) NewUser(user model.User) (model.User, error) {
	var dbuser dbmodel.Users
	err := table.Users.
		INSERT(table.Users.AllColumns).
		MODEL(convertUserFromDomain(user)).RETURNING(table.Users.AllColumns).
		Query(s.db, &dbuser)
	if err != nil {
		return model.User{}, err
	}
	return convertUserToDomain(dbuser), nil
}

func convertUserFromDomain(user model.User) dbmodel.Users {
	return dbmodel.Users{
		ID:        int32(user.ID),
		FirstName: user.FirstName,
		Username:  user.Username,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}
}

func (s *Storage) GetUser(id int) (model.User, error) {
	var dbUser dbmodel.Users
	err := table.Users.
		SELECT(table.Users.AllColumns).
		WHERE(table.Users.ID.EQ(sqlite.Int(int64(id)))).
		Query(s.db, &dbUser)
	if err != nil {
		return model.User{}, err
	}
	return convertUserToDomain(dbUser), nil
}

func convertUserToDomain(user dbmodel.Users) model.User {
	return model.User{
		ID:        int(user.ID),
		FirstName: user.FirstName,
		Username:  user.Username,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}
}

func (s *Storage) Log(user model.User, msg string) error {
	message := dbmodel.Log{
		UserID:    int32(user.ID),
		Message:   msg,
		CreatedAt: time.Now(),
	}
	_, err := table.Log.
		INSERT(table.Log.UserID, table.Log.Message, table.Log.CreatedAt).
		MODEL(message).
		Exec(s.db)
	if err != nil {
		return err
	}
	return nil
}

type EventType string

const (
	NewMatch = "new_match"
)

func (s *Storage) Subscribe(user model.User) error {
	userEvents := dbmodel.UserEvents{
		UserID: int32(user.ID),
		Event:  NewMatch,
	}
	_, err := table.UserEvents.
		INSERT(table.UserEvents.AllColumns).
		MODEL(userEvents).
		Exec(s.db)
	if err != nil {
		if strings.HasPrefix(err.Error(), "UNIQUE constraint failed") {
			return nil
		}
		return err
	}
	return nil
}

func (s *Storage) Unsubscribe(user model.User) error {
	_, err := table.UserEvents.
		DELETE().
		WHERE(
			table.UserEvents.UserID.EQ(sqlite.Int(int64(user.ID))).
				AND(table.UserEvents.Event.EQ(sqlite.String(NewMatch))),
		).Exec(s.db)
	if err != nil {
		return err
	}
	return nil
}
