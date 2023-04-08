package tgbot

import (
	"context"
	"fmt"
	"log"
	"os"
	"ratingserver/bot/botstorage"
	botmodel "ratingserver/bot/model"
	"ratingserver/internal/domain"
	"ratingserver/internal/service"
	"strconv"
	"strings"
	"time"

	mapset "github.com/deckarep/golang-set/v2"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Bot struct {
	bot *tgbotapi.BotAPI

	botStorage botstorage.BotStorage

	playerService *service.PlayerService

	// cancel func to stop the bot
	cancel func()

	subscriptions map[botmodel.EventType]mapset.Set[int]
}

func New(ps *service.PlayerService, bs botstorage.BotStorage) (Bot, error) {
	bot, err := tgbotapi.NewBotAPI(os.Getenv("TELEGRAM_APITOKEN"))
	if err != nil {
		return Bot{}, fmt.Errorf("env TELEGRAM_APITOKEN: %w", err)
	}

	bot.Debug = true
	_, err = bot.GetMe()
	if err != nil {
		return Bot{}, err
	}
	subs := make(map[botmodel.EventType]mapset.Set[int])
	users, err := bs.ListUsers()
	if err != nil {
		return Bot{}, err
	}
	for i := range users {
		for _, subscription := range users[i].Subscriptions {
			if subs[subscription] == nil {
				subs[subscription] = mapset.NewSet[int]()
			}
			subs[subscription].Add(users[i].ID)
		}
	}
	return Bot{
		bot:           bot,
		playerService: ps,
		botStorage:    bs,
		subscriptions: subs,
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
			tgUser := update.SentFrom()
			if tgUser == nil {
				continue
			}
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
					fmt.Println("ERRRRRR", err)
					continue
				}
			}

			err = b.botStorage.Log(user, update.Message.Text)
			if err != nil {
				fmt.Println("CAN'T LOG TO DB", err)
			}

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
			case "help", "start":
				msg.Text = `Доступные команды "/top", "/help", "/status" и "/info имя".`
			case "status":
				sticker := tgbotapi.NewSticker(msg.ChatID, tgbotapi.FileID("CAACAgIAAxkBAAEIek5kLqgKrk6cRxw0uUy2CNY-VYdyBQACdxEAAjyzxQdiXqFFBrRFjy8E"))
				_, err := b.bot.Send(sticker)
				if err != nil {
					log.Println(err.Error())
					continue
				}
				continue
			case "info":
				msg.Text = b.processInfo(update.Message.CommandArguments())
			case "game":
				msg.Text = b.processAddMatch(update.Message.CommandArguments())
				for _, userIDs := range b.subscriptions {
					b.sendMatchNotification(userIDs.ToSlice(), update.Message.CommandArguments())
				}
			case "top":
				ratings, err := b.playerService.GetRatings()
				if err != nil {
					log.Println(err.Error())
					continue
				}
				var buffer strings.Builder
				for i := range ratings {
					if i > 9 {
						break
					}
					buffer.WriteString(strconv.Itoa(ratings[i].RatingRank))
					buffer.WriteString(". ")
					buffer.WriteString(ratings[i].Name)
					buffer.WriteString("(")
					buffer.WriteString(strconv.Itoa(ratings[i].EloRating))
					buffer.WriteString(")\n")
				}
				msg.Text = buffer.String()
			case "sub":
				msg.Text = "Подписка оформленна, чтобы отписаться от уведомлений: /unsub"
				err := b.botStorage.Subscribe(user)
				if err != nil {
					log.Println(err.Error())
					msg.Text = err.Error()
				}
				b.subscriptions[botmodel.NewMatch].Add(user.ID)
			case "unsub":
				msg.Text = "Подписка отменена, чтобы подписаться на уведомления: /sub"
				err := b.botStorage.Unsubscribe(user)
				if err != nil {
					log.Println(err.Error())
					msg.Text = err.Error()
				}
				b.subscriptions[botmodel.NewMatch].Remove(user.ID)
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
	buf.WriteString("Место в рейтинге: ")
	buf.WriteString(prettifyRank(player))
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

func prettifyRank(player domain.Player) string {
	if player.RatingRank == 1 {
		return "🥇"
	}
	if player.RatingRank == 2 {
		return "🥈"
	}
	if player.RatingRank == 3 {
		return "🥉"
	}
	return strconv.Itoa(player.RatingRank)
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
	switch strings.ToLower(fields[winnerIndex]) {
	case strings.ToLower(playerAName):
		newMatch.Winner = playerA
	case strings.ToLower(playerBName):
		newMatch.Winner = playerB
	}
	err = b.playerService.CreateMatch(newMatch)
	if err != nil {
		return "ошибка создания матча"
	}
	return "матч успешно создан"
}

func (b *Bot) sendMatchNotification(userIDs []int, arguments string) {
	for _, userID := range userIDs {
		msg := tgbotapi.NewMessage(int64(userID), "матч создан: "+arguments)
		if _, err := b.bot.Send(msg); err != nil {
			log.Println("BOT ERROR", err.Error())
			return
		}
	}
}
