package sqlite

import (
	"database/sql"
	"github.com/google/uuid"
	"ratingserver/bot/botstorage"
	dbmodel "ratingserver/bot/gen/model"
	"ratingserver/bot/gen/table"
	"ratingserver/bot/model"
	"ratingserver/internal/config"
	"ratingserver/internal/domain"
	sqlite3 "ratingserver/internal/migrate"
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

func New(l *logrus.Logger, cfg config.TgBot) (*Storage, error) {
	log := l.WithFields(map[string]interface{}{
		"from": "bot-storage",
	})
	db, err := sql.Open("sqlite3", buildSource(cfg.SqliteFile))
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)

	err = sqlite3.UpBotDB(db)
	if err != nil {
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		return nil, err
	}
	log.Info("bot storage connected")
	return &Storage{
		db:  db,
		log: log,
	}, nil
}

func buildSource(fileName string) string {
	return "file:" + fileName + "?cache=shared"
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

type GetUserModel struct {
	dbmodel.Users
	UserEvents []struct {
		dbmodel.UserEvents
	}
	UserRole struct {
		dbmodel.UserRoles
	}
}

func (s *Storage) GetUser(id int) (model.User, error) {
	var dest GetUserModel
	err := table.Users.
		SELECT(table.Users.AllColumns, table.UserEvents.AllColumns, table.UserRoles.AllColumns).
		FROM(table.Users.
			FULL_JOIN(table.UserEvents, table.UserEvents.UserID.EQ(sqlite.Int(int64(id)))).
			FULL_JOIN(table.UserRoles, table.UserRoles.UserID.EQ(sqlite.Int(int64(id)))),
		).
		WHERE(table.Users.ID.EQ(sqlite.Int(int64(id)))).
		Query(s.db, &dest)
	if err != nil {
		return model.User{}, err
	}
	if dest.UserRole.RoleID == 0 {
		dest.UserRole.RoleID = 3
		_, err = table.UserRoles.
			INSERT(table.UserRoles.AllColumns).
			MODEL(dbmodel.UserRoles{
				UserID: int32(id),
				RoleID: 3,
			}).
			Exec(s.db)
		if err != nil {
			return model.User{}, err
		}
	}
	return convertGetUserModelToDomain(dest), nil
}

func convertGetUserModelToDomain(user GetUserModel) model.User {
	converted := convertUserToDomain(user.Users)
	for i := range user.UserEvents {
		converted.Subscriptions = append(converted.Subscriptions, model.EventType(user.UserEvents[i].Event))
	}
	converted.Role = model.UserRole(user.UserRole.RoleID)
	return converted
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

func (s *Storage) Subscribe(user model.User) error {
	userEvents := dbmodel.UserEvents{
		UserID: int32(user.ID),
		Event:  string(model.NewMatch),
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
				AND(table.UserEvents.Event.EQ(sqlite.String(string(model.NewMatch)))),
		).Exec(s.db)
	if err != nil {
		return err
	}
	return nil
}

func (s *Storage) ListUsers() ([]model.User, error) {
	var dest []GetUserModel
	err := table.Users.
		SELECT(table.Users.AllColumns, table.UserEvents.AllColumns).
		FROM(table.Users.
			FULL_JOIN(table.UserEvents, table.UserEvents.UserID.EQ(table.Users.ID)),
		).
		Query(s.db, &dest)
	if err != nil {
		return nil, err
	}
	return convertGetUsersModelToDomain(dest), nil
}

func convertGetUsersModelToDomain(dest []GetUserModel) []model.User {
	var converted []model.User
	for i := range dest {
		converted = append(converted, convertGetUserModelToDomain(dest[i]))
	}
	return converted
}

func (s *Storage) UpdateUserRole(user model.User) error {
	_, err := table.UserRoles.
		UPDATE(table.UserRoles.RoleID).
		SET(user.Role).
		WHERE(table.UserRoles.UserID.EQ(sqlite.Int(int64(user.ID)))).
		Exec(s.db)
	if err != nil {
		return err
	}
	return nil
}

func (s *Storage) GetMyPlayer(user model.User) (uuid.UUID, error) {
	var up dbmodel.UserPlayers
	err := table.UserPlayers.
		SELECT(table.UserPlayers.AllColumns).
		WHERE(table.UserPlayers.UserID.EQ(sqlite.Int(int64(user.ID)))).
		Query(s.db, &up)
	if err != nil {
		return uuid.Nil, err
	}
	id, err := uuid.Parse(up.PlayerID)
	if err != nil {
		return [16]byte{}, err
	}
	return id, nil
}

func (s *Storage) LinkPlayer(user model.User, player domain.Player) error {
	up := dbmodel.UserPlayers{
		UserID:   int32(user.ID),
		PlayerID: player.ID.String(),
	}
	_, err := table.UserPlayers.
		INSERT(table.UserPlayers.AllColumns).
		MODEL(up).
		ON_CONFLICT(table.UserPlayers.UserID).
		DO_UPDATE(sqlite.SET(table.UserPlayers.PlayerID.SET(sqlite.String(player.ID.String())))).
		Exec(s.db)
	if err != nil {
		return err
	}
	return nil
}
