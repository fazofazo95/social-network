package main

import (
	"context"
	"fmt"
	"log"

	queries "backend/pkg/db/queries"
	database "backend/pkg/db/sqlite"
	models "backend/pkg/models"
)

func main() {
	ctx := context.Background()
	dbPath := "pkg/db/social_network.db"
	if err := database.Init(dbPath); err != nil {
		log.Fatalf("init db: %v", err)
	}
	defer database.DB.Close()

	// login as alice
	loginInput := models.LoginRequest{Email: "alice@example.com", Password: "Password123!"}
	userID, err := queries.LogIn(ctx, database.DB, loginInput)
	if err != nil {
		log.Fatalf("login failed: %v", err)
	}
	fmt.Printf("Logged in as userID=%d\n", userID)

	// unfollow target id 5
	targetID := 5
	rows, err := queries.DeleteFollow(ctx, database.DB, userID, targetID)
	if err != nil {
		log.Fatalf("delete follow failed: %v", err)
	}
	fmt.Printf("DeleteFollow affected rows=%d\n", rows)

	// show remaining followers rows for this follower
	fmt.Println("Remaining follows for follower:")
	rows2, err := database.DB.QueryContext(ctx, "SELECT follower_id, followed_id, status FROM followers WHERE follower_id = ?", userID)
	if err != nil {
		log.Fatalf("query failed: %v", err)
	}
	defer rows2.Close()
	for rows2.Next() {
		var f, t int
		var s string
		if err := rows2.Scan(&f, &t, &s); err != nil {
			log.Fatalf("scan failed: %v", err)
		}
		fmt.Printf("%d -> %d status=%s\n", f, t, s)
	}
}
