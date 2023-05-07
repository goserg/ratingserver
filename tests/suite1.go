package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	auth "ratingserver/auth/service"
	"ratingserver/internal/config"
	"ratingserver/internal/web/webpath"
	"time"

	mapset "github.com/deckarep/golang-set/v2"

	"github.com/chromedp/cdproto/cdp"

	"github.com/chromedp/chromedp"
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
	s.T().Logf("process finished with code %d", exitCode)
}

func (s *TestSuite1) TestHandlers() {
	ctx, _ := context.WithTimeout(context.Background(), time.Second*5)

	// create context
	ctx, cancel := chromedp.NewContext(ctx)
	defer cancel()
	// run task list
	var logo string
	err := chromedp.Run(ctx,
		s.CheckAccessDenied(s.addr+webpath.ApiNewMatch),
		s.CheckAccessDenied(s.addr+webpath.ApiNewPlayer),
		s.CheckAccessGranted(s.addr),
		s.CheckAccessGranted(s.addr+webpath.Api),
		s.CheckAccessGranted(s.addr+webpath.ApiMatchesList),
		s.CheckAccessGranted(s.addr+webpath.Signin),
		s.CheckAccessGranted(s.addr+webpath.Signup),
		s.CheckAccessGranted(s.addr+webpath.Signout),
		s.Login(auth.Root, s.config.Server.Auth.RootPassword),
		s.CheckAccessGranted(s.addr+webpath.ApiNewMatch),
		s.CheckAccessGranted(s.addr+webpath.ApiNewPlayer),
		s.CheckAccessGranted(s.addr),
		s.CheckAccessGranted(s.addr+webpath.Api),
		s.CheckAccessGranted(s.addr+webpath.ApiMatchesList),
		s.CheckAccessGranted(s.addr+webpath.Signin),
		s.CheckAccessGranted(s.addr+webpath.Signup),
		s.CheckAccessGranted(s.addr+webpath.Signout),
		s.Login(auth.Root, s.config.Server.Auth.RootPassword),
		s.NewPlayer("Иван"),
		s.NewPlayer("Артём"),
		s.NewPlayer("Мария"),
		s.CheckPlayersExist("Иван", "Артём", "Мария"),
		chromedp.Navigate(s.addr),
		chromedp.Text(`.brand-logo`, &logo),
		chromedp.ActionFunc(func(ctx context.Context) error {
			if logo != "Эш-рейтинг" {
				err := errors.New("invalid logo text: " + logo)
				var screenShot []byte
				if errS := chromedp.FullScreenshot(&screenShot, 80).Do(ctx); errS != nil {
					return errors.Join(errS, err)
				}
				if errW := os.WriteFile("invalid_logo.png", screenShot, 0o644); errW != nil {
					return errors.Join(errW, err)
				}
				return err
			}
			return nil
		}),
	)

	if err != nil {
		s.T().Fatalf(err.Error())
	}
	s.Equal("Эш-рейтинг", logo)
}

func (s *TestSuite1) CheckAccessDenied(path string) chromedp.Tasks {
	s.T().Logf("Проверка на закрытый доступ к %s", path)
	return []chromedp.Action{
		chromedp.ActionFunc(func(ctx context.Context) error {
			resp, err := chromedp.RunResponse(ctx,
				chromedp.Navigate(path))
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
				chromedp.Navigate(path))
			if err != nil {
				return err
			}
			if resp.Status != http.StatusOK {
				s.T().Errorf("Доступ к %s для гостей должен быть разрешен (статус 200), сервер ответил статуом %d", path, resp.Status)
			}
			return nil
		}),
	}
}

func (s *TestSuite1) Screenshot(filename string) chromedp.ActionFunc {
	return func(ctx context.Context) error {
		var screenShot []byte
		if err := chromedp.FullScreenshot(&screenShot, 80).Do(ctx); err != nil {
			return err
		}
		if err := os.WriteFile(filename, screenShot, 0o644); err != nil {
			return err
		}
		return nil
	}
}

func (s *TestSuite1) NewPlayer(name string) chromedp.Tasks {
	s.T().Logf("Создание игрока %s", name)
	return []chromedp.Action{
		chromedp.Navigate(s.addr + webpath.ApiNewPlayer),
		chromedp.SendKeys("#new-player-form-name", name),
		chromedp.Submit("#new-player-form-submit"),
		chromedp.WaitVisible(`.brand-logo`),
	}
}

func (s *TestSuite1) Login(user, password string) chromedp.Tasks {
	s.T().Logf("Логин как %s", user)
	return []chromedp.Action{
		chromedp.Navigate(s.addr + webpath.Signin),
		chromedp.SendKeys("#username-field", user),
		chromedp.SendKeys("#password-field", password),
		chromedp.Submit("#signin-form-submit"),
		chromedp.WaitVisible(`.brand-logo`),
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
		chromedp.WaitVisible(`.brand-logo`),
		chromedp.ActionFunc(func(ctx context.Context) error {
			var nodes []*cdp.Node
			err := chromedp.Nodes("#player-list-row-name", &nodes, chromedp.NodeVisible).Do(ctx)
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
