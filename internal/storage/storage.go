package storage

import "database/sql"

func New() (*sql.DB, error) {
	db, err := sql.Open("sqlite3", "file:rating.sqlite?cache=shared")
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)
	err = db.Ping()
	if err != nil {
		return nil, err
	}
	return db, nil
}
