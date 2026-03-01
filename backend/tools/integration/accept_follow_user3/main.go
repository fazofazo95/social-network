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

	// login as carol (user 3)
	loginInput := models.LoginRequest{Email: "carol@example.com", Password: "Password123!"}
	userID, err := queries.LogIn(ctx, database.DB, loginInput)
	if err != nil {
		log.Fatalf("login failed: %v", err)
	}
	fmt.Printf("Logged in as userID=%d\n", userID)

	// accept pending follow from followerID=1 -> followedID=userID(3)
	followerID := 1
	followedID := userID
	updated, err := queries.AcceptFollow(ctx, database.DB, followerID, followedID)
	if err != nil {
		log.Fatalf("AcceptFollow failed: %v", err)
	}
	fmt.Printf("AcceptFollow updated rows=%d\n", updated)

	// show the specific relationship
	row := database.DB.QueryRowContext(ctx, "SELECT follower_id, followed_id, status FROM followers WHERE follower_id = ? AND followed_id = ?", followerID, followedID)
	var f, t int
	var s string
	err = row.Scan(&f, &t, &s)
	if err != nil {
		fmt.Printf("relationship not found or error: %v\n", err)
	} else {
		fmt.Printf("relationship: %d -> %d status=%s\n", f, t, s)
	}
}
