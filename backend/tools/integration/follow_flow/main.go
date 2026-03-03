package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	queries "backend/pkg/db/queries"
	database "backend/pkg/db/sqlite"
	models "backend/pkg/models"
	repository "backend/pkg/repository"
	services "backend/pkg/services"
)

func main() {
	// init DB
	dbPath := "pkg/db/social_network.db"
	if err := database.Init(dbPath); err != nil {
		log.Fatalf("init db: %v", err)
	}
	defer database.DB.Close()

	ctx := context.Background()

	// 1) login as alice
	email := "alice@example.com"
	password := "Password123!"

	loginInput := models.LoginRequest{Email: email, Password: password}
	userID, err := queries.LogIn(ctx, database.DB, loginInput)
	if err != nil {
		log.Fatalf("login failed: %v", err)
	}
	fmt.Printf("Logged in: userID=%d\n", userID)

	// create session (not strictly needed for service calls but show it)
	sess, err := queries.CreateSession(ctx, database.DB, userID)
	if err != nil {
		log.Fatalf("create session failed: %v", err)
	}
	fmt.Printf("Created session: %s\n", sess)

	// 2) discover users
	users, err := queries.DiscoverUsers(ctx, database.DB, userID, 10)
	if err != nil {
		log.Fatalf("discover failed: %v", err)
	}
	b, _ := json.MarshalIndent(users, "", "  ")
	fmt.Printf("Discovered users:\n%s\n", string(b))

	// 3) send follow requests to discovered ids
	followRepo := repository.NewFollowRepository(database.DB)
	profileRepo := repository.NewProfileRepository(database.DB)
	followSvc := services.NewFollowService(followRepo, profileRepo, nil)
	for _, u := range users {
		req := models.FollowRequest{FollowerID: userID, FollowedID: u.ID}
		status, err := followSvc.FollowUser(ctx, req)
		if err != nil {
			fmt.Printf("Follow %d -> %d error: %v\n", userID, u.ID, err)
			continue
		}
		fmt.Printf("Follow %d -> %d status=%s\n", userID, u.ID, status)
	}
}
