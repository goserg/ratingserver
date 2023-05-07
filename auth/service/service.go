package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"errors"
	"ratingserver/auth/storage"
	"ratingserver/auth/users"
	"regexp"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
)

type Service struct {
	storage storage.AuthStorage
	cfg     Config
}

var (
	ErrForbidden     = errors.New("access denied")
	ErrNotAuthorized = errors.New("unauthorized")
)

func New(ctx context.Context, cfg Config, storage storage.AuthStorage) (*Service, error) {
	s := Service{
		cfg:     cfg,
		storage: storage,
	}
	_, err := s.storage.GetUserSecret(ctx, users.User{Name: "root"})
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
			Name:         "root",
			Roles:        []string{"admin"},
			RegisteredAt: time.Now(),
		}, secret)
		if err != nil {
			return nil, err
		}
	}
	return &s, nil
}

func (s *Service) Login(ctx context.Context, name string, password string) (users.User, error) {
	userSecret, err := s.storage.GetUserSecret(ctx, users.User{Name: name})
	if err != nil {
		return users.User{}, err
	}
	secret := generateSecret(password, s.cfg.PasswordPepper, userSecret.Salt)
	if err != nil {
		return users.User{}, err
	}
	return s.storage.SignIn(ctx, name, secret.PasswordHash)
}

func (s *Service) GenerateJWTCookie(userID uuid.UUID, host string) (*fiber.Cookie, error) {
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
		Domain:      host,
		Expires:     expirationTime,
		Secure:      false,
		HTTPOnly:    true,
		SameSite:    "",
		SessionOnly: false,
	}, nil
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
							if role == userRole {
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
		return users.User{}, nil
	}
	token, err := jwt.ParseWithClaims(cookie, &jwt.StandardClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(s.cfg.Token), nil
	})
	if err != nil {
		return users.User{}, err
	}
	if token.Valid {
		claims, ok := token.Claims.(*jwt.StandardClaims)
		if !ok {
			return users.User{}, errors.New("bad request")
		}
		userID := claims.Subject
		id, err := uuid.Parse(userID)
		if err != nil {
			return users.User{}, err
		}
		user, err := s.storage.GetUser(ctx, id)
		if err != nil {
			return users.User{}, err
		}
		return user, nil
	}
	ve := jwt.ValidationError{}
	if ok := errors.As(err, &ve); !ok {
		return users.User{}, err
	}
	if ve.Errors&jwt.ValidationErrorMalformed != 0 {
		return users.User{}, errors.New("bad request")
	} else if ve.Errors&(jwt.ValidationErrorExpired|jwt.ValidationErrorNotValidYet) != 0 {
		return users.User{}, errors.New("token expired")
	} else {
		return users.User{}, err
	}
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
		Roles:        nil, // TODO roles?
		RegisteredAt: time.Now(),
	}, secret)
	if err != nil {
		return err
	}
	return nil
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
