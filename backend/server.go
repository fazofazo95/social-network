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

	if err := os.MkdirAll("pkg/db", 0755); err != nil {
		log.Fatalf("Failed to create db directory: %v", err)
	}

	err := database.Init(dbPath)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	fmt.Println("Database initialized successfully!")
	defer database.DB.Close()

	// Serve the frontend static files from ../frontend at the root path.
	fs := http.FileServer(http.Dir("../frontend"))
	http.Handle("/", fs)

	// API route(s)
	http.HandleFunc("/api/signup", handlers.SignUpHandler)
	http.HandleFunc("/api/login", handlers.LogInHandler)
	http.HandleFunc("/api/verify-session", handlers.VerifySession)
	http.HandleFunc("/api/logout", middleware.WithAuth(handlers.LogOutHandler))
	http.HandleFunc("/api/follow", middleware.WithAuth(handlers.FollowRequestHandler))
	http.Handle("/uploads/", handlers.UploadsFileServer())

	// Convenience: make /signup (browser) show the frontend signup page.
	http.HandleFunc("/signup", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "../frontend/index.html")
	})
	// http.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
	// 	http.ServeFile(w, r, "../frontend/index.html")
	// })

	port := 8080
	fmt.Printf("Server running on http://localhost:%d\n", port)
	if err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
