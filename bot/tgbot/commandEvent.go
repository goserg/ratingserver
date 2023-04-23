package tgbot

import (
	mapset "github.com/deckarep/golang-set/v2"
	"ratingserver/bot/model"
	"ratingserver/internal/domain"
	"ratingserver/internal/service"
	"strings"
)

type EventState int

const (
	EventStateStart = iota
	EventStateWaitForPlayers
)

type EventCommand struct {
	playerService *service.PlayerService
	state         EventState
	players       mapset.Set[domain.Player]
}

func NewEventCommand(ps *service.PlayerService) *EventCommand {
	return &EventCommand{
		playerService: ps,
		state:         EventStateStart,
		players:       mapset.NewSet[domain.Player](),
	}
}

func (c *EventCommand) Run(user model.User, text string) (string, bool, error) {
	switch c.state {
	case EventStateStart:
		c.state = EventStateWaitForPlayers
		return "start event", true, nil
	case EventStateWaitForPlayers:
		if text == "" {
			return "waiting for players names", true, nil
		}
		names := strings.Fields(text)
		for _, name := range names {
			player, err := c.playerService.GetByName(name)
			if err != nil {
				return "", false, err
			}
			c.players.Add(player)
		}
		return "event registered", true, nil
	}
	return "internal error, command aborted", false, nil
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
