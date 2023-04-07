package sqlite

import (
	"database/sql"
	"ratingserver/bot/botstorage"
	dbmodel "ratingserver/bot/gen/model"
	"ratingserver/bot/gen/table"
	"ratingserver/bot/model"
	"time"
)

type Storage struct {
	db *sql.DB
}

var _ botstorage.BotStorage = (*Storage)(nil)

func New() (*Storage, error) {
	db, err := sql.Open("sqlite3", "file:bot.sqlite?cache=shared")
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)
	err = db.Ping()
	if err != nil {
		return nil, err
	}
	return &Storage{db: db}, nil
}

func (s *Storage) NewUser(user model.User) error {
	_, err := table.Users.
		INSERT(table.Users.AllColumns).
		MODEL(convertUserFromDomain(user)).
		Exec(s.db)
	if err != nil {
		return err
	}
	return nil
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

func (s *Storage) GetUser() (model.User, error) {
	var dbUser dbmodel.Users
	err := table.Users.
		SELECT(table.Users.AllColumns).
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