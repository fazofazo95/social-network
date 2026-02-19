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

	var version int
	var dirty int
	err = db.QueryRow("SELECT version, dirty FROM schema_migrations").Scan(&version, &dirty)
	if err != nil {
		log.Fatalf("query schema_migrations: %v", err)
	}

	fmt.Printf("schema_migrations: version=%d dirty=%d\n", version, dirty)
}
