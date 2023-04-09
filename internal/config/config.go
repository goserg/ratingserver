package config

import "os"

type TgBot struct {
	TelegramApiToken string
}

func New() TgBot {
	token := os.Getenv("TELEGRAM_APITOKEN")
	return TgBot{
		TelegramApiToken: token,
	}
}
