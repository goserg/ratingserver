package tgbot

import (
	mapset "github.com/deckarep/golang-set/v2"
	"ratingserver/bot/botstorage"
	"ratingserver/bot/model"
)

type UnsubCommand struct {
	botStorage botstorage.BotStorage
	unsub      func(int)
}

func (c *UnsubCommand) Run(user model.User, _ string) (string, bool, error) {
	err := c.botStorage.Unsubscribe(user)
	if err != nil {
		return "", false, err
	}
	c.unsub(user.ID)
	return "Подписка отменена, чтобы подписаться на уведомления: /sub", false, nil
}

func (c *UnsubCommand) Help() string {
	return `Отписаться от уведомлений`
}

func (c *UnsubCommand) Permission() mapset.Set[model.UserRole] {
	return mapset.NewSet[model.UserRole](model.RoleAdmin, model.RoleModerator, model.RoleUser)
}
func (c *UnsubCommand) Visibility() mapset.Set[model.UserRole] {
	return mapset.NewSet[model.UserRole](model.RoleAdmin, model.RoleModerator, model.RoleUser)
}
