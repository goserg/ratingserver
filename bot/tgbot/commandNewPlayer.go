package tgbot

import (
	"errors"
	"strings"
	"unicode"

	mapset "github.com/deckarep/golang-set/v2"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/goserg/ratingserver/bot/model"
	"github.com/goserg/ratingserver/internal/service"
)

type NewPlayerCommand struct {
	playerService *service.PlayerService
}

func (c *NewPlayerCommand) Reset() {}

func (c *NewPlayerCommand) Run(_ model.User, args string, resp *tgbotapi.MessageConfig) (bool, error) {
	resp.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
	if args == "" {
		return false, errors.New("имя должно быть не пустое")
	}
	if strings.EqualFold(args, draw) {
		return false, errors.New("имя " + draw + " запрещено")
	}
	for i, r := range args {
		if i == 0 {
			if !unicode.IsLetter(r) {
				return false, errors.New("имя должно начинать с буквы")
			}
			continue
		}
		if !unicode.IsPrint(r) || unicode.IsSpace(r) {
			return false, errors.New("имя должно содержать только печатные символы")
		}
	}
	p, err := c.playerService.CreatePlayer(args)
	if err != nil {
		return false, err
	}
	resp.Text = "Добавлен игрок " + p.Name + " (ID " + p.ID.String() + ")"
	return false, nil
}

func (c *NewPlayerCommand) Help() string {
	return `Добавить нового игрока. Использование: /new_player <имя игрок>`
}

func (c *NewPlayerCommand) Permission() mapset.Set[model.UserRole] {
	return mapset.NewSet[model.UserRole](model.RoleAdmin, model.RoleModerator)
}

func (c *NewPlayerCommand) Visibility() mapset.Set[model.UserRole] {
	return mapset.NewSet[model.UserRole](model.RoleAdmin, model.RoleModerator)
}
