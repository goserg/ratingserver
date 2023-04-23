package tgbot

import (
	mapset "github.com/deckarep/golang-set/v2"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"ratingserver/bot/model"
	"strings"
)

type HelpCommand struct {
	commands map[string]Command
}

func (c *HelpCommand) Reset() {}

func (c *HelpCommand) Run(user model.User, args string, resp *tgbotapi.MessageConfig) (string, bool, error) {
	resp.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
	for s, command := range c.commands {
		if !command.Visibility().Contains(user.Role) {
			continue
		}
		if args == s {
			return command.Help(), false, nil
		}
	}
	var b strings.Builder
	b.WriteString("Доступные команды:\n")
	for commandName, command := range c.commands {
		if !command.Visibility().Contains(user.Role) {
			continue
		}
		b.WriteString("/")
		b.WriteString(commandName)
		b.WriteString("\n")
	}
	b.WriteString("Подробная помощь по команде /help и имя команды")
	return b.String(), false, nil
}

func (c *HelpCommand) Help() string {
	return "Выводит список доступных комманд"
}

func (c *HelpCommand) Permission() mapset.Set[model.UserRole] {
	return mapset.NewSet[model.UserRole](model.RoleAdmin, model.RoleModerator, model.RoleUser)
}
func (c *HelpCommand) Visibility() mapset.Set[model.UserRole] {
	return mapset.NewSet[model.UserRole](model.RoleAdmin, model.RoleModerator, model.RoleUser)
}
