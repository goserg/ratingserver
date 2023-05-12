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
	"regexp"
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
	return ctx.Render("index", Data("Рейтинг").
		WithUser(user).
		With("Button", "rating").
		With("Players", globalRating),
		"layouts/main")
}

func (s *Server) handleMatches(ctx *fiber.Ctx) error {
	user, _ := ctx.Context().UserValue(userKey).(users.User)
	matches, err := s.playerService.GetMatches()
	if err != nil {
		return err
	}
	return ctx.Render("matches",
		Data("Список матчей").
			WithUser(user).
			With("Button", "matches").
			With("Matches", matches),
		"layouts/main")
}

func (s *Server) handleCreateMatchGet(ctx *fiber.Ctx) error {
	user, _ := ctx.Context().UserValue(userKey).(users.User)
	return ctx.Render("newMatch", Data("Добавить игру").WithUser(user), "layouts/main")
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
	card, err := s.playerService.GetPlayerData(id)
	if err != nil {
		return err
	}
	return ctx.Render("playerCard",
		Data(card.Player.Name).
			WithUser(user).
			With("PlayerCard", card).
			With("Button", "playerCard"),
		"layouts/main")
}

func (s *Server) HandleGetSignIn(ctx *fiber.Ctx) error {
	return ctx.Render("signin", Data("Войти"), "layouts/main")
}

type signInRequest struct {
	name     string
	password string
}

func parseSignInRequest(ctx *fiber.Ctx) (signInRequest, []string) {
	var errs []string
	name := ctx.FormValue("username", "")
	if name == "" {
		errs = append(errs, "имя пользователя не должно быть пустое")
	}
	nameRegexp := regexp.MustCompile("^[A-Za-z]\\w+$")
	if !nameRegexp.MatchString(name) {
		errs = append(errs, "имя пользователя должно начинаться с латинской буквы и содержать только латинские буквы, цифры и знаки подчеркивания")
	}
	password := ctx.FormValue("password", "")
	if password == "" {
		errs = append(errs, "пароль пользователя не должн быть пустым")
	}
	if errs != nil {
		return signInRequest{}, errs
	}
	return signInRequest{
		name:     name,
		password: password,
	}, nil
}

func (s *Server) HandlePostSignIn(ctx *fiber.Ctx) error {
	req, errs := parseSignInRequest(ctx)
	if errs != nil {
		ctx.Status(fiber.StatusBadRequest)
		return ctx.Render("signin", Data("Войти").WithErrors(errs...), "layouts/main")
	}
	user, err := s.auth.Login(ctx.Context(), req.name, req.password)
	if err != nil {
		ctx.Status(fiber.StatusUnauthorized)
		return ctx.Render("signin",
			Data("Войти").
				WithErrors(err.Error()),
			"layouts/main")
	}
	cookie, err := s.auth.GenerateJWTCookie(user.ID, s.cfg.Host)
	if err != nil {
		ctx.Status(fiber.StatusUnauthorized)
		return ctx.Render("signin",
			Data("Войти").
				WithErrors(err.Error()),
			"layouts/main")
	}
	ctx.Cookie(cookie)
	return ctx.Redirect(webpath.ApiHome)
}

func (s *Server) HandleGetSignup(ctx *fiber.Ctx) error {
	return ctx.Render("signup", Data("Зарегистрироваться"), "layouts/main")
}

func (s *Server) HandlePostSignup(ctx *fiber.Ctx) error {
	name := ctx.FormValue("username", "")
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
	return ctx.Render("newPlayer", Data("Добавить игрока"), "layouts/main")
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
