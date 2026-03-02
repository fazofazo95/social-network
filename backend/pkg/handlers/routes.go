package handlers

import (
	"net/http"

	"backend/pkg/middleware"
)

var auth = middleware.WithAuth

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
