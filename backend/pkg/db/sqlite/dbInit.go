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

	return runMigrations(DB)
}
