package embedded

import "embed"

//go:embed "views"
var Views embed.FS

//go:embed "migrations"
var ServerMigrations embed.FS

//go:embed "bot/migrations"
var BotMigrations embed.FS
