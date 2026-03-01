// handlers.UserRoutes(mux)
// handlers.AuthRoutes(mux)
// handlers.PostRoutes(mux)
// handlers.CommentRoutes(mux)
// handlers.FollowRoutes(mux)

package handlers

import (
	"net/http"

	"backend/pkg/middleware"
)

var auth = middleware.WithAuth

func SettingsRoutes(mux *http.ServeMux){
	mux.Handle("GET /api/users/settings", middleware.Chain(GetVisibilitySettingsHandler, auth))
	mux.Handle("PATCH /api/users/settings", middleware.Chain(UpdateVisibilitySettingsHandler, auth))
	mux.Handle("PUT /api/users/settings", middleware.Chain(UpdateVisibilitySettingsHandler, auth))
	mux.Handle("GET /api/users/settings/content", middleware.Chain(GetUserContentSettingsHandler, auth))
	mux.Handle("PATCH /api/users/settings/content", middleware.Chain(UpdateUserContentSettingsHandler, auth))
	mux.Handle("PUT /api/users/settings/content", middleware.Chain(UpdateUserContentSettingsHandler, auth))
}

func RelationRoutes(mux *http.ServeMux){
	mux.Handle("GET /api/users/following", middleware.Chain(FollowingHandler, auth))
	mux.Handle("GET /api/users/followers", middleware.Chain(FollowersHandler, auth))
	mux.Handle("GET /api/users/{id}/following", middleware.Chain(FollowingByUserHandler, auth))
	mux.Handle("GET /api/users/{id}/followers", middleware.Chain(FollowersByUserHandler, auth))
	mux.Handle("GET /api/users/blocked", middleware.Chain(BlockedHandler, auth))
	mux.Handle("GET /api/users/pending", middleware.Chain(PendingRequestsHandler, auth))
}

func ProfileRoutes(mux *http.ServeMux){
	mux.Handle("GET /api/users/{id}", middleware.Chain(GetUserHandler, auth))
	mux.Handle("PUT /api/users/{id}", middleware.Chain(UpdateUserHandler, auth))
	mux.Handle("GET /api/users/{id}/posts", middleware.Chain(GetUserPostsHandler, auth))
}

// func AuthRoutes(mux *http.ServeMux) {
// 	mux.HandleFunc("POST /api/users", CreateUserHandler)
// 	mux.HandleFunc("POST /api/login", LogInHandler)
// 	mux.Handle("DELETE /api/logout", middleware.Chain(LogOutHandler, auth))
// 	mux.HandleFunc("GET /api/verify-session", VerifySession)
// }

func FeedRoutes(mux *http.ServeMux) {
	mux.Handle("GET /api/feed", middleware.Chain(GetFeedHandler, auth))
	mux.Handle("GET /api/discover", middleware.Chain(DiscoverHandler, auth))
}

func FollowRoutes(mux *http.ServeMux) {
	mux.Handle("POST /api/users/{id}/follow", middleware.Chain(FollowUserHandler, auth))
	mux.Handle("DELETE /api/users/{id}/unfollow", middleware.Chain(UnfollowUserHandler, auth))
	mux.Handle("DELETE /api/users/{id}/remove-follower", middleware.Chain(RemoveFollowerHandler, auth))
	mux.Handle("POST /api/users/{id}/follow/accept", middleware.Chain(AcceptFollowHandler, auth))
	mux.Handle("DELETE /api/users/{id}/follow/reject", middleware.Chain(RejectFollowHandler, auth))
	mux.Handle("POST /api/users/{id}/block", middleware.Chain(BlockUserHandler, auth))
	mux.Handle("DELETE /api/users/{id}/unblock", middleware.Chain(UnblockUserHandler, auth))
}

func GroupRoutes(mux *http.ServeMux) {
	mux.Handle("POST /api/groups", middleware.Chain(CreateGroupHandler, auth))
	mux.Handle("GET /api/groups/active", middleware.Chain(ActiveGroupsHandler, auth))
	mux.Handle("GET /api/groups/requests/pending", middleware.Chain(PendingMyGroupRequestsHandler, auth))
	mux.Handle("GET /api/groups/invites/pending", middleware.Chain(PendingMyGroupInvitesHandler, auth))
	mux.Handle("GET /api/groups/discover", middleware.Chain(DiscoverGroupsHandler, auth))
	mux.Handle("GET /api/groups/{id}", middleware.Chain(GetGroupPageHandler, auth))
	mux.Handle("POST /api/groups/{id}/events", middleware.Chain(CreateGroupEventHandler, auth))
	mux.Handle("GET /api/groups/{id}/events/{event_id}/inviteable", middleware.Chain(GroupEventInviteableMembersHandler, auth))
	mux.Handle("POST /api/groups/{id}/events/{event_id}/invites/all", middleware.Chain(InviteAllGroupEventMembersHandler, auth))
	mux.Handle("POST /api/groups/{id}/events/{event_id}/invites/{user_id}", middleware.Chain(InviteGroupEventMemberHandler, auth))
	mux.Handle("POST /api/groups/{id}/events/{event_id}/respond", middleware.Chain(RespondGroupEventInviteHandler, auth))
	mux.Handle("PATCH /api/groups/{id}/events/{event_id}/respond", middleware.Chain(ChangeGroupEventResponseHandler, auth))
	mux.Handle("DELETE /api/groups/{id}/events/{event_id}", middleware.Chain(CancelGroupEventHandler, auth))
	mux.Handle("POST /api/groups/{id}/join", middleware.Chain(RequestToJoinGroupHandler, auth))
	mux.Handle("POST /api/groups/{id}/chat/messages", middleware.Chain(SendGroupMessageHandler, auth))
	mux.Handle("POST /api/groups/{id}/invite/{user_id}", middleware.Chain(InviteToGroupHandler, auth))
	mux.Handle("POST /api/groups/{id}/invite/accept", middleware.Chain(AcceptInviteHandler, auth))
	mux.Handle("POST /api/groups/{id}/invite/reject", middleware.Chain(RejectInviteHandler, auth))
	mux.Handle("POST /api/groups/{id}/members/{user_id}/kick", middleware.Chain(KickMemberHandler, auth))
	mux.Handle("POST /api/groups/{id}/members/{user_id}/promote", middleware.Chain(PromoteModeratorHandler, auth))
	mux.Handle("POST /api/groups/{id}/members/{user_id}/demote", middleware.Chain(DemoteModeratorHandler, auth))
	mux.Handle("GET /api/groups/{id}/members", middleware.Chain(GroupMembersListHandler, auth))
	mux.Handle("GET /api/groups/{id}/requests/pending", middleware.Chain(PendingGroupRequestsHandler, auth))
	mux.Handle("GET /api/groups/{id}/invites/pending", middleware.Chain(PendingGroupInvitesHandler, auth))
	mux.Handle("DELETE /api/groups/{id}/invites/{user_id}", middleware.Chain(RemoveInviteHandler, auth))
	mux.Handle("DELETE /api/groups/{id}/requests/me", middleware.Chain(RemoveOwnRequestHandler, auth))
	mux.Handle("GET /api/groups/{id}/settings", middleware.Chain(GetGroupSettingsHandler, auth))
	mux.Handle("PATCH /api/groups/{id}/settings", middleware.Chain(UpdateGroupSettingsHandler, auth))
	mux.Handle("PUT /api/groups/{id}/settings", middleware.Chain(UpdateGroupSettingsHandler, auth))
	mux.Handle("POST /api/groups/{id}/leave", middleware.Chain(LeaveGroupHandler, auth))
	mux.Handle("POST /api/groups/{id}/requests/{user_id}/accept", middleware.Chain(AcceptGroupRequestHandler, auth))
	mux.Handle("POST /api/groups/{id}/requests/{user_id}/reject", middleware.Chain(RejectGroupRequestHandler, auth))
	mux.Handle("DELETE /api/groups/{id}", middleware.Chain(DeleteGroupHandler, auth))
}

func ReactionRoutes(mux *http.ServeMux) {
	mux.Handle("POST /api/posts/{id}/reactions", middleware.Chain(AddReactionHandler, auth))
	mux.Handle("DELETE /api/posts/{id}/reactions", middleware.Chain(RemoveReactionHandler, auth))
}

// func ChatRoutes(mux *http.ServeMux, hub *websocket.Hub) {
// 	mux.Handle("/ws", middleware.Chain((http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 		ws.ServeWs(hub, w, r)
// 	})), auth))
// 	mux.Handle("GET /api/chats", middleware.Chain(ListChatsHandler, auth))
// 	mux.Handle("GET /api/chats/{chat_id}/messages", middleware.Chain(GetChatMessagesHandler, auth))
// 	mux.Handle("POST /api/chats/direct/{user_id}/messages", middleware.Chain(SendDirectMessageHandler(hub), auth))
// 	mux.Handle("POST /api/chats/{chat_id}/read", middleware.Chain(MarkChatReadHandler, auth))
// }

// func PostRoutes(mux *http.ServeMux, db *sql.DB) {
// 	checkOwner := middleware.OwnershipMiddleware(db, "posts")

// 	mux.Handle("POST /api/posts", middleware.Chain(CreatePostHandler, auth))

// 	mux.Handle("PUT /api/posts/{id}", middleware.Chain(UpdatePostHandler, auth, checkOwner))
// 	mux.Handle("DELETE /api/posts/{id}", middleware.Chain(DeletePostHandler, auth, checkOwner))
// 	mux.Handle("PUT /api/posts/{id}/restore", middleware.Chain(RestorePostHandler, auth, checkOwner))

// 	mux.Handle("GET /api/posts/{id}", middleware.Chain(GetPostHandler, auth))
// }

// func CommentRoutes(mux *http.ServeMux, db *sql.DB) {
// 	checkOwner := middleware.OwnershipMiddleware(db, "comments")

// 	mux.Handle("GET /api/posts/{id}/comments", middleware.Chain(GetPostCommentsHandler, auth))

// 	mux.Handle("POST /api/comments", middleware.Chain(CreateCommentHandler, auth))
// 	mux.Handle("PUT /api/comments/{id}", middleware.Chain(UpdateCommentHandler, auth, checkOwner))
// 	mux.Handle("PUT /api/comments/{id}/delete", middleware.Chain(DeleteCommentHandler, auth, checkOwner))
// 	mux.Handle("PUT /api/comments/{id}/restore", middleware.Chain(RestoreCommentHandler, auth, checkOwner))
// }
