package postgres

import (
	"context"
	"database/sql"
	"encoding/hex"
	"errors"
	"time"

	"github.com/go-jet/jet/v2/sqlite"

	"github.com/go-jet/jet/v2/postgres"
	"github.com/go-jet/jet/v2/qrm"
	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/google/uuid"
	"github.com/goserg/ratingserver/auth/service"
	"github.com/goserg/ratingserver/auth/storage"
	"github.com/goserg/ratingserver/auth/users"
	"github.com/goserg/ratingserver/gen/auth/public/model"
	"github.com/goserg/ratingserver/gen/auth/public/table"
	"github.com/sirupsen/logrus"
)

type Storage struct {
	db  *sql.DB
	log *logrus.Entry
}

var _ storage.AuthStorage = (*Storage)(nil)

func New(config service.StorageConfig) (*Storage, error) {
	db, err := sql.Open("pgx", "postgres://postgres:postgres@localhost:5431/auth")
	if err != nil {
		return nil, err
	}
	if err := db.PingContext(context.Background()); err != nil {
		return nil, err
	}
	return &Storage{
		db: db,
	}, nil
}

func (s Storage) GetUserSecret(ctx context.Context, user users.User) (users.Secret, error) {
	return inTx(ctx, s.db, func(tx *sql.Tx) (users.Secret, error) {
		var where sqlite.BoolExpression
		switch {
		case user.ID != uuid.Nil:
			where = table.Users.ID.EQ(postgres.UUID(user.ID))
		case user.Name != "":
			where = table.Users.Username.EQ(postgres.String(user.Name))
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
			WHERE(where).QueryContext(ctx, tx, &dbUser)
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
	})
}

func (s Storage) CreateUser(ctx context.Context, user users.User, secret users.Secret) error {
	return inTxSimple(ctx, s.db, func(tx *sql.Tx) error {
		dbUser := model.Users{
			ID:           user.ID,
			Username:     user.Name,
			PasswordHash: bytesToHex(secret.PasswordHash),
			PasswordSalt: bytesToHex(secret.Salt),
			CreatedAt:    time.Now(),
		}
		_, err := table.Users.INSERT(table.Users.AllColumns).MODEL(dbUser).ExecContext(ctx, tx)
		if err != nil {
			return err
		}
		for _, role := range user.Roles {
			userRoleDB := model.UserRoles{
				UserID: user.ID,
				Role:   role,
			}
			_, err = table.UserRoles.INSERT(table.UserRoles.AllColumns).MODEL(userRoleDB).ExecContext(ctx, tx)
			if err != nil {
				return err
			}
		}
		return nil
	})
}

func (s Storage) SignIn(ctx context.Context, name string, passwordHash []byte) (uuid.UUID, error) {
	return inTx(ctx, s.db, func(tx *sql.Tx) (uuid.UUID, error) {
		var user model.Users
		err := table.Users.
			SELECT(
				table.Users.AllColumns.Except(
					table.Users.PasswordHash,
					table.Users.PasswordSalt,
				),
			).
			WHERE(
				table.Users.Username.EQ(postgres.String(name)).
					AND(table.Users.DeletedAt.IS_NULL()).
					AND(table.Users.PasswordHash.EQ(postgres.String(bytesToHex(passwordHash)))),
			).
			QueryContext(ctx, tx, &user)
		if err != nil {
			if errors.Is(err, qrm.ErrNoRows) {
				return uuid.Nil, sql.ErrNoRows
			}
			return uuid.Nil, err
		}
		token := uuid.New()
		now := time.Now()
		tokenDB := model.Tokens{
			UserID:       user.ID,
			Token:        token,
			CreatedAt:    now,
			LastActiveAt: now,
		}
		_, err = table.Tokens.
			INSERT(table.Tokens.AllColumns.Except(table.Tokens.ID, table.Tokens.DeletedAt)).
			MODEL(tokenDB).
			Exec(tx)
		if err != nil {
			return uuid.Nil, err
		}
		return token, nil
	})
}

func (s Storage) GetUser(ctx context.Context, id uuid.UUID) (users.User, error) {
	return inTx(ctx, s.db, func(tx *sql.Tx) (users.User, error) {
		var dest struct {
			model.Users
			UserRoles []model.UserRoles
		}
		err := table.Users.
			SELECT(
				table.Users.AllColumns.Except(
					table.Users.PasswordHash,
					table.Users.PasswordSalt,
				),
				table.UserRoles.AllColumns).
			FROM(
				table.Users.INNER_JOIN(table.UserRoles, table.UserRoles.UserID.EQ(table.Users.ID))).
			WHERE(
				table.Users.ID.EQ(postgres.UUID(id)).
					AND(table.Users.DeletedAt.IS_NULL())).
			QueryContext(ctx, tx, &dest)
		if err != nil {
			return users.User{}, err
		}
		return convertDBUserToModel(dest.Users, dest.UserRoles), nil
	})
}

func (s Storage) Me(ctx context.Context, token uuid.UUID) (users.User, error) {
	return inTx(ctx, s.db, func(tx *sql.Tx) (users.User, error) {
		var dest struct {
			model.Users
			UserRoles []model.UserRoles
		}
		err := table.Users.
			SELECT(
				table.Users.AllColumns.Except(
					table.Users.PasswordHash,
					table.Users.PasswordSalt,
				),
				table.UserRoles.AllColumns).
			FROM(
				table.Users.
					INNER_JOIN(table.UserRoles, table.UserRoles.UserID.EQ(table.Users.ID)).
					INNER_JOIN(table.Tokens, table.Tokens.UserID.EQ(table.Users.ID))).
			WHERE(
				table.Tokens.Token.EQ(postgres.UUID(token)).
					AND(table.Users.DeletedAt.IS_NULL()).
					AND(table.Tokens.DeletedAt.IS_NULL())).
			QueryContext(ctx, tx, &dest)
		if err != nil {
			return users.User{}, err
		}
		return convertDBUserToModel(dest.Users, dest.UserRoles), nil
	})
}

func convertDBUserToModel(user model.Users, roles []model.UserRoles) users.User {
	u := users.User{
		ID:           user.ID,
		Name:         user.Username,
		Roles:        []string{},
		RegisteredAt: user.CreatedAt,
	}

	for _, role := range roles {
		u.Roles = append(u.Roles, role.Role)
	}
	return u
}

func bytesToHex(b []byte) string {
	return hex.EncodeToString(b)
}

func hexToBytes(s string) ([]byte, error) {
	return hex.DecodeString(s)
}

func inTx[T any](ctx context.Context, db *sql.DB, fn func(tx *sql.Tx) (T, error)) (T, error) {
	var zero T
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return zero, err
	}
	value, err := fn(tx)
	if err != nil {
		return zero, errors.Join(err, tx.Rollback())
	}
	return value, tx.Commit()
}

func inTxSimple(ctx context.Context, db *sql.DB, fn func(tx *sql.Tx) error) error {
	_, err := inTx(ctx, db, func(tx *sql.Tx) (struct{}, error) { return struct{}{}, fn(tx) })
	return err
}
