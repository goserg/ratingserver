package tgbot

import (
	mapset "github.com/deckarep/golang-set/v2"
	"ratingserver/bot/botstorage"
	"ratingserver/bot/model"
)

type SubCommand struct {
	botStorage botstorage.BotStorage
	sub        func(int)
}

func (c *SubCommand) Run(user model.User, _ string) (string, error) {
	err := c.botStorage.Subscribe(user)
	if err != nil {
		return "", err
	}
	c.sub(user.ID)
	return "Подписка оформленна, чтобы отписаться от уведомлений: /unsub", nil
}

func (c *SubCommand) Help() string {
	return `Подписаться на уведомления`
}

func (c *SubCommand) Permission() mapset.Set[model.UserRole] {
	return mapset.NewSet[model.UserRole](model.RoleAdmin, model.RoleModerator, model.RoleUser)
}
func (c *SubCommand) Visibility() mapset.Set[model.UserRole] {
	return mapset.NewSet[model.UserRole](model.RoleAdmin, model.RoleModerator, model.RoleUser)
}
