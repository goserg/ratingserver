package tgbot

import (
	mapset "github.com/deckarep/golang-set/v2"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"ratingserver/bot/botstorage"
	"ratingserver/bot/model"
)

type SubCommand struct {
	botStorage botstorage.BotStorage
	sub        func(int)
}

func (c *SubCommand) Reset() {}

func (c *SubCommand) Run(user model.User, _ string, resp *tgbotapi.MessageConfig) (string, bool, error) {
	resp.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
	err := c.botStorage.Subscribe(user)
	if err != nil {
		return "", false, err
	}
	c.sub(user.ID)
	return "Подписка оформленна, чтобы отписаться от уведомлений: /unsub", false, nil
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
