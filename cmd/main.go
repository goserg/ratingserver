package main

import (
	"context"
	"log"
	"os"

	"github.com/goserg/ratingserver/auth/storage/postgres"

	authservice "github.com/goserg/ratingserver/auth/service"
	botstorage "github.com/goserg/ratingserver/bot/botstorage/sqlite"
	"github.com/goserg/ratingserver/bot/tgbot"
	"github.com/goserg/ratingserver/internal/cache/mem"
	"github.com/goserg/ratingserver/internal/config"
	"github.com/goserg/ratingserver/internal/logger"
	"github.com/goserg/ratingserver/internal/service"
	"github.com/goserg/ratingserver/internal/storage/sqlite"
	"github.com/goserg/ratingserver/internal/web"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	if err := run(); err != nil {
		log.Println(err.Error())
		os.Exit(1)
	}
}

func run() error {
	ctx := context.Background()
	cfg, err := config.New()
	if err != nil {
		return err
	}
	log := logger.New()

	storage, err := sqlite.New(log, cfg.Server)
	if err != nil {
		return err
	}

	botStorage, err := botstorage.New(log, cfg.TgBot)
	if err != nil {
		return err
	}

	playerService, err := service.New(storage, storage, mem.New())
	if err != nil {
		return err
	}

	if !cfg.Server.TgBotDisable {
		bot, err := tgbot.New(playerService, botStorage, cfg, log)
		if err != nil {
			return err
		}
		go bot.Run()
		defer bot.Stop()
	}

	authStorage, err := postgres.New(cfg.Server.Auth.Storage, cfg.Server.Auth.TokenTTL)
	if err != nil {
		return err
	}
	auth, err := authservice.New(ctx, cfg.Server.Auth, authStorage)
	if err != nil {
		return err
	}

	server, err := web.New(playerService, cfg.Server, auth)
	if err != nil {
		return err
	}
	return server.Serve()
}
