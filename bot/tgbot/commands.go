package tgbot

import (
	mapset "github.com/deckarep/golang-set/v2"
	"ratingserver/bot/botstorage"
	"ratingserver/bot/model"
	"ratingserver/internal/service"
)

type Command interface {
	Run(user model.User, args string) (string, error)
	Help() string
	Permission() mapset.Set[model.UserRole]
	Visibility() mapset.Set[model.UserRole]
}

type Commands struct {
	list map[string]Command
}

func NewCommands(
	ps *service.PlayerService,
	bs botstorage.BotStorage,
	adminPass string,
	subFn func(id int),
	unsubFn func(id int),
	sendNotifFn func(msg string),
) *Commands {
	hc := &HelpCommand{}
	uc := Commands{
		list: map[string]Command{
			"help":  hc,
			"start": hc,
			"top": &TopCommand{
				playerService: ps,
			},
			"gtop": &Glicko2TopCommand{
				playerService: ps,
			},
			"me": &MeCommand{
				playerService: ps,
				botStorage:    bs,
			},
			"info": &InfoCommand{
				playerService: ps,
			},
			"role": &RoleCommand{
				adminPassword: adminPass,
				botStorage:    bs,
			},
			"game": &NewGameCommand{
				playerService: ps,
				notify:        sendNotifFn,
			},
			"new_player": &NewPlayerCommand{
				playerService: ps,
			},
			"sub": &SubCommand{
				botStorage: bs,
				sub:        subFn,
			},
			"unsub": &UnsubCommand{
				botStorage: bs,
				unsub:      unsubFn,
			},
		},
	}
	hc.commands = uc.list
	return &uc
}

func (uc *Commands) RunCommand(user model.User, cmd string, args string) (string, error) {
	for s, command := range uc.list {
		if cmd == s {
			if command.Permission().Contains(user.Role) {
				return command.Run(user, args)
			}
		}
	}
	return "", ErrBadRequest
}
