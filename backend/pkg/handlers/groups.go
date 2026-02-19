package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"backend/pkg/db/queries"
	database "backend/pkg/db/sqlite"
	"backend/pkg/middleware"
	"backend/pkg/models"
	"backend/pkg/responses"
)

func CreateGroupHandler(w http.ResponseWriter, r *http.Request) {
	ownerID, err := middleware.UserIDFromContext(r.Context())
	if err != nil {
		responses.SendError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var in models.CreateGroupInput
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		responses.SendError(w, http.StatusBadRequest, "invalid JSON")
		return
	}

	in.Name = strings.TrimSpace(in.Name)
	in.Description = strings.TrimSpace(in.Description)
	in.Visibility = strings.ToLower(strings.TrimSpace(in.Visibility))
	in.JoinMode = strings.ToLower(strings.TrimSpace(in.JoinMode))

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
