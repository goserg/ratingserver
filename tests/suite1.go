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
func (suite *TestSuite1) SetupSuite() {
	fmt.Println("setupSuite")
	suite.Require().NotEmpty(serverConfigPath, "-server-config MUST be set")
	suite.Require().NotEmpty(botConfigPath, "-bot-config MUST be set")
	p := NewProcess(context.Background(), "../bin/server",
		"-server-config", serverConfigPath,
		"-bot-config", botConfigPath)
	suite.process = p
	err := p.Start(context.Background())
	if err != nil {
		suite.T().Errorf("cant start process: %v", err)
	}

	if err := waitForStartup(time.Second * 5); err != nil {
		suite.T().Fatalf("unable to start app: %v", err)
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
func (suite *TestSuite1) TearDownSuite() {
	fmt.Println("teardown Suite1")
	exitCode, err := suite.process.Stop()
	if err != nil {
		suite.T().Logf("cant stop process: %v", err)
	}

	suite.T().Logf("process finished with code %d", exitCode)
}

func (suite *TestSuite1) TestHandlers() {
	fmt.Println("test handlers")
	defer fmt.Println("test finished")

	ctx, _ := context.WithTimeout(context.Background(), time.Second*5)

	// create context
	ctx, cancel := chromedp.NewContext(ctx)
	defer cancel()
	// run task list
	var logo string
	err := chromedp.Run(ctx,
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
		suite.T().Fatalf(err.Error())
	}
	suite.Equal("Эш-рейтинг", logo)
}
