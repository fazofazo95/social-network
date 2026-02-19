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
	loginInput := models.LoginInput{Email: "alice@example.com", Password: "Password123!"}
	userID, err := queries.LogIn(ctx, database.DB, loginInput)
	if err != nil {
		log.Fatalf("login failed: %v", err)
	}
	fmt.Printf("Logged in user %d\n", userID)

	// Update all visibility flags to visible (1)
	one := 1
	if err := queries.UpdateUserVisibilitySettings(ctx, database.DB, userID,
		&one, // email_vis
		&one, // birthday_date_vis
		&one, // relationship_status_vis
		&one, // employed_at_vis
		&one, // phone_number_vis
		&one, // about_me_vis
		&one, // nickname_vis
		&one, // follow_vis
		nil,  // profile_type (nil = no change)
	); err != nil {
		log.Fatalf("update settings failed: %v", err)
	}
	fmt.Println("Updated all visibility flags -> visible")

	// Read back via GetUserVisibilitySettings
	settings, err := queries.GetUserVisibilitySettings(ctx, database.DB, userID)
	if err != nil {
		log.Fatalf("get settings failed: %v", err)
	}
	b, _ := json.MarshalIndent(settings, "", "  ")
	fmt.Printf("visibility settings:\n%s\n", string(b))

	// Show raw DB value
	var phone int
	err = database.DB.QueryRowContext(ctx, "SELECT COALESCE(phone_number_vis,0) FROM user_settings WHERE id = ?", userID).Scan(&phone)
	if err != nil {
		log.Fatalf("db query failed: %v", err)
	}
	fmt.Printf("DB raw phone_number_vis = %d\n", phone)
}
