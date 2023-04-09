package main

import (
	"fmt"
	"os"
	botstorage "ratingserver/bot/botstorage/sqlite"
	"ratingserver/bot/tgbot"
	"ratingserver/internal/config"
	"ratingserver/internal/logger"
	"ratingserver/internal/service"
	"ratingserver/internal/storage/sqlite"
	"ratingserver/internal/web"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	if err := run(); err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.New()
	if err != nil {
		return err
	}
	log := logger.New()

	storage, err := sqlite.New(log)
	if err != nil {
		return err
	}

	botStorage, err := botstorage.New(log)
	if err != nil {
		return err
	}

	playerService := service.New(storage, storage)

	if cfg.Server.TgBotEnabled {
		bot, err := tgbot.New(playerService, botStorage, cfg.TgBot)
		if err != nil {
			return err
		}
		go bot.Run()
		defer bot.Stop()
	}

	server := web.New(playerService)
	return server.Serve()
}
