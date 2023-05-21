package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"errors"
	"regexp"
	"strings"
	"time"

	"github.com/goserg/ratingserver/auth/storage"
	"github.com/goserg/ratingserver/auth/users"

	"github.com/google/uuid"
)

type Service struct {
	storage storage.AuthStorage
	cfg     Config
}

const Root = "root"

var (
	ErrForbidden     = errors.New("access denied")
	ErrNotAuthorized = errors.New("unauthorized")
	ErrAlreadyExists = errors.New("already exists")
)

func New(ctx context.Context, cfg Config, storage storage.AuthStorage) (*Service, error) {
	s := Service{
		cfg:     cfg,
		storage: storage,
	}
	_, err := s.storage.GetUserSecret(ctx, users.User{Name: Root})
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			return nil, err
		}
		salt, err := randomSalt()
		if err != nil {
			return nil, err
		}
		secret := generateSecret(cfg.RootPassword, cfg.PasswordPepper, salt)
		err = s.storage.CreateUser(ctx, users.User{
			ID:           uuid.New(),
			Name:         Root,
			Roles:        []string{"admin"},
			RegisteredAt: time.Now(),
		}, secret)
		if err != nil {
			return nil, err
		}
	}
	return &s, nil
}

func (s *Service) Login(ctx context.Context, name string, password string) (token uuid.UUID, err error) {
	userSecret, err := s.storage.GetUserSecret(ctx, users.User{Name: name})
	if err != nil {
		return uuid.Nil, err
	}
	secret := generateSecret(password, s.cfg.PasswordPepper, userSecret.Salt)
	if err != nil {
		return uuid.Nil, err
	}
	return s.storage.SignIn(ctx, name, secret.PasswordHash)
}

func (s *Service) Auth(ctx context.Context, cookie string, method string, url string) (users.User, error) {
	user, err := s.getUserFromToken(ctx, cookie)
	if err != nil {
		return users.User{}, ErrNotAuthorized
	}

	for _, rule := range s.cfg.Rules {
		r, err := regexp.Compile(rule.Path)
		if err != nil {
			return users.User{}, err
		}
		if r.MatchString(url) {
			for _, ruleMethod := range rule.Method {
				if ruleMethod == "*" || ruleMethod == method {
					for _, role := range rule.Allow {
						if role == "*" {
							return user, nil
						}
						for _, userRole := range user.Roles {
							if role == strings.TrimSpace(userRole) { // "admin     " ??
								return user, nil
							}
						}
					}
					return users.User{}, ErrForbidden
				}
			}
		}
	}
	return users.User{}, ErrForbidden
}

func (s *Service) getUserFromToken(ctx context.Context, cookie string) (users.User, error) {
	if cookie == "" {
		return users.User{
			Roles: []string{"user"},
		}, nil
	}
	token, err := uuid.Parse(cookie)
	if err != nil {
		return users.User{}, err
	}
	user, err := s.storage.Me(ctx, token)
	if err != nil {
		return users.User{}, err
	}
	if user.ID == uuid.Nil {
		return user, ErrNotAuthorized
	}
	return user, nil
}

func (s *Service) SignUp(ctx context.Context, name string, password string) error {
	salt, err := randomSalt()
	if err != nil {
		return err
	}
	secret := generateSecret(password, s.cfg.PasswordPepper, salt)
	err = s.storage.CreateUser(ctx, users.User{
		ID:           uuid.New(),
		Name:         name,
		Roles:        []string{"user"},
		RegisteredAt: time.Now(),
	}, secret)
	if err != nil {
		if strings.HasPrefix(err.Error(), "UNIQUE constraint failed") {
			return ErrAlreadyExists
		}
		return err
	}
	return nil
}

func (s *Service) Logout(ctx context.Context, token uuid.UUID) error {
	return s.storage.LogOut(ctx, token)
}

func randomSalt() ([]byte, error) {
	salt := make([]byte, 8)
	_, err := rand.Read(salt)
	if err != nil {
		return nil, err
	}
	return salt, nil
}

func generateSecret(password string, pepper string, salt []byte) users.Secret {
	sha := sha256.New()
	sha.Write([]byte(pepper + password))

	sha.Write(salt)
	return users.Secret{
		PasswordHash: sha.Sum(nil),
		Salt:         salt,
	}
}
