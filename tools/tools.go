//go:build tools
// +build tools

package main

import (
	_ "github.com/go-jet/jet/v2/cmd/jet"
	_ "github.com/golangci/golangci-lint/cmd/golangci-lint"
)
