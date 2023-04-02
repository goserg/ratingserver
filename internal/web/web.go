package web

import (
	"fmt"
	"os"
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
	engine.AddFunc("FormatDate", formatDate)

	app := fiber.New(fiber.Config{
		Views: engine,
	})
	app.Static("/", "./public")
	app.Get("/", server.handleMain)
	app.Get("/matches", server.handleMatches)
	app.Get("/export", server.HandleExport)
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
	matches, err := s.playerService.GetMatches()
	if err != nil {
		return err
	}
	fmt.Println(matches)

	return c.Render("matches", fiber.Map{
		"Matches": matches,
	}, "layouts/main")
}

func (s *Server) HandleExport(c *fiber.Ctx) error {
	fileData, err := s.playerService.Export()
	if err != nil {
		return err
	}
	f, err := os.CreateTemp("", "rating_export_*.json")
	if err != nil {
		return err
	}
	_, err = f.Write(fileData)
	if err != nil {
		return err
	}
	defer os.Remove(f.Name())
	return c.SendFile(f.Name())
}

func formatDate(t time.Time) string {
	return t.Format("02.01.2006г.")
}
