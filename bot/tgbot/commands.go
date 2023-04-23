package tgbot

import (
	mapset "github.com/deckarep/golang-set/v2"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"ratingserver/bot/botstorage"
	"ratingserver/bot/model"
	"ratingserver/internal/service"
)

type Command interface {
	Run(user model.User, args string, resp *tgbotapi.MessageConfig) (string, bool, error)
	Help() string
	Permission() mapset.Set[model.UserRole]
	Visibility() mapset.Set[model.UserRole]

	Reset()
}

type Commands struct {
	list               map[string]Command
	userCurrentCommand map[int]Command
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
			"event": NewEventCommand(ps, sendNotifFn),
		},
		userCurrentCommand: make(map[int]Command),
	}
	hc.commands = uc.list
	return &uc
}

func (uc *Commands) RunCommand(
	user model.User,
	msg *tgbotapi.Message,
	resp *tgbotapi.MessageConfig,
) error {
	for s, command := range uc.list {
		if msg.Command() == s {
			if command.Permission().Contains(user.Role) {
				command.Reset()
				text, needContinue, err := command.Run(user, msg.CommandArguments(), resp)
				if err != nil {
					return err
				}
				if needContinue {
					uc.userCurrentCommand[user.ID] = command
				} else {
					uc.userCurrentCommand[user.ID] = nil
				}
				resp.Text = text
				return nil
			}
		}
	}
	command := uc.userCurrentCommand[user.ID]
	if command != nil {
		text, needContinue, err := command.Run(user, msg.Text, resp)
		if err != nil {
			return err
		}
		if needContinue {
			uc.userCurrentCommand[user.ID] = command
		} else {
			uc.userCurrentCommand[user.ID] = nil
		}
		resp.Text = text
		return nil
	}
	return ErrBadRequest
}
