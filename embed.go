package embedded

import "embed"

//go:embed "views"
var Views embed.FS

//go:embed "migrations"
var ServerMigrations embed.FS

//go:embed "bot/migrations"
var BotMigrations embed.FS

//go:embed "auth/migrations"
var AuthMigrations embed.FS

//go:embed "configs_default"
var DefaultConfigs embed.FS

//go:embed "static"
var WebStatic embed.FS
