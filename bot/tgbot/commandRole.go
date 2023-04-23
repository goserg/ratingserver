package tgbot

import (
	"errors"
	mapset "github.com/deckarep/golang-set/v2"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"ratingserver/bot/botstorage"
	"ratingserver/bot/model"
	"strings"
)

type RoleCommand struct {
	adminPassword string
	botStorage    botstorage.BotStorage
}

func (c *RoleCommand) Reset() {}

func (c *RoleCommand) Run(user model.User, args string, resp *tgbotapi.MessageConfig) (string, bool, error) {
	resp.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
	return c.handleRole(user, args)
}

func (c *RoleCommand) Help() string {
	return `Изменение роли. Использование: /role user или /role admin <pass>`
}

func (c *RoleCommand) handleRole(user model.User, args string) (string, bool, error) {
	a := strings.SplitN(args, " ", 2)
	switch a[0] {
	case "admin":
		if user.Role == model.RoleAdmin {
			return "", false, errors.New("эта роль уже задана")
		}
		if len(a) != 2 {
			return "", false, ErrBadRequest
		}
		if a[1] != c.adminPassword { // wrong admin password
			return "", false, ErrBadRequest
		}
		user.Role = model.RoleAdmin
	case "user":
		if user.Role == model.RoleUser {
			return "", false, errors.New("эта роль уже задана")
		}
		user.Role = model.RoleUser
	default:
		return "", false, ErrBadRequest
	}
	err := c.botStorage.UpdateUserRole(user)
	if err != nil {
		return "", false, err
	}
	return "role updated", false, nil
}

func (c *RoleCommand) Permission() mapset.Set[model.UserRole] {
	return mapset.NewSet[model.UserRole](model.RoleAdmin, model.RoleModerator, model.RoleUser)
}
func (c *RoleCommand) Visibility() mapset.Set[model.UserRole] {
	return mapset.NewSet[model.UserRole](model.RoleAdmin, model.RoleModerator)
}
