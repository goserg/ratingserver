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
	"ratingserver/internal/web/webpath"
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
	app.Use(webpath.Api, func(c *fiber.Ctx) error {
		tokenCookie := c.Cookies("token")
		user, err := authService.Auth(c.Context(), tokenCookie, c.Method(), c.OriginalURL())
		if err != nil {
			switch {
			case errors.Is(err, authservice.ErrForbidden):
				c.Status(fiber.StatusForbidden)
			case errors.Is(err, authservice.ErrNotAuthorized):
				c.Status(fiber.StatusUnauthorized)
			default:
				c.Status(fiber.StatusInternalServerError)
			}
			return nil
		}
		c.Context().SetUserValue(userKey, user)
		return c.Next()
	})
	app.Get(webpath.Signin, server.HandleGetSignIn)
	app.Post(webpath.Signin, server.HandlePostSignIn)
	app.Get(webpath.Signup, server.HandleGetSignup)
	app.Post(webpath.Signup, server.HandlePostSignup)
	app.Get(webpath.Signout, server.HandleSignOut)
	app.Get(webpath.Home, func(ctx *fiber.Ctx) error {
		return ctx.Redirect(webpath.Api)
	})

	app.Get(webpath.ApiHome, server.handleMain)
	app.Get(webpath.ApiMatchesList, server.handleMatches)
	app.Get(webpath.ApiNewMatch, server.handleCreateMatchGet)
	app.Post(webpath.ApiNewMatch, server.handleCreateMatchPost)
	app.Get(webpath.ApiGetPlayers, server.HandlePlayerInfo)
	app.Get(webpath.ApiNewPlayer, server.handleNewPlayerGet)
	app.Post(webpath.ApiNewPlayer, server.handleNewPlayerPost)
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
		"Path":    webpath.Path(),
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
		"Path":    webpath.Path(),
		"User":    user,
	}, "layouts/main")
}

func (s *Server) handleCreateMatchGet(ctx *fiber.Ctx) error {
	user, _ := ctx.Context().UserValue(userKey).(users.User)
	return ctx.Render("newMatch", fiber.Map{
		"Title": "Добавить игру",
		"Path":  webpath.Path(),
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
	err = ctx.Redirect(webpath.ApiHome)
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
		"Path":   webpath.Path(),
		"User":   user,
	}, "layouts/main")
}

func (s *Server) HandleGetSignIn(ctx *fiber.Ctx) error {
	return ctx.Render("signin", fiber.Map{
		"Title": "Войти",
		"Path":  webpath.Path(),
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
	return ctx.Redirect(webpath.ApiHome)
}

func (s *Server) HandleGetSignup(ctx *fiber.Ctx) error {
	return ctx.Render("signup", fiber.Map{
		"Title": "Зарегистрироваться",
		"Path":  webpath.Path(),
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
	return ctx.Redirect(webpath.Signin)
}

func (s *Server) HandleSignOut(ctx *fiber.Ctx) error {
	ctx.ClearCookie("token")
	return ctx.Redirect(webpath.ApiHome)
}

func (s *Server) handleNewPlayerGet(ctx *fiber.Ctx) error {
	return ctx.Render("newPlayer", fiber.Map{
		"Title": "Добавить игрока",
		"Path":  webpath.Path(),
	}, "layouts/main")
}

func (s *Server) handleNewPlayerPost(ctx *fiber.Ctx) error {
	name := ctx.FormValue("name", "")
	if name == "" {
		return errors.New("empty player name")
	}
	_, err := s.playerService.CreatePlayer(name)
	if err != nil {
		return err
	}
	return ctx.Redirect(webpath.ApiHome)
}

func formatDate(t time.Time) string {
	return t.Format("02.01.2006г.")
}
