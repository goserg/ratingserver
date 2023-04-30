package service

import (
	"context"
	"errors"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
	"ratingserver/auth/storage/sqlite"
	"ratingserver/auth/users"
	"ratingserver/internal/config"
	"time"
)

type Service struct {
	storage *sqlite.Storage
	cfg     config.Server
}

func New(cgf config.Server, storage *sqlite.Storage) *Service {
	return &Service{
		cfg:     cgf,
		storage: storage,
	}
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
		ExpiresAt: expirationTime.Unix(),
		IssuedAt:  time.Now().Unix(),
		Subject:   userID.String(),
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

func (s *Service) Auth(cookie string) error {
	token, err := jwt.ParseWithClaims(cookie, &jwt.StandardClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(s.cfg.Auth.AuthToken), nil
	})
	if err != nil {
		return err
	}
	if token.Valid {
		claims, ok := token.Claims.(*jwt.StandardClaims)
		if !ok {
			return errors.New("bad request")
		}
		user := claims.Subject
		fmt.Println("user:", user) // TODO
		return nil
	}
	ve := jwt.ValidationError{}
	if ok := errors.As(err, &ve); !ok {
		return err
	}
	if ve.Errors&jwt.ValidationErrorMalformed != 0 {
		return errors.New("bad request")
	} else if ve.Errors&(jwt.ValidationErrorExpired|jwt.ValidationErrorNotValidYet) != 0 {
		return errors.New("token expired")
	} else {
		return err
	}
}
