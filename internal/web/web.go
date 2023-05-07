package web

import (
	"errors"
	"io/fs"
	"net/http"
	embedded "ratingserver"
	authservice "ratingserver/auth/service"
	"ratingserver/auth/users"
	"ratingserver/internal/config"
	"ratingserver/internal/domain"
	"ratingserver/internal/normalize"
	"ratingserver/internal/service"
	"strconv"
	"time"

	"github.com/google/uuid"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/template/html"
)

type Server struct {
	auth          *authservice.Service
	playerService *service.PlayerService
	app           *fiber.App
	cfg           config.Server
}

func New(ps *service.PlayerService, cfg config.Server, authService *authservice.Service) (*Server, error) {
	server := Server{
		playerService: ps,
		auth:          authService,
		cfg:           cfg,
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
	app.Use(api, func(c *fiber.Ctx) error {
		tokenCookie := c.Cookies("token")
		user, err := authService.Auth(c.Context(), tokenCookie, c.Method(), c.OriginalURL())
		if err != nil {
			switch err {
			case authservice.ErrForbidden:
				c.Status(fiber.StatusForbidden)
			case authservice.ErrNotAuthorized:
				c.Status(fiber.StatusUnauthorized)
			default:
				c.Status(fiber.StatusInternalServerError)
			}
			return nil
		}
		c.Context().SetUserValue(userKey, user)
		return c.Next()
	})
	app.Get(signin, server.HandleGetSignIn)
	app.Post(signin, server.HandlePostSignIn)
	app.Get(signup, server.HandleGetSignup)
	app.Post(signup, server.HandlePostSignup)
	app.Get(signout, server.HandleSignOut)
	app.Get(home, func(ctx *fiber.Ctx) error {
		return ctx.Redirect(api)
	})

	app.Get(apiHome, server.handleMain)
	app.Get(apiMatchesList, server.handleMatches)
	app.Get(apiMatches, server.handleCreateMatchGet)
	app.Post(apiMatches, server.handleCreateMatchPost)
	app.Get(apiGetPlayers, server.HandlePlayerInfo)
	app.Get(apiAddPlayer, server.handleAddPlayerGet)
	app.Post(apiAddPlayer, server.handleAddPlayerPost)
	server.app = app
	return &server, nil
}

func (s *Server) Serve() error {
	return s.app.Listen(s.cfg.Host + ":" + strconv.Itoa(s.cfg.Port))
}

const userKey = "user"

func (s *Server) handleMain(ctx *fiber.Ctx) error {
	user, _ := ctx.Context().UserValue(userKey).(users.User)
	globalRating := s.playerService.GetRatings()
	return ctx.Render("index", fiber.Map{
		"Button":  "rating",
		"Title":   "Рейтинг",
		"Players": globalRating,
		"Path":    Path(),
		"User":    user,
	}, "layouts/main")
}

func (s *Server) handleMatches(ctx *fiber.Ctx) error {
	user, _ := ctx.Context().UserValue(userKey).(users.User)
	matches, err := s.playerService.GetMatches()
	if err != nil {
		return err
	}
	return ctx.Render("matches", fiber.Map{
		"Button":  "matches",
		"Title":   "Список матчей",
		"Matches": matches,
		"Path":    Path(),
		"User":    user,
	}, "layouts/main")
}

func (s *Server) handleCreateMatchGet(ctx *fiber.Ctx) error {
	user, _ := ctx.Context().UserValue(userKey).(users.User)
	return ctx.Render("newMatch", fiber.Map{
		"Title": "Добавить игру",
		"Path":  Path(),
		"User":  user,
	}, "layouts/main")
}

func (s *Server) handleCreateMatchPost(ctx *fiber.Ctx) error {
	winner, err := s.playerService.GetByName(normalize.Name(ctx.FormValue("winner")))
	if err != nil {
		return err
	}
	loser, err := s.playerService.GetByName(normalize.Name(ctx.FormValue("loser")))
	if err != nil {
		return err
	}
	m := domain.Match{
		PlayerA: winner,
		PlayerB: loser,
		Winner:  winner,
	}
	if ctx.FormValue("draw") == "on" {
		m.Winner = domain.Player{}
	}
	_, err = s.playerService.CreateMatch(m)
	if err != nil {
		return err
	}
	err = ctx.Redirect(apiHome)
	if err != nil {
		return err
	}
	return nil
}

func (s *Server) HandlePlayerInfo(ctx *fiber.Ctx) error {
	user, _ := ctx.Context().UserValue(userKey).(users.User)
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
		"Path":   Path(),
		"User":   user,
	}, "layouts/main")
}

func (s *Server) HandleGetSignIn(ctx *fiber.Ctx) error {
	return ctx.Render("signin", fiber.Map{
		"Title": "Войти",
		"Path":  Path(),
	}, "layouts/main")
}

func (s *Server) HandlePostSignIn(ctx *fiber.Ctx) error {
	name := ctx.FormValue("username", "")
	password := ctx.FormValue("password", "")
	user, err := s.auth.Login(ctx.Context(), name, password)
	if err != nil {
		return err
	}
	cookie, err := s.auth.GenerateJWTCookie(user.ID, s.cfg.Host)
	if err != nil {
		return err
	}
	ctx.Cookie(cookie)
	return ctx.Redirect(apiHome)
}

func (s *Server) HandleGetSignup(ctx *fiber.Ctx) error {
	return ctx.Render("signup", fiber.Map{
		"Title": "Зарегистрироваться",
		"Path":  Path(),
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
	return ctx.Redirect(signin)
}

func (s *Server) HandleSignOut(ctx *fiber.Ctx) error {
	ctx.ClearCookie("token")
	return ctx.Redirect(apiHome)
}

func (s *Server) handleAddPlayerGet(ctx *fiber.Ctx) error {
	return ctx.Render("newPlayer", fiber.Map{
		"Title": "Добавить игрока",
		"Path":  Path(),
	}, "layouts/main")
}

func (s *Server) handleAddPlayerPost(ctx *fiber.Ctx) error {
	name := ctx.FormValue("name", "")
	if name == "" {
		return errors.New("empty player name")
	}
	_, err := s.playerService.CreatePlayer(name)
	if err != nil {
		return err
	}
	return ctx.Redirect(apiHome)
}

func formatDate(t time.Time) string {
	return t.Format("02.01.2006г.")
}
