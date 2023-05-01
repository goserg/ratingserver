package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"errors"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
	"ratingserver/auth/storage"
	"ratingserver/auth/users"
	"time"
)

type Service struct {
	storage storage.AuthStorage
	cfg     Config
}

func New(ctx context.Context, cfg Config, storage storage.AuthStorage) (*Service, error) {
	s := Service{
		cfg:     cfg,
		storage: storage,
	}
	_, err := s.storage.GetUserByName(ctx, "root")
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			return nil, err
		}
		secret, err := generateSecret(cfg.RootPassword, cfg.PasswordPepper)
		if err != nil {
			return nil, err
		}
		err = s.storage.CreateUser(ctx, users.User{
			ID:           uuid.New(),
			Name:         "root",
			Roles:        nil,
			RegisteredAt: time.Now(),
		}, secret)
		if err != nil {
			return nil, err
		}
	}
	return &s, nil
}

func (s *Service) GetUserByName(ctx context.Context, name string) (users.User, error) {
	return s.storage.GetUserByName(ctx, name)
}

func (s *Service) Login(ctx context.Context, name string, password string) (users.User, error) {
	return s.storage.GetUserByName(ctx, name) // TODO check password
}

func (s *Service) GenerateJWTCookie(userID uuid.UUID) (*fiber.Cookie, error) {
	expiresIn, err := time.ParseDuration(s.cfg.Expiration)
	if err != nil {
		return nil, err
	}
	expirationTime := time.Now().Add(expiresIn)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.StandardClaims{
		ExpiresAt: expirationTime.Unix(),
		IssuedAt:  time.Now().Unix(),
		Subject:   userID.String(),
	})
	tokenString, err := token.SignedString([]byte(s.cfg.Token))
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

func (s *Service) Auth(cookie string) (uuid.UUID, error) {
	token, err := jwt.ParseWithClaims(cookie, &jwt.StandardClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(s.cfg.Token), nil
	})
	if err != nil {
		return uuid.Nil, err
	}
	if token.Valid {
		claims, ok := token.Claims.(*jwt.StandardClaims)
		if !ok {
			return uuid.Nil, errors.New("bad request")
		}
		userID := claims.Subject
		id, err := uuid.Parse(userID)
		if err != nil {
			return uuid.Nil, err
		}
		return id, nil
	}
	ve := jwt.ValidationError{}
	if ok := errors.As(err, &ve); !ok {
		return uuid.Nil, err
	}
	if ve.Errors&jwt.ValidationErrorMalformed != 0 {
		return uuid.Nil, errors.New("bad request")
	} else if ve.Errors&(jwt.ValidationErrorExpired|jwt.ValidationErrorNotValidYet) != 0 {
		return uuid.Nil, errors.New("token expired")
	} else {
		return uuid.Nil, err
	}
}

func (s *Service) SignUp(ctx context.Context, name string, password string) error {
	secret, err := generateSecret(password, s.cfg.PasswordPepper)
	if err != nil {
		return err
	}
	err = s.storage.CreateUser(ctx, users.User{
		ID:           uuid.New(),
		Name:         name,
		Roles:        nil, // TODO roles?
		RegisteredAt: time.Now(),
	}, secret)
	if err != nil {
		return err
	}
	return nil
}

func generateSecret(password string, pepper string) (users.Secret, error) {
	sha := sha256.New()
	sha.Write([]byte(pepper + password))
	salt := make([]byte, 8)
	_, err := rand.Read(salt)
	if err != nil {
		return users.Secret{}, err
	}
	sha.Write(salt)
	return users.Secret{
		PasswordHash: sha.Sum(nil),
		Salt:         salt,
	}, nil
}
