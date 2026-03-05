package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	queries "backend/pkg/db/queries"
	database "backend/pkg/db/sqlite"
	models "backend/pkg/models"
)

func main() {
	dbPath := "pkg/db/social_network.db"
	if err := database.Init(dbPath); err != nil {
		log.Fatalf("init db: %v", err)
	}
	defer database.DB.Close()

	ctx := context.Background()
	// login as alice
	userID, err := queries.LogIn(ctx, database.DB, models.LoginRequest{Email: "alice@example.com", Password: "Password123!"})
	if err != nil {
		log.Fatalf("login failed: %v", err)
	}
	fmt.Printf("Logged in user %d\n", userID)

	settings, err := queries.GetUserVisibilitySettings(ctx, database.DB, userID)
	if err != nil {
		log.Fatalf("get settings failed: %v", err)
	}
	b, _ := json.MarshalIndent(settings, "", "  ")
	fmt.Printf("visibility settings:\n%s\n", string(b))
}
