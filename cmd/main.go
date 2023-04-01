package main

import (
	"database/sql"
	"fmt"
	"os"
	"ratingserver/internal/service"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/template/html"
	_ "github.com/mattn/go-sqlite3"
)

func main() {
	if err := run(); err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
}

func run() error {
	db, err := sql.Open("sqlite3", "file:rating.sqlite?cache=shared")
	if err != nil {
		return err
	}
	db.SetMaxOpenConns(1)
	err = db.Ping()
	if err != nil {
		return err
	}
	playerService := service.New(db)
	players, err := playerService.List()
	if err != nil {
		return err
	}
	globalRating, err := playerService.GetRatings()
	if err != nil {
		return err
	}
	fmt.Println(globalRating)

	engine := html.New("./views", ".html")
	engine.Reload(true)
	engine.Debug(true)

	app := fiber.New(fiber.Config{
		Views: engine,
	})
	app.Static("/", "./public")
	app.Get("/", func(c *fiber.Ctx) error {
		return c.Render("index", fiber.Map{
			"Title":   "Hellow, World!",
			"Players": players,
		}, "layouts/main")
	})

	err = app.Listen(":3000")
	if err != nil {
		return err
	}
	return nil
}
