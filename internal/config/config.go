package config

import (
	"errors"
	"io"
	"os"
	configsdefault "ratingserver/configs_default"

	"github.com/BurntSushi/toml"
)

const (
	botCfgName       = "bot.toml"
	serverCftName    = "server.toml"
	defaultEmbedPath = "default/"
	cftFolder        = "configs/"
)

type TgBot struct {
	TelegramApiToken string `toml:"telegream_atitoken"`
}

type Server struct {
	TgBotDisable bool `toml:"disable_tg_bot"`
	Debug        bool `toml:"debug_mode"`
}

type Config struct {
	TgBot  TgBot
	Server Server
}

func New() (Config, error) {
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

func serverConfig() (Server, error) {
	var serverCfg Server
	_, err := os.Stat(cftFolder + serverCftName)
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			if err != nil {
				return Server{}, err
			}
		}
		defaultCfg, err := configsdefault.Files.Open(defaultEmbedPath + serverCftName)
		if err != nil {
			return Server{}, err
		}
		defer defaultCfg.Close()
		newCfg, err := os.Create(cftFolder + serverCftName)
		if err != nil {
			return Server{}, err
		}
		defer newCfg.Close()
		_, err = io.Copy(newCfg, defaultCfg)
		if err != nil {
			return Server{}, err
		}
	}
	_, err = toml.DecodeFile(cftFolder+serverCftName, &serverCfg)
	if err != nil {
		return Server{}, err
	}
	return serverCfg, nil
}

func tgBotConfig() (TgBot, error) {
	var tgBotCfg TgBot
	_, err := os.Stat(cftFolder + botCfgName)
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			if err != nil {
				return TgBot{}, err
			}
		}
		defaultCfg, err := configsdefault.Files.Open(defaultEmbedPath + botCfgName)
		if err != nil {
			return TgBot{}, err
		}
		defer defaultCfg.Close()
		newCfg, err := os.Create(cftFolder + botCfgName)
		if err != nil {
			return TgBot{}, err
		}
		defer newCfg.Close()
		_, err = io.Copy(newCfg, defaultCfg)
		if err != nil {
			return TgBot{}, err
		}
	}
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
