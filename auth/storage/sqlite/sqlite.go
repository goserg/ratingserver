package sqlite

import (
	"context"
	"database/sql"
	"encoding/hex"
	"errors"
	"github.com/go-jet/jet/v2/qrm"
	"github.com/go-jet/jet/v2/sqlite"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"ratingserver/auth/gen/model"
	"ratingserver/auth/gen/table"
	"ratingserver/auth/storage"
	"ratingserver/auth/users"
	"ratingserver/internal/config"
	sqlite3 "ratingserver/internal/migrate"
	"time"
)

type Storage struct {
	db  *sql.DB
	log *logrus.Entry
}

func (s *Storage) GetUserSecret(ctx context.Context, user users.User) (users.Secret, error) {
	var where sqlite.BoolExpression
	switch {
	case user.ID != uuid.Nil:
		where = table.Users.ID.EQ(sqlite.UUID(user.ID))
	case user.Name != "":
		where = table.Users.Username.EQ(sqlite.String(user.Name))
	default:
		return users.Secret{}, errors.New("empty user")
	}

	var dbUser model.Users
	err := table.Users.
		SELECT(
			table.Users.PasswordHash,
			table.Users.PasswordSalt,
		).
		FROM(table.Users).
		WHERE(where).QueryContext(ctx, s.db, &dbUser)
	if err != nil {
		if errors.Is(err, qrm.ErrNoRows) {
			return users.Secret{}, sql.ErrNoRows
		}
		return users.Secret{}, err
	}
	hash, err := hexToBytes(dbUser.PasswordHash)
	if err != nil {
		return users.Secret{}, err
	}
	salt, err := hexToBytes(dbUser.PasswordSalt)
	if err != nil {
		return users.Secret{}, err
	}
	return users.Secret{
		PasswordHash: hash,
		Salt:         salt,
	}, nil
}

func (s *Storage) CreateUser(ctx context.Context, user users.User, secret users.Secret) error {
	dbUser := model.Users{
		ID:           user.ID.String(),
		Username:     user.Name,
		PasswordHash: bytesToHex(secret.PasswordHash),
		PasswordSalt: bytesToHex(secret.Salt),
		CreatedAt:    time.Now(),
	}
	_, err := table.Users.INSERT(table.Users.AllColumns).MODEL(dbUser).ExecContext(ctx, s.db)
	if err != nil {
		return err
	}
	return nil
}

func (s *Storage) SignIn(ctx context.Context, name string, passwordHash []byte) (users.User, error) {
	var dbUser model.Users
	err := table.Users.
		SELECT(
			table.Users.AllColumns.Except(
				table.Users.PasswordHash,
				table.Users.PasswordSalt,
			),
		).
		WHERE(
			table.Users.Username.EQ(sqlite.String(name)).
				AND(table.Users.DeletedAt.IS_NULL()).
				AND(table.Users.PasswordHash.EQ(sqlite.String(bytesToHex(passwordHash)))),
		).
		QueryContext(ctx, s.db, &dbUser)
	if err != nil {
		if errors.Is(err, qrm.ErrNoRows) { // TODO better handle ErrNoRows
			return users.User{}, sql.ErrNoRows
		}
		return users.User{}, err
	}
	return convertUserToModel(dbUser)
}

func convertUserToModel(user model.Users) (users.User, error) {
	id, err := uuid.Parse(user.ID)
	if err != nil {
		return users.User{}, err
	}
	return users.User{
		ID:           id,
		Name:         user.Username,
		Roles:        nil, // TODO
		RegisteredAt: user.CreatedAt,
	}, nil
}

var _ storage.AuthStorage = (*Storage)(nil)

func New(l *logrus.Logger, cfg config.Server) (*Storage, error) {
	log := l.WithFields(map[string]interface{}{
		"from": "auth-storage",
	})
	db, err := sql.Open("sqlite3", buildSource(cfg.Auth.SqliteFile))
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)

	err = sqlite3.UpAuthDB(db)
	if err != nil {
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		return nil, err
	}
	log.Info("auth storage connected")
	return &Storage{
		db:  db,
		log: log,
	}, nil
}

func buildSource(fileName string) string {
	return "file:" + fileName + "?cache=shared"
}

func bytesToHex(b []byte) string {
	return hex.EncodeToString(b)
}

func hexToBytes(s string) ([]byte, error) {
	return hex.DecodeString(s)
}
