package database

import (
	"database/sql"

	_ "github.com/mattn/go-sqlite3"
)

var DB *sql.DB

func Init(dbPath string) error {
	var err error

	DB, err = sql.Open("sqlite3", dbPath)
	if err != nil {
		return err
	}

	// SQLite best practice
	if _, err := DB.Exec(`PRAGMA foreign_keys = ON;`); err != nil {
		return err
	}

	_, err = DB.Exec("PRAGMA busy_timeout=5000;")
    if err != nil {
        return err
    }

	_, err = DB.Exec("PRAGMA journal_mode=WAL;")
    if err != nil {
        return  err
    }

	DB.SetMaxOpenConns(1)

	return runMigrations(DB)
}
