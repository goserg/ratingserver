package sqlite

import (
	"database/sql"
	"github.com/sirupsen/logrus"
	"ratingserver/auth/storage"
	"ratingserver/internal/config"
	sqlite3 "ratingserver/internal/migrate"
)

type Storage struct {
	db  *sql.DB
	log *logrus.Entry
}

var _ storage.AuthStorage = (*Storage)(nil)

func New(l *logrus.Logger, cfg config.Server) (*Storage, error) {
	log := l.WithFields(map[string]interface{}{
		"from": "auth-storage",
	})
	db, err := sql.Open("sqlite3", buildSource(cfg.Auth.SqliteFile))
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)

	err = sqlite3.UpAuthDB(db)
	if err != nil {
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		return nil, err
	}
	log.Info("auth storage connected")
	return &Storage{
		db:  db,
		log: log,
	}, nil
}

func buildSource(fileName string) string {
	return "file:" + fileName + "?cache=shared"
}
