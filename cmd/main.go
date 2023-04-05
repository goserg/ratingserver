package main

import (
	"fmt"
	"os"
	"ratingserver/internal/service"
	"ratingserver/internal/storage/sqlite"
	"ratingserver/internal/tgbot"
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
	storage, err := sqlite.New()
	if err != nil {
		return err
	}

	playerService := service.New(storage, storage)

	bot, err := tgbot.New(playerService)
	if err != nil {
		return err
	}
	go bot.Run()
	defer bot.Stop()

	server := web.New(playerService)
	return server.Serve()
}
