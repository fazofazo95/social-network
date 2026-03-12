package main

import (
	"fmt"
	"os"
)

func usage() {
	fmt.Println("usage:")
	fmt.Println("  go run main.go           # start server")
	fmt.Println("  go run main.go server    # start server")
	fmt.Println("  go run main.go reset     # delete DB (destructive) and restart server")
	fmt.Println()
}

func main() {
	if len(os.Args) < 2 {
		runServer()
		return
	}

	usage()
}
