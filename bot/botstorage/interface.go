package botstorage

import "ratingserver/bot/model"

type BotStorage interface {
	NewUser(user model.User) (model.User, error)
	GetUser(id int) (model.User, error)
	Log(model.User, string) error
}
