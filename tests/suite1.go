package main

import (
	"context"
	"fmt"
	"time"

	"github.com/stretchr/testify/suite"
)

type TestSuite1 struct {
	suite.Suite
	process *Process
}

// SetupSuite подготавливает необходимые зависимости
func (suite *TestSuite1) SetupSuite() {
	fmt.Println("setupSuite")
	p := NewProcess(context.Background(), "../bin/server",
		"-server-config", "/home/serg/src/ratingserver/configs/server.toml",
		"-bot-config", "/home/serg/src/ratingserver/configs/bot.toml")
	suite.process = p
	err := p.Start(context.Background())
	if err != nil {
		suite.T().Errorf("cant start process: %v", err)
	}
}

// TearDownSuite высвобождает имеющиеся зависимости
func (suite *TestSuite1) TearDownSuite() {
	fmt.Println("teardown Suite1")
	exitCode, err := suite.process.Stop()
	if err != nil {
		suite.T().Logf("cant stop process: %v", err)
	}
	out := suite.process.stdout.String()
	suite.T().Logf("out: %s", out)

	stderr := suite.process.stderr.String()
	suite.T().Logf("err: %s", stderr)

	suite.T().Logf("process finished with code %d", exitCode)
}

func (suite *TestSuite1) TestHandlers() {
	fmt.Println("test handlers")
	defer fmt.Println("test finished")
	time.Sleep(time.Second * 2)
}
