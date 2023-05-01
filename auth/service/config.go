package service

type Config struct {
	SqliteFile   string `toml:"sqlite_file"`
	Token        string `toml:"token"`
	Expiration   string `toml:"expiration"`
	RootPassword string `toml:"root_password"`
	Roles        string `toml:"roles"` // TODO create roles on first start
}
