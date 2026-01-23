package main

import (
	"fmt"
	"log"
	"os"

	adminpkg "backend/tools/admin"
	seedpkg "backend/tools/seed"
)

func usage() {
	fmt.Println("usage:")
	fmt.Println("  go run main.go           # start server")
	fmt.Println("  go run main.go server    # start server")
	fmt.Println("  go run main.go reset     # delete DB (destructive) and restart server")
	fmt.Println("  go run main.go populate  # fill DB with seed data and exit")
	fmt.Println()
}

func main() {
	if len(os.Args) < 2 {
		// default: start server
		runServer()
		return
	}

	switch os.Args[1] {
	case "server", "run":
		runServer()
	case "reset":
		fmt.Println("Resetting DB (destructive)...")
		if err := adminpkg.ResetDB(true); err != nil {
			log.Fatalf("reset failed: %v", err)
		}
	case "populate":
		fmt.Println("Seeding DB...")
		created, err := seedpkg.SeedFromJSON("tools/seed/signup_seed.json")
		if err != nil {
			log.Fatalf("seeding failed: %v", err)
		}
		fmt.Printf("done. created %d users\n", created)
	default:
		usage()
	}
}
