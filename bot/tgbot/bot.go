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

	"github.com/sirupsen/logrus"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Bot struct {
	bot *tgbotapi.BotAPI

	adminPassword string

	botStorage    botstorage.BotStorage
	playerService *service.PlayerService
	log           *logrus.Entry

	// cancel func to stop the bot
	cancel func()

	subs subscriptions
}

const draw = "ничья"

var ErrBadRequest = errors.New("bad request")

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
	return Bot{
		bot:           bot,
		playerService: ps,
		botStorage:    bs,
		subs:          subs,
		log:           log.WithField("name", "tg_bot"),
		adminPassword: cfg.TgBot.AdminPass,
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
					continue
				}
			}

			err = b.botStorage.Log(user, update.Message.Text)
			if err != nil {
				log.WithError(err).Error("Can't log to db")
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
					log.WithError(err).Error("send failed")
					continue
				}
				continue
			case "info":
				msg.Text = b.processInfo(update.Message.CommandArguments())
			case "game", "match":
				if user.Role != botmodel.RoleAdmin && user.Role != botmodel.RoleModerator {
					msg.Text = ErrBadRequest.Error()
					break
				}
				newMatch, err := b.processAddMatch(update.Message.CommandArguments())
				msg.Text = "матч создан"
				if err != nil {
					msg.Text = err.Error()
					log.WithError(err).Error("match creation failed")
				}
				b.sendMatchNotification(b.subs.GetUserIDs(botmodel.NewMatch), newMatch)
			case "top":
				ratings, err := b.playerService.GetRatings()
				if err != nil {
					log.WithError(err).Error("/top failed")
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
					log.WithError(err).Error("subscribe failed")
					msg.Text = err.Error()
				}
				b.subs.Add(botmodel.NewMatch, user.ID)
			case "unsub":
				msg.Text = "Подписка отменена, чтобы подписаться на уведомления: /sub"
				err := b.botStorage.Unsubscribe(user)
				if err != nil {
					log.WithError(err).Error("unsubscribe failed")
					msg.Text = err.Error()
				}
				b.subs.Remove(botmodel.NewMatch, user.ID)
			case "new_player":
				if user.Role != botmodel.RoleAdmin {
					msg.Text = ErrBadRequest.Error()
					break
				}
				if strings.ToLower(update.Message.CommandArguments()) == draw {
					msg.Text = "имя " + draw + " запрещено"
					break
				}
				p, err := b.playerService.CreatePlayer(update.Message.CommandArguments())
				msg.Text = "Добавлен игрок " + p.Name + " (ID " + p.ID.String() + ")"
				if err != nil {
					log.WithError(err).Error("can't create new player")
					msg.Text = err.Error()
				}
			case "role":
				text, err := b.handleRole(user, update.Message.CommandArguments())
				if err != nil {
					msg.Text = err.Error()
					break
				}
				msg.Text = text
			default:
				msg.Text = "I don't know that command"
			}

			if _, err := b.bot.Send(msg); err != nil {
				log.WithError(err).Error("send error")
				return
			}
		}

	}
}

func (b *Bot) handleRole(user botmodel.User, args string) (string, error) {
	a := strings.SplitN(args, " ", 2)
	switch a[0] {
	case "admin":
		if user.Role == botmodel.RoleAdmin {
			return "", errors.New("эта роль уже задана")
		}
		if len(a) != 2 {
			return "", ErrBadRequest
		}
		if a[1] != b.adminPassword { // wrong admin password
			return "", ErrBadRequest
		}
		user.Role = botmodel.RoleAdmin
	case "user":
		if user.Role == botmodel.RoleUser {
			return "", errors.New("эта роль уже задана")
		}
		user.Role = botmodel.RoleUser
	default:
		return "", ErrBadRequest
	}
	fmt.Println("USER: ", user)
	err := b.botStorage.UpdateUserRole(user)
	if err != nil {
		return "", err
	}
	return "role updated", nil
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
	case draw:
		newMatch.Winner = domain.Player{}
	default:
		return domain.Match{}, errors.New("winner unknown")
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
