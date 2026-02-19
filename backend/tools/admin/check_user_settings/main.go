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

	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM user_settings;").Scan(&count)
	if err != nil {
		log.Fatalf("count query failed: %v", err)
	}
	fmt.Printf("user_settings rows = %d\n", count)

	rows, err := db.Query("SELECT id, CAST(COALESCE(email_vis,0) AS INTEGER), CAST(COALESCE(birthday_date_vis,1) AS INTEGER), CAST(COALESCE(relationship_status_vis,1) AS INTEGER), CAST(COALESCE(employed_at_vis,1) AS INTEGER), CAST(COALESCE(phone_number_vis,0) AS INTEGER), CAST(COALESCE(about_me_vis,1) AS INTEGER), CAST(COALESCE(nickname_vis,1) AS INTEGER) FROM user_settings ORDER BY id LIMIT 5;")
	if err != nil {
		log.Fatalf("select sample failed: %v", err)
	}
	defer rows.Close()

	fmt.Println("sample rows:")
	for rows.Next() {
		var id int
		var emailVis, bdVis, relVis, empVis, phoneVis, aboutVis, nickVis int
		if err := rows.Scan(&id, &emailVis, &bdVis, &relVis, &empVis, &phoneVis, &aboutVis, &nickVis); err != nil {
			log.Fatalf("scan failed: %v", err)
		}
		fmt.Printf("id=%d email_vis=%v birthday_date_vis=%v relationship_status_vis=%v employed_at_vis=%v phone_number_vis=%v about_me_vis=%v nickname_vis=%v\n",
			id, emailVis, bdVis, relVis, empVis, phoneVis, aboutVis, nickVis)
	}
	if err := rows.Err(); err != nil {
		log.Fatalf("rows error: %v", err)
	}

	fmt.Println("done")
}
