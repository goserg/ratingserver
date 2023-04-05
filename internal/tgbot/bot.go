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
			switch update.Message.Command() {
			case "help":
				msg.Text = `Доступные команды "/sayhi" "/status" и "/info имя".`
			case "sayhi":
				msg.Text = "Привет :)"
			case "status":
				msg.Text = "Всё норм."
			case "info":
				msg.Text = b.processInfo(update.Message.CommandArguments())
			case "game":
				msg.Text = b.processAddMatch(update.Message.CommandArguments())
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

const (
	playerAIndex int = iota
	playerBIndex
	winnerIndex
)

func (b *Bot) processAddMatch(arguments string) string {
	fields := strings.Fields(arguments)
	if len(fields) < 3 {
		return `Неверный запрос. Пример: "Вася петя вася" - играли вася и петя, победил вася`
	}
	playerAName := fields[playerAIndex]
	playerA, err := b.playerService.GetByName(playerAName)
	if err != nil {
		return playerAName + " не найден"
	}
	playerBName := fields[playerBIndex]
	playerB, err := b.playerService.GetByName(playerBName)
	if err != nil {
		return playerBName + " не найден"
	}

	newMatch := domain.Match{
		PlayerA: playerA,
		PlayerB: playerB,
		Date:    time.Now(),
	}
	switch fields[winnerIndex] {
	case playerAName:
		newMatch.Winner = playerA
	case playerBName:
		newMatch.Winner = playerB
	}
	err = b.playerService.CreateMatch(newMatch)
	if err != nil {
		return "ошибка создания матча"
	}
	return "матч успешно создан"
}
