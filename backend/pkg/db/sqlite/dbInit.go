package database

import (
	"database/sql"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

var DB *sql.DB

func Init(dbPath string) error {
	log.Printf("[INFO] Init: Opening database at path: %s", dbPath)
	var err error

	DB, err = sql.Open("sqlite3", dbPath)
	if err != nil {
		log.Printf("[ERROR] Init: Failed to open database: %v", err)
		return err
	}

	log.Printf("[INFO] Init: Setting PRAGMA foreign_keys = ON")
	if _, err := DB.Exec(`PRAGMA foreign_keys = ON;`); err != nil {
		log.Printf("[ERROR] Init: Failed to set foreign_keys: %v", err)
		return err
	}

	log.Printf("[INFO] Init: Setting PRAGMA busy_timeout=5000")
	_, err = DB.Exec("PRAGMA busy_timeout=5000;")
	if err != nil {
		log.Printf("[ERROR] Init: Failed to set busy_timeout: %v", err)
		return err
	}

	log.Printf("[INFO] Init: Setting PRAGMA journal_mode=WAL")
	_, err = DB.Exec("PRAGMA journal_mode=WAL;")
	if err != nil {
		log.Printf("[ERROR] Init: Failed to set journal_mode: %v", err)
		return err
	}

	log.Printf("[INFO] Init: Setting MaxOpenConns(1) for SQLite consistency")
	DB.SetMaxOpenConns(1)

	log.Printf("[INFO] Init: Running database migrations")
	err = runMigrations(DB)
	if err != nil {
		log.Printf("[ERROR] Init: Migrations failed: %v", err)
		return err
	}

	log.Printf("[SUCCESS] Init: Database initialized and ready")
	return nil
}
