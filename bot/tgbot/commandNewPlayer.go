package tgbot

import (
	"errors"
	mapset "github.com/deckarep/golang-set/v2"
	"ratingserver/bot/model"
	"ratingserver/internal/service"
	"strings"
	"unicode"
)

type NewPlayerCommand struct {
	playerService *service.PlayerService
}

func (c *NewPlayerCommand) Run(_ model.User, args string) (string, bool, error) {
	if len(args) == 0 {
		return "", false, errors.New("имя должно быть не пустое")
	}
	if strings.ToLower(args) == draw {
		return "", false, errors.New("имя " + draw + " запрещено")
	}
	for i, r := range args {
		if i == 0 {
			if !unicode.IsLetter(r) {
				return "", false, errors.New("имя должно начинать с буквы")
			}
			continue
		}
		if !unicode.IsPrint(r) || unicode.IsSpace(r) {
			return "", false, errors.New("имя должно содержать только печатные символы")
		}
	}
	p, err := c.playerService.CreatePlayer(args)
	if err != nil {
		return "", false, err
	}
	return "Добавлен игрок " + p.Name + " (ID " + p.ID.String() + ")", false, nil
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
