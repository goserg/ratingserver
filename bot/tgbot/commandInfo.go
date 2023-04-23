package tgbot

import (
	"errors"
	mapset "github.com/deckarep/golang-set/v2"
	"ratingserver/bot/model"
	"ratingserver/internal/domain"
	"ratingserver/internal/service"
	"strconv"
	"strings"
	"time"
)

type InfoCommand struct {
	playerService *service.PlayerService
}

func (c *InfoCommand) Run(_ model.User, args string) (string, bool, error) {
	return c.processInfo(args)
}

func (c *InfoCommand) Help() string {
	return `Информация об игроке. Использование - /info и имя игрока.`
}

func (c *InfoCommand) processInfo(command string) (string, bool, error) {
	fields := strings.Fields(command)
	if len(fields) < 1 {
		return "", false, errors.New(`после /info имя игрока необходимо указывать в этом же соощении. Например "/info джон"`)
	}
	player, err := c.playerService.GetByName(fields[0])
	if err != nil {
		return "", false, err
	}
	return printPlayer(player), false, nil
}

func printPlayer(player domain.Player) string {
	var buf strings.Builder
	buf.WriteString("ID: ")
	buf.WriteString(player.ID.String())
	buf.WriteString("\n")
	buf.WriteString("Имя: ")
	buf.WriteString(player.Name)
	buf.WriteString("\n")
	buf.WriteString("Место в рейтинге: ")
	buf.WriteString(prettifyRank(player))
	buf.WriteString("\n")
	buf.WriteString("Рейтинг: ")
	buf.WriteString(strconv.Itoa(player.EloRating))
	buf.WriteString("\n")
	buf.WriteString("Сыграно игр: ")
	buf.WriteString(strconv.Itoa(player.GamesPlayed))
	buf.WriteString("\n")
	buf.WriteString("Зарегистрирован: ")
	buf.WriteString(player.RegisteredAt.Format(time.RFC1123))
	return buf.String()
}

func prettifyRank(player domain.Player) string {
	if player.RatingRank == 1 {
		return "🥇"
	}
	if player.RatingRank == 2 {
		return "🥈"
	}
	if player.RatingRank == 3 {
		return "🥉"
	}
	return strconv.Itoa(player.RatingRank)
}

func (c *InfoCommand) Permission() mapset.Set[model.UserRole] {
	return mapset.NewSet[model.UserRole](model.RoleAdmin, model.RoleModerator, model.RoleUser)
}
func (c *InfoCommand) Visibility() mapset.Set[model.UserRole] {
	return mapset.NewSet[model.UserRole](model.RoleAdmin, model.RoleModerator, model.RoleUser)
}
