package config

import (
	"errors"
	"os"

	"github.com/BurntSushi/toml"
)

const botCfgPath = "configs/bot2.toml"

type TgBot struct {
	TelegramApiToken string `toml:"telegream_atitoken"`
}

const serverCftPath = "configs/server.toml"

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
	_, err := toml.DecodeFile(serverCftPath, &serverCfg)
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			if err != nil {
				return Server{}, err
			}
		}
		_, err := os.Create(serverCftPath)
		if err != nil {
			return Server{}, err
		}
	}
	return serverCfg, nil
}

func tgBotConfig() (TgBot, error) {
	var tgBotCfg TgBot
	_, err := toml.DecodeFile(botCfgPath, &tgBotCfg)
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			if err != nil {
				return TgBot{}, err
			}
		}
		_, err := os.Create(botCfgPath)
		if err != nil {
			return TgBot{}, err
		}
	}
	token := os.Getenv("TELEGRAM_APITOKEN")
	if token != "" {
		tgBotCfg.TelegramApiToken = token
	}
	return tgBotCfg, err
}
