package tgbot

import (
	"context"
	"errors"
	"fmt"
	"github.com/sirupsen/logrus"
	"log"
	"ratingserver/bot/botstorage"
	botmodel "ratingserver/bot/model"
	"ratingserver/internal/config"
	"ratingserver/internal/service"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Bot struct {
	bot *tgbotapi.BotAPI

	botStorage botstorage.BotStorage
	log        *logrus.Entry

	// cancel func to stop the bot
	cancel func()

	subs subscriptions

	commands *Commands
}

const draw = "ничья"

var ErrBadRequest = errors.New("неизвестная команда")

func New(ps *service.PlayerService, bs botstorage.BotStorage, cfg config.Config, log *logrus.Logger) (Bot, error) {
	bot, err := tgbotapi.NewBotAPI(cfg.TgBot.TelegramApiToken)
	if err != nil {
		return Bot{}, fmt.Errorf("env TELEGRAM_APITOKEN: %w", err)
	}

	bot.Debug = cfg.Server.Debug
	_, err = bot.GetMe()
	if err != nil {
		return Bot{}, err
	}
	subs := newSubs()
	users, err := bs.ListUsers()
	if err != nil {
		return Bot{}, err
	}
	for i := range users {
		for _, subType := range users[i].Subscriptions {
			subs.Add(subType, users[i].ID)
		}
	}

	b := Bot{
		bot:        bot,
		botStorage: bs,
		log:        log.WithField("name", "tg_bot"),
		subs:       subs,
	}

	b.commands = NewCommands(
		ps,
		bs,
		cfg.TgBot.AdminPass,
		func(id int) {
			b.subs.Add(botmodel.NewMatch, id)
		},
		func(id int) {
			b.subs.Remove(botmodel.NewMatch, id)
		},
		func(msg string) {
			b.sendMatchNotification(botmodel.NewMatch, msg)
		},
	)

	return b, nil
}

func (b *Bot) Run() {
	ctx, cancel := context.WithCancel(context.Background())
	b.cancel = cancel

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := b.bot.GetUpdatesChan(u)

	for {
		select {
		case <-ctx.Done():
			return
		case update := <-updates:
			b.handleMessage(update)
		}

	}
}

func (b *Bot) handleMessage(update tgbotapi.Update) {
	if update.Message == nil { // ignore any non-Message updates
		return
	}
	tgUser := update.SentFrom()
	if tgUser == nil {
		return
	}
	log := b.log.WithFields(map[string]interface{}{
		"user_id": tgUser.ID,
		"text":    update.Message.Text,
	})
	user, err := b.botStorage.GetUser(int(tgUser.ID))
	if err != nil {
		user, err = b.botStorage.NewUser(botmodel.User{
			ID:        int(tgUser.ID),
			FirstName: tgUser.FirstName,
			Username:  tgUser.UserName,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		})
		if err != nil {
			log.WithError(err).Error("unable to get user from db")
			return
		}
	}

	err = b.botStorage.Log(user, update.Message.Text)
	if err != nil {
		log.WithError(err).Error("Can't log to db")
	}

	// Create a new MessageConfig. We don't have text yet,
	// so we leave it empty.
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, "")

	err = b.commands.RunCommand(user, update.Message, &msg)
	if err != nil {
		msg.Text = err.Error()
	}
	if _, err := b.bot.Send(msg); err != nil {
		log.WithError(err).Error("send error")
		return
	}
}

func (b *Bot) Stop() {
	if b.cancel != nil {
		b.cancel()
	}
}

func (b *Bot) sendMatchNotification(event botmodel.EventType, text string) {
	for _, userID := range b.subs.GetUserIDs(event) {
		msg := tgbotapi.NewMessage(int64(userID), text)
		if _, err := b.bot.Send(msg); err != nil {
			log.Println("BOT ERROR", err.Error())
			return
		}
	}
}
