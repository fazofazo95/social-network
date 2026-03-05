package main

import (
	"context"
	"fmt"
	"log"

	queries "backend/pkg/db/queries"
	database "backend/pkg/db/sqlite"
	models "backend/pkg/models"
	repository "backend/pkg/repository"
	services "backend/pkg/services"
)

func main() {
	dbPath := "pkg/db/social_network.db"
	if err := database.Init(dbPath); err != nil {
		log.Fatalf("init db: %v", err)
	}
	defer database.DB.Close()

	ctx := context.Background()

	// login as alice
	email := "alice@example.com"
	password := "Password123!"
	loginInput := models.LoginRequest{Email: email, Password: password}
	userID, err := queries.LogIn(ctx, database.DB, loginInput)
	if err != nil {
		log.Fatalf("login failed: %v", err)
	}
	fmt.Printf("Logged in: userID=%d\n", userID)

	followRepo := repository.NewFollowRepository(database.DB)
	profileRepo := repository.NewProfileRepository(database.DB)
	followSvc := services.NewFollowService(followRepo, profileRepo, nil)
	targets := []int{5, 6}
	for _, t := range targets {
		req := models.FollowRequest{FollowerID: userID, FollowedID: t}
		status, err := followSvc.FollowUser(ctx, req)
		if err != nil {
			fmt.Printf("Follow %d -> %d error: %v\n", userID, t, err)
			continue
		}
		fmt.Printf("Follow %d -> %d status=%s\n", userID, t, status)
	}
}
