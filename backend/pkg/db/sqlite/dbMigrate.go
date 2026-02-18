package database

import (
	"database/sql"
	"log"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite3"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func runMigrations(db *sql.DB) error {
	log.Println("[INFO] runMigrations: Starting database migrations")

	driver, err := sqlite3.WithInstance(db, &sqlite3.Config{})
	if err != nil {
		log.Printf("[ERROR] runMigrations: Failed to create driver instance: %v", err)
		return err
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://pkg/db/migrations",
		"sqlite3",
		driver,
	)
	if err != nil {
		log.Printf("[ERROR] runMigrations: Failed to initialize migrate instance: %v", err)
		return err
	}

	log.Println("[INFO] runMigrations: Checking for pending migrations")
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		log.Printf("[ERROR] runMigrations: Failed to apply migrations: %v", err)
		return err
	}

	if err == migrate.ErrNoChange {
		log.Println("[INFO] runMigrations: Database is already up to date (no changes)")
	} else {
		log.Println("[SUCCESS] runMigrations: Migrations applied successfully")
	}

	return nil
}
