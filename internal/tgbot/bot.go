package tgbot

import (
	"context"
	"fmt"
	"log"
	"os"
	"ratingserver/internal/domain"
	"ratingserver/internal/service"
	"strconv"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Bot struct {
	bot           *tgbotapi.BotAPI
	playerService *service.PlayerService
	cancel        func()
}

func New(playerService *service.PlayerService) (Bot, error) {
	bot, err := tgbotapi.NewBotAPI(os.Getenv("TELEGRAM_APITOKEN"))
	if err != nil {
		return Bot{}, fmt.Errorf("env TELEGRAM_APITOKEN: %w", err)
	}

	bot.Debug = true
	user, err := bot.GetMe()
	if err != nil {
		return Bot{}, err
	}
	fmt.Println(user)
	return Bot{
		bot:           bot,
		playerService: playerService,
	}, nil
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
			if update.Message == nil { // ignore any non-Message updates
				continue
			}

			if !update.Message.IsCommand() { // ignore any non-command Messages
				continue
			}

			// Create a new MessageConfig. We don't have text yet,
			// so we leave it empty.
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "")

			// Extract the command from the Message.
			switch {
			case update.Message.Command() == "help":
				msg.Text = `Доступные команды "/sayhi" "/status" и "/info имя".`
			case update.Message.Command() == "sayhi":
				msg.Text = "Привет :)"
			case update.Message.Command() == "status":
				msg.Text = "Всё норм."
			case strings.HasPrefix(update.Message.Command(), "info"):
				msg.Text = b.processInfo(update.Message.CommandArguments())
			default:
				msg.Text = "I don't know that command"
			}

			if _, err := b.bot.Send(msg); err != nil {
				log.Println("BOT ERROR", err.Error())
				return
			}
		}

	}
}

func (b *Bot) processInfo(command string) string {
	fields := strings.Fields(command)
	if len(fields) < 1 {
		return "Укажите имя"
	}
	player, err := b.playerService.GetByName(fields[0])
	if err != nil {
		return "Неожиданная ошибка"
	}
	return printPlayer(player)
}

func printPlayer(player domain.Player) string {
	var buf strings.Builder
	buf.WriteString("ID: ")
	buf.WriteString(player.ID.String())
	buf.WriteString("\n")
	buf.WriteString("Имя: ")
	buf.WriteString(player.Name)
	buf.WriteString("\n")
	buf.WriteString("Рейтинг: ")
	buf.WriteString(strconv.Itoa(player.EloRating))
	buf.WriteString("\n")
	buf.WriteString("Сыграно игр: ")
	buf.WriteString(strconv.Itoa(player.GamesPlayed))
	buf.WriteString("\n")
	buf.WriteString("Зарегистрирован: ")
	buf.WriteString(player.RegisteredAt.Format(time.RFC1123))
	return buf.String()
}

func (b *Bot) Stop() {
	if b.cancel != nil {
		b.cancel()
	}
}
