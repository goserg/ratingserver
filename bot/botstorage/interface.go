package botstorage

import (
	"github.com/google/uuid"
	"ratingserver/bot/model"
	"ratingserver/internal/domain"
)

type BotStorage interface {
	NewUser(user model.User) (model.User, error)
	GetUser(id int) (model.User, error)
	Log(model.User, string) error
	Subscribe(user model.User) error
	Unsubscribe(user model.User) error
	ListUsers() ([]model.User, error)
	UpdateUserRole(user model.User) error
	GetMyPlayer(user model.User) (uuid.UUID, error)
	LinkPlayer(user model.User, player domain.Player) error
}
