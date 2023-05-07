package main

import (
	"context"
	"fmt"
	"net"
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
	p := NewProcess(context.Background(), "../bin/server")
	suite.process = p
	err := p.Start(context.Background())
	if err != nil {
		suite.T().Errorf("cant start process: %v", err)
	}
	t := time.NewTicker(time.Second / 2)
	//wait for start
	for {
		select {
		case <-t.C:
			suite.T().Logf("test connection")
			conn, _ := net.DialTimeout("http", "127.0.0.1:3300", time.Second/4)
			if conn != nil {
				conn.Close()
				return
			}
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
	out := suite.process.stdout.String()
	suite.T().Logf("out: %s", out)

	stderr := suite.process.stdout.String()
	suite.T().Logf("err: %s", stderr)

	suite.T().Logf("process finished with code %d", exitCode)
}

func (suite *TestSuite1) TestHandlers() {
	fmt.Println("test handlers")
	defer fmt.Println("test finished")
	time.Sleep(time.Second * 2)
}
