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

	if err := queries.RebuildAllFollowCounts(context.Background(), database.DB); err != nil {
		log.Fatalf("rebuild follow counts: %v", err)
	}

	fmt.Println("rebuild follow counts: done")
}
