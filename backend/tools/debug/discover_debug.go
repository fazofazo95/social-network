package main

import (
	"context"
	"encoding/json"
	"fmt"

	queries "backend/pkg/db/queries"
	database "backend/pkg/db/sqlite"
)

func main() {
    dbPath := "pkg/db/social_network.db"
    if err := database.Init(dbPath); err != nil {
        fmt.Printf("init db error: %v\n", err)
        return
    }

    users, err := queries.DiscoverUsers(context.Background(), database.DB, 3, 5)
    if err != nil {
        fmt.Printf("DiscoverUsers error: %v\n", err)
        return
    }

    b, _ := json.MarshalIndent(users, "", "  ")
    fmt.Println(string(b))
}
