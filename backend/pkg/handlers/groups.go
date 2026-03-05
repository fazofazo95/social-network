package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"

	"backend/pkg/middleware"
	"backend/pkg/models"
	"backend/pkg/responses"
	"backend/pkg/services"
	"backend/pkg/utils"
)

type GroupHandler struct {
	Service services.GroupService
}

func NewGroupHandler(s services.GroupService) *GroupHandler {
	return &GroupHandler{Service: s}
}

func (h *GroupHandler) RegisterRoutes(mux *http.ServeMux) {
	auth := middleware.WithAuth

	mux.Handle("POST /api/groups", middleware.Chain(h.CreateGroup, auth))
	mux.Handle("GET /api/groups/active", middleware.Chain(h.ActiveGroups, auth))
	mux.Handle("GET /api/groups/requests/pending", middleware.Chain(h.PendingMyGroupRequests, auth))
	mux.Handle("GET /api/groups/invites/pending", middleware.Chain(h.PendingMyGroupInvites, auth))
	mux.Handle("GET /api/groups/discover", middleware.Chain(h.DiscoverGroups, auth))
	mux.Handle("GET /api/groups/{id}", middleware.Chain(h.GetGroupPage, auth))
	mux.Handle("POST /api/groups/{id}/events", middleware.Chain(h.CreateGroupEvent, auth))
	mux.Handle("GET /api/groups/{id}/events/{event_id}/inviteable", middleware.Chain(h.GroupEventInviteableMembers, auth))
	mux.Handle("POST /api/groups/{id}/events/{event_id}/invites/all", middleware.Chain(h.InviteAllGroupEventMembers, auth))
	mux.Handle("POST /api/groups/{id}/events/{event_id}/invites/{user_id}", middleware.Chain(h.InviteGroupEventMember, auth))
	mux.Handle("POST /api/groups/{id}/events/{event_id}/respond", middleware.Chain(h.RespondGroupEventInvite, auth))
	mux.Handle("PATCH /api/groups/{id}/events/{event_id}/respond", middleware.Chain(h.ChangeGroupEventResponse, auth))
	mux.Handle("DELETE /api/groups/{id}/events/{event_id}", middleware.Chain(h.CancelGroupEvent, auth))
	mux.Handle("POST /api/groups/{id}/join", middleware.Chain(h.RequestToJoinGroup, auth))
	mux.Handle("POST /api/groups/{id}/invite/{user_id}", middleware.Chain(h.InviteToGroup, auth))
	mux.Handle("POST /api/groups/{id}/invite/accept", middleware.Chain(h.AcceptInvite, auth))
	mux.Handle("POST /api/groups/{id}/invite/reject", middleware.Chain(h.RejectInvite, auth))
	mux.Handle("POST /api/groups/{id}/members/{user_id}/kick", middleware.Chain(h.KickMember, auth))
	mux.Handle("POST /api/groups/{id}/members/{user_id}/promote", middleware.Chain(h.PromoteModerator, auth))
	mux.Handle("POST /api/groups/{id}/members/{user_id}/demote", middleware.Chain(h.DemoteModerator, auth))
	mux.Handle("GET /api/groups/{id}/members", middleware.Chain(h.GroupMembersList, auth))
	mux.Handle("GET /api/groups/{id}/requests/pending", middleware.Chain(h.PendingGroupRequests, auth))
	mux.Handle("GET /api/groups/{id}/invites/pending", middleware.Chain(h.PendingGroupInvites, auth))
	mux.Handle("DELETE /api/groups/{id}/invites/{user_id}", middleware.Chain(h.RemoveInvite, auth))
	mux.Handle("DELETE /api/groups/{id}/requests/me", middleware.Chain(h.RemoveOwnRequest, auth))
	mux.Handle("GET /api/groups/{id}/settings", middleware.Chain(h.GetGroupSettings, auth))
	mux.Handle("PATCH /api/groups/{id}/settings", middleware.Chain(h.UpdateGroupSettings, auth))
	mux.Handle("PUT /api/groups/{id}/settings", middleware.Chain(h.UpdateGroupSettings, auth))
	mux.Handle("POST /api/groups/{id}/leave", middleware.Chain(h.LeaveGroup, auth))
	mux.Handle("POST /api/groups/{id}/requests/{user_id}/accept", middleware.Chain(h.AcceptGroupRequest, auth))
	mux.Handle("POST /api/groups/{id}/requests/{user_id}/reject", middleware.Chain(h.RejectGroupRequest, auth))
	mux.Handle("DELETE /api/groups/{id}", middleware.Chain(h.DeleteGroup, auth))
}

// CreateGroup creates a new group
func (h *GroupHandler) CreateGroup(w http.ResponseWriter, r *http.Request) {
	ownerID, err := middleware.UserIDFromContext(r.Context())
	if err != nil {
		responses.SendError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	if err := r.ParseMultipartForm(20 << 20); err != nil {
		log.Printf("[ERROR] CreateGroup: ParseMultipartForm failed: %v", err)
		responses.SendError(w, http.StatusBadRequest, "Invalid Form")
		return
	}

	var in models.CreateGroupInput
	in.Name = r.FormValue("name")
	in.Description = r.FormValue("description")
	in.Visibility = r.FormValue("visibility")
	in.JoinMode = r.FormValue("join_mode")

	imageURL, err := utils.AttachGroupImage(r)
	if err != nil {
		log.Printf("[ERROR] CreateGroup: File attachment failed: %v", err)
		responses.SendError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if imageURL != "" {
		in.Picture = imageURL
	}

	group, err := h.Service.CreateGroup(r.Context(), ownerID, in)
	if err != nil {
		switch err.Error() {
		case "group name already in use":
			responses.SendError(w, http.StatusConflict, "group name already in use")
		default:
			responses.SendError(w, http.StatusInternalServerError, "failed to create group: "+err.Error())
		}
		return
	}

	responses.SendCreated(w, "group created successfully", group)
}

// DeleteGroup deletes a group
func (h *GroupHandler) DeleteGroup(w http.ResponseWriter, r *http.Request) {
	requesterID, err := middleware.UserIDFromContext(r.Context())
	if err != nil {
		responses.SendError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	groupID, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || groupID <= 0 {
		responses.SendError(w, http.StatusBadRequest, "invalid group id")
		return
	}

	err = h.Service.DeleteGroup(r.Context(), requesterID, groupID)
	if err != nil {
		switch err.Error() {
		case "group not found":
			responses.SendError(w, http.StatusNotFound, "group not found")
		case "only the group owner can delete the group":
			responses.SendError(w, http.StatusForbidden, "only the group owner can delete this group")
		default:
			responses.SendError(w, http.StatusInternalServerError, "failed to delete group: "+err.Error())
		}
		return
	}

	responses.SendSuccess(w, "group deleted successfully", map[string]int{"group_id": groupID})
}

// RequestToJoinGroup requests to join a group
func (h *GroupHandler) RequestToJoinGroup(w http.ResponseWriter, r *http.Request) {
	requesterID, err := middleware.UserIDFromContext(r.Context())
	if err != nil {
		responses.SendError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	groupID, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || groupID <= 0 {
		responses.SendError(w, http.StatusBadRequest, "invalid group id")
		return
	}

	membershipStatus, err := h.Service.RequestToJoinGroup(r.Context(), requesterID, groupID)
	if err != nil {
		switch err.Error() {
		case "group not found":
			responses.SendError(w, http.StatusNotFound, "group not found")
		case "cannot join private group from this endpoint":
			responses.SendError(w, http.StatusForbidden, "cannot request to join a private group")
		case "group is invite-only":
			responses.SendError(w, http.StatusForbidden, "group is invite-only")
		case "user is already a group member":
			responses.SendError(w, http.StatusConflict, "already a group member")
		case "user has already requested to join":
			responses.SendError(w, http.StatusConflict, "join request already exists")
		default:
			responses.SendError(w, http.StatusInternalServerError, "failed to process join request: "+err.Error())
		}
		return
	}

	message := "join request submitted"
	if membershipStatus == "active" {
		message = "joined group successfully"
	}

	responses.SendSuccess(w, message, map[string]interface{}{
		"group_id":          groupID,
		"user_id":           requesterID,
		"membership_status": membershipStatus,
	})
}

// DiscoverGroups discovers public groups
func (h *GroupHandler) DiscoverGroups(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.UserIDFromContext(r.Context())
	if err != nil {
		responses.SendError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	page := 1
	if pageStr := strings.TrimSpace(r.URL.Query().Get("page")); pageStr != "" {
		p, err := strconv.Atoi(pageStr)
		if err != nil || p <= 0 {
			responses.SendError(w, http.StatusBadRequest, "page must be a positive integer")
			return
		}
		page = p
	}

	const limit = 10
	offset := (page - 1) * limit

	items, err := h.Service.DiscoverGroups(r.Context(), userID, limit, offset)
	if err != nil {
		responses.SendError(w, http.StatusInternalServerError, "failed to discover groups: "+err.Error())
		return
	}

	responses.SendSuccess(w, "discover groups", map[string]interface{}{
		"page":  page,
		"limit": limit,
		"items": items,
	})
}

// CreateGroupEvent creates a group event
func (h *GroupHandler) CreateGroupEvent(w http.ResponseWriter, r *http.Request) {
	actorID, err := middleware.UserIDFromContext(r.Context())
	if err != nil {
		responses.SendError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	groupID, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || groupID <= 0 {
		responses.SendError(w, http.StatusBadRequest, "invalid group id")
		return
	}

	var in models.GroupEventCreateInput
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		responses.SendError(w, http.StatusBadRequest, "invalid JSON")
		return
	}

	created, err := h.Service.CreateGroupEvent(r.Context(), actorID, groupID, in)
	if err != nil {
		switch err.Error() {
		case "group not found":
			responses.SendError(w, http.StatusNotFound, "group not found")
		case "only group owner or moderators can approve requests":
			responses.SendError(w, http.StatusForbidden, "only group owner or moderators can create events")
		default:
			responses.SendError(w, http.StatusInternalServerError, "failed to create group event: "+err.Error())
		}
		return
	}

	responses.SendCreated(w, "group event created", created)
}

// GroupEventInviteableMembers returns members that can be invited to an event
func (h *GroupHandler) GroupEventInviteableMembers(w http.ResponseWriter, r *http.Request) {
	actorID, err := middleware.UserIDFromContext(r.Context())
	if err != nil {
		responses.SendError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	groupID, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || groupID <= 0 {
		responses.SendError(w, http.StatusBadRequest, "invalid group id")
		return
	}

	eventID, err := strconv.Atoi(r.PathValue("event_id"))
	if err != nil || eventID <= 0 {
		responses.SendError(w, http.StatusBadRequest, "invalid event id")
		return
	}

	members, err := h.Service.GetGroupEventInviteableMembers(r.Context(), actorID, groupID, eventID)
	if err != nil {
		switch err.Error() {
		case "user is not an active group member":
			responses.SendError(w, http.StatusForbidden, "only active group members can access")
			return
		case "user is not invited or responded to event":
			responses.SendError(w, http.StatusForbidden, "user is not invited to this event")
			return
		case "group event not found":
			responses.SendError(w, http.StatusNotFound, "event not found")
			return
		default:
			responses.SendError(w, http.StatusInternalServerError, "failed to get inviteable members: "+err.Error())
			return
		}
	}

	responses.SendSuccess(w, "inviteable members", members)
}

// InviteGroupEventMember invites a member to an event
func (h *GroupHandler) InviteGroupEventMember(w http.ResponseWriter, r *http.Request) {
	actorID, err := middleware.UserIDFromContext(r.Context())
	if err != nil {
		responses.SendError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	groupID, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || groupID <= 0 {
		responses.SendError(w, http.StatusBadRequest, "invalid group id")
		return
	}

	eventID, err := strconv.Atoi(r.PathValue("event_id"))
	if err != nil || eventID <= 0 {
		responses.SendError(w, http.StatusBadRequest, "invalid event id")
		return
	}

	targetUserID, err := strconv.Atoi(r.PathValue("user_id"))
	if err != nil || targetUserID <= 0 {
		responses.SendError(w, http.StatusBadRequest, "invalid target user id")
		return
	}

	if err := h.Service.InviteGroupEventMember(r.Context(), actorID, groupID, eventID, targetUserID); err != nil {
		switch err.Error() {
		case "cannot invite yourself":
			responses.SendError(w, http.StatusBadRequest, "cannot invite yourself")
		case "user is not an active group member":
			responses.SendError(w, http.StatusForbidden, "only active group members can invite")
		case "user is not invited or responded to event":
			responses.SendError(w, http.StatusForbidden, "user is not invited to this event")
		case "target user is not an active group member":
			responses.SendError(w, http.StatusForbidden, "target user is not an active group member")
		case "user already invited or responded to event":
			responses.SendError(w, http.StatusConflict, "user is already invited or responded")
		default:
			responses.SendError(w, http.StatusInternalServerError, "failed to invite member: "+err.Error())
		}
		return
	}

	responses.SendSuccess(w, "member invited", map[string]interface{}{
		"group_id":   groupID,
		"event_id":   eventID,
		"invited_to": targetUserID,
	})
}

// InviteAllGroupEventMembers invites all members to an event
func (h *GroupHandler) InviteAllGroupEventMembers(w http.ResponseWriter, r *http.Request) {
	actorID, err := middleware.UserIDFromContext(r.Context())
	if err != nil {
		responses.SendError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	groupID, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || groupID <= 0 {
		responses.SendError(w, http.StatusBadRequest, "invalid group id")
		return
	}

	eventID, err := strconv.Atoi(r.PathValue("event_id"))
	if err != nil || eventID <= 0 {
		responses.SendError(w, http.StatusBadRequest, "invalid event id")
		return
	}

	count, err := h.Service.InviteAllGroupEventMembers(r.Context(), actorID, groupID, eventID)
	if err != nil {
		switch err.Error() {
		case "user is not an active group member":
			responses.SendError(w, http.StatusForbidden, "only active group members can invite")
		case "user is not invited or responded to event":
			responses.SendError(w, http.StatusForbidden, "user is not invited to this event")
		default:
			responses.SendError(w, http.StatusInternalServerError, "failed to invite members: "+err.Error())
		}
		return
	}

	responses.SendSuccess(w, "members invited", map[string]interface{}{
		"group_id":      groupID,
		"event_id":      eventID,
		"invited_count": count,
	})
}

// RespondGroupEventInvite responds to an event invitation
func (h *GroupHandler) RespondGroupEventInvite(w http.ResponseWriter, r *http.Request) {
	actorID, err := middleware.UserIDFromContext(r.Context())
	if err != nil {
		responses.SendError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	groupID, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || groupID <= 0 {
		responses.SendError(w, http.StatusBadRequest, "invalid group id")
		return
	}

	eventID, err := strconv.Atoi(r.PathValue("event_id"))
	if err != nil || eventID <= 0 {
		responses.SendError(w, http.StatusBadRequest, "invalid event id")
		return
	}

	var in models.GroupEventResponseInput
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		responses.SendError(w, http.StatusBadRequest, "invalid JSON")
		return
	}

	if err := h.Service.RespondToGroupEventInvite(r.Context(), actorID, groupID, eventID, in.ReactionType); err != nil {
		switch err.Error() {
		case "group event not found":
			responses.SendError(w, http.StatusNotFound, "group event not found")
		case "group not found":
			responses.SendError(w, http.StatusNotFound, "group not found")
		case "user is not an active group member":
			responses.SendError(w, http.StatusForbidden, "only active group members can respond")
		case "user is not invited or responded to event":
			responses.SendError(w, http.StatusForbidden, "user is not invited to this event")
		case "user already responded to event":
			responses.SendError(w, http.StatusConflict, "event response already recorded")
		default:
			responses.SendError(w, http.StatusInternalServerError, "failed to respond to event: "+err.Error())
		}
		return
	}

	responses.SendSuccess(w, "group event response recorded", map[string]interface{}{
		"group_id":      groupID,
		"event_id":      eventID,
		"reaction_type": in.ReactionType,
	})
}

// ChangeGroupEventResponse changes an event response
func (h *GroupHandler) ChangeGroupEventResponse(w http.ResponseWriter, r *http.Request) {
	actorID, err := middleware.UserIDFromContext(r.Context())
	if err != nil {
		responses.SendError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	groupID, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || groupID <= 0 {
		responses.SendError(w, http.StatusBadRequest, "invalid group id")
		return
	}

	eventID, err := strconv.Atoi(r.PathValue("event_id"))
	if err != nil || eventID <= 0 {
		responses.SendError(w, http.StatusBadRequest, "invalid event id")
		return
	}

	var in models.GroupEventResponseInput
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		responses.SendError(w, http.StatusBadRequest, "invalid JSON")
		return
	}

	if err := h.Service.ChangeGroupEventResponse(r.Context(), actorID, groupID, eventID, in.ReactionType); err != nil {
		switch err.Error() {
		case "group event not found":
			responses.SendError(w, http.StatusNotFound, "group event not found")
		case "group not found":
			responses.SendError(w, http.StatusNotFound, "group not found")
		case "user is not an active group member":
			responses.SendError(w, http.StatusForbidden, "only active group members can respond")
		case "no event response to change":
			responses.SendError(w, http.StatusConflict, "no existing response to change")
		case "event response already set":
			responses.SendError(w, http.StatusConflict, "event response already set")
		default:
			responses.SendError(w, http.StatusInternalServerError, "failed to change event response: "+err.Error())
		}
		return
	}

	responses.SendSuccess(w, "group event response updated", map[string]interface{}{
		"group_id":      groupID,
		"event_id":      eventID,
		"reaction_type": in.ReactionType,
	})
}

// CancelGroupEvent cancels a group event
func (h *GroupHandler) CancelGroupEvent(w http.ResponseWriter, r *http.Request) {
	actorID, err := middleware.UserIDFromContext(r.Context())
	if err != nil {
		responses.SendError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	groupID, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || groupID <= 0 {
		responses.SendError(w, http.StatusBadRequest, "invalid group id")
		return
	}

	eventID, err := strconv.Atoi(r.PathValue("event_id"))
	if err != nil || eventID <= 0 {
		responses.SendError(w, http.StatusBadRequest, "invalid event id")
		return
	}

	if err := h.Service.DeleteGroupEvent(r.Context(), actorID, groupID, eventID); err != nil {
		switch err.Error() {
		case "group event not found":
			responses.SendError(w, http.StatusNotFound, "group event not found")
		case "group not found":
			responses.SendError(w, http.StatusNotFound, "group not found")
		case "only group owner or moderators can approve requests":
			responses.SendError(w, http.StatusForbidden, "only group owner or moderators can cancel events")
		default:
			responses.SendError(w, http.StatusInternalServerError, "failed to cancel group event: "+err.Error())
		}
		return
	}

	responses.SendSuccess(w, "group event cancelled", map[string]int{
		"group_id": groupID,
		"event_id": eventID,
	})
}

// ActiveGroups returns active groups for user
func (h *GroupHandler) ActiveGroups(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.UserIDFromContext(r.Context())
	if err != nil {
		responses.SendError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	items, err := h.Service.GetActiveGroupsForUser(r.Context(), userID)
	if err != nil {
		responses.SendError(w, http.StatusInternalServerError, "failed to load active groups: "+err.Error())
		return
	}

	responses.SendSuccess(w, "active groups", items)
}

// PendingMyGroupRequests returns pending requests made by user
func (h *GroupHandler) PendingMyGroupRequests(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.UserIDFromContext(r.Context())
	if err != nil {
		responses.SendError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	items, err := h.Service.GetUserPendingGroupRequests(r.Context(), userID)
	if err != nil {
		responses.SendError(w, http.StatusInternalServerError, "failed to load pending group requests: "+err.Error())
		return
	}

	responses.SendSuccess(w, "pending group requests", items)
}

// PendingMyGroupInvites returns pending invites for user
func (h *GroupHandler) PendingMyGroupInvites(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.UserIDFromContext(r.Context())
	if err != nil {
		responses.SendError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	items, err := h.Service.GetUserPendingGroupInvites(r.Context(), userID)
	if err != nil {
		responses.SendError(w, http.StatusInternalServerError, "failed to load pending group invites: "+err.Error())
		return
	}

	responses.SendSuccess(w, "pending group invites", items)
}

// AcceptGroupRequest accepts a join request
func (h *GroupHandler) AcceptGroupRequest(w http.ResponseWriter, r *http.Request) {
	approverID, err := middleware.UserIDFromContext(r.Context())
	if err != nil {
		responses.SendError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	groupID, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || groupID <= 0 {
		responses.SendError(w, http.StatusBadRequest, "invalid group id")
		return
	}

	requesterID, err := strconv.Atoi(r.PathValue("user_id"))
	if err != nil || requesterID <= 0 {
		responses.SendError(w, http.StatusBadRequest, "invalid requester user id")
		return
	}

	err = h.Service.AcceptGroupJoinRequest(r.Context(), approverID, groupID, requesterID)
	if err != nil {
		switch err.Error() {
		case "only group owner or moderators can approve requests":
			responses.SendError(w, http.StatusForbidden, "only group owner or moderators can approve requests")
		case "group join request not found":
			responses.SendError(w, http.StatusNotFound, "group join request not found")
		default:
			responses.SendError(w, http.StatusInternalServerError, "failed to approve group request: "+err.Error())
		}
		return
	}

	responses.SendSuccess(w, "group join request approved", map[string]int{
		"group_id":    groupID,
		"approved_by": approverID,
		"user_id":     requesterID,
	})
}

// RejectGroupRequest rejects a join request
func (h *GroupHandler) RejectGroupRequest(w http.ResponseWriter, r *http.Request) {
	approverID, err := middleware.UserIDFromContext(r.Context())
	if err != nil {
		responses.SendError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	groupID, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || groupID <= 0 {
		responses.SendError(w, http.StatusBadRequest, "invalid group id")
		return
	}

	requesterID, err := strconv.Atoi(r.PathValue("user_id"))
	if err != nil || requesterID <= 0 {
		responses.SendError(w, http.StatusBadRequest, "invalid requester user id")
		return
	}

	err = h.Service.RejectGroupJoinRequest(r.Context(), approverID, groupID, requesterID)
	if err != nil {
		switch err.Error() {
		case "only group owner or moderators can approve requests":
			responses.SendError(w, http.StatusForbidden, "only group owner or moderators can reject requests")
		case "group join request not found":
			responses.SendError(w, http.StatusNotFound, "group join request not found")
		default:
			responses.SendError(w, http.StatusInternalServerError, "failed to reject group request: "+err.Error())
		}
		return
	}

	responses.SendSuccess(w, "group join request rejected", map[string]int{
		"group_id":    groupID,
		"rejected_by": approverID,
		"user_id":     requesterID,
	})
}

// InviteToGroup invites a user to a group
func (h *GroupHandler) InviteToGroup(w http.ResponseWriter, r *http.Request) {
	inviterID, err := middleware.UserIDFromContext(r.Context())
	if err != nil {
		responses.SendError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	groupID, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || groupID <= 0 {
		responses.SendError(w, http.StatusBadRequest, "invalid group id")
		return
	}

	targetUserID, err := strconv.Atoi(r.PathValue("user_id"))
	if err != nil || targetUserID <= 0 {
		responses.SendError(w, http.StatusBadRequest, "invalid target user id")
		return
	}

	membershipStatus, err := h.Service.InviteUserToGroup(r.Context(), inviterID, groupID, targetUserID)
	if err != nil {
		switch err.Error() {
		case "group not found":
			responses.SendError(w, http.StatusNotFound, "group not found")
		case "cannot join private group from this endpoint":
			responses.SendError(w, http.StatusForbidden, "cannot invite to private group")
		case "only group owner or moderators can approve requests":
			responses.SendError(w, http.StatusForbidden, "only group owner or moderators can invite")
		case "user is already a group member":
			responses.SendError(w, http.StatusConflict, "user is already a member")
		case "user has already been invited":
			responses.SendError(w, http.StatusConflict, "user has already been invited")
		default:
			responses.SendError(w, http.StatusInternalServerError, "failed to invite user: "+err.Error())
		}
		return
	}

	responses.SendSuccess(w, "user invited to group", map[string]interface{}{
		"group_id":          groupID,
		"invited_user_id":   targetUserID,
		"membership_status": membershipStatus,
	})
}

// AcceptInvite accepts a group invitation
func (h *GroupHandler) AcceptInvite(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.UserIDFromContext(r.Context())
	if err != nil {
		responses.SendError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	groupID, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || groupID <= 0 {
		responses.SendError(w, http.StatusBadRequest, "invalid group id")
		return
	}

	if err := h.Service.AcceptGroupInvite(r.Context(), userID, groupID); err != nil {
		switch err.Error() {
		case "group invite not found":
			responses.SendError(w, http.StatusNotFound, "group invite not found")
		default:
			responses.SendError(w, http.StatusInternalServerError, "failed to accept invite: "+err.Error())
		}
		return
	}

	responses.SendSuccess(w, "invite accepted", map[string]int{
		"group_id": groupID,
		"user_id":  userID,
	})
}

// RejectInvite rejects a group invitation
func (h *GroupHandler) RejectInvite(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.UserIDFromContext(r.Context())
	if err != nil {
		responses.SendError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	groupID, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || groupID <= 0 {
		responses.SendError(w, http.StatusBadRequest, "invalid group id")
		return
	}

	if err := h.Service.RejectGroupInvite(r.Context(), userID, groupID); err != nil {
		switch err.Error() {
		case "group invite not found":
			responses.SendError(w, http.StatusNotFound, "group invite not found")
		default:
			responses.SendError(w, http.StatusInternalServerError, "failed to reject invite: "+err.Error())
		}
		return
	}

	responses.SendSuccess(w, "invite rejected", map[string]int{
		"group_id": groupID,
		"user_id":  userID,
	})
}

// KickMember removes a member from group
func (h *GroupHandler) KickMember(w http.ResponseWriter, r *http.Request) {
	actorID, err := middleware.UserIDFromContext(r.Context())
	if err != nil {
		responses.SendError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	groupID, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || groupID <= 0 {
		responses.SendError(w, http.StatusBadRequest, "invalid group id")
		return
	}

	targetUserID, err := strconv.Atoi(r.PathValue("user_id"))
	if err != nil || targetUserID <= 0 {
		responses.SendError(w, http.StatusBadRequest, "invalid target user id")
		return
	}

	if err := h.Service.KickGroupMember(r.Context(), actorID, groupID, targetUserID); err != nil {
		switch err.Error() {
		case "only group owner or moderators can approve requests":
			responses.SendError(w, http.StatusForbidden, "only group owner or moderators can kick members")
		case "group member not found":
			responses.SendError(w, http.StatusNotFound, "group member not found")
		case "cannot kick owner or moderator":
			responses.SendError(w, http.StatusForbidden, "cannot kick owner or moderator")
		default:
			responses.SendError(w, http.StatusInternalServerError, "failed to kick member: "+err.Error())
		}
		return
	}

	responses.SendSuccess(w, "member removed from group", map[string]interface{}{
		"group_id":       groupID,
		"kicked_user_id": targetUserID,
	})
}

// LeaveGroup leaves a group
func (h *GroupHandler) LeaveGroup(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.UserIDFromContext(r.Context())
	if err != nil {
		responses.SendError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	groupID, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || groupID <= 0 {
		responses.SendError(w, http.StatusBadRequest, "invalid group id")
		return
	}

	result, err := h.Service.LeaveGroup(r.Context(), userID, groupID)
	if err != nil {
		switch err.Error() {
		case "group member not found":
			responses.SendError(w, http.StatusNotFound, "user is not a group member")
		default:
			responses.SendError(w, http.StatusInternalServerError, "failed to leave group: "+err.Error())
		}
		return
	}

	resp := map[string]interface{}{
		"group_id": groupID,
		"user_id":  userID,
	}
	if result.GroupDeleted {
		resp["status"] = "group deleted"
	} else if result.OwnerTransferred {
		resp["status"] = "ownership transferred"
		resp["new_owner_id"] = result.NewOwnerID
	} else {
		resp["status"] = "left group"
	}

	responses.SendSuccess(w, "left group", resp)
}

// PromoteModerator promotes a member to moderator
func (h *GroupHandler) PromoteModerator(w http.ResponseWriter, r *http.Request) {
	ownerID, err := middleware.UserIDFromContext(r.Context())
	if err != nil {
		responses.SendError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	groupID, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || groupID <= 0 {
		responses.SendError(w, http.StatusBadRequest, "invalid group id")
		return
	}

	targetUserID, err := strconv.Atoi(r.PathValue("user_id"))
	if err != nil || targetUserID <= 0 {
		responses.SendError(w, http.StatusBadRequest, "invalid target user id")
		return
	}

	if err := h.Service.PromoteGroupModerator(r.Context(), ownerID, groupID, targetUserID); err != nil {
		switch err.Error() {
		case "group not found":
			responses.SendError(w, http.StatusNotFound, "group not found")
		case "only the group owner can delete the group":
			responses.SendError(w, http.StatusForbidden, "only the group owner can promote moderators")
		case "group member not found":
			responses.SendError(w, http.StatusNotFound, "group member not found")
		case "group member role mismatch":
			responses.SendError(w, http.StatusConflict, "user is not a regular member")
		default:
			responses.SendError(w, http.StatusInternalServerError, "failed to promote moderator: "+err.Error())
		}
		return
	}

	responses.SendSuccess(w, "member promoted to moderator", map[string]int{
		"group_id":         groupID,
		"promoted_user_id": targetUserID,
	})
}

// DemoteModerator demotes a moderator to member
func (h *GroupHandler) DemoteModerator(w http.ResponseWriter, r *http.Request) {
	ownerID, err := middleware.UserIDFromContext(r.Context())
	if err != nil {
		responses.SendError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	groupID, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || groupID <= 0 {
		responses.SendError(w, http.StatusBadRequest, "invalid group id")
		return
	}

	targetUserID, err := strconv.Atoi(r.PathValue("user_id"))
	if err != nil || targetUserID <= 0 {
		responses.SendError(w, http.StatusBadRequest, "invalid target user id")
		return
	}

	if err := h.Service.DemoteGroupModerator(r.Context(), ownerID, groupID, targetUserID); err != nil {
		switch err.Error() {
		case "group not found":
			responses.SendError(w, http.StatusNotFound, "group not found")
		case "only the group owner can delete the group":
			responses.SendError(w, http.StatusForbidden, "only the group owner can demote moderators")
		case "group member not found":
			responses.SendError(w, http.StatusNotFound, "group member not found")
		case "group member role mismatch":
			responses.SendError(w, http.StatusConflict, "user is not a moderator")
		default:
			responses.SendError(w, http.StatusInternalServerError, "failed to demote moderator: "+err.Error())
		}
		return
	}

	responses.SendSuccess(w, "moderator demoted to member", map[string]int{
		"group_id":        groupID,
		"demoted_user_id": targetUserID,
	})
}

// UpdateGroupSettings updates group settings
func (h *GroupHandler) UpdateGroupSettings(w http.ResponseWriter, r *http.Request) {
	ownerID, err := middleware.UserIDFromContext(r.Context())
	if err != nil {
		responses.SendError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	groupID, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || groupID <= 0 {
		responses.SendError(w, http.StatusBadRequest, "invalid group id")
		return
	}

	var in struct {
		Visibility *string `json:"visibility"`
		JoinMode   *string `json:"join_mode"`
	}

	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		responses.SendError(w, http.StatusBadRequest, "invalid JSON")
		return
	}

	result, err := h.Service.UpdateGroupSettings(r.Context(), ownerID, groupID, in.Visibility, in.JoinMode)
	if err != nil {
		switch err.Error() {
		case "group not found":
			responses.SendError(w, http.StatusNotFound, "group not found")
		case "only the group owner can delete the group":
			responses.SendError(w, http.StatusForbidden, "only the group owner can update settings")
		default:
			responses.SendError(w, http.StatusInternalServerError, "failed to update group settings: "+err.Error())
		}
		return
	}

	responses.SendSuccess(w, "group settings updated", result)
}

// GetGroupSettings retrieves group settings
func (h *GroupHandler) GetGroupSettings(w http.ResponseWriter, r *http.Request) {
	ownerID, err := middleware.UserIDFromContext(r.Context())
	if err != nil {
		responses.SendError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	groupID, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || groupID <= 0 {
		responses.SendError(w, http.StatusBadRequest, "invalid group id")
		return
	}

	result, err := h.Service.GetGroupSettings(r.Context(), ownerID, groupID)
	if err != nil {
		switch err.Error() {
		case "group not found":
			responses.SendError(w, http.StatusNotFound, "group not found")
		case "only the group owner can delete the group":
			responses.SendError(w, http.StatusForbidden, "only the group owner can view settings")
		default:
			responses.SendError(w, http.StatusInternalServerError, "failed to get group settings: "+err.Error())
		}
		return
	}

	responses.SendSuccess(w, "group settings", result)
}

// GroupMembersList retrieves group members
func (h *GroupHandler) GroupMembersList(w http.ResponseWriter, r *http.Request) {
	groupID, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || groupID <= 0 {
		responses.SendError(w, http.StatusBadRequest, "invalid group id")
		return
	}

	members, err := h.Service.GetActiveGroupMembers(r.Context(), groupID)
	if err != nil {
		switch err.Error() {
		case "group not found":
			responses.SendError(w, http.StatusNotFound, "group not found")
		default:
			responses.SendError(w, http.StatusInternalServerError, "failed to load members: "+err.Error())
		}
		return
	}

	responses.SendSuccess(w, "group members", members)
}

// PendingGroupRequests retrieves pending join requests for a group
func (h *GroupHandler) PendingGroupRequests(w http.ResponseWriter, r *http.Request) {
	actorID, err := middleware.UserIDFromContext(r.Context())
	if err != nil {
		responses.SendError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	groupID, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || groupID <= 0 {
		responses.SendError(w, http.StatusBadRequest, "invalid group id")
		return
	}

	items, err := h.Service.GetPendingGroupJoinRequests(r.Context(), actorID, groupID)
	if err != nil {
		switch err.Error() {
		case "only group owner or moderators can approve requests":
			responses.SendError(w, http.StatusForbidden, "only group owner or moderators can view requests")
		default:
			responses.SendError(w, http.StatusInternalServerError, "failed to load pending requests: "+err.Error())
		}
		return
	}

	responses.SendSuccess(w, "pending group requests", items)
}

// PendingGroupInvites retrieves pending invites for a group
func (h *GroupHandler) PendingGroupInvites(w http.ResponseWriter, r *http.Request) {
	actorID, err := middleware.UserIDFromContext(r.Context())
	if err != nil {
		responses.SendError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	groupID, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || groupID <= 0 {
		responses.SendError(w, http.StatusBadRequest, "invalid group id")
		return
	}

	items, err := h.Service.GetPendingGroupInvites(r.Context(), actorID, groupID)
	if err != nil {
		switch err.Error() {
		case "only group owner or moderators can approve requests":
			responses.SendError(w, http.StatusForbidden, "only group owner or moderators can view invites")
		default:
			responses.SendError(w, http.StatusInternalServerError, "failed to load pending invites: "+err.Error())
		}
		return
	}

	responses.SendSuccess(w, "pending group invites", items)
}

// RemoveInvite removes a pending invite
func (h *GroupHandler) RemoveInvite(w http.ResponseWriter, r *http.Request) {
	actorID, err := middleware.UserIDFromContext(r.Context())
	if err != nil {
		responses.SendError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	groupID, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || groupID <= 0 {
		responses.SendError(w, http.StatusBadRequest, "invalid group id")
		return
	}

	targetUserID, err := strconv.Atoi(r.PathValue("user_id"))
	if err != nil || targetUserID <= 0 {
		responses.SendError(w, http.StatusBadRequest, "invalid target user id")
		return
	}

	if err := h.Service.RemovePendingGroupInvite(r.Context(), actorID, groupID, targetUserID); err != nil {
		switch err.Error() {
		case "group not found":
			responses.SendError(w, http.StatusNotFound, "group not found")
		case "only group owner or moderators can approve requests":
			responses.SendError(w, http.StatusForbidden, "only group owner or moderators can remove invites")
		case "group invite not found":
			responses.SendError(w, http.StatusNotFound, "group invite not found")
		case "group member is active":
			responses.SendError(w, http.StatusConflict, "user is an active member")
		default:
			responses.SendError(w, http.StatusInternalServerError, "failed to remove invite: "+err.Error())
		}
		return
	}

	responses.SendSuccess(w, "invite removed", map[string]int{
		"group_id": groupID,
		"user_id":  targetUserID,
	})
}

// RemoveOwnRequest removes own pending join request
func (h *GroupHandler) RemoveOwnRequest(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.UserIDFromContext(r.Context())
	if err != nil {
		responses.SendError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	groupID, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || groupID <= 0 {
		responses.SendError(w, http.StatusBadRequest, "invalid group id")
		return
	}

	if err := h.Service.RemoveOwnPendingGroupRequest(r.Context(), userID, groupID); err != nil {
		switch err.Error() {
		case "group join request not found":
			responses.SendError(w, http.StatusNotFound, "no pending request found")
		case "group member is active":
			responses.SendError(w, http.StatusConflict, "you are already an active member")
		default:
			responses.SendError(w, http.StatusInternalServerError, "failed to cancel request: "+err.Error())
		}
		return
	}

	responses.SendSuccess(w, "request cancelled", map[string]int{
		"group_id": groupID,
		"user_id":  userID,
	})
}

// GetGroupPage retrieves group page view
func (h *GroupHandler) GetGroupPage(w http.ResponseWriter, r *http.Request) {
	viewerID, err := middleware.UserIDFromContext(r.Context())
	if err != nil {
		responses.SendError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	groupID, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || groupID <= 0 {
		responses.SendError(w, http.StatusBadRequest, "invalid group id")
		return
	}

	view, err := h.Service.GetGroupPageView(r.Context(), viewerID, groupID)
	if err != nil {
		switch err.Error() {
		case "group not found":
			responses.SendError(w, http.StatusNotFound, "group not found")
		default:
			responses.SendError(w, http.StatusInternalServerError, "failed to load group page: "+err.Error())
		}
		return
	}

	if view.IsActive {
		responses.SendSuccess(w, "group page", map[string]interface{}{
			"id":            view.ID,
			"name":          view.Name,
			"description":   view.Description,
			"visibility":    view.Visibility,
			"join_mode":     view.JoinMode,
			"group_picture": view.GroupPicture,
			"group_members": view.GroupMembers,
			"created_at":    view.CreatedAt,
			"role":          view.Role,
		})
		return
	}

	if view.Visibility == "public" {
		canRequest := false
		if view.PendingType == "none" {
			canRequest = view.JoinMode == "auto" || view.JoinMode == "request" || view.JoinMode == "request_and_invite"
		}

		responses.SendSuccess(w, "group page", map[string]interface{}{
			"id":            view.ID,
			"name":          view.Name,
			"description":   view.Description,
			"visibility":    view.Visibility,
			"join_mode":     view.JoinMode,
			"group_picture": view.GroupPicture,
			"group_members": view.GroupMembers,
			"created_at":    view.CreatedAt,
			"pending_type":  view.PendingType,
			"can_request":   canRequest,
		})
		return
	}

	responses.SendSuccess(w, "group page", map[string]interface{}{
		"id":            view.ID,
		"name":          view.Name,
		"description":   view.Description,
		"visibility":    view.Visibility,
		"join_mode":     view.JoinMode,
		"group_picture": view.GroupPicture,
	})
}
