package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	database "backend/pkg/db/sqlite"
	handlers "backend/pkg/handlers"
	"backend/pkg/middleware"
)

func runServer() {
	// 1️⃣ Initialize the database
	dbPath := "pkg/db/social_network.db"

	if err := os.MkdirAll("pkg/db", 0o755); err != nil {
		log.Fatalf("Failed to create db directory: %v", err)
	}

	err := database.Init(dbPath)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	fmt.Println("Database initialized successfully!")
	defer database.DB.Close()

	mux := http.NewServeMux()

	// API route(s)
	mux.HandleFunc("/api/signup", handlers.SignUpHandler)
	http.HandleFunc("/api/login", handlers.LogInHandler)

	http.HandleFunc("/api/logout", middleware.WithAuth(handlers.LogOutHandler))

	http.HandleFunc("/api/verify-session", handlers.VerifySession)
	http.HandleFunc("/api/follow", middleware.WithAuth(handlers.FollowRequestHandler))
	http.HandleFunc("/api/feed", middleware.WithAuth(handlers.FeedHandler))
	http.Handle("/uploads/", handlers.UploadsFileServer())

	port := 8080
	fmt.Printf("Server running on http://localhost:%d\n", port)
	if err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
