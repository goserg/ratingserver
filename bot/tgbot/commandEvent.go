package tgbot

import (
	mapset "github.com/deckarep/golang-set/v2"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
	"ratingserver/bot/model"
	"ratingserver/internal/domain"
	"ratingserver/internal/service"
	"strings"
	"time"
)

type EventState int

const (
	EventStateStart = iota
	EventStateWaitForPlayers
	EventStateWinner
	EventStateLooser
	EventStateDraw
)

type EventCommand struct {
	playerService *service.PlayerService
	state         EventState
	players       mapset.Set[domain.Player]
	winner        string
	notify        func(msg string)
}

func NewEventCommand(ps *service.PlayerService, notify func(msg string)) *EventCommand {
	return &EventCommand{
		playerService: ps,
		state:         EventStateStart,
		players:       mapset.NewSet[domain.Player](),
		notify:        notify,
	}
}

func (c *EventCommand) Reset() {
	c.state = EventStateStart
	c.players = mapset.NewSet[domain.Player]()
	c.winner = ""
}

func (c *EventCommand) Run(
	_ model.User,
	text string,
	resp *tgbotapi.MessageConfig,
) (needContinue bool, err error) {
	defer func() {
		if err != nil {
			c.Reset()
		}
	}()
	switch c.state {
	case EventStateStart:
		c.state = EventStateWaitForPlayers
		resp.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
		resp.Text = "start event"
		return true, nil
	case EventStateWaitForPlayers:
		if text == "" {
			resp.Text = "waiting for players names"
			return true, nil
		}
		names := strings.Fields(text)
		for _, name := range names {
			player, err := c.playerService.GetByName(name)
			if err != nil {
				return false, err
			}
			c.players.Add(player)
		}
		keyboard := tgbotapi.NewReplyKeyboard()
		for i, player := range c.players.ToSlice() {
			d := i % 3
			if d == 0 {
				keyboard.Keyboard = append(
					keyboard.Keyboard,
					tgbotapi.NewKeyboardButtonRow(
						tgbotapi.NewKeyboardButton(player.Name),
					),
				)
				continue
			}
			row := i / 3
			keyboard.Keyboard[row] = append(keyboard.Keyboard[row], tgbotapi.NewKeyboardButton(player.Name))
		}
		if c.players.Cardinality()%3 == 0 {
			keyboard.Keyboard = append(
				keyboard.Keyboard,
				tgbotapi.NewKeyboardButtonRow(
					tgbotapi.NewKeyboardButton(draw),
				),
			)
		} else {
			row := c.players.Cardinality() / 3
			keyboard.Keyboard[row] = append(keyboard.Keyboard[row], tgbotapi.NewKeyboardButton(draw))
		}
		resp.ReplyMarkup = keyboard
		c.state = EventStateWinner
		resp.Text = "event registered\nwinner:"
		return true, nil
	case EventStateWinner:
		if text == "" {
			resp.Text = "winner:"
			return true, nil
		}
		if text == draw {
			c.state = EventStateDraw
			resp.Text = "first"
			return true, nil
		}
		c.winner = text
		c.state = EventStateLooser
		resp.Text = "loser"
		return true, nil
	case EventStateLooser:
		if text == "" {
			resp.Text = "loser:"
			return true, nil
		}
		if text == draw {
			c.state = EventStateDraw
			resp.Text = "second:"
			return true, nil
		}
		winner, err := c.playerService.GetByName(c.winner)
		if err != nil {
			return true, err
		}
		loser, err := c.playerService.GetByName(text)
		if err != nil {
			return false, err
		}
		match, err := c.playerService.CreateMatch(domain.Match{
			PlayerA: winner,
			PlayerB: loser,
			Winner:  winner,
			Date:    time.Now(),
		})
		if err != nil {
			return true, err
		}
		c.sendMatchNotification(match)
		c.state = EventStateWinner
		c.winner = ""
		resp.Text = "match registered\nwinner:"
		return true, nil
	case EventStateDraw:
		if c.winner == "" {
			c.winner = text
			resp.Text = "second:"
			return true, nil
		}
		winner, err := c.playerService.GetByName(c.winner)
		if err != nil {
			return true, err
		}
		loser, err := c.playerService.GetByName(text)
		if err != nil {
			return false, err
		}
		match, err := c.playerService.CreateMatch(domain.Match{
			PlayerA: winner,
			PlayerB: loser,
			Winner:  domain.Player{},
			Date:    time.Now(),
		})
		if err != nil {
			return true, err
		}
		c.sendMatchNotification(match)
		c.state = EventStateWinner
		c.winner = ""
		resp.Text = "draw registered\nwinner:"
		return true, nil
	}
	resp.Text = "internal error, command aborted"
	return false, nil
}

func (c *EventCommand) Help() string {
	return `Управление событием`
}

func (c *EventCommand) Permission() mapset.Set[model.UserRole] {
	return mapset.NewSet[model.UserRole](model.RoleAdmin, model.RoleModerator)
}
func (c *EventCommand) Visibility() mapset.Set[model.UserRole] {
	return mapset.NewSet[model.UserRole](model.RoleAdmin, model.RoleModerator)
}

func (c *EventCommand) sendMatchNotification(match domain.Match) {
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
