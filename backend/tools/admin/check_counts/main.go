package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	dbPath := "pkg/db/social_network.db"
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		log.Fatalf("open db: %v", err)
	}
	defer db.Close()

	var loginCount, usersCount, settingsCount int

	err = db.QueryRow("SELECT COUNT(*) FROM login_users;").Scan(&loginCount)
	if err != nil {
		log.Fatalf("count login_users failed: %v", err)
	}
	err = db.QueryRow("SELECT COUNT(*) FROM users;").Scan(&usersCount)
	if err != nil {
		log.Fatalf("count users failed: %v", err)
	}
	err = db.QueryRow("SELECT COUNT(*) FROM user_settings;").Scan(&settingsCount)
	if err != nil {
		log.Fatalf("count user_settings failed: %v", err)
	}

	fmt.Printf("counts -> login_users=%d users=%d user_settings=%d\n", loginCount, usersCount, settingsCount)

	// show missing user ids: users that don't have settings
	rows, err := db.Query(`SELECT u.id FROM users u LEFT JOIN user_settings s ON u.id = s.id WHERE s.id IS NULL ORDER BY u.id LIMIT 20;`)
	if err != nil {
		log.Fatalf("query missing settings failed: %v", err)
	}
	defer rows.Close()

	fmt.Println("users without settings (sample up to 20):")
	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err != nil {
			log.Fatalf("scan failed: %v", err)
		}
		fmt.Printf("%d ", id)
	}
	fmt.Println()

	if err := rows.Err(); err != nil {
		log.Fatalf("rows error: %v", err)
	}

	fmt.Println("done")
}
