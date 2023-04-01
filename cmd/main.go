package main

import (
	"fmt"
	"os"
	"ratingserver/internal/service"
	"ratingserver/internal/storage"
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
	db, err := storage.New()
	if err != nil {
		return err
	}
	playerService := service.New(db)
	server := web.New(playerService)
	return server.Serve()
}
