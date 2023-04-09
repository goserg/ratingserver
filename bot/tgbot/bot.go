package tgbot

import (
	"context"
	"errors"
	"fmt"
	"log"
	"ratingserver/bot/botstorage"
	botmodel "ratingserver/bot/model"
	"ratingserver/internal/config"
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

func New(ps *service.PlayerService, bs botstorage.BotStorage, cfg config.Config) (Bot, error) {
	bot, err := tgbotapi.NewBotAPI(cfg.TgBot.TelegramApiToken)
	if err != nil {
		return Bot{}, fmt.Errorf("env TELEGRAM_APITOKEN: %w", err)
	}

	bot.Debug = cfg.Server.Debug
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
			if update.Message == nil { // ignore any non-Message updates
				continue
			}
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

			if !update.Message.IsCommand() { // ignore any non-command Messages
				continue
			}

			// Create a new MessageConfig. We don't have text yet,
			// so we leave it empty.
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "")

			// Extract the command from the Message.
			switch update.Message.Command() {
			case "help", "start":
				msg.Text = `Доступные команды /top, /help, /sub, /unsub, /status и /info.`
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
				newMatch, err := b.processAddMatch(update.Message.CommandArguments())
				msg.Text = "матч создан"
				if err != nil {
					msg.Text = err.Error()
				}
				for _, userIDs := range b.subscriptions {
					b.sendMatchNotification(userIDs.ToSlice(), newMatch)
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
		return `После /info имя игрока необходимо указывать в этом же соощении. Например "/info джон"`
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

func (b *Bot) processAddMatch(arguments string) (domain.Match, error) {
	fields := strings.Fields(arguments)
	if len(fields) < 3 {
		return domain.Match{}, errors.New(`неверный запрос. Пример: "Вася петя вася" - играли вася и петя, победил вася`)
	}
	playerAName := fields[playerAIndex]
	playerA, err := b.playerService.GetByName(playerAName)
	if err != nil {
		return domain.Match{}, errors.New(playerAName + " не найден")
	}
	playerBName := fields[playerBIndex]
	playerB, err := b.playerService.GetByName(playerBName)
	if err != nil {
		return domain.Match{}, errors.New(playerBName + " не найден")
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
	return b.playerService.CreateMatch(newMatch)
}

func (b *Bot) sendMatchNotification(userIDs []int, match domain.Match) {
	matches, err := b.playerService.GetMatches()
	if err != nil {
		log.Println("ERRRRRR", err.Error())
		return
	}
	for i := range matches {
		if matches[i].ID == match.ID {
			match := matches[i]
			for _, userID := range userIDs {
				msg := tgbotapi.NewMessage(int64(userID), formatMatchResult(match))
				if _, err := b.bot.Send(msg); err != nil {
					log.Println("BOT ERROR", err.Error())
					return
				}
			}
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
