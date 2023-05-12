package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	auth "ratingserver/auth/service"
	"ratingserver/internal/config"
	"ratingserver/internal/web/webpath"
	sel "ratingserver/tests/selectors"

	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/chromedp"
	mapset "github.com/deckarep/golang-set/v2"
	"github.com/stretchr/testify/suite"
)

type TestSuite1 struct {
	suite.Suite
	process *Process

	addr   string
	config config.Config
}

// SetupSuite подготавливает необходимые зависимости
func (s *TestSuite1) SetupSuite() {
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

func (s *TestSuite1) waitForStartup(duration time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), duration)
	defer cancel()

	ticker := time.NewTicker(time.Second / 2)
	for {
		select {
		case <-ticker.C:
			r, _ := http.Get(s.addr)
			if r != nil && r.StatusCode == http.StatusOK {
				return nil
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

// TearDownSuite высвобождает имеющиеся зависимости
func (s *TestSuite1) TearDownSuite() {
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

const globalTestsTimeoutSeconds = 5

func (s *TestSuite1) TestHandlers() {
	ctx, _ := context.WithTimeout(context.Background(), time.Second*globalTestsTimeoutSeconds)
	ctx, cancel := chromedp.NewContext(ctx)
	defer cancel()
	err := chromedp.Run(ctx,
		s.CheckAccessDenied(webpath.ApiNewMatch),
		s.CheckAccessDenied(webpath.ApiNewPlayer),
		s.CheckAccessGranted(webpath.Home),
		s.CheckAccessGranted(webpath.Api),
		s.CheckAccessGranted(webpath.ApiMatchesList),
		s.CheckAccessGranted(webpath.Signin),
		s.CheckAccessGranted(webpath.Signup),
		s.CheckAccessGranted(webpath.Signout),
		s.SignIn(auth.Root, s.config.Server.Auth.RootPassword),
		s.CheckAccessGranted(webpath.ApiNewMatch),
		s.CheckAccessGranted(webpath.ApiNewPlayer),
		s.CheckAccessGranted(webpath.Home),
		s.CheckAccessGranted(webpath.Api),
		s.CheckAccessGranted(webpath.ApiMatchesList),
		s.CheckAccessGranted(webpath.Signin),
		s.CheckAccessGranted(webpath.Signup),
		s.CheckAccessGranted(webpath.Signout),
		s.SignIn(auth.Root, s.config.Server.Auth.RootPassword),
		s.NewPlayer("Иван"),
		s.NewPlayer("Артём"),
		s.NewPlayer("Мария"),
		s.CheckPlayersExist("Иван", "Артём", "Мария"),
		s.NewGame("Иван", "Артём", false),
		s.NewGame("Иван", "Мария", false),
		s.NewGame("Иван", "Мария", true),
		s.NewGame("Артём", "Иван", false),
		s.NewGame("Мария", "Артём", false),
		s.CheckPlayersStats(),
		s.CreateUser("user1", "qwerty"),
		s.SignIn("user1", "qwerty"),
		s.CheckAccessDenied(webpath.ApiNewMatch),
		s.CheckAccessDenied(webpath.ApiNewPlayer),
		s.CheckAccessGranted(webpath.Home),
		s.CheckAccessGranted(webpath.Api),
		s.CheckAccessGranted(webpath.ApiMatchesList),
		s.CheckAccessGranted(webpath.Signin),
		s.CheckAccessGranted(webpath.Signup),
		s.CheckAccessGranted(webpath.Signout),
		s.CheckLink(webpath.ApiHome, sel.NavMatchesLink, webpath.ApiMatchesList),
		s.CheckLink(webpath.ApiHome, sel.NavPlayersLink, webpath.ApiHome),
		s.CheckLink(webpath.ApiMatchesList, sel.NavMatchesLink, webpath.ApiMatchesList),
		s.CheckLink(webpath.ApiMatchesList, sel.NavPlayersLink, webpath.ApiHome),
		s.CheckLink(webpath.Signup, sel.SignUpToSignInLink, webpath.Signin),
		s.CheckLink(webpath.Signin, sel.SignInToSignUpLink, webpath.Signup),
		s.CheckInvalidSignInForm(),
	)
	if err != nil {
		s.T().Fatalf(err.Error())
	}
}

func (s *TestSuite1) CheckAccessDenied(path string) chromedp.Tasks {
	s.T().Logf("Проверка на закрытый доступ к %s", path)
	return []chromedp.Action{
		chromedp.ActionFunc(func(ctx context.Context) error {
			resp, err := chromedp.RunResponse(ctx,
				chromedp.Navigate(s.addr+path))
			if err != nil {
				return err
			}
			if resp.Status != http.StatusForbidden {
				s.T().Errorf("Доступ к %s должен быть запрещен (статус 403), сервер ответил статуом %d", path, resp.Status)
			}
			return nil
		}),
	}
}

func (s *TestSuite1) CheckAccessGranted(path string) chromedp.Tasks {
	s.T().Logf("Проверка на разрешенный доступ к %s", path)
	return []chromedp.Action{
		chromedp.ActionFunc(func(ctx context.Context) error {
			resp, err := chromedp.RunResponse(ctx,
				chromedp.Navigate(s.addr+path))
			if err != nil {
				return err
			}
			if resp.Status != http.StatusOK {
				s.T().Errorf("Доступ к %s должен быть разрешен (статус 200), сервер ответил статуом %d", path, resp.Status)
			}
			return nil
		}),
	}
}

const screenshotQuality = 80
const screenshotFilePermission = 0o644

func (s *TestSuite1) Screenshot(filename string) chromedp.ActionFunc {
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

func (s *TestSuite1) NewPlayer(name string) chromedp.Tasks {
	s.T().Logf("Создание игрока %s", name)
	return []chromedp.Action{
		chromedp.Navigate(s.addr + webpath.ApiNewPlayer),
		chromedp.SendKeys(sel.NewPlayerFormName, name),
		s.wait200(chromedp.Submit(sel.NewPlayerFormSubmit)),
	}
}

func (s *TestSuite1) SignIn(user, password string) chromedp.Tasks {
	s.T().Logf("Логин как %q", user)
	return []chromedp.Action{
		chromedp.Navigate(s.addr + webpath.Signin),
		chromedp.SendKeys(sel.SignInFormUsername, user),
		chromedp.SendKeys(sel.SignInFormPassword, password),
		s.wait200(chromedp.Submit(sel.SignInFormSubmit)),
	}
}

func (s *TestSuite1) cleanupDB() error {
	err := errors.Join(
		os.Remove(s.config.Server.SqliteFile),
		os.Remove(s.config.Server.Auth.SqliteFile),
	)
	if !s.config.Server.TgBotDisable {
		err = errors.Join(err, os.Remove(s.config.TgBot.SqliteFile))
	}
	return err
}

func (s *TestSuite1) CheckPlayersExist(names ...string) chromedp.Tasks {
	expectedNames := mapset.NewSet(names...)
	s.T().Log("Проверяем, что игроки создались")
	return []chromedp.Action{
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
	}
}

func (s *TestSuite1) NewGame(winner string, loser string, draw bool) chromedp.Tasks {
	winnerPoints := 1.0
	loserPoints := 0.0
	if draw {
		winnerPoints = 0.5
		loserPoints = 0.5
	}
	s.T().Logf("Создание новой игры: %s %.1f:%.1f %s", winner, winnerPoints, loserPoints, loser)
	return []chromedp.Action{
		chromedp.Navigate(s.addr + webpath.ApiNewMatch),
		chromedp.SendKeys(sel.NewMatchFormWinner, winner),
		chromedp.SendKeys(sel.NewMatchFormLoser, loser),
		chromedp.ActionFunc(func(ctx context.Context) error {
			if !draw {
				return nil
			}
			return chromedp.Click(sel.NewMatchFormDraw).Do(ctx)
		}),
		s.wait200(chromedp.Submit(sel.NewMatchFormSubmit)),
	}
}

func (s *TestSuite1) wait200(action chromedp.Action) chromedp.Action {
	return chromedp.ActionFunc(func(ctx context.Context) error {
		resp, err := chromedp.RunResponse(ctx, action)
		if err != nil {
			return err
		}
		if resp.Status != http.StatusOK {
			return fmt.Errorf("ожидалось 200, результат %d", resp.Status)
		}
		return nil
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

func (s *TestSuite1) CheckPlayersStats() chromedp.Tasks {
	s.T().Logf("Проверка правильности содания игр")
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
	return []chromedp.Action{
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
	}
}

func (s *TestSuite1) CreateUser(name, password string) chromedp.Tasks {
	s.T().Logf("Создание нового пользователя %q", name)
	return []chromedp.Action{
		chromedp.Navigate(s.addr + webpath.Signup),
		chromedp.SendKeys(sel.SignUpFormUsername, name),
		chromedp.SendKeys(sel.SignUpFormPassword, password),
		chromedp.SendKeys(sel.SignUpFormPasswordRepeat, password),
		s.wait200(chromedp.Submit(sel.SignUpFormSubmit)),
	}
}

func (s *TestSuite1) CheckLink(from string, selector string, expectTarget string) chromedp.Tasks {
	s.T().Logf("Проверка ссылки %q на странице %s", selector, from)
	return []chromedp.Action{
		s.wait200(chromedp.Navigate(s.addr + from)),
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
	}
}

func (s *TestSuite1) expectBadRequest(action chromedp.Action) chromedp.Action {
	return chromedp.ActionFunc(func(ctx context.Context) error {
		resp, err := chromedp.RunResponse(ctx, action)
		if err != nil {
			return err
		}
		if resp.Status != http.StatusBadRequest {
			s.T().Logf("Одижается статус 400, пришёл статус %d", resp.Status)
			return err
		}
		return nil
	})
}

func (s *TestSuite1) CheckInvalidSignInForm() chromedp.Tasks {
	s.T().Logf("Проверка на невалидный ввод в форме SignIn")
	return []chromedp.Action{
		// пустые поля
		chromedp.Navigate(s.addr + webpath.Signin),
		s.expectBadRequest(chromedp.Submit(sel.SignInFormSubmit)),

		// пустой пароль
		chromedp.Navigate(s.addr + webpath.Signin),
		chromedp.SendKeys(sel.SignInFormUsername, auth.Root),
		s.expectBadRequest(chromedp.Submit(sel.SignInFormSubmit)),

		// пустое имя пользователя
		chromedp.Navigate(s.addr + webpath.Signin),
		chromedp.SendKeys(sel.SignInFormPassword, s.config.Server.Auth.RootPassword),
		s.expectBadRequest(chromedp.Submit(sel.SignInFormSubmit)),

		// имя пользователя начинающееся не с буквы
		chromedp.Navigate(s.addr + webpath.Signin),
		chromedp.SendKeys(sel.SignInFormUsername, "1awd"),
		chromedp.SendKeys(sel.SignInFormPassword, s.config.Server.Auth.RootPassword),
		s.expectBadRequest(chromedp.Submit(sel.SignInFormSubmit)),

		// имя пользователя сщдержащее спец символы
		chromedp.Navigate(s.addr + webpath.Signin),
		chromedp.SendKeys(sel.SignInFormUsername, "Dawd%awd"),
		chromedp.SendKeys(sel.SignInFormPassword, s.config.Server.Auth.RootPassword),
		s.expectBadRequest(chromedp.Submit(sel.SignInFormSubmit)),
	}
}
