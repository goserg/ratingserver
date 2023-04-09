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

	if !cfg.Server.TgBotDisable {
		bot, err := tgbot.New(playerService, botStorage, cfg)
		if err != nil {
			return err
		}
		go bot.Run()
		defer bot.Stop()
	}

	server, err := web.New(playerService, cfg.Server)
	if err != nil {
		return err
	}
	return server.Serve()
}
