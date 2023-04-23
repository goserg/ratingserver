package tgbot

import (
	mapset "github.com/deckarep/golang-set/v2"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"ratingserver/bot/botstorage"
	"ratingserver/bot/model"
	"ratingserver/internal/service"
)

type MeCommand struct {
	playerService *service.PlayerService
	botStorage    botstorage.BotStorage
}

func (c *MeCommand) Reset() {}

func (c *MeCommand) Run(user model.User, args string, resp *tgbotapi.MessageConfig) (string, bool, error) {
	resp.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
	if args == "" {
		return c.processMe(user)
	}
	return c.connectMe(user, args)
}

func (c *MeCommand) Help() string {
	return `Информация об избранном игроке.`
}

func (c *MeCommand) processMe(user model.User) (string, bool, error) {
	playerID, err := c.botStorage.GetMyPlayer(user)
	if err != nil {
		return "", false, err
	}
	player, err := c.playerService.Get(playerID)
	if err != nil {
		return "", false, err
	}
	return printPlayer(player), false, nil
}

func (c *MeCommand) Permission() mapset.Set[model.UserRole] {
	return mapset.NewSet[model.UserRole](model.RoleAdmin, model.RoleModerator, model.RoleUser)
}
func (c *MeCommand) Visibility() mapset.Set[model.UserRole] {
	return mapset.NewSet[model.UserRole](model.RoleAdmin, model.RoleModerator, model.RoleUser)
}

func (c *MeCommand) connectMe(user model.User, playerName string) (string, bool, error) {
	player, err := c.playerService.GetByName(playerName)
	if err != nil {
		return "", false, err
	}
	err = c.botStorage.LinkPlayer(user, player)
	if err != nil {
		return "", false, err
	}
	return "игрок " + player.Name + " задан, теперь можно вызвать /me", false, nil
}
