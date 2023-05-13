package tgbot

import (
	"log"
	"strings"
	"time"

	mapset "github.com/deckarep/golang-set/v2"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/goserg/ratingserver/bot/model"
	"github.com/goserg/ratingserver/internal/domain"
	"github.com/goserg/ratingserver/internal/service"
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
		return c.handleStateStart(resp), nil
	case EventStateWaitForPlayers:
		return c.handleStateWaitForPlayers(text, resp)
	case EventStateWinner:
		c.handleStateWinner(text, resp)
		return true, nil
	case EventStateLooser:
		return c.handleStateLoser(text, resp)
	case EventStateDraw:
		return c.handleStateDraw(text, resp)
	}
	resp.Text = "internal error, command aborted"
	return false, nil
}

func (c *EventCommand) handleStateDraw(text string, resp *tgbotapi.MessageConfig) (bool, error) {
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

func (c *EventCommand) handleStateLoser(text string, resp *tgbotapi.MessageConfig) (bool, error) {
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
}

func (c *EventCommand) handleStateWinner(text string, resp *tgbotapi.MessageConfig) {
	if text == "" {
		resp.Text = "winner:"
		return
	}
	if text == draw {
		c.state = EventStateDraw
		resp.Text = "first"
		return
	}
	c.winner = text
	c.state = EventStateLooser
	resp.Text = "loser"
}

func (c *EventCommand) handleStateWaitForPlayers(text string, resp *tgbotapi.MessageConfig) (bool, error) {
	if text == "" {
		resp.Text = "waiting for players names"
		return true, nil
	}
	names := strings.Fields(text)
	if len(names) <= 1 {
		resp.Text = "need more then 1 player"
		return true, nil
	}
	for _, name := range names {
		player, err := c.playerService.GetByName(name)
		if err != nil {
			return false, err
		}
		c.players.Add(player)
	}
	resp.ReplyMarkup = generateKeyboard(c.players)
	c.state = EventStateWinner
	resp.Text = "event registered\nwinner:"
	return true, nil
}

func (c *EventCommand) handleStateStart(resp *tgbotapi.MessageConfig) bool {
	c.state = EventStateWaitForPlayers
	resp.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
	resp.Text = "start event"
	return true
}

const rowWidth = 4

func generateKeyboard(players mapset.Set[domain.Player]) tgbotapi.ReplyKeyboardMarkup {
	keyboard := tgbotapi.NewReplyKeyboard()
	addPlayersToKeyboard(players.ToSlice(), &keyboard)
	addDrawToKeyboard(players.Cardinality(), &keyboard)
	return keyboard
}

func addPlayersToKeyboard(players []domain.Player, keyboard *tgbotapi.ReplyKeyboardMarkup) {
	for i, player := range players {
		if i%rowWidth == 0 {
			addRowToKeyboard(keyboard)
		}
		addPlayerToKeyboard(i/rowWidth, keyboard, player)
	}
}

func addRowToKeyboard(keyboard *tgbotapi.ReplyKeyboardMarkup) {
	keyboard.Keyboard = append(
		keyboard.Keyboard,
		tgbotapi.NewKeyboardButtonRow(),
	)
}

func addPlayerToKeyboard(row int, keyboard *tgbotapi.ReplyKeyboardMarkup, player domain.Player) {
	keyboard.Keyboard[row] = append(keyboard.Keyboard[row], tgbotapi.NewKeyboardButton(player.Name))
}

func addDrawToKeyboard(playersLen int, keyboard *tgbotapi.ReplyKeyboardMarkup) {
	if playersLen%rowWidth == 0 {
		keyboard.Keyboard = append(
			keyboard.Keyboard,
			tgbotapi.NewKeyboardButtonRow(),
		)
	}
	row := playersLen / rowWidth
	keyboard.Keyboard[row] = append(keyboard.Keyboard[row], tgbotapi.NewKeyboardButton(draw))
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
