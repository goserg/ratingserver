package botstorage

import "ratingserver/bot/model"

type BotStorage interface {
	NewUser(user model.User) error
	GetUser() (model.User, error)
	Log(model.User, string) error
}
