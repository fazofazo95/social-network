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

	rows, err := db.Query("SELECT version, dirty FROM schema_migrations LIMIT 1;")
	if err != nil {
		log.Fatalf("query schema_migrations: %v", err)
	}
	defer rows.Close()

	var version int
	var dirty int
	if rows.Next() {
		if err := rows.Scan(&version, &dirty); err != nil {
			log.Fatalf("scan: %v", err)
		}
		fmt.Printf("schema_migrations: version=%d dirty=%d\n", version, dirty)
	} else {
		log.Fatalf("schema_migrations table empty or missing")
	}

	if dirty == 0 {
		fmt.Println("Not dirty, nothing to do")
		return
	}

	// Set to previous successful version (18) and clear dirty flag
	target := 18
	res, err := db.Exec("UPDATE schema_migrations SET version = ?, dirty = 0;", target)
	if err != nil {
		log.Fatalf("update schema_migrations: %v", err)
	}
	n, _ := res.RowsAffected()
	fmt.Printf("Updated schema_migrations rows=%d -> version=%d dirty=0\n", n, target)
}
