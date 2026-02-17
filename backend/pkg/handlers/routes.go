// handlers.UserRoutes(mux)
// handlers.AuthRoutes(mux)
// handlers.PostRoutes(mux)
// handlers.CommentRoutes(mux)
// handlers.FollowRoutes(mux)

package handlers

import (
	"backend/pkg/middleware"
	"database/sql"
	"net/http"
)

var auth = middleware.WithAuth

func UserRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/signup", CreateUserHandler)
	mux.Handle("GET /api/users/{id}", middleware.Chain(GetUserHandler, auth))
	mux.Handle("PUT /api/users/{id}", middleware.Chain(UpdateUserHandler, auth))
	mux.Handle("GET /api/users/{id}/posts", middleware.Chain(GetUserPostsHandler, auth))
	mux.Handle("POST /api/users/{id}/follow", middleware.Chain(FollowUserHandler, auth))
	mux.Handle("DELETE /api/users/{id}/unfollow", middleware.Chain(UnfollowUserHandler, auth))
	mux.Handle("POST /api/users/{id}/block", middleware.Chain(BlockUserHandler, auth))
	mux.Handle("DELETE /api/users/{id}/unblock", middleware.Chain(UnblockUserHandler, auth))
	// follow list endpoints for the authenticated user
	mux.Handle("GET /api/users/following", middleware.Chain(FollowingHandler, auth))
	mux.Handle("GET /api/users/followers", middleware.Chain(FollowersHandler, auth))
	mux.Handle("GET /api/users/blocked", middleware.Chain(BlockedHandler, auth))
	mux.Handle("GET /api/users/pending", middleware.Chain(PendingRequestsHandler, auth))
}

func AuthRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/login", LogInHandler)
	mux.Handle("DELETE /api/logout", middleware.Chain(LogOutHandler, auth))
	mux.HandleFunc("GET /api/verify-session", VerifySession)
}

func FeedRoutes(mux *http.ServeMux) {
	mux.Handle("GET /api/feed", middleware.Chain(FeedHandler, auth))
	mux.Handle("GET /api/discover", middleware.Chain(DiscoverHandler, auth))
}

func FollowRoutes(mux *http.ServeMux) {
	mux.Handle("POST /api/users/{id}/follow", middleware.Chain(FollowUserHandler, auth))
	mux.Handle("DELETE /api/users/{id}/unfollow", middleware.Chain(UnfollowUserHandler, auth))
	mux.Handle("POST /api/users/{id}/follow/accept", middleware.Chain(AcceptFollowHandler, auth))
}

func PostRoutes(mux *http.ServeMux, db *sql.DB) {
	checkOwner := middleware.OwnershipMiddleware(db, "posts")

	mux.Handle("POST /api/posts", middleware.Chain(CreatePostHandler, auth))

	mux.Handle("PUT /api/posts/{id}", middleware.Chain(UpdatePostHandler, auth, checkOwner))
	mux.Handle("PUT /api/posts/{id}/delete", middleware.Chain(DeletePostHandler, auth, checkOwner))
	mux.Handle("PUT /api/posts/{id}/restore", middleware.Chain(RestorePostHandler, auth, checkOwner))

	mux.Handle("GET /api/posts/{id}", middleware.Chain(GetPostHandler, auth))
}

func CommentRoutes(mux *http.ServeMux, db *sql.DB) {
	auth := middleware.WithAuth
	checkOwner := middleware.OwnershipMiddleware(db, "comments")

	mux.Handle("GET /api/posts/{id}/comments", middleware.Chain(GetPostCommentsHandler, auth))
	mux.Handle("POST /api/comments", middleware.Chain(CreateCommentHandler, auth))
	mux.Handle("PUT /api/comments/{id}", middleware.Chain(UpdateCommentHandler, auth, checkOwner))
	mux.Handle("DELETE /api/comments/{id}", middleware.Chain(DeleteCommentHandler, auth, checkOwner))

	mux.Handle("PUT /api/comments/{id}/restore", middleware.Chain(RestoreCommentHandler, auth, checkOwner))
}
