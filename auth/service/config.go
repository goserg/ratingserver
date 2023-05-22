package service

import "time"

type Config struct {
	SqliteFile     string   `toml:"sqlite_file"`
	Token          string   `toml:"token"`
	Expiration     string   `toml:"expiration"`
	RootPassword   string   `toml:"root_password"`
	PasswordPepper string   `toml:"password_pepper"`
	Roles          []string `toml:"roles"` // TODO create roles on first start
	Rules          []struct {
		Name   string   `toml:"name"`
		Path   string   `toml:"path"`
		Method []string `toml:"method"`
		Allow  []string `toml:"allow"`
		Order  int      `toml:"order"`
	} `toml:"rules"`
	TokenTTL time.Duration `toml:"tokenTTL"`
	Storage  StorageConfig `toml:"db"`
}

type StorageConfig struct {
	Host     string `toml:"host"`
	Port     int    `toml:"port"`
	DBName   string `toml:"dbname"`
	Username string `toml:"username"`
	Password string `toml:"password"`
}
