package config

import (
	"os"

	"github.com/BurntSushi/toml"
)

type TgBot struct {
	TelegramApiToken string `toml:"telegream_atitoken"`
}

type Server struct {
	TgBotEnabled bool `toml:"tg_bot_enabled"`
	Debug        bool `toml:"debug_mode"`
}

type Config struct {
	TgBot  TgBot
	Server Server
}

func New() (Config, error) {
	var tgBotCfg TgBot
	_, err := toml.DecodeFile("configs/bot.toml", &tgBotCfg)
	if err != nil {
		return Config{}, err
	}
	token := os.Getenv("TELEGRAM_APITOKEN")
	if token != "" {
		tgBotCfg.TelegramApiToken = token
	}

	var serverCfg Server
	_, err = toml.DecodeFile("configs/server.toml", &serverCfg)
	if err != nil {
		return Config{}, err
	}

	return Config{
		TgBot:  tgBotCfg,
		Server: serverCfg,
	}, nil
}
