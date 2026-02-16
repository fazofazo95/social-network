// handlers.UserRoutes(mux)
// handlers.AuthRoutes(mux)
// handlers.PostRoutes(mux)
// handlers.CommentRoutes(mux)
// handlers.FollowRoutes(mux)

package handlers

import (
	"backend/pkg/middleware"
	"net/http"
)

func UserRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/signup", CreateUserHandler)
	mux.HandleFunc("GET /api/users/{id}", middleware.WithAuth(GetUserHandler))
	mux.HandleFunc("PUT /api/users/{id}", middleware.WithAuth(UpdateUserHandler))
	mux.HandleFunc("POST /api/users/{id}/follow", middleware.WithAuth(FollowUserHandler))
	mux.HandleFunc("DELETE /api/users/{id}/unfollow", middleware.WithAuth(UnfollowUserHandler))
}

func AuthRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/login", LogInHandler)
	mux.HandleFunc("DELETE /api/logout", middleware.WithAuth(LogOutHandler))
	mux.HandleFunc("GET /api/verify-session", VerifySession)
}

func FeedRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/feed", middleware.WithAuth(FeedHandler))
}

func PostRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/posts", middleware.WithAuth(CreatePostHandler))
	mux.HandleFunc("PUT /api/posts/{id}", middleware.WithAuth(UpdatePostHandler))
	mux.HandleFunc("DELETE /api/posts/{id}", middleware.WithAuth(DeletePostHandler))
	mux.HandleFunc("PUT /api/posts/{id}/restore", middleware.WithAuth(RestorePostHandler))
	mux.HandleFunc("GET /api/posts/{id}", middleware.WithAuth(GetPostHandler))
}

func CommentRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/comments", middleware.WithAuth(CreateCommentHandler))
	mux.HandleFunc("PUT /api/comments/{id}", middleware.WithAuth(UpdateCommentHandler))
	mux.HandleFunc("DELETE /api/comments/{id}", middleware.WithAuth(DeleteCommentHandler))
	mux.HandleFunc("PUT /api/comments/{id}/restore", middleware.WithAuth(RestoreCommentHandler))
	mux.HandleFunc("GET /api/posts/{id}/comments", middleware.WithAuth(GetPostCommentsHandler))
}
