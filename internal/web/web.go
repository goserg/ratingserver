package web

import (
	"errors"
	"io/fs"
	"net/http"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/template/html"
	"github.com/google/uuid"
	embedded "github.com/goserg/ratingserver"
	authservice "github.com/goserg/ratingserver/auth/service"
	"github.com/goserg/ratingserver/auth/users"
	"github.com/goserg/ratingserver/internal/config"
	"github.com/goserg/ratingserver/internal/domain"
	"github.com/goserg/ratingserver/internal/normalize"
	"github.com/goserg/ratingserver/internal/service"
	"github.com/goserg/ratingserver/internal/web/webpath"
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
	app.Static("/", "./static")
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
	app.Get(webpath.Signin, server.handleGetSignIn)
	app.Post(webpath.Signin, server.handlePostSignIn)
	app.Get(webpath.Signup, server.handleGetSignup)
	app.Post(webpath.Signup, server.handlePostSignup)
	app.Get(webpath.Signout, server.handleSignOut)
	app.Get(webpath.Home, func(ctx *fiber.Ctx) error {
		return ctx.Redirect(webpath.Api)
	})

	app.Get(webpath.ApiHome, server.handleMain)
	app.Get(webpath.ApiMatchesList, server.handleMatches)
	app.Get(webpath.ApiNewMatch, server.handleCreateMatchGet)
	app.Post(webpath.ApiNewMatch, server.handleCreateMatchPost)
	app.Get(webpath.ApiGetPlayers, server.handlePlayerInfo)
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
	user, ok := ctx.Context().UserValue(userKey).(users.User)
	if !ok {
		return errors.New("assertion failed")
	}
	globalRating := s.playerService.GetRatings()
	return ctx.Render("index", newData("Рейтинг").
		WithUser(user).
		With("Button", "rating").
		With("Players", globalRating),
		"layouts/main",
	)
}

func (s *Server) handleMatches(ctx *fiber.Ctx) error {
	user, ok := ctx.Context().UserValue(userKey).(users.User)
	if !ok {
		return errors.New("assertion failed")
	}
	matches, err := s.playerService.GetMatches()
	if err != nil {
		return err
	}
	return ctx.Render("matches",
		newData("Список матчей").
			WithUser(user).
			With("Button", "matches").
			With("Matches", matches),
		"layouts/main")
}

func (s *Server) handleCreateMatchGet(ctx *fiber.Ctx) error {
	user, ok := ctx.Context().UserValue(userKey).(users.User)
	if !ok {
		return errors.New("assertion failed")
	}
	return ctx.Render("newMatch", newData("Добавить игру").WithUser(user), "layouts/main")
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

func (s *Server) handlePlayerInfo(ctx *fiber.Ctx) error {
	user, ok := ctx.Context().UserValue(userKey).(users.User)
	if !ok {
		return errors.New("assertion failed")
	}
	strID := ctx.Params("id")
	id, err := uuid.Parse(strID)
	if err != nil {
		return err
	}
	card, err := s.playerService.GetPlayerData(id)
	if err != nil {
		return err
	}
	return ctx.Render("playerCard",
		newData(card.Player.Name).
			WithUser(user).
			With("PlayerCard", card).
			With("Button", "playerCard"),
		"layouts/main")
}

func (s *Server) handleGetSignIn(ctx *fiber.Ctx) error {
	return ctx.Render("signin", newData("Войти"), "layouts/main")
}

func (s *Server) handlePostSignIn(ctx *fiber.Ctx) error {
	req, errs := parseSignInRequest(ctx)
	if errs != nil {
		ctx.Status(fiber.StatusBadRequest)
		return ctx.Render("signin", newData("Войти").WithErrors(errs...), "layouts/main")
	}
	user, err := s.auth.Login(ctx.Context(), req.name, req.password)
	if err != nil {
		ctx.Status(fiber.StatusUnauthorized)
		return ctx.Render("signin",
			newData("Войти").
				WithErrors(err.Error()),
			"layouts/main")
	}
	cookie, err := s.auth.GenerateJWTCookie(user.ID, s.cfg.Host)
	if err != nil {
		ctx.Status(fiber.StatusUnauthorized)
		return ctx.Render("signin",
			newData("Войти").
				WithErrors(err.Error()),
			"layouts/main")
	}
	ctx.Cookie(cookie)
	return ctx.Redirect(webpath.ApiHome)
}

func (s *Server) handleGetSignup(ctx *fiber.Ctx) error {
	return ctx.Render("signup", newData("Зарегистрироваться"), "layouts/main")
}

func (s *Server) handlePostSignup(ctx *fiber.Ctx) error {
	req, errs := parseSignUpRequest(ctx)
	if errs != nil {
		ctx.Status(fiber.StatusBadRequest)
		return ctx.Render("signup",
			newData("Зарегистрироваться").
				WithErrors(errs...),
			"layouts/main",
		)
	}
	err := s.auth.SignUp(ctx.Context(), req.name, req.password)
	if err != nil {
		ctx.Status(fiber.StatusBadRequest)
		errMsg := "Неизвестная ошибка" // TODO log
		if errors.Is(err, authservice.ErrAlreadyExists) {
			errMsg = "Пользователь с таким именем уже существует."
		}
		return ctx.Render("signup",
			newData("Зарегистрироваться").
				WithErrors(errMsg),
			"layouts/main",
		)
	}
	return ctx.Redirect(webpath.Signin)
}

func (s *Server) handleSignOut(ctx *fiber.Ctx) error {
	ctx.ClearCookie("token")
	return ctx.Redirect(webpath.ApiHome)
}

func (s *Server) handleNewPlayerGet(ctx *fiber.Ctx) error {
	return ctx.Render("newPlayer", newData("Добавить игрока"), "layouts/main")
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
