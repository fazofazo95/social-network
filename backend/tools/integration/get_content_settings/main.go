package main

import (
	"context"
	"encoding/json"
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

	data, err := queries.GetUserContentSettings(context.Background(), database.DB, 1)
	if err != nil {
		log.Fatalf("get content settings failed: %v", err)
	}

	b, _ := json.MarshalIndent(data, "", "  ")
	fmt.Printf("content settings for user 1:\n%s\n", string(b))
}
