package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	db, err := sql.Open("sqlite3", "pkg/db/social_network.db")
	if err != nil {
		log.Fatalf("open db: %v", err)
	}
	defer db.Close()

	rows, err := db.Query(`
		SELECT
			u.id,
			u.Followers,
			u.Following,
			(SELECT COUNT(*) FROM followers f WHERE f.followed_id = u.id AND f.status = 'accepted') AS expected_followers,
			(SELECT COUNT(*) FROM followers f WHERE f.follower_id = u.id AND f.status = 'accepted') AS expected_following
		FROM users u
		ORDER BY u.id;
	`)
	if err != nil {
		log.Fatalf("query failed: %v", err)
	}
	defer rows.Close()

	fmt.Println("id | Followers | Following | expected_followers | expected_following")
	for rows.Next() {
		var id, followers, following, expFollowers, expFollowing int
		if err := rows.Scan(&id, &followers, &following, &expFollowers, &expFollowing); err != nil {
			log.Fatalf("scan failed: %v", err)
		}
		fmt.Printf("%d | %d | %d | %d | %d\n", id, followers, following, expFollowers, expFollowing)
	}
	if err := rows.Err(); err != nil {
		log.Fatalf("rows error: %v", err)
	}
}
