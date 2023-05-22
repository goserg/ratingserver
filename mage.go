//go:build mage

package main

import (
	"os"

	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
)

const (
	jetOutput                 = "gen"
	jetBotOutput              = "bot/gen"
	jetAuthOutput             = "auth/gen"
	sqliteRatingsFileLocation = "rating.sqlite"
	sqliteBotFileLocation     = "bot.sqlite"
	sqliteAuthFileLocation    = "auth.sqlite"
	serverBin                 = "./bin/server"
)

const (
	toolsDir     = "tools/"
	toolsModfile = toolsDir + "go.mod"
	toolsBinDir  = toolsDir + "bin/"
	lintTool     = toolsBinDir + "golangci-lint"
	jetTool      = toolsBinDir + "jet"
	atlasTool    = toolsBinDir + "atlas"
)

const (
	testServerConfigPath = "test_configs/server.toml"
	testBotConfigPath    = "test_configs/bot.toml"
)

func goModDownload() error {
	return sh.Run("go", "mod", "download")
}

// Build builds server binary
func Build() error {
	mg.Deps(goModDownload)
	mg.Deps(genJet)
	return sh.Run("go", "build", "-o", serverBin, "cmd/main.go")
}

// Run starts server
func Run() error {
	mg.Deps(Build)
	mg.Deps(atlasApply)
	return sh.Run(serverBin)
}

func genJet() error {
	mg.Deps(buildJetTool)
	if err := sh.Run(jetTool, "-source", "sqlite", "-dsn", sqliteRatingsFileLocation, "-path", jetOutput); err != nil {
		return err
	}
	if err := sh.Run(jetTool, "-source", "sqlite", "-dsn", sqliteAuthFileLocation, "-path", jetAuthOutput); err != nil {
		return err
	}
	if err := sh.Run(jetTool, "-source", "sqlite", "-dsn", sqliteBotFileLocation, "-path", jetBotOutput); err != nil {
		return err
	}
	if err := sh.Run(jetTool, "-source", "postgres", "-dsn", "postgres://postgres:postgres@localhost:5431/auth?sslmode=disable", "-path", "gen"); err != nil {
		return err
	}
	return nil
}

func buildJetTool() error {
	sh.RunWith(map[string]string{
		"CGO_ENABLED": "1",
	}, "go", "build", "-modfile", toolsModfile, "-o", jetTool, "github.com/go-jet/jet/v2/cmd/jet")
	return nil
}

func atlasApply() error {
	mg.Deps(buildToolsAtlas)
	return sh.Run(
		atlasTool, "schema", "apply",
		"--auto-approve",
		"-u", "postgres://postgres:postgres@localhost:5431/auth?sslmode=disable",
		"--to", "file://auth/migrations/init.hcl",
	)
}

func buildToolsAtlas() error {
	return sh.Run(
		"go", "build",
		"-modfile", toolsModfile,
		"-o", atlasTool,
		"ariga.io/atlas/cmd/atlas",
	)
}

func AtlasSchemaInspect() error {
	mg.Deps(buildToolsAtlas)
	initHcl, err := os.OpenFile("auth/migrations/init.hcl", os.O_RDWR|os.O_CREATE, 0o0755)
	if err != nil {
		return err
	}
	defer initHcl.Close()
	_, err = sh.Exec(nil, initHcl, nil,
		atlasTool, "schema", "inspect",
		"-u", "postgres://postgres:postgres@localhost:5431/auth?sslmode=disable",
	)
	if err != nil {
		return err
	}
	return nil
}

func Lint() error {
	mg.Deps(buildLintTool)
	return sh.Run(lintTool, "run", "./...")
}

func buildLintTool() error {
	return sh.Run(
		"go", "build",
		"-modfile", toolsModfile,
		"-o", lintTool,
		"github.com/golangci/golangci-lint/cmd/golangci-lint",
	)
}

func AutoTest() error {
	mg.Deps(Build)
	if err := sh.Run(
		"psql", "postgres://postgres:postgres@localhost:5431",
		"-c", "drop database if exists \"auth-test\";",
	); err != nil {
		return err
	}
	if err := sh.Run(
		"psql", "postgres://postgres:postgres@localhost:5431",
		"-c", "create database \"auth-test\";",
	); err != nil {
		return err
	}
	if err := sh.Run(
		atlasTool, "schema", "apply",
		"--auto-approve",
		"-u", `postgres://postgres:postgres@localhost:5431/auth-test?sslmode=disable`,
		"--to", "file://auth/migrations/init.hcl",
	); err != nil {
		return err
	}
	if err := os.Chdir("tests"); err != nil {
		return err
	}
	if err := sh.Run(
		"go", "test", "-v", "-server-config", testServerConfigPath, "-bot-config", testBotConfigPath, "./...",
	); err != nil {
		return err
	}
	return nil
}
