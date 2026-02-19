package main

import (
	"context"
	"fmt"
	"log"

	queries "backend/pkg/db/queries"
	database "backend/pkg/db/sqlite"
)

func main() {
	if err := database.Init("pkg/db/social_network.db"); err != nil {
		log.Fatalf("init db: %v", err)
	}
	defer database.DB.Close()

	ctx := context.Background()
	userID := 1
	one := 1

	if err := queries.UpdateUserVisibilitySettings(
		ctx,
		database.DB,
		userID,
		&one, // email_vis
		&one, // birthday_date_vis
		&one, // relationship_status_vis
		&one, // employed_at_vis
		&one, // phone_number_vis
		&one, // about_me_vis
		&one, // nickname_vis
		&one, // follow_vis
		&one, // profile_type
	); err != nil {
		log.Fatalf("update visibility/profile_type failed: %v", err)
	}

	fmt.Println("updated user 1: all visibility fields=1 and profile_type=1")
}
