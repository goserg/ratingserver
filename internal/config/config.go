package config

import (
	"errors"
	"flag"
	"io"
	"os"
	"sort"

	"github.com/BurntSushi/toml"
	embedded "github.com/goserg/ratingserver"
	authservice "github.com/goserg/ratingserver/auth/service"
)

const (
	botCfgName       = "bot.toml"
	serverCfgName    = "server.toml"
	defaultEmbedPath = "default/"
	cfgFolder        = "configs/"
)

type TgBot struct {
	TelegramAPIToken string `toml:"telegram_apitoken"`
	SqliteFile       string `toml:"sqlite_file"`
	AdminPass        string `toml:"admin_pass"`
}

type Server struct {
	TgBotDisable bool               `toml:"disable_tg_bot"`
	Debug        bool               `toml:"debug_mode"`
	SqliteFile   string             `toml:"sqlite_file"`
	Host         string             `toml:"host"`
	Port         int                `toml:"port"`
	Auth         authservice.Config `toml:"auth"`
}

type Config struct {
	TgBot  TgBot
	Server Server
}

var (
	ServerConfigPath string
	BotConfigPath    string
)

func init() {
	flag.StringVar(&ServerConfigPath, "server-config", cfgFolder+serverCfgName, "server config path")
	flag.StringVar(&BotConfigPath, "bot-config", cfgFolder+botCfgName, "bot config path")
}

func New() (Config, error) {
	flag.Parse()
	err := createCfgFolderIfNotExists()
	if err != nil {
		return Config{}, err
	}

	serverCfg, err := serverConfig(ServerConfigPath)
	if err != nil {
		return Config{}, err
	}

	var tgBotCfg TgBot
	if !serverCfg.TgBotDisable {
		tgBotCfg, err = tgBotConfig(BotConfigPath)
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
	_, err := os.Stat(cfgFolder)
	if err == nil {
		return nil
	}
	if !errors.Is(err, os.ErrNotExist) {
		return err
	}
	return os.Mkdir(cfgFolder, 0o750)
}

func serverConfig(path string) (Server, error) {
	if path != "" {
		var serverCfg Server
		_, err := toml.DecodeFile(path, &serverCfg)
		if err != nil {
			return Server{}, err
		}
		sort.SliceStable(serverCfg.Auth.Rules, func(i, j int) bool {
			return serverCfg.Auth.Rules[i].Order < serverCfg.Auth.Rules[j].Order
		})
		return serverCfg, nil
	}
	err := createConfigFileIfNotExists(serverCfgName)
	if err != nil {
		return Server{}, err
	}
	var serverCfg Server
	_, err = toml.DecodeFile(cfgFolder+serverCfgName, &serverCfg)
	if err != nil {
		return Server{}, err
	}
	sort.SliceStable(serverCfg.Auth.Rules, func(i, j int) bool {
		return serverCfg.Auth.Rules[i].Order < serverCfg.Auth.Rules[j].Order
	})
	return serverCfg, nil
}

func tgBotConfig(path string) (TgBot, error) {
	if path != "" {
		var tgBotCfg TgBot
		_, err := toml.DecodeFile(path, &tgBotCfg)
		if err != nil {
			return TgBot{}, err
		}
		token := os.Getenv("TELEGRAM_APITOKEN")
		if token != "" {
			tgBotCfg.TelegramAPIToken = token
		}
		return tgBotCfg, err
	}
	err := createConfigFileIfNotExists(botCfgName)
	if err != nil {
		return TgBot{}, err
	}
	var tgBotCfg TgBot
	_, err = toml.DecodeFile(cfgFolder+botCfgName, &tgBotCfg)
	if err != nil {
		return TgBot{}, err
	}
	token := os.Getenv("TELEGRAM_APITOKEN")
	if token != "" {
		tgBotCfg.TelegramAPIToken = token
	}
	return tgBotCfg, err
}

func createConfigFileIfNotExists(filename string) error {
	_, err := os.Stat(cfgFolder + filename)
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
	newCfg, err := os.Create(cfgFolder + filename)
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
