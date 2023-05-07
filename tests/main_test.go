package main

import (
	"os"
	"testing"

	"github.com/stretchr/testify/suite"
)

func TestMain(m *testing.M) {
	os.Exit(m.Run())
}

func Test1(t *testing.T) {
	t.Log("start autotests")
	suite.Run(t, &TestSuite1{})
}
