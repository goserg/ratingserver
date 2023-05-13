package main

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/goserg/ratingserver/internal/web/webpath"
	sel "github.com/goserg/ratingserver/tests/selectors"

	auth "github.com/goserg/ratingserver/auth/service"
	"github.com/goserg/ratingserver/internal/config"

	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/chromedp"
	mapset "github.com/deckarep/golang-set/v2"
	"github.com/stretchr/testify/suite"
)

type Suite struct {
	suite.Suite
	process *Process

	addr   string
	config config.Config
}

// SetupSuite подготавливает необходимые зависимости.
func (s *Suite) SetupSuite() {
	cfg, err := config.New()
	if err != nil {
		s.T().Fatalf("can't get configs")
	}
	s.config = cfg
	s.addr = fmt.Sprintf("http://%s:%d", cfg.Server.Host, cfg.Server.Port)

	p := NewProcess(context.Background(), "../bin/server",
		"-server-config", config.ServerConfigPath,
		"-bot-config", config.BotConfigPath)
	s.process = p
	err = p.Start(context.Background())
	if err != nil {
		s.T().Errorf("cant start process: %v", err)
	}

	if err := s.waitForStartup(time.Second * 5); err != nil {
		s.T().Fatalf("unable to start app: %v", err)
	}
}

func (s *Suite) waitForStartup(duration time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), duration)
	defer cancel()

	ticker := time.NewTicker(time.Second / 2)
	for {
		select {
		case <-ticker.C:
			req, err := http.NewRequestWithContext(ctx, http.MethodGet, s.addr, http.NoBody)
			if err != nil {
				s.T().Fatal("can not create request", err)
			}
			response, err := http.DefaultClient.Do(req)
			if err != nil {
				s.T().Log("waiting server to startup", err)
				continue
			}
			if response != nil && response.StatusCode == http.StatusOK {
				return nil
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

// TearDownSuite высвобождает имеющиеся зависимости.
func (s *Suite) TearDownSuite() {
	exitCode, err := s.process.Stop()
	if err != nil {
		s.T().Logf("cant stop process: %v", err)
	}
	if err := s.cleanupDB(); err != nil {
		s.T().Logf("can't clean db files")
	}
	if exitCode > 0 {
		s.T().Logf("process finished with code %d", exitCode)
	}
}

const globalTestsTimeoutSeconds = 5000

func (s *Suite) TestRatings() {
	ctx, _ := context.WithTimeout(context.Background(), time.Second*globalTestsTimeoutSeconds)
	ctx, cancel := chromedp.NewContext(ctx)
	defer cancel()

	s.SignIn(ctx, auth.Root, s.config.Server.Auth.RootPassword)
	s.Run("creating players", func() {
		s.NewPlayer(ctx, "Иван")
		s.NewPlayer(ctx, "Артём")
		s.NewPlayer(ctx, "Мария")
		s.CheckPlayersExist(ctx, "Иван", "Артём", "Мария")
	})
	s.Run("creating games", func() {
		s.NewGame(ctx, "Иван", "Артём", false)
		s.NewGame(ctx, "Иван", "Мария", false)
		s.NewGame(ctx, "Иван", "Мария", true)
		s.NewGame(ctx, "Артём", "Иван", false)
		s.NewGame(ctx, "Мария", "Артём", false)
	})
	s.Run("check ratings", func() {
		s.CheckPlayersStats(ctx)
	})
}

func (s *Suite) TestInvalidFormInput() {
	ctx, _ := context.WithTimeout(context.Background(), time.Second*globalTestsTimeoutSeconds)
	ctx, cancel := chromedp.NewContext(ctx)
	defer cancel()

	s.Run("signin form", func() {
		s.CheckInvalidSignInForm(ctx)
	})
}

func (s *Suite) TestAccessRoot() {
	ctx, _ := context.WithTimeout(context.Background(), time.Second*globalTestsTimeoutSeconds)
	ctx, cancel := chromedp.NewContext(ctx)
	defer cancel()

	s.SignIn(ctx, auth.Root, s.config.Server.Auth.RootPassword)
	s.CheckAccessGranted(ctx, webpath.ApiNewMatch)
	s.CheckAccessGranted(ctx, webpath.ApiNewPlayer)
	s.CheckAccessGranted(ctx, webpath.Home)
	s.CheckAccessGranted(ctx, webpath.Api)
	s.CheckAccessGranted(ctx, webpath.ApiMatchesList)
	s.CheckAccessGranted(ctx, webpath.Signin)
	s.CheckAccessGranted(ctx, webpath.Signup)
	s.CheckAccessGranted(ctx, webpath.Signout)
}

func (s *Suite) TestAccessGuest() {
	ctx, _ := context.WithTimeout(context.Background(), time.Second*globalTestsTimeoutSeconds)
	ctx, cancel := chromedp.NewContext(ctx)
	defer cancel()
	// TODO прверить что это действительно гость
	s.CheckAccessDenied(ctx, webpath.ApiNewMatch)
	s.CheckAccessDenied(ctx, webpath.ApiNewPlayer)
	s.CheckAccessGranted(ctx, webpath.Home)
	s.CheckAccessGranted(ctx, webpath.Api)
	s.CheckAccessGranted(ctx, webpath.ApiMatchesList)
	s.CheckAccessGranted(ctx, webpath.Signin)
	s.CheckAccessGranted(ctx, webpath.Signup)
	s.CheckAccessGranted(ctx, webpath.Signout)
}

func (s *Suite) TestAccessUser() {
	ctx, _ := context.WithTimeout(context.Background(), time.Second*globalTestsTimeoutSeconds)
	ctx, cancel := chromedp.NewContext(ctx)
	defer cancel()

	u := randomUsername()
	s.CreateUser(ctx, u, "qwerty")
	s.SignIn(ctx, u, "qwerty")
	s.CheckAccessDenied(ctx, webpath.ApiNewMatch)
	s.CheckAccessDenied(ctx, webpath.ApiNewPlayer)
	s.CheckAccessGranted(ctx, webpath.Home)
	s.CheckAccessGranted(ctx, webpath.Api)
	s.CheckAccessGranted(ctx, webpath.ApiMatchesList)
	s.CheckAccessGranted(ctx, webpath.Signin)
	s.CheckAccessGranted(ctx, webpath.Signup)
	s.CheckAccessGranted(ctx, webpath.Signout)
}

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func randomUsername() string {
	b := make([]byte, 15)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}

func (s *Suite) TestNewUser() {
	ctx, _ := context.WithTimeout(context.Background(), time.Second*globalTestsTimeoutSeconds)
	ctx, cancel := chromedp.NewContext(ctx)
	defer cancel()

	u := randomUsername()
	s.CreateUser(ctx, u, "qwerty")
	s.SignIn(ctx, u, "qwerty")
}

func (s *Suite) TestUserCreateUniqueName() {
	ctx, _ := context.WithTimeout(context.Background(), time.Second*globalTestsTimeoutSeconds)
	ctx, cancel := chromedp.NewContext(ctx)
	defer cancel()

	u := randomUsername()
	s.CreateUser(ctx, u, "qwerty")
	s.CreateUserMustFail(ctx, u, "qwerty2")
}

func (s *Suite) TestLinks() {
	ctx, _ := context.WithTimeout(context.Background(), time.Second*globalTestsTimeoutSeconds)
	ctx, cancel := chromedp.NewContext(ctx)
	defer cancel()
	s.Run("main to matches", func() {
		s.CheckLink(ctx, webpath.ApiHome, sel.NavMatchesLink, webpath.ApiMatchesList)
	})

	s.Run("main to players", func() {
		s.CheckLink(ctx, webpath.ApiHome, sel.NavPlayersLink, webpath.ApiHome)
	})

	s.Run("matches to matches", func() {
		s.CheckLink(ctx, webpath.ApiMatchesList, sel.NavMatchesLink, webpath.ApiMatchesList)
	})
	s.Run("matches to players", func() {
		s.CheckLink(ctx, webpath.ApiMatchesList, sel.NavPlayersLink, webpath.ApiHome)
	})
	s.Run("signup to signin", func() {
		s.CheckLink(ctx, webpath.Signup, sel.SignUpToSignInLink, webpath.Signin)
	})
	s.Run("signin to signup", func() {
		s.CheckLink(ctx, webpath.Signin, sel.SignInToSignUpLink, webpath.Signup)
	})
}

func (s *Suite) CheckAccessDenied(ctx context.Context, path string) {
	s.T().Helper()
	resp, err := chromedp.RunResponse(ctx,
		chromedp.Navigate(s.addr+path))
	s.Require().NoError(err)
	s.Require().EqualValues(http.StatusForbidden, resp.Status)
}

func (s *Suite) CheckAccessGranted(ctx context.Context, path string) {
	s.T().Helper()
	resp, err := chromedp.RunResponse(ctx,
		chromedp.Navigate(s.addr+path))
	s.Require().NoError(err)
	s.Require().EqualValues(http.StatusOK, resp.Status)
}

const (
	screenshotQuality        = 80
	screenshotFilePermission = 0o644
)

func (s *Suite) Screenshot(filename string) chromedp.ActionFunc {
	return func(ctx context.Context) error {
		var screenShot []byte
		if err := chromedp.FullScreenshot(&screenShot, screenshotQuality).Do(ctx); err != nil {
			return err
		}
		if err := os.WriteFile(filename, screenShot, screenshotFilePermission); err != nil {
			return err
		}
		return nil
	}
}

func (s *Suite) NewPlayer(ctx context.Context, name string) {
	s.T().Helper()
	err := chromedp.Run(ctx, chromedp.Tasks([]chromedp.Action{
		chromedp.Navigate(s.addr + webpath.ApiNewPlayer),
		chromedp.SendKeys(sel.NewPlayerFormName, name),
		s.waitStatus(chromedp.Submit(sel.NewPlayerFormSubmit), http.StatusOK),
	}))
	s.Require().NoError(err)
}

func (s *Suite) SignIn(ctx context.Context, user, password string) {
	s.T().Helper()
	err := chromedp.Run(ctx, chromedp.Tasks([]chromedp.Action{
		chromedp.Navigate(s.addr + webpath.Signin),
		chromedp.SendKeys(sel.SignInFormUsername, user),
		chromedp.SendKeys(sel.SignInFormPassword, password),
		s.waitStatus(chromedp.Submit(sel.SignInFormSubmit), http.StatusOK),
	}))
	s.Require().NoError(err)
}

func (s *Suite) cleanupDB() error {
	err := errors.Join(
		os.Remove(s.config.Server.SqliteFile),
		os.Remove(s.config.Server.Auth.SqliteFile),
	)
	if !s.config.Server.TgBotDisable {
		err = errors.Join(err, os.Remove(s.config.TgBot.SqliteFile))
	}
	return err
}

func (s *Suite) CheckPlayersExist(ctx context.Context, names ...string) {
	s.T().Helper()
	expectedNames := mapset.NewSet(names...)
	err := chromedp.Run(ctx, chromedp.Tasks([]chromedp.Action{
		chromedp.Navigate(s.addr + webpath.ApiHome),
		chromedp.ActionFunc(func(ctx context.Context) error {
			var nodes []*cdp.Node
			err := chromedp.Nodes(sel.PlayerListRowName, &nodes, chromedp.NodeVisible).Do(ctx)
			if err != nil {
				return err
			}
			actualNames := mapset.NewSet[string]()
			for _, node := range nodes {
				var name string
				err := chromedp.TextContent(node.FullXPath(), &name, chromedp.NodeVisible).Do(ctx)
				if err != nil {
					return err
				}
				actualNames.Add(name)
			}
			if !actualNames.Equal(expectedNames) {
				s.T().Errorf("Ожидались: %s; Найдено: %s", expectedNames.String(), actualNames.String())
			}
			return nil
		}),
	}))
	s.Require().NoError(err)
}

func (s *Suite) NewGame(ctx context.Context, winner, loser string, draw bool) {
	s.T().Helper()
	err := chromedp.Run(ctx, chromedp.Tasks([]chromedp.Action{
		chromedp.Navigate(s.addr + webpath.ApiNewMatch),
		chromedp.SendKeys(sel.NewMatchFormWinner, winner),
		chromedp.SendKeys(sel.NewMatchFormLoser, loser),
		chromedp.ActionFunc(func(ctx context.Context) error {
			if !draw {
				return nil
			}
			return chromedp.Click(sel.NewMatchFormDraw).Do(ctx)
		}),
		s.waitStatus(chromedp.Submit(sel.NewMatchFormSubmit), http.StatusOK),
	}))
	s.Require().NoError(err)
}

func (s *Suite) waitStatus(action chromedp.Action, status int) chromedp.Action {
	return chromedp.ActionFunc(func(ctx context.Context) error {
		resp, err := chromedp.RunResponse(ctx, action)
		s.Assert().NoError(err)
		s.Assert().EqualValues(status, resp.Status)
		return err
	})
}

type PlayerStats struct {
	Name        string
	GamesPlayed int
	EloRating   int
}

func PlayerStatsFromString(str string) (PlayerStats, error) {
	s := strings.Split(strings.TrimSpace(str), "\n")
	if len(s) < 3 {
		return PlayerStats{}, errors.New("invalid player stats")
	}
	gp, err := strconv.Atoi(strings.TrimSpace(s[1]))
	if err != nil {
		return PlayerStats{}, err
	}
	er, err := strconv.Atoi(strings.TrimSpace(s[2]))
	if err != nil {
		return PlayerStats{}, err
	}

	return PlayerStats{
		Name:        strings.TrimSpace(s[0]),
		GamesPlayed: gp,
		EloRating:   er,
	}, nil
}

func (s *Suite) CheckPlayersStats(ctx context.Context) {
	s.T().Helper()
	expectedStats := mapset.NewSet[PlayerStats](
		PlayerStats{
			Name:        "Иван",
			GamesPlayed: 4,
			EloRating:   1013,
		},
		PlayerStats{
			Name:        "Артём",
			GamesPlayed: 3,
			EloRating:   982,
		},
		PlayerStats{
			Name:        "Мария",
			GamesPlayed: 3,
			EloRating:   1005,
		},
	)
	err := chromedp.Run(ctx, chromedp.Tasks([]chromedp.Action{
		chromedp.Navigate(s.addr),
		chromedp.ActionFunc(func(ctx context.Context) error {
			var nodes []*cdp.Node
			err := chromedp.Nodes(sel.PlayerListRow, &nodes, chromedp.NodeVisible).Do(ctx)
			if err != nil {
				return err
			}
			actualStats := mapset.NewSet[PlayerStats]()
			for _, node := range nodes {
				var text string
				err := chromedp.TextContent(node.FullXPath(), &text, chromedp.NodeVisible).Do(ctx)
				if err != nil {
					return err
				}
				stat, err := PlayerStatsFromString(text)
				if err != nil {
					return err
				}
				actualStats.Add(stat)
			}
			if !actualStats.Equal(expectedStats) {
				s.T().Errorf("Ошибка\nОжидалось: %s.\nПолучено: %s.\nРазница: %s", expectedStats, actualStats, expectedStats.Difference(actualStats))
			}
			return nil
		}),
	}))
	s.Require().NoError(err)
}

func (s *Suite) CreateUser(ctx context.Context, name, password string) {
	s.T().Helper()
	err := chromedp.Run(ctx, chromedp.Tasks([]chromedp.Action{
		chromedp.Navigate(s.addr + webpath.Signup),
		chromedp.SendKeys(sel.SignUpFormUsername, name),
		chromedp.SendKeys(sel.SignUpFormPassword, password),
		chromedp.SendKeys(sel.SignUpFormPasswordRepeat, password),
		s.waitStatus(chromedp.Submit(sel.SignUpFormSubmit), http.StatusOK),
	}))
	s.Require().NoError(err)
}

func (s *Suite) CreateUserMustFail(ctx context.Context, name, password string) {
	s.T().Helper()
	err := chromedp.Run(ctx, chromedp.Tasks([]chromedp.Action{
		chromedp.Navigate(s.addr + webpath.Signup),
		chromedp.SendKeys(sel.SignUpFormUsername, name),
		chromedp.SendKeys(sel.SignUpFormPassword, password),
		chromedp.SendKeys(sel.SignUpFormPasswordRepeat, password),
		s.waitStatus(chromedp.Submit(sel.SignUpFormSubmit), http.StatusBadRequest),
	}))
	s.Require().NoError(err)
}

func (s *Suite) CheckLink(ctx context.Context, from, selector, expectTarget string) {
	s.T().Helper()
	err := chromedp.Run(ctx, chromedp.Tasks([]chromedp.Action{
		s.waitStatus(chromedp.Navigate(s.addr+from), http.StatusOK),
		chromedp.ActionFunc(func(ctx context.Context) error {
			resp, err := chromedp.RunResponse(ctx, chromedp.Click(selector))
			if err != nil {
				return err
			}
			if resp.Status != http.StatusOK {
				s.T().Errorf("Ссылка %s на странице %s вернула статус %d", selector, from, resp.Status)
			}
			if resp.URL != s.addr+expectTarget {
				s.T().Errorf("Ссылка %s ведет на %s, ожидалось %s", selector, resp.URL, expectTarget)
			}
			return nil
		}),
	}))
	if err != nil {
		s.T().Fatal(err.Error())
	}
}

func (s *Suite) CheckInvalidSignInForm(ctx context.Context) {
	s.T().Helper()
	err := chromedp.Run(ctx, chromedp.Tasks([]chromedp.Action{
		// валидные поля
		chromedp.Navigate(s.addr + webpath.Signin),
		chromedp.SendKeys(sel.SignInFormUsername, "root"),
		chromedp.SendKeys(sel.SignInFormPassword, s.config.Server.Auth.RootPassword),
		s.waitStatus(chromedp.Submit(sel.SignInFormSubmit), http.StatusOK),

		// пустые поля
		chromedp.Navigate(s.addr + webpath.Signin),
		s.waitStatus(chromedp.Submit(sel.SignInFormSubmit), http.StatusBadRequest),

		// пустой пароль
		chromedp.Navigate(s.addr + webpath.Signin),
		chromedp.SendKeys(sel.SignInFormUsername, auth.Root),
		s.waitStatus(chromedp.Submit(sel.SignInFormSubmit), http.StatusBadRequest),

		// пустое имя пользователя
		chromedp.Navigate(s.addr + webpath.Signin),
		chromedp.SendKeys(sel.SignInFormPassword, s.config.Server.Auth.RootPassword),
		s.waitStatus(chromedp.Submit(sel.SignInFormSubmit), http.StatusBadRequest),

		// имя пользователя начинающееся не с буквы
		chromedp.Navigate(s.addr + webpath.Signin),
		chromedp.SendKeys(sel.SignInFormUsername, "1awd"),
		chromedp.SendKeys(sel.SignInFormPassword, s.config.Server.Auth.RootPassword),
		s.waitStatus(chromedp.Submit(sel.SignInFormSubmit), http.StatusBadRequest),

		// имя пользователя сщдержащее спец символы
		chromedp.Navigate(s.addr + webpath.Signin),
		chromedp.SendKeys(sel.SignInFormUsername, "Dawd%awd"),
		chromedp.SendKeys(sel.SignInFormPassword, s.config.Server.Auth.RootPassword),
		s.waitStatus(chromedp.Submit(sel.SignInFormSubmit), http.StatusBadRequest),
	}))
	s.Require().NoError(err)
}
