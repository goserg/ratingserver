package storage

import (
	"context"
	"ratingserver/auth/users"
)

type AuthStorage interface {
	CreateUser(ctx context.Context, user users.User, secret users.Secret) error
	GetUserByName(ctx context.Context, name string) (users.User, error)
}
