package storage

import (
	"context"

	"github.com/google/uuid"
	"github.com/goserg/ratingserver/auth/users"
)

type AuthStorage interface {
	GetUserSecret(ctx context.Context, user users.User) (users.Secret, error)
	CreateUser(ctx context.Context, user users.User, secret users.Secret) error
	SignIn(ctx context.Context, name string, pwdHash []byte) (users.User, error)
	GetUser(ctx context.Context, id uuid.UUID) (users.User, error)
}
