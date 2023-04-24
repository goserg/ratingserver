package tgbot

import (
	"errors"
	mapset "github.com/deckarep/golang-set/v2"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
	"ratingserver/bot/model"
	"ratingserver/internal/domain"
	"ratingserver/internal/normalize"
	"ratingserver/internal/service"
	"strconv"
	"strings"
	"time"
)

type NewGameCommand struct {
	playerService *service.PlayerService
	notify        func(msg string)
}

func (c *NewGameCommand) Reset() {}

func (c *NewGameCommand) Run(_ model.User, args string, resp *tgbotapi.MessageConfig) (bool, error) {
	resp.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
	match, err := c.processAddMatch(args)
	if err != nil {
		return false, err
	}
	c.sendMatchNotification(match)
	resp.Text = "матч создан"
	return false, nil
}

func (c *NewGameCommand) Help() string {
	return `Добавить игру. Использование: /game <игрок1> <игрок1> <победитель / "ничья">`
}

func (c *NewGameCommand) Permission() mapset.Set[model.UserRole] {
	return mapset.NewSet[model.UserRole](model.RoleAdmin, model.RoleModerator)
}
func (c *NewGameCommand) Visibility() mapset.Set[model.UserRole] {
	return mapset.NewSet[model.UserRole](model.RoleAdmin, model.RoleModerator)
}

const (
	playerAIndex int = iota
	playerBIndex
	winnerIndex
)

func (c *NewGameCommand) processAddMatch(arguments string) (domain.Match, error) {
	fields := strings.Fields(arguments)
	if len(fields) < 3 {
		return domain.Match{}, errors.New(`неверный запрос. Пример: "Вася петя вася" - играли вася и петя, победил вася`)
	}
	playerAName := fields[playerAIndex]
	playerA, err := c.playerService.GetByName(playerAName)
	if err != nil {
		return domain.Match{}, errors.New(playerAName + " не найден")
	}
	playerBName := fields[playerBIndex]
	playerB, err := c.playerService.GetByName(playerBName)
	if err != nil {
		return domain.Match{}, errors.New(playerBName + " не найден")
	}

	newMatch := domain.Match{
		PlayerA: playerA,
		PlayerB: playerB,
		Date:    time.Now(),
	}
	switch normalize.Name(fields[winnerIndex]) {
	case normalize.Name(playerAName):
		newMatch.Winner = playerA
	case normalize.Name(playerBName):
		newMatch.Winner = playerB
	case draw:
		newMatch.Winner = domain.Player{}
	default:
		return domain.Match{}, errors.New("winner unknown")
	}
	return c.playerService.CreateMatch(newMatch)
}

func (c *NewGameCommand) sendMatchNotification(match domain.Match) {
	matches, err := c.playerService.GetMatches()
	if err != nil {
		log.Println("ERRRRRR", err.Error())
		return
	}
	for i := range matches {
		if matches[i].ID == match.ID {
			match := matches[i]
			c.notify(formatMatchResult(match))
			return
		}
	}
}

func formatMatchResult(match domain.Match) string {
	var buf strings.Builder
	if match.Winner.ID == match.PlayerA.ID {
		buf.WriteString("🏆")
	} else if match.Winner.ID == match.PlayerB.ID {
		buf.WriteString("😖")
	}
	buf.WriteString(match.PlayerA.Name)
	buf.WriteString(" vs ")
	buf.WriteString(match.PlayerB.Name)
	if match.Winner.ID == match.PlayerB.ID {
		buf.WriteString("🏆")
	} else if match.Winner.ID == match.PlayerA.ID {
		buf.WriteString("😖")
	}
	buf.WriteString("\n")
	if match.Winner.ID != match.PlayerA.ID && match.Winner.ID != match.PlayerB.ID {
		buf.WriteString("Ничья\n")
	}
	buf.WriteString("Рейтинг:\n")

	buf.WriteString(match.PlayerA.Name)
	buf.WriteString(": ")
	buf.WriteString(strconv.Itoa(match.PlayerA.EloRating))
	buf.WriteString("(")
	buf.WriteString(strconv.Itoa(match.PlayerA.RatingChange))
	buf.WriteString(")\n")
	buf.WriteString(match.PlayerB.Name)
	buf.WriteString(": ")
	buf.WriteString(strconv.Itoa(match.PlayerB.EloRating))
	buf.WriteString("(")
	buf.WriteString(strconv.Itoa(match.PlayerB.RatingChange))
	buf.WriteString(")\n")

	return buf.String()
}
