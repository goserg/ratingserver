package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/chromedp/chromedp"
	"github.com/stretchr/testify/suite"
)

type TestSuite1 struct {
	suite.Suite
	process *Process
}

var (
	serverConfigPath string
	botConfigPath    string
)

func init() {
	flag.StringVar(&serverConfigPath, "server-config", "", "path to server configs")
	flag.StringVar(&botConfigPath, "bot-config", "", "path to bot configs")
}

// SetupSuite подготавливает необходимые зависимости
func (s *TestSuite1) SetupSuite() {
	fmt.Println("setupSuite")
	s.Require().NotEmpty(serverConfigPath, "-server-config MUST be set")
	s.Require().NotEmpty(botConfigPath, "-bot-config MUST be set")
	p := NewProcess(context.Background(), "../bin/server",
		"-server-config", serverConfigPath,
		"-bot-config", botConfigPath)
	s.process = p
	err := p.Start(context.Background())
	if err != nil {
		s.T().Errorf("cant start process: %v", err)
	}

	if err := waitForStartup(time.Second * 5); err != nil {
		s.T().Fatalf("unable to start app: %v", err)
	}
}

func waitForStartup(duration time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), duration)
	defer cancel()

	ticker := time.NewTicker(time.Second / 2)
	for {
		select {
		case <-ticker.C:
			r, _ := http.Get("http://0.0.0.0:3000/")
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
	// TODO clean DB files
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
		s.CheckGuestAccessDenied(`http://0.0.0.0:3000/api/matches`),
		s.CheckGuestAccessGranted(`http://0.0.0.0:3000/`),
		s.CheckGuestAccessGranted(`http://0.0.0.0:3000/api`),
		s.CheckGuestAccessGranted(`http://0.0.0.0:3000/api/matches-list`),
		s.CheckGuestAccessGranted(`http://0.0.0.0:3000/signin`),
		s.CheckGuestAccessGranted(`http://0.0.0.0:3000/signout`),
		s.CheckGuestAccessGranted(`http://0.0.0.0:3000/signup`),
		chromedp.Navigate(`http://0.0.0.0:3000/`),
		chromedp.Text(`.brand-logo`, &logo),
		chromedp.ActionFunc(func(ctx context.Context) error {
			if logo != "Эш-рейтинг" {
				err := errors.New("invalid logo text: " + logo)
				var screenShot []byte
				chromedp.FullScreenshot(&screenShot, 80).Do(ctx)
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

func (s *TestSuite1) CheckGuestAccessDenied(path string) chromedp.Tasks {
	return []chromedp.Action{
		chromedp.ActionFunc(func(ctx context.Context) error {
			resp, err := chromedp.RunResponse(ctx,
				chromedp.Navigate(path))
			if err != nil {
				return err
			}
			if resp.Status != http.StatusForbidden {
				s.T().Errorf("Доступ к %s для гостей должен быть запрещен (статус 403), сервер ответил статуом %d", path, resp.Status)
			}
			return nil
		}),
	}
}

func (s *TestSuite1) CheckGuestAccessGranted(path string) chromedp.Tasks {
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
