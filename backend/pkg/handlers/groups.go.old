package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"

	"backend/pkg/db/queries"
	database "backend/pkg/db/sqlite"
	"backend/pkg/middleware"
	"backend/pkg/models"
	"backend/pkg/responses"
	"backend/pkg/utils"
)

func CreateGroupHandler(w http.ResponseWriter, r *http.Request) {
	ownerID, err := middleware.UserIDFromContext(r.Context())
	if err != nil {
		responses.SendError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var in models.CreateGroupInput

	if err := r.ParseMultipartForm(20 << 20); err != nil {
		log.Printf("[ERROR] CreateGroupHandler: ParseMultipartForm failed: %v", err)
		responses.SendError(w, http.StatusBadRequest, "Invalid Form")
		return
	}

	in.Name = r.FormValue("name")
	in.Description = r.FormValue("description")
	in.Visibility = r.FormValue("visibility")
	in.JoinMode = r.FormValue("join_mode")

	imageURL, err := utils.AttachGroupImage(r)
	if err != nil {
		log.Printf("[ERROR] CreateGroupHandler: File attachment failed: %v", err)
		responses.SendError(w, http.StatusInternalServerError, err.Error())
		return
	}

	if imageURL != "" {
		log.Printf("[INFO] CreateGroupHandler: Image attached: %s", imageURL)
		in.Picture = imageURL
	}

	if in.Name == "" {
		responses.SendError(w, http.StatusBadRequest, "name is required")
		return
	}
	if in.Visibility != "public" && in.Visibility != "private" {
		responses.SendError(w, http.StatusBadRequest, "visibility must be public or private")
		return
	}
	switch in.JoinMode {
	case "auto", "request", "invite", "request_and_invite":
	default:
		responses.SendError(w, http.StatusBadRequest, "join_mode must be auto, request, invite, or request_and_invite")
		return
	}

	group, err := queries.CreateGroup(r.Context(), database.DB, ownerID, in)
	if err != nil {
		if err == queries.ErrGroupNameTaken {
			responses.SendError(w, http.StatusConflict, "group name already in use")
			return
		}
		responses.SendError(w, http.StatusInternalServerError, "failed to create group: "+err.Error())
		return
	}

	responses.SendCreated(w, "group created successfully", group)
}

func DeleteGroupHandler(w http.ResponseWriter, r *http.Request) {
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

	err = queries.DeleteGroup(r.Context(), database.DB, requesterID, groupID)
	if err != nil {
		switch err {
		case queries.ErrGroupNotFound:
			responses.SendError(w, http.StatusNotFound, "group not found")
			return
		case queries.ErrNotGroupOwner:
			responses.SendError(w, http.StatusForbidden, "only the group owner can delete this group")
			return
		default:
			responses.SendError(w, http.StatusInternalServerError, "failed to delete group: "+err.Error())
			return
		}
	}

	responses.SendSuccess(w, "group deleted successfully", map[string]int{"group_id": groupID})
}

func RequestToJoinGroupHandler(w http.ResponseWriter, r *http.Request) {
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

	membershipStatus, err := queries.RequestToJoinGroup(r.Context(), database.DB, requesterID, groupID)
	if err != nil {
		switch err {
		case queries.ErrGroupNotFound:
			responses.SendError(w, http.StatusNotFound, "group not found")
			return
		case queries.ErrPrivateGroup:
			responses.SendError(w, http.StatusForbidden, "cannot request to join a private group")
			return
		case queries.ErrInviteOnlyGroup:
			responses.SendError(w, http.StatusForbidden, "group is invite-only")
			return
		case queries.ErrAlreadyGroupMember:
			responses.SendError(w, http.StatusConflict, "already a group member")
			return
		case queries.ErrAlreadyRequestedToJoin:
			responses.SendError(w, http.StatusConflict, "join request already exists")
			return
		default:
			responses.SendError(w, http.StatusInternalServerError, "failed to process join request: "+err.Error())
			return
		}
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

func DiscoverGroupsHandler(w http.ResponseWriter, r *http.Request) {
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

	items, err := queries.DiscoverGroups(r.Context(), database.DB, userID, limit, offset)
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

func CreateGroupEventHandler(w http.ResponseWriter, r *http.Request) {
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

	in.Title = strings.TrimSpace(in.Title)
	in.Description = strings.TrimSpace(in.Description)
	in.EventDay = strings.TrimSpace(in.EventDay)
	in.EventTime = strings.TrimSpace(in.EventTime)

	if in.Title == "" {
		responses.SendError(w, http.StatusBadRequest, "title is required")
		return
	}
	if in.EventDay == "" {
		responses.SendError(w, http.StatusBadRequest, "event_day is required")
		return
	}
	if in.EventTime == "" {
		responses.SendError(w, http.StatusBadRequest, "event_time is required")
		return
	}

	created, err := queries.CreateGroupEvent(r.Context(), database.DB, actorID, groupID, in)
	if err != nil {
		switch err {
		case queries.ErrGroupNotFound:
			responses.SendError(w, http.StatusNotFound, "group not found")
			return
		case queries.ErrNotGroupModeratorOrOwner:
			responses.SendError(w, http.StatusForbidden, "only group owner or moderators can create events")
			return
		default:
			responses.SendError(w, http.StatusInternalServerError, "failed to create group event: "+err.Error())
			return
		}
	}

	responses.SendCreated(w, "group event created", created)
}

func GroupEventInviteableMembersHandler(w http.ResponseWriter, r *http.Request) {
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

	items, err := queries.GetGroupEventInviteableMembers(r.Context(), database.DB, actorID, groupID, eventID)
	if err != nil {
		switch err {
		case queries.ErrGroupEventNotFound:
			responses.SendError(w, http.StatusNotFound, "group event not found")
			return
		case queries.ErrNotActiveGroupMember:
			responses.SendError(w, http.StatusForbidden, "only active group members can invite")
			return
		case queries.ErrNotInvitedToEvent:
			responses.SendError(w, http.StatusForbidden, "only invited/responded members can invite")
			return
		case queries.ErrGroupNotFound:
			responses.SendError(w, http.StatusNotFound, "group not found")
			return
		default:
			responses.SendError(w, http.StatusInternalServerError, "failed to load inviteable members: "+err.Error())
			return
		}
	}

	responses.SendSuccess(w, "group event inviteable members", items)
}

func InviteGroupEventMemberHandler(w http.ResponseWriter, r *http.Request) {
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

	if err := queries.InviteGroupEventMember(r.Context(), database.DB, actorID, groupID, eventID, targetUserID); err != nil {
		switch err {
		case queries.ErrGroupEventNotFound:
			responses.SendError(w, http.StatusNotFound, "group event not found")
			return
		case queries.ErrGroupNotFound:
			responses.SendError(w, http.StatusNotFound, "group not found")
			return
		case queries.ErrNotActiveGroupMember:
			responses.SendError(w, http.StatusForbidden, "only active group members can invite")
			return
		case queries.ErrNotInvitedToEvent:
			responses.SendError(w, http.StatusForbidden, "only invited/responded members can invite")
			return
		case queries.ErrTargetNotActiveMember:
			responses.SendError(w, http.StatusNotFound, "target user is not an active group member")
			return
		case queries.ErrCannotInviteSelf:
			responses.SendError(w, http.StatusBadRequest, "cannot invite yourself")
			return
		case queries.ErrGroupEventAlreadyAnswered:
			responses.SendError(w, http.StatusConflict, "user already invited or responded")
			return
		default:
			responses.SendError(w, http.StatusInternalServerError, "failed to invite member: "+err.Error())
			return
		}
	}

	responses.SendSuccess(w, "group event invite sent", map[string]int{
		"group_id": groupID,
		"event_id": eventID,
		"user_id":  targetUserID,
	})
}

func InviteAllGroupEventMembersHandler(w http.ResponseWriter, r *http.Request) {
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

	count, err := queries.InviteAllGroupEventMembers(r.Context(), database.DB, actorID, groupID, eventID)
	if err != nil {
		switch err {
		case queries.ErrGroupEventNotFound:
			responses.SendError(w, http.StatusNotFound, "group event not found")
			return
		case queries.ErrGroupNotFound:
			responses.SendError(w, http.StatusNotFound, "group not found")
			return
		case queries.ErrNotActiveGroupMember:
			responses.SendError(w, http.StatusForbidden, "only active group members can invite")
			return
		case queries.ErrNotInvitedToEvent:
			responses.SendError(w, http.StatusForbidden, "only invited/responded members can invite")
			return
		default:
			responses.SendError(w, http.StatusInternalServerError, "failed to invite members: "+err.Error())
			return
		}
	}

	responses.SendSuccess(w, "group event invites sent", map[string]int{
		"group_id": groupID,
		"event_id": eventID,
		"invited":  count,
	})
}

func RespondGroupEventInviteHandler(w http.ResponseWriter, r *http.Request) {
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

	in.ReactionType = strings.TrimSpace(strings.ToLower(in.ReactionType))
	if in.ReactionType != "going" && in.ReactionType != "not_going" {
		responses.SendError(w, http.StatusBadRequest, "reaction_type must be going or not_going")
		return
	}

	if err := queries.RespondToGroupEventInvite(r.Context(), database.DB, actorID, groupID, eventID, in.ReactionType); err != nil {
		switch err {
		case queries.ErrGroupEventNotFound:
			responses.SendError(w, http.StatusNotFound, "group event not found")
			return
		case queries.ErrGroupNotFound:
			responses.SendError(w, http.StatusNotFound, "group not found")
			return
		case queries.ErrNotActiveGroupMember:
			responses.SendError(w, http.StatusForbidden, "only active group members can respond")
			return
		case queries.ErrNotInvitedToEvent:
			responses.SendError(w, http.StatusForbidden, "user is not invited to this event")
			return
		case queries.ErrGroupEventAlreadyResponded:
			responses.SendError(w, http.StatusConflict, "event response already recorded")
			return
		default:
			responses.SendError(w, http.StatusInternalServerError, "failed to respond to event: "+err.Error())
			return
		}
	}

	responses.SendSuccess(w, "group event response recorded", map[string]interface{}{
		"group_id":      groupID,
		"event_id":      eventID,
		"reaction_type": in.ReactionType,
	})
}

func ChangeGroupEventResponseHandler(w http.ResponseWriter, r *http.Request) {
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

	in.ReactionType = strings.TrimSpace(strings.ToLower(in.ReactionType))
	if in.ReactionType != "going" && in.ReactionType != "not_going" {
		responses.SendError(w, http.StatusBadRequest, "reaction_type must be going or not_going")
		return
	}

	if err := queries.ChangeGroupEventResponse(r.Context(), database.DB, actorID, groupID, eventID, in.ReactionType); err != nil {
		switch err {
		case queries.ErrGroupEventNotFound:
			responses.SendError(w, http.StatusNotFound, "group event not found")
			return
		case queries.ErrGroupNotFound:
			responses.SendError(w, http.StatusNotFound, "group not found")
			return
		case queries.ErrNotActiveGroupMember:
			responses.SendError(w, http.StatusForbidden, "only active group members can respond")
			return
		case queries.ErrGroupEventNoResponseToChange:
			responses.SendError(w, http.StatusConflict, "no existing response to change")
			return
		case queries.ErrGroupEventResponseUnchanged:
			responses.SendError(w, http.StatusConflict, "event response already set")
			return
		default:
			responses.SendError(w, http.StatusInternalServerError, "failed to change event response: "+err.Error())
			return
		}
	}

	responses.SendSuccess(w, "group event response updated", map[string]interface{}{
		"group_id":      groupID,
		"event_id":      eventID,
		"reaction_type": in.ReactionType,
	})
}

func CancelGroupEventHandler(w http.ResponseWriter, r *http.Request) {
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

	if err := queries.DeleteGroupEvent(r.Context(), database.DB, actorID, groupID, eventID); err != nil {
		switch err {
		case queries.ErrGroupEventNotFound:
			responses.SendError(w, http.StatusNotFound, "group event not found")
			return
		case queries.ErrGroupNotFound:
			responses.SendError(w, http.StatusNotFound, "group not found")
			return
		case queries.ErrNotGroupModeratorOrOwner:
			responses.SendError(w, http.StatusForbidden, "only group owner or moderators can cancel events")
			return
		default:
			responses.SendError(w, http.StatusInternalServerError, "failed to cancel group event: "+err.Error())
			return
		}
	}

	responses.SendSuccess(w, "group event cancelled", map[string]int{
		"group_id": groupID,
		"event_id": eventID,
	})
}

func ActiveGroupsHandler(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.UserIDFromContext(r.Context())
	if err != nil {
		responses.SendError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	items, err := queries.GetActiveGroupsForUser(r.Context(), database.DB, userID)
	if err != nil {
		responses.SendError(w, http.StatusInternalServerError, "failed to load active groups: "+err.Error())
		return
	}

	responses.SendSuccess(w, "active groups", items)
}

func PendingMyGroupRequestsHandler(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.UserIDFromContext(r.Context())
	if err != nil {
		responses.SendError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	items, err := queries.GetUserPendingGroupRequests(r.Context(), database.DB, userID)
	if err != nil {
		responses.SendError(w, http.StatusInternalServerError, "failed to load pending group requests: "+err.Error())
		return
	}

	responses.SendSuccess(w, "pending group requests", items)
}

func PendingMyGroupInvitesHandler(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.UserIDFromContext(r.Context())
	if err != nil {
		responses.SendError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	items, err := queries.GetUserPendingGroupInvites(r.Context(), database.DB, userID)
	if err != nil {
		responses.SendError(w, http.StatusInternalServerError, "failed to load pending group invites: "+err.Error())
		return
	}

	responses.SendSuccess(w, "pending group invites", items)
}

func AcceptGroupRequestHandler(w http.ResponseWriter, r *http.Request) {
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

	err = queries.AcceptGroupJoinRequest(r.Context(), database.DB, approverID, groupID, requesterID)
	if err != nil {
		switch err {
		case queries.ErrNotGroupModeratorOrOwner:
			responses.SendError(w, http.StatusForbidden, "only group owner or moderators can approve requests")
			return
		case queries.ErrGroupJoinRequestNotFound:
			responses.SendError(w, http.StatusNotFound, "group join request not found")
			return
		default:
			responses.SendError(w, http.StatusInternalServerError, "failed to approve group request: "+err.Error())
			return
		}
	}

	responses.SendSuccess(w, "group join request approved", map[string]int{
		"group_id":    groupID,
		"approved_by": approverID,
		"user_id":     requesterID,
	})
}

func RejectGroupRequestHandler(w http.ResponseWriter, r *http.Request) {
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

	err = queries.RejectGroupJoinRequest(r.Context(), database.DB, approverID, groupID, requesterID)
	if err != nil {
		switch err {
		case queries.ErrNotGroupModeratorOrOwner:
			responses.SendError(w, http.StatusForbidden, "only group owner or moderators can reject requests")
			return
		case queries.ErrGroupJoinRequestNotFound:
			responses.SendError(w, http.StatusNotFound, "group join request not found")
			return
		default:
			responses.SendError(w, http.StatusInternalServerError, "failed to reject group request: "+err.Error())
			return
		}
	}

	responses.SendSuccess(w, "group join request rejected", map[string]int{
		"group_id":    groupID,
		"rejected_by": approverID,
		"user_id":     requesterID,
	})
}

func InviteToGroupHandler(w http.ResponseWriter, r *http.Request) {
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

	membershipStatus, err := queries.InviteUserToGroup(r.Context(), database.DB, inviterID, groupID, targetUserID)
	if err != nil {
		switch err {
		case queries.ErrGroupNotFound:
			responses.SendError(w, http.StatusNotFound, "group not found")
			return
		case queries.ErrPrivateGroup:
			responses.SendError(w, http.StatusForbidden, "cannot invite to a private group from this endpoint")
			return
		case queries.ErrInviteOnlyGroup:
			responses.SendError(w, http.StatusForbidden, "group is invite-only")
			return
		case queries.ErrNotGroupModeratorOrOwner:
			responses.SendError(w, http.StatusForbidden, "only group owner or moderators can invite users")
			return
		case queries.ErrAlreadyGroupMember:
			responses.SendError(w, http.StatusConflict, "target user is already a group member")
			return
		case queries.ErrAlreadyInvitedToGroup, queries.ErrAlreadyRequestedToJoin:
			responses.SendError(w, http.StatusConflict, "target user already has a pending group invite/request")
			return
		default:
			responses.SendError(w, http.StatusInternalServerError, "failed to invite user: "+err.Error())
			return
		}
	}

	message := "group invitation sent"
	if membershipStatus == "active" {
		message = "user added to group successfully"
	}

	responses.SendSuccess(w, message, map[string]interface{}{
		"group_id":          groupID,
		"invited_by":        inviterID,
		"user_id":           targetUserID,
		"membership_status": membershipStatus,
	})
}

func AcceptInviteHandler(w http.ResponseWriter, r *http.Request) {
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

	err = queries.AcceptGroupInvite(r.Context(), database.DB, userID, groupID)
	if err != nil {
		switch err {
		case queries.ErrGroupInviteNotFound:
			responses.SendError(w, http.StatusNotFound, "group invite not found")
			return
		default:
			responses.SendError(w, http.StatusInternalServerError, "failed to accept group invite: "+err.Error())
			return
		}
	}

	responses.SendSuccess(w, "group invite accepted", map[string]int{
		"group_id": groupID,
		"user_id":  userID,
	})
}

func RejectInviteHandler(w http.ResponseWriter, r *http.Request) {
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

	err = queries.RejectGroupInvite(r.Context(), database.DB, userID, groupID)
	if err != nil {
		switch err {
		case queries.ErrGroupInviteNotFound:
			responses.SendError(w, http.StatusNotFound, "group invite not found")
			return
		default:
			responses.SendError(w, http.StatusInternalServerError, "failed to reject group invite: "+err.Error())
			return
		}
	}

	responses.SendSuccess(w, "group invite rejected", map[string]int{
		"group_id": groupID,
		"user_id":  userID,
	})
}

func KickMemberHandler(w http.ResponseWriter, r *http.Request) {
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

	err = queries.KickGroupMember(r.Context(), database.DB, actorID, groupID, targetUserID)
	if err != nil {
		switch err {
		case queries.ErrNotGroupModeratorOrOwner:
			responses.SendError(w, http.StatusForbidden, "only group owner or moderators can kick members")
			return
		case queries.ErrCannotKickGroupStaff:
			responses.SendError(w, http.StatusForbidden, "cannot kick owner or moderator")
			return
		case queries.ErrGroupMemberNotFound:
			responses.SendError(w, http.StatusNotFound, "group member not found")
			return
		default:
			responses.SendError(w, http.StatusInternalServerError, "failed to kick member: "+err.Error())
			return
		}
	}

	responses.SendSuccess(w, "group member kicked", map[string]int{
		"group_id":  groupID,
		"kicked_by": actorID,
		"user_id":   targetUserID,
	})
}

func LeaveGroupHandler(w http.ResponseWriter, r *http.Request) {
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

	result, err := queries.LeaveGroup(r.Context(), database.DB, userID, groupID)
	if err != nil {
		switch err {
		case queries.ErrGroupMemberNotFound:
			responses.SendError(w, http.StatusNotFound, "group membership not found")
			return
		case queries.ErrGroupNotFound:
			responses.SendError(w, http.StatusNotFound, "group not found")
			return
		default:
			responses.SendError(w, http.StatusInternalServerError, "failed to leave group: "+err.Error())
			return
		}
	}

	data := map[string]interface{}{
		"group_id": groupID,
		"user_id":  userID,
	}
	if result.GroupDeleted {
		data["group_deleted"] = true
		responses.SendSuccess(w, "left group and group was deleted", data)
		return
	}
	if result.OwnerTransferred {
		data["owner_transferred"] = true
		data["new_owner_id"] = result.NewOwnerID
		responses.SendSuccess(w, "left group and ownership transferred", data)
		return
	}

	responses.SendSuccess(w, "left group successfully", data)
}

func PromoteModeratorHandler(w http.ResponseWriter, r *http.Request) {
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

	err = queries.PromoteGroupModerator(r.Context(), database.DB, ownerID, groupID, targetUserID)
	if err != nil {
		switch err {
		case queries.ErrGroupNotFound:
			responses.SendError(w, http.StatusNotFound, "group not found")
			return
		case queries.ErrNotGroupOwner:
			responses.SendError(w, http.StatusForbidden, "only the group owner can promote moderators")
			return
		case queries.ErrGroupMemberNotFound:
			responses.SendError(w, http.StatusNotFound, "group member not found")
			return
		case queries.ErrGroupMemberRoleMismatch:
			responses.SendError(w, http.StatusConflict, "target user must be an active member")
			return
		default:
			responses.SendError(w, http.StatusInternalServerError, "failed to promote moderator: "+err.Error())
			return
		}
	}

	responses.SendSuccess(w, "member promoted to moderator", map[string]int{
		"group_id":    groupID,
		"promoted_by": ownerID,
		"user_id":     targetUserID,
	})
}

func DemoteModeratorHandler(w http.ResponseWriter, r *http.Request) {
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

	err = queries.DemoteGroupModerator(r.Context(), database.DB, ownerID, groupID, targetUserID)
	if err != nil {
		switch err {
		case queries.ErrGroupNotFound:
			responses.SendError(w, http.StatusNotFound, "group not found")
			return
		case queries.ErrNotGroupOwner:
			responses.SendError(w, http.StatusForbidden, "only the group owner can demote moderators")
			return
		case queries.ErrGroupMemberNotFound:
			responses.SendError(w, http.StatusNotFound, "group member not found")
			return
		case queries.ErrGroupMemberRoleMismatch:
			responses.SendError(w, http.StatusConflict, "target user must be an active moderator")
			return
		default:
			responses.SendError(w, http.StatusInternalServerError, "failed to demote moderator: "+err.Error())
			return
		}
	}

	responses.SendSuccess(w, "moderator demoted to member", map[string]int{
		"group_id":   groupID,
		"demoted_by": ownerID,
		"user_id":    targetUserID,
	})
}

func UpdateGroupSettingsHandler(w http.ResponseWriter, r *http.Request) {
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

	if in.Visibility == nil && in.JoinMode == nil {
		responses.SendError(w, http.StatusBadRequest, "at least one of visibility or join_mode is required")
		return
	}

	if in.Visibility != nil {
		v := strings.ToLower(strings.TrimSpace(*in.Visibility))
		if v != "public" && v != "private" {
			responses.SendError(w, http.StatusBadRequest, "visibility must be public or private")
			return
		}
		in.Visibility = &v
	}

	if in.JoinMode != nil {
		jm := strings.ToLower(strings.TrimSpace(*in.JoinMode))
		switch jm {
		case "auto", "request", "invite", "request_and_invite":
		default:
			responses.SendError(w, http.StatusBadRequest, "join_mode must be auto, request, invite, or request_and_invite")
			return
		}
		in.JoinMode = &jm
	}

	updated, err := queries.UpdateGroupSettings(r.Context(), database.DB, ownerID, groupID, in.Visibility, in.JoinMode)
	if err != nil {
		switch err {
		case queries.ErrGroupNotFound:
			responses.SendError(w, http.StatusNotFound, "group not found")
			return
		case queries.ErrNotGroupOwner:
			responses.SendError(w, http.StatusForbidden, "only the group owner can update group settings")
			return
		default:
			responses.SendError(w, http.StatusInternalServerError, "failed to update group settings: "+err.Error())
			return
		}
	}

	responses.SendSuccess(w, "group settings updated successfully", map[string]interface{}{
		"group_id":   updated.GroupID,
		"visibility": updated.Visibility,
		"join_mode":  updated.JoinMode,
	})
}

func GetGroupSettingsHandler(w http.ResponseWriter, r *http.Request) {
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

	settings, err := queries.GetGroupSettings(r.Context(), database.DB, ownerID, groupID)
	if err != nil {
		switch err {
		case queries.ErrGroupNotFound:
			responses.SendError(w, http.StatusNotFound, "group not found")
			return
		case queries.ErrNotGroupOwner:
			responses.SendError(w, http.StatusForbidden, "only the group owner can view group settings")
			return
		default:
			responses.SendError(w, http.StatusInternalServerError, "failed to get group settings: "+err.Error())
			return
		}
	}

	responses.SendSuccess(w, "group settings retrieved successfully", map[string]interface{}{
		"group_id":   settings.GroupID,
		"visibility": settings.Visibility,
		"join_mode":  settings.JoinMode,
	})
}

func GroupMembersListHandler(w http.ResponseWriter, r *http.Request) {
	_, err := middleware.UserIDFromContext(r.Context())
	if err != nil {
		responses.SendError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	groupID, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || groupID <= 0 {
		responses.SendError(w, http.StatusBadRequest, "invalid group id")
		return
	}

	members, err := queries.GetActiveGroupMembers(r.Context(), database.DB, groupID)
	if err != nil {
		switch err {
		case queries.ErrGroupNotFound:
			responses.SendError(w, http.StatusNotFound, "group not found")
			return
		default:
			responses.SendError(w, http.StatusInternalServerError, "failed to get group members: "+err.Error())
			return
		}
	}

	responses.SendSuccess(w, "group active members", members)
}

func PendingGroupRequestsHandler(w http.ResponseWriter, r *http.Request) {
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

	items, err := queries.GetPendingGroupJoinRequests(r.Context(), database.DB, actorID, groupID)
	if err != nil {
		switch err {
		case queries.ErrGroupNotFound:
			responses.SendError(w, http.StatusNotFound, "group not found")
			return
		case queries.ErrNotGroupModeratorOrOwner:
			responses.SendError(w, http.StatusForbidden, "only group owner or moderators can view pending requests")
			return
		default:
			responses.SendError(w, http.StatusInternalServerError, "failed to get pending requests: "+err.Error())
			return
		}
	}

	responses.SendSuccess(w, "group pending requests", items)
}

func PendingGroupInvitesHandler(w http.ResponseWriter, r *http.Request) {
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

	items, err := queries.GetPendingGroupInvites(r.Context(), database.DB, actorID, groupID)
	if err != nil {
		switch err {
		case queries.ErrGroupNotFound:
			responses.SendError(w, http.StatusNotFound, "group not found")
			return
		case queries.ErrNotGroupModeratorOrOwner:
			responses.SendError(w, http.StatusForbidden, "only group owner or moderators can view sent invites")
			return
		default:
			responses.SendError(w, http.StatusInternalServerError, "failed to get pending invites: "+err.Error())
			return
		}
	}

	responses.SendSuccess(w, "group pending invites", items)
}

func RemoveInviteHandler(w http.ResponseWriter, r *http.Request) {
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

	err = queries.RemovePendingGroupInvite(r.Context(), database.DB, actorID, groupID, targetUserID)
	if err != nil {
		switch err {
		case queries.ErrGroupNotFound:
			responses.SendError(w, http.StatusNotFound, "group not found")
			return
		case queries.ErrNotGroupModeratorOrOwner:
			responses.SendError(w, http.StatusForbidden, "only group owner or moderators can remove invites")
			return
		case queries.ErrGroupInviteNotFound:
			responses.SendError(w, http.StatusNotFound, "pending invite not found")
			return
		case queries.ErrGroupMemberIsActive:
			responses.SendError(w, http.StatusConflict, "cannot remove invite for active member")
			return
		default:
			responses.SendError(w, http.StatusInternalServerError, "failed to remove invite: "+err.Error())
			return
		}
	}

	responses.SendSuccess(w, "pending invite removed", map[string]int{
		"group_id":   groupID,
		"removed_by": actorID,
		"user_id":    targetUserID,
	})
}

func RemoveOwnRequestHandler(w http.ResponseWriter, r *http.Request) {
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

	err = queries.RemoveOwnPendingGroupRequest(r.Context(), database.DB, userID, groupID)
	if err != nil {
		switch err {
		case queries.ErrGroupJoinRequestNotFound:
			responses.SendError(w, http.StatusNotFound, "pending request not found")
			return
		case queries.ErrGroupMemberIsActive:
			responses.SendError(w, http.StatusConflict, "active members cannot remove pending requests")
			return
		default:
			responses.SendError(w, http.StatusInternalServerError, "failed to remove request: "+err.Error())
			return
		}
	}

	responses.SendSuccess(w, "pending request removed", map[string]int{
		"group_id": groupID,
		"user_id":  userID,
	})
}

func GetGroupPageHandler(w http.ResponseWriter, r *http.Request) {
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

	view, err := queries.GetGroupPageView(r.Context(), database.DB, viewerID, groupID)
	if err != nil {
		switch err {
		case queries.ErrGroupNotFound:
			responses.SendError(w, http.StatusNotFound, "group not found")
			return
		default:
			responses.SendError(w, http.StatusInternalServerError, "failed to load group page: "+err.Error())
			return
		}
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
