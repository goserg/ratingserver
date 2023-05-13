package tgbot

import (
	"errors"
	"strconv"
	"strings"
	"time"

	mapset "github.com/deckarep/golang-set/v2"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/goserg/ratingserver/bot/model"
	"github.com/goserg/ratingserver/internal/domain"
	"github.com/goserg/ratingserver/internal/service"
)

type InfoCommand struct {
	playerService *service.PlayerService
}

func (c *InfoCommand) Reset() {}

func (c *InfoCommand) Run(_ model.User, args string, resp *tgbotapi.MessageConfig) (bool, error) {
	resp.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
	text, err := c.processInfo(args)
	if err != nil {
		return false, err
	}
	resp.Text = text
	return false, nil
}

func (c *InfoCommand) Help() string {
	return `Информация об игроке. Использование - /info и имя игрока.`
}

func (c *InfoCommand) processInfo(command string) (string, error) {
	fields := strings.Fields(command)
	if len(fields) < 1 {
		return "", errors.New(`после /info имя игрока необходимо указывать в этом же соощении. Например "/info джон"`)
	}
	player, err := c.playerService.GetByName(fields[0])
	if err != nil {
		return "", err
	}
	return printPlayer(player), nil
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
