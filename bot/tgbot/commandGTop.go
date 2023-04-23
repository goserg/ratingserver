package tgbot

import (
	mapset "github.com/deckarep/golang-set/v2"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"ratingserver/bot/model"
	"ratingserver/internal/service"
	"strconv"
	"strings"
)

type Glicko2TopCommand struct {
	playerService *service.PlayerService
}

func (c *Glicko2TopCommand) Reset() {}

func (c *Glicko2TopCommand) Run(_ model.User, _ string, resp *tgbotapi.MessageConfig) (string, bool, error) {
	resp.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
	ratings, err := c.playerService.GetGlicko2()
	if err != nil {
		return "", false, err
	}
	var buffer strings.Builder
	for i := range ratings {
		if i > 9 {
			break
		}
		buffer.WriteString(strconv.Itoa(ratings[i].RatingRank))
		buffer.WriteString(". ")
		buffer.WriteString(ratings[i].Name)
		buffer.WriteString(" - ")
		buffer.WriteString(strconv.Itoa(int(ratings[i].Glicko2Rating.Rating)))
		buffer.WriteString(" (")
		buffer.WriteString(strconv.Itoa(int(ratings[i].Glicko2Rating.Interval.Min)))
		buffer.WriteString("-")
		buffer.WriteString(strconv.Itoa(int(ratings[i].Glicko2Rating.Interval.Max)))
		buffer.WriteString(")\n")
	}
	return buffer.String(), false, nil
}

func (c *Glicko2TopCommand) Help() string {
	return `Список лучших в рейтинге Glicko2 (beta)`
}

func (c *Glicko2TopCommand) Permission() mapset.Set[model.UserRole] {
	return mapset.NewSet[model.UserRole](model.RoleAdmin, model.RoleModerator, model.RoleUser)
}
func (c *Glicko2TopCommand) Visibility() mapset.Set[model.UserRole] {
	return mapset.NewSet[model.UserRole](model.RoleAdmin, model.RoleModerator)
}
