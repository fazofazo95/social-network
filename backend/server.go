package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	database "backend/pkg/db/sqlite"
	handlers "backend/pkg/handlers"
	"backend/pkg/middleware"
	"backend/pkg/repository"
	"backend/pkg/services"
	"backend/pkg/sse"
	websocket "backend/pkg/ws"
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

	defer database.DB.Close()

	mux := http.NewServeMux()

	hub := websocket.NewHub()
	notificationHub := sse.NewHub()
	go hub.Run()

	authRepo := repository.NewAuthRepository(database.DB)
	authServ := services.NewAuthService(authRepo)
	authHandl := handlers.NewAuthHandler(authServ)
	authHandl.RegisterRoutes(mux)

	mux.Handle("/ws", middleware.Chain((http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		websocket.ServeWs(hub, w, r)
	})), middleware.WithAuth))

	chatRepo := repository.NewChatRepository(database.DB)
	chatServ := services.NewChatService(chatRepo, hub)
	chatHandl := handlers.NewChatHandler(chatServ)
	chatHandl.RegisterRoutes(mux)

	postRepo := repository.NewPostRepository(database.DB)
	postServ := services.NewPostService(postRepo)
	postHandl := handlers.NewPostHandler(postServ)
	postHandl.RegisterRoutes(mux)

	commentRepo := repository.NewCommentRepository(database.DB)
	commentServ := services.NewCommentService(commentRepo, postRepo)
	commentHandl := handlers.NewCommentHandler(commentServ)
	commentHandl.RegisterRoutes(mux)

	profileRepo := repository.NewProfileRepository(database.DB)
	profileServ := services.NewProfileService(profileRepo)
	profileHandl := handlers.NewProfileHandler(profileServ, postServ)
	profileHandl.RegisterRoutes(mux)

	settingsHandl := handlers.NewSettingsHandler(profileServ)
	settingsHandl.RegisterRoutes(mux)

	followRepo := repository.NewFollowRepository(database.DB)
	notificationRepo := repository.NewNotificationRepository(database.DB)
	notificationServ := services.NewNotificationService(notificationRepo, notificationHub)

	followServ := services.NewFollowService(followRepo, profileRepo, notificationServ)
	followHandl := handlers.NewFollowHandler(followServ)
	followHandl.RegisterRoutes(mux)

	relationHandl := handlers.NewRelationsHandler(followServ)
	relationHandl.RegisterRoutes(mux)

	reactionRepo := repository.NewReactionRepository(database.DB)
	reactionServ := services.NewReactionService(reactionRepo)
	reactionHandl := handlers.NewReactionHandler(reactionServ)
	reactionHandl.RegisterRoutes(mux)

	feedHandl := handlers.NewFeedHandler(postServ, profileServ)
	feedHandl.RegisterRoutes(mux)

	groupRepo := repository.NewGroupRepository(database.DB)
	groupServ := services.NewGroupService(groupRepo, notificationServ)
	groupHandl := handlers.NewGroupHandler(groupServ)
	groupHandl.RegisterRoutes(mux)

	notificationHandl := handlers.NewNotificationHandler(notificationServ, followServ, groupServ, notificationHub)
	notificationHandl.RegisterRoutes(mux)

	fs := http.FileServer(http.Dir("./uploads"))
	mux.Handle("GET /uploads/", http.StripPrefix("/uploads/", fs))

	handlerWithCORS := middleware.CorsMiddleware(mux)

	port := 8080
	fmt.Printf("Server running on http://localhost:%d\n", port)
	if err := http.ListenAndServe(fmt.Sprintf(":%d", port), handlerWithCORS); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
