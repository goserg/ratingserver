package web

import (
	"ratingserver/internal/service"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/template/html"
)

type Server struct {
	playerService *service.PlayerService
	app           *fiber.App
}

func New(ps *service.PlayerService) *Server {
	server := Server{
		playerService: ps,
	}

	engine := html.New("./views", ".html")
	engine.Reload(true)
	engine.Debug(true)

	app := fiber.New(fiber.Config{
		Views: engine,
	})
	app.Static("/", "./public")
	app.Get("/", server.handleMain)
	app.Get("/matches", server.handleMatches)
	server.app = app
	return &server
}

func (s *Server) Serve() error {
	return s.app.Listen(":3000")
}

func (s *Server) handleMain(c *fiber.Ctx) error {
	globalRating, err := s.playerService.GetRatings()
	if err != nil {
		return err
	}
	return c.Render("index", fiber.Map{
		"Players": globalRating,
	}, "layouts/main")
}

func (s *Server) handleMatches(c *fiber.Ctx) error {
	matches := []struct {
		PlayerA    string
		PlayerAWin bool
		PlayerB    string
		Date       string
	}{
		{
			PlayerA:    "Lol",
			PlayerAWin: true,
			PlayerB:    "kek",
			Date:       time.Now().Format("02.01.2006г"),
		},
		{
			PlayerA:    "Lol",
			PlayerAWin: false,
			PlayerB:    "kek",
			Date:       time.Now().Add(time.Hour * 24).Format("02.01.2006г"),
		},
	}
	return c.Render("matches", fiber.Map{
		"Matches": matches,
	}, "layouts/main")
}
