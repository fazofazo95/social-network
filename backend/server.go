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
	handlers.UserRoutes(mux)
	handlers.AuthRoutes(mux)
	handlers.PostRoutes(mux)
	handlers.CommentRoutes(mux)
	handlers.FeedRoutes(mux)

	fs := http.FileServer(http.Dir("./uploads"))
	mux.Handle("GET /uploads/", http.StripPrefix("/uploads/", fs))

	handlerWithCORS := middleware.CorsMiddleware(mux)

	port := 8080
	fmt.Printf("Server running on http://localhost:%d\n", port)
	if err := http.ListenAndServe(fmt.Sprintf(":%d", port), handlerWithCORS); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
