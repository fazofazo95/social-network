package main

import (
	"fmt"
	"log"

	"backend/tools/seed"
)

func main() {
	fmt.Println("Seeding database with test users...")
	
	fmt.Println("seed: starting signup seeding")
	count, err := seed.SeedFromJSON("tools/seed/signup_seed.json")
	if err != nil {
		log.Fatalf("Seeding failed: %v", err)
	}
	fmt.Printf("seed: Successfully seeded %d users\n", count)

	fmt.Println("seed: starting profile seeding")
	pcount, perr := seed.SeedProfilesFromJSON("tools/seed/user_seed.json")
	if perr != nil {
		log.Fatalf("Failed to seed profiles: %v", perr)
	}
	fmt.Printf("seed: Profiles created/updated: %d\n", pcount)

	fmt.Println("seed: starting followers seeding")
	fcount, ferr := seed.SeedFollowersFromJSON("tools/seed/followers_seed.json")
	if ferr != nil {
		log.Fatalf("Failed to seed followers: %v", ferr)
	}
	fmt.Printf("seed: Followers created: %d\n", fcount)
}
