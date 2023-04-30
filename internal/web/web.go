package web

import (
	"encoding/json"
	"errors"
	"io/fs"
	"net/http"
	"os"
	embedded "ratingserver"
	authservice "ratingserver/auth/service"
	"ratingserver/internal/config"
	"ratingserver/internal/service"
	"time"

	"github.com/google/uuid"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/template/html"
)

type Server struct {
	auth          *authservice.Service
	playerService *service.PlayerService
	app           *fiber.App
}

func New(ps *service.PlayerService, cfg config.Server, auth *authservice.Service) (*Server, error) {
	server := Server{
		playerService: ps,
		auth:          auth,
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
	app.Use("/api", func(c *fiber.Ctx) error {
		tokenCookie := c.Cookies("token")
		userID, err := auth.Auth(tokenCookie)
		if err != nil {
			return c.Redirect("/login")
		}
		c.Context().SetUserValue(userIDKey, userID)
		return c.Next()
	})
	app.Get("/signin", server.HandleGetLogin)
	app.Post("/signin", server.HandlePostLogin)
	app.Get("/signup", server.HandleGetSignup)
	app.Post("/signup", server.HandlePostSignup)
	app.Get("/", func(ctx *fiber.Ctx) error {
		return ctx.Redirect("/api")
	})

	api := app.Group("/api")
	api.Get("/", server.handleMain)
	api.Get("/matches", server.handleMatches)
	api.Post("/matches", server.handleCreateMatch)
	api.Get("/export", server.HandleExport)
	api.Post("/import", server.HandleImport)
	api.Get("/players/:id", server.HandlePlayerInfo)
	server.app = app
	return &server, nil
}

func (s *Server) Serve() error {
	return s.app.Listen(":3000")
}

const userIDKey = "user-id"

func (s *Server) handleMain(ctx *fiber.Ctx) error {
	userID, ok := ctx.Context().UserValue(userIDKey).(uuid.UUID)
	if !ok {
		return errors.New("unauthorized")
	}
	globalRating := s.playerService.GetRatings()
	return ctx.Render("index", fiber.Map{
		"Button":  "rating",
		"Title":   "Рейтинг",
		"Players": globalRating,
		"UserID":  userID.String(),
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

func (s *Server) HandleGetLogin(ctx *fiber.Ctx) error {
	return ctx.Render("signin", fiber.Map{
		"Title": "Войти",
	}, "layouts/main")
}

func (s *Server) HandlePostLogin(ctx *fiber.Ctx) error {
	name := ctx.FormValue("username", "")
	password := ctx.FormValue("password", "")
	user, err := s.auth.Login(ctx.Context(), name, password)
	if err != nil {
		return err
	}
	cookie, err := s.auth.GenerateJWTCookie(user.ID)
	if err != nil {
		return err
	}
	ctx.Cookie(cookie)
	return ctx.Redirect("/")
}

func (s *Server) HandleGetSignup(ctx *fiber.Ctx) error {
	return ctx.Render("signup", fiber.Map{
		"Title": "Зарегистрироваться",
	}, "layouts/main")
}

func (s *Server) HandlePostSignup(ctx *fiber.Ctx) error {
	name := ctx.FormValue("name", "")
	password := ctx.FormValue("password", "")
	passwordRepeat := ctx.FormValue("password-repeat", "")
	if password != passwordRepeat {
		return errors.New("passwords don't match")
	}
	err := s.auth.SignUp(ctx.Context(), name, password)
	if err != nil {
		return err
	}
	return ctx.Redirect("/login")
}

func formatDate(t time.Time) string {
	return t.Format("02.01.2006г.")
}
