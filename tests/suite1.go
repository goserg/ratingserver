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
	fmt.Println("setupSuite")

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
	fmt.Println("teardown Suite1")
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
	fmt.Println("test handlers")
	defer fmt.Println("test finished")

	ctx, _ := context.WithTimeout(context.Background(), time.Second*5)

	// create context
	ctx, cancel := chromedp.NewContext(ctx)
	defer cancel()
	// run task list
	var logo string
	err := chromedp.Run(ctx,
		s.CheckAccessDenied(s.addr+webpath.ApiNewMatch),
		s.CheckAccessGranted(s.addr),
		s.CheckAccessGranted(s.addr+webpath.Api),
		s.CheckAccessGranted(s.addr+webpath.ApiMatchesList),
		s.CheckAccessGranted(s.addr+webpath.Signin),
		s.CheckAccessGranted(s.addr+webpath.Signout),
		s.CheckAccessGranted(s.addr+webpath.Signup),
		s.Login(auth.Root, s.config.Server.Auth.RootPassword),
		s.CheckAccessGranted(s.addr+webpath.ApiNewMatch),
		s.CheckAccessGranted(s.addr),
		s.CheckAccessGranted(s.addr+webpath.Api),
		s.CheckAccessGranted(s.addr+webpath.ApiMatchesList),
		s.CheckAccessGranted(s.addr+webpath.Signin),
		s.CheckAccessGranted(s.addr+webpath.Signout),
		s.CheckAccessGranted(s.addr+webpath.Signup),
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

func (s *TestSuite1) Login(user, password string) chromedp.Tasks {
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
