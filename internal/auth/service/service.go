package service

import (
	"context"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
	"ratingserver/internal/auth/users"
	"ratingserver/internal/config"
	"time"
)

type Service struct {
	cfg config.Server
}

func New(cgf config.Server) *Service {
	return &Service{cfg: cgf}
}

func (s *Service) Login(ctx context.Context, name string, password string) (users.User, error) {
	return users.User{}, nil
}

func (s *Service) GenerateJWTCookie(userID uuid.UUID) (*fiber.Cookie, error) {
	expiresIn, err := time.ParseDuration(s.cfg.Auth.AuthExpiration)
	if err != nil {
		return nil, err
	}
	expirationTime := time.Now().Add(expiresIn)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.StandardClaims{
		IssuedAt: expirationTime.Unix(),
		Subject:  userID.String(),
	})
	tokenString, err := token.SignedString([]byte(s.cfg.Auth.AuthToken))
	if err != nil {
		return nil, err
	}
	return &fiber.Cookie{
		Name:        "token",
		Value:       tokenString,
		Path:        "/",
		Domain:      "127.0.0.1",
		Expires:     expirationTime,
		Secure:      false,
		HTTPOnly:    true,
		SameSite:    "",
		SessionOnly: false,
	}, nil
}
