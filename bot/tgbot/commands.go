package tgbot

import (
	mapset "github.com/deckarep/golang-set/v2"
	"ratingserver/bot/model"
)

type Command interface {
	Run(user model.User, args string) (string, error)
	Help() string
	Permission() mapset.Set[model.UserRole]
	Visibility() mapset.Set[model.UserRole]
}

type Commands map[string]Command

func (uc Commands) RunCommand(user model.User, cmd string, args string) (string, error) {
	for s, command := range uc {
		if cmd == s {
			if command.Permission().Contains(user.Role) {
				return command.Run(user, args)
			}
		}
	}
	return "", ErrBadRequest
}
