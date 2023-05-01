package config

import (
	"errors"
	"io"
	"os"
	embedded "ratingserver"
	authservice "ratingserver/auth/service"

	"github.com/BurntSushi/toml"
)

const (
	botCfgName       = "bot.toml"
	serverCftName    = "server.toml"
	defaultEmbedPath = "default/"
	cftFolder        = "configs/"
)

type TgBot struct {
	TelegramApiToken string `toml:"telegram_apitoken"`
	SqliteFile       string `toml:"sqlite_file"`
	AdminPass        string `toml:"admin_pass"`
}

type Server struct {
	TgBotDisable bool               `toml:"disable_tg_bot"`
	Debug        bool               `toml:"debug_mode"`
	SqliteFile   string             `toml:"sqlite_file"`
	Auth         authservice.Config `toml:"auth"`
}

type Config struct {
	TgBot  TgBot
	Server Server
}

func New() (Config, error) {
	err := createCfgFolderIfNotExists()
	if err != nil {
		return Config{}, err
	}

	serverCfg, err := serverConfig()
	if err != nil {
		return Config{}, err
	}

	var tgBotCfg TgBot
	if !serverCfg.TgBotDisable {
		tgBotCfg, err = tgBotConfig()
		if err != nil {
			return Config{}, err
		}
	}

	return Config{
		TgBot:  tgBotCfg,
		Server: serverCfg,
	}, nil
}

func createCfgFolderIfNotExists() error {
	_, err := os.Stat(cftFolder)
	if err == nil {
		return nil
	}
	if !errors.Is(err, os.ErrNotExist) {
		return err
	}
	return os.Mkdir(cftFolder, 0750)
}

func serverConfig() (Server, error) {
	err := createConfigFileIfNotExists(serverCftName)
	if err != nil {
		return Server{}, err
	}
	var serverCfg Server
	_, err = toml.DecodeFile(cftFolder+serverCftName, &serverCfg)
	if err != nil {
		return Server{}, err
	}
	return serverCfg, nil
}

func tgBotConfig() (TgBot, error) {
	err := createConfigFileIfNotExists(botCfgName)
	if err != nil {
		return TgBot{}, err
	}
	var tgBotCfg TgBot
	_, err = toml.DecodeFile(cftFolder+botCfgName, &tgBotCfg)
	if err != nil {
		return TgBot{}, err
	}
	token := os.Getenv("TELEGRAM_APITOKEN")
	if token != "" {
		tgBotCfg.TelegramApiToken = token
	}
	return tgBotCfg, err
}

func createConfigFileIfNotExists(filename string) error {
	_, err := os.Stat(cftFolder + filename)
	if err == nil {
		return nil
	}
	if !errors.Is(err, os.ErrNotExist) {
		return err
	}
	defaultCfg, err := embedded.DefaultConfigs.Open(defaultEmbedPath + filename)
	if err != nil {
		return err
	}
	defer defaultCfg.Close()
	newCfg, err := os.Create(cftFolder + filename)
	if err != nil {
		return err
	}
	defer newCfg.Close()
	_, err = io.Copy(newCfg, defaultCfg)
	if err != nil {
		return err
	}
	return nil
}
