package admin

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	database "backend/pkg/db/sqlite"

	_ "github.com/mattn/go-sqlite3"
)

const dbPath = "pkg/db/social_network.db"

func openDB() (*sql.DB, error) {
	abs, _ := filepath.Abs(dbPath)
	return sql.Open("sqlite3", abs)
}

// ShowUsers prints up to 100 rows from login_users to stdout. Returns any error encountered.
func ShowUsers() error {
	db, err := openDB()
	if err != nil {
		return fmt.Errorf("open db: %w", err)
	}
	defer db.Close()

	rows, err := db.Query(`SELECT id, username, email, created_at FROM login_users ORDER BY id DESC LIMIT 100;`)
	if err != nil {
		return fmt.Errorf("query: %w", err)
	}
	defer rows.Close()

	fmt.Println("id | username | email | created_at")
	for rows.Next() {
		var id int
		var username, email, createdAt string
		if err := rows.Scan(&id, &username, &email, &createdAt); err != nil {
			return fmt.Errorf("scan: %w", err)
		}
		fmt.Printf("%d | %s | %s | %s\n", id, username, email, createdAt)
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("rows err: %w", err)
	}
	return nil
}

// WipeUsers deletes all rows from login_users and users. If confirm is false, the function returns without making changes.
func WipeUsers(confirm bool) error {
	if !confirm {
		return errors.New("wipe not confirmed")
	}

	db, err := openDB()
	if err != nil {
		return fmt.Errorf("open db: %w", err)
	}
	defer db.Close()

	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("begin: %w", err)
	}

	if _, err := tx.Exec("DELETE FROM login_users;"); err != nil {
		tx.Rollback()
		return fmt.Errorf("delete login_users: %w", err)
	}
	if _, err := tx.Exec("DELETE FROM users;"); err != nil {
		tx.Rollback()
		return fmt.Errorf("delete users: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit: %w", err)
	}

	if _, err := db.Exec("VACUUM;"); err != nil {
		log.Printf("vacuum failed: %v", err)
	}

	fmt.Println("Wipe complete.")
	return nil
}

// ResetDB removes the DB file and re-runs migrations. The caller must ensure the server is stopped.
func ResetDB(force bool) error {
	if !force {
		return errors.New("reset-db is destructive; pass force=true to proceed")
	}

	abs, _ := filepath.Abs(dbPath)
	if err := os.Remove(abs); err != nil {
		if !os.IsNotExist(err) {
			return fmt.Errorf("remove db: %w", err)
		}
	}
	fmt.Printf("Removed DB file: %s\n", abs)

	// Re-run migrations to recreate the DB
	if err := database.Init(dbPath); err != nil {
		return fmt.Errorf("re-init db: %w", err)
	}
	fmt.Println("Database reinitialized and migrations applied.")
	return nil
}

// WipeAllData deletes all rows from every non-sqlite_ table while keeping schema intact.
// It temporarily disables FK checks to allow deleting in arbitrary order, then vacuums.
func WipeAllData(confirm bool) error {
	if !confirm {
		return errors.New("wipe not confirmed")
	}

	db, err := openDB()
	if err != nil {
		return fmt.Errorf("open db: %w", err)
	}
	defer db.Close()

	// Turn off foreign keys for the connection so deletes don't fail due to constraints.
	if _, err := db.Exec("PRAGMA foreign_keys = OFF;"); err != nil {
		return fmt.Errorf("disable foreign_keys: %w", err)
	}

	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("begin: %w", err)
	}

	rows, err := tx.Query("SELECT name FROM sqlite_master WHERE type='table' AND name NOT LIKE 'sqlite_%';")
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("list tables: %w", err)
	}
	defer rows.Close()

	var tables []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			tx.Rollback()
			return fmt.Errorf("scan table name: %w", err)
		}
		tables = append(tables, name)
	}
	if err := rows.Err(); err != nil {
		tx.Rollback()
		return fmt.Errorf("table rows err: %w", err)
	}

	for _, t := range tables {
		// Safety: skip any sqlite_ internal tables and migration bookkeeping table(s)
		if strings.HasPrefix(t, "sqlite_") {
			continue
		}
		if t == "schema_migrations" {
			// golang-migrate stores applied migration versions here; don't remove
			continue
		}
		// Quote table name by doubling quotes to be safe
		q := strings.ReplaceAll(t, "\"", "\"\"")
		if _, err := tx.Exec(fmt.Sprintf("DELETE FROM \"%s\";", q)); err != nil {
			tx.Rollback()
			return fmt.Errorf("delete table %s: %w", t, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit: %w", err)
	}

	// Reset sqlite_sequence to make AUTOINCREMENT start over (ignore if not present)
	if _, err := db.Exec("DELETE FROM sqlite_sequence;"); err != nil {
		if !strings.Contains(err.Error(), "no such table") {
			log.Printf("warning: could not clear sqlite_sequence: %v", err)
		}
	}

	// Re-enable foreign keys and vacuum
	if _, err := db.Exec("PRAGMA foreign_keys = ON;"); err != nil {
		log.Printf("warning: enable foreign_keys failed: %v", err)
	}
	if _, err := db.Exec("VACUUM;"); err != nil {
		log.Printf("vacuum failed: %v", err)
	}

	fmt.Println("Wipe-all complete: all table rows deleted.")
	return nil
}
