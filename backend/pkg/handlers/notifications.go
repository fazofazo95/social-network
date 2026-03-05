package handlers

import (
	"backend/pkg/middleware"
	"backend/pkg/models"
	"backend/pkg/repository"
	"backend/pkg/responses"
	"backend/pkg/services"
	"backend/pkg/sse"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type NotificationHandler struct {
	Service       services.NotificationService
	FollowService services.FollowService
	GroupService  services.GroupService
	Hub           *sse.Hub
}

func NewNotificationHandler(ns services.NotificationService, fs services.FollowService, gs services.GroupService, hub *sse.Hub) *NotificationHandler {
	return &NotificationHandler{Service: ns, FollowService: fs, GroupService: gs, Hub: hub}
}

func (h *NotificationHandler) RegisterRoutes(mux *http.ServeMux) {
	auth := middleware.WithAuth

	mux.Handle("GET /api/notifications", middleware.Chain(h.ListNotifications, auth))
	mux.Handle("POST /api/notifications/read-all", middleware.Chain(h.MarkAllSeen, auth))
	mux.Handle("POST /api/notifications/{id}/read", middleware.Chain(h.MarkSeen, auth))
	mux.Handle("POST /api/notifications/{id}/action", middleware.Chain(h.RespondToNotification, auth))
	mux.Handle("GET /api/notifications/stream", middleware.Chain(h.Stream, auth))
}

func (h *NotificationHandler) ListNotifications(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.UserIDFromContext(r.Context())
	if err != nil {
		responses.SendError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	limit := 20
	if raw := strings.TrimSpace(r.URL.Query().Get("limit")); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil || parsed <= 0 {
			responses.SendError(w, http.StatusBadRequest, "invalid limit")
			return
		}
		limit = parsed
	}

	offset := 0
	if raw := strings.TrimSpace(r.URL.Query().Get("offset")); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil || parsed < 0 {
			responses.SendError(w, http.StatusBadRequest, "invalid offset")
			return
		}
		offset = parsed
	}

	items, err := h.Service.ListByUser(r.Context(), userID, limit, offset)
	if err != nil {
		responses.SendError(w, http.StatusInternalServerError, "failed to list notifications: "+err.Error())
		return
	}

	responses.SendSuccess(w, "notifications fetched", map[string]any{
		"items":    items,
		"limit":    limit,
		"offset":   offset,
		"has_more": len(items) == limit,
	})
}

func (h *NotificationHandler) MarkSeen(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.UserIDFromContext(r.Context())
	if err != nil {
		responses.SendError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	notificationID, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || notificationID <= 0 {
		responses.SendError(w, http.StatusBadRequest, "invalid notification id")
		return
	}

	if err := h.Service.MarkSeen(r.Context(), userID, notificationID); err != nil {
		if err == repository.ErrNotificationNotFound {
			responses.SendError(w, http.StatusNotFound, "notification not found")
			return
		}
		responses.SendError(w, http.StatusInternalServerError, "failed to mark notification as seen: "+err.Error())
		return
	}

	responses.SendSuccess(w, "notification marked as seen", map[string]int{"notification_id": notificationID})
}

func (h *NotificationHandler) MarkAllSeen(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.UserIDFromContext(r.Context())
	if err != nil {
		responses.SendError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	count, err := h.Service.MarkAllSeen(r.Context(), userID)
	if err != nil {
		responses.SendError(w, http.StatusInternalServerError, "failed to mark notifications as seen: "+err.Error())
		return
	}

	responses.SendSuccess(w, "notifications marked as seen", map[string]int64{"updated": count})
}

func (h *NotificationHandler) RespondToNotification(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.UserIDFromContext(r.Context())
	if err != nil {
		responses.SendError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	notificationID, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || notificationID <= 0 {
		responses.SendError(w, http.StatusBadRequest, "invalid notification id")
		return
	}

	var in models.NotificationActionInput
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		responses.SendError(w, http.StatusBadRequest, "invalid JSON")
		return
	}
	in.Action = strings.ToLower(strings.TrimSpace(in.Action))
	if in.Action != "accept" && in.Action != "reject" {
		responses.SendError(w, http.StatusBadRequest, "action must be accept or reject")
		return
	}

	notification, err := h.Service.GetByIDForUser(r.Context(), userID, notificationID)
	if err != nil {
		switch err {
		case repository.ErrNotificationNotFound:
			responses.SendError(w, http.StatusNotFound, "notification not found")
		case repository.ErrNotificationUnauthorized:
			responses.SendError(w, http.StatusForbidden, "notification does not belong to user")
		default:
			responses.SendError(w, http.StatusInternalServerError, "failed to load notification: "+err.Error())
		}
		return
	}

	if notification.Status != models.NotificationStatusPending && notification.Status != models.NotificationStatusRead {
		responses.SendError(w, http.StatusConflict, "notification is already resolved")
		return
	}

	if err := h.applyNotificationAction(r.Context(), userID, notification, in.Action); err != nil {
		responses.SendError(w, http.StatusBadRequest, err.Error())
		return
	}

	status := models.NotificationStatusAccepted
	message := "notification accepted"
	if in.Action == "reject" {
		status = models.NotificationStatusRejected
		message = "notification rejected"
	}

	if err := h.Service.SetStatusForUser(r.Context(), userID, notificationID, status); err != nil {
		responses.SendError(w, http.StatusInternalServerError, "failed to update notification status: "+err.Error())
		return
	}

	responses.SendSuccess(w, message, map[string]any{
		"notification_id": notificationID,
		"status":          status,
	})
}

func (h *NotificationHandler) applyNotificationAction(ctx context.Context, userID int, notification models.NotificationWithActor, action string) error {
	accept := action == "accept"

	switch notification.Type {
	case models.NotificationTypeFollowRequest:
		if notification.ActorID == nil {
			return fmt.Errorf("invalid follow request notification")
		}
		if accept {
			_, err := h.FollowService.AcceptFollow(ctx, *notification.ActorID, userID)
			return err
		}
		_, err := h.FollowService.RejectFollow(ctx, *notification.ActorID, userID)
		return err

	case models.NotificationTypeGroupInvite:
		if notification.GroupID == nil {
			return fmt.Errorf("invalid group invite notification")
		}
		if accept {
			return h.GroupService.AcceptGroupInvite(ctx, userID, *notification.GroupID)
		}
		return h.GroupService.RejectGroupInvite(ctx, userID, *notification.GroupID)

	case models.NotificationTypeGroupJoin:
		if notification.GroupID == nil || notification.ActorID == nil {
			return fmt.Errorf("invalid group join request notification")
		}
		if accept {
			return h.GroupService.AcceptGroupJoinRequest(ctx, userID, *notification.GroupID, *notification.ActorID)
		}
		return h.GroupService.RejectGroupJoinRequest(ctx, userID, *notification.GroupID, *notification.ActorID)

	default:
		return fmt.Errorf("notification type does not support action")
	}
}

func (h *NotificationHandler) Stream(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.UserIDFromContext(r.Context())
	if err != nil {
		responses.SendError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	flusher, ok := w.(http.Flusher)
	if !ok {
		responses.SendError(w, http.StatusInternalServerError, "streaming not supported")
		return
	}

	stream := h.Hub.Subscribe(userID)
	defer h.Hub.Unsubscribe(userID, stream)

	fmt.Fprintf(w, "event: ready\ndata: {\"status\":\"ok\"}\n\n")
	flusher.Flush()

	heartbeat := time.NewTicker(25 * time.Second)
	defer heartbeat.Stop()

	for {
		select {
		case <-r.Context().Done():
			return
		case payload, ok := <-stream:
			if !ok {
				return
			}
			encoded, err := json.Marshal(payload)
			if err != nil {
				continue
			}
			fmt.Fprintf(w, "event: %s\ndata: %s\n\n", payload.Event, encoded)
			flusher.Flush()
		case <-heartbeat.C:
			fmt.Fprintf(w, "event: ping\ndata: {}\n\n")
			flusher.Flush()
		}
	}
}
