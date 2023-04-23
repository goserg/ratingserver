package tgbot

import (
	mapset "github.com/deckarep/golang-set/v2"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"ratingserver/bot/model"
	"ratingserver/internal/service"
	"strconv"
	"strings"
)

type TopCommand struct {
	playerService *service.PlayerService
}

func (c *TopCommand) Reset() {}

func (c *TopCommand) Run(_ model.User, _ string, resp *tgbotapi.MessageConfig) (string, bool, error) {
	resp.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
	ratings, err := c.playerService.GetRatings()
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
		buffer.WriteString("(")
		buffer.WriteString(strconv.Itoa(ratings[i].EloRating))
		buffer.WriteString(")\n")
	}
	return buffer.String(), false, nil
}

func (c *TopCommand) Help() string {
	return `Список лучших в рейтинге`
}

func (c *TopCommand) Permission() mapset.Set[model.UserRole] {
	return mapset.NewSet[model.UserRole](model.RoleAdmin, model.RoleModerator, model.RoleUser)
}
func (c *TopCommand) Visibility() mapset.Set[model.UserRole] {
	return mapset.NewSet[model.UserRole](model.RoleAdmin, model.RoleModerator, model.RoleUser)
}
