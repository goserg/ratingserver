package service

import (
	"context"
	"ratingserver/internal/auth/users"
)

type Service struct {
}

func New() *Service {
	return &Service{}
}

func (s *Service) Login(ctx context.Context, name string, password string) (users.User, error) {
	return users.User{}, nil
}
