package main

import (
	"fmt"
	"log"

	"backend/tools/seed"
)

func main() {
	fmt.Println("Seeding database with test data (users, profiles, followers)...")

	users, profiles, followers, err := seed.SeedAll("tools/seed/signup_seed.json", "tools/seed/user_seed.json", "tools/seed/followers_seed.json")
	if err != nil {
		log.Fatalf("Seeding failed: %v", err)
	}

	fmt.Printf("seed: completed: users=%d profiles=%d followers=%d\n", users, profiles, followers)
}
