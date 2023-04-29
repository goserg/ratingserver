package web

import (
	"encoding/json"
	"io/fs"
	"net/http"
	"os"
	embedded "ratingserver"
	"ratingserver/internal/config"
	"ratingserver/internal/service"
	"time"

	"github.com/google/uuid"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/template/html"
)

type Server struct {
	playerService *service.PlayerService
	app           *fiber.App
}

func New(ps *service.PlayerService, cfg config.Server) (*Server, error) {
	server := Server{
		playerService: ps,
	}

	fsFS, err := fs.Sub(embedded.Views, "views")
	if err != nil {
		return nil, err
	}
	engine := html.NewFileSystem(http.FS(fsFS), ".html")
	engine.Reload(cfg.Debug)
	engine.Debug(cfg.Debug)
	engine.AddFunc("FormatDate", formatDate)

	app := fiber.New(fiber.Config{
		Views: engine,
	})
	app.Get("/", server.handleMain)
	app.Get("/matches", server.handleMatches)
	app.Post("/matches", server.handleCreateMatch)
	app.Get("/export", server.HandleExport)
	app.Post("/import", server.HandleImport)
	app.Get("/players/:id", server.HandlePlayerInfo)
	server.app = app
	return &server, nil
}

func (s *Server) Serve() error {
	return s.app.Listen(":3000")
}

func (s *Server) handleMain(ctx *fiber.Ctx) error {
	globalRating := s.playerService.GetRatings()
	return ctx.Render("index", fiber.Map{
		"Button":  "rating",
		"Title":   "Рейтинг",
		"Players": globalRating,
	}, "layouts/main")
}

func (s *Server) handleMatches(ctx *fiber.Ctx) error {
	matches, err := s.playerService.GetMatches()
	if err != nil {
		return err
	}

	return ctx.Render("matches", fiber.Map{
		"Button":  "matches",
		"Title":   "Список матчей",
		"Matches": matches,
	}, "layouts/main")
}

func (s *Server) HandleExport(ctx *fiber.Ctx) error {
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
	return ctx.SendFile(f.Name())
}

func (s *Server) HandleImport(ctx *fiber.Ctx) error {
	err := s.playerService.Import(ctx.Body())
	if err != nil {
		return err
	}
	return nil
}

func (s *Server) handleCreateMatch(ctx *fiber.Ctx) error {
	var newMatch createMatch
	err := json.Unmarshal(ctx.Body(), &newMatch)
	if err != nil {
		return err
	}
	err = newMatch.Validate()
	if err != nil {
		return err
	}
	dMatch := newMatch.convertToDomainMatch()
	_, err = s.playerService.CreateMatch(dMatch)
	if err != nil {
		return err
	}
	err = ctx.Redirect("/")
	if err != nil {
		return err
	}
	return nil
}

func (s *Server) HandlePlayerInfo(ctx *fiber.Ctx) error {
	strID := ctx.Params("id")
	id, err := uuid.Parse(strID)
	if err != nil {
		return err
	}
	data, err := s.playerService.GetPlayerData(id)
	if err != nil {
		return err
	}
	return ctx.Render("playerCard", fiber.Map{
		"Button": "playerCard",
		"Title":  data.Player.Name,
		"Data":   data,
	}, "layouts/main")
}

func formatDate(t time.Time) string {
	return t.Format("02.01.2006г.")
}
