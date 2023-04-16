package tgbot

import (
	mapset "github.com/deckarep/golang-set/v2"
	"ratingserver/bot/model"
	"strings"
)

type HelpCommand struct {
	commands map[string]Command
}

func (c *HelpCommand) Run(user model.User, args string) (string, error) {
	for s, command := range c.commands {
		if !command.Visibility().Contains(user.Role) {
			continue
		}
		if args == s {
			return command.Help(), nil
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
	return b.String(), nil
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
