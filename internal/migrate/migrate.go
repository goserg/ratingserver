package sqlite3

import (
	"database/sql"
	"errors"
	embedded "ratingserver"

	"github.com/golang-migrate/migrate/v4/database/sqlite3"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

func UpServerDB(db *sql.DB) error {
	sourceDriver, err := iofs.New(embedded.ServerMigrations, "migrations")
	if err != nil {
		return err
	}
	databaseDriver, err := sqlite3.WithInstance(db, &sqlite3.Config{})
	if err != nil {
		return err
	}
	m, err := migrate.NewWithInstance("iofs",
		sourceDriver,
		"rating", databaseDriver)
	if err != nil {
		return err
	}
	err = m.Up()
	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return err
	}
	return nil
}

func UpBotDB(db *sql.DB) error {
	sourceDriver, err := iofs.New(embedded.BotMigrations, "bot/migrations")
	if err != nil {
		return err
	}
	databaseDriver, err := sqlite3.WithInstance(db, &sqlite3.Config{})
	if err != nil {
		return err
	}
	m, err := migrate.NewWithInstance("iofs",
		sourceDriver,
		"bot", databaseDriver)
	if err != nil {
		return err
	}
	err = m.Up()
	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return err
	}
	return nil
}
