package main

import (
	"fmt"
	"os"

	seedpkg "backend/tools/seed"
)

func main() {
	userID := 1
	iterations := 4
	limit := 10

	fmt.Printf("Running DiscoverUsers for user %d, %d times (limit=%d)\n", userID, iterations, limit)
	if err := seedpkg.DebugDiscover(userID, iterations, limit); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
