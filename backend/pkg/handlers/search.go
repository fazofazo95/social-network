package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"backend/pkg/middleware"
	"backend/pkg/responses"
	"backend/pkg/services"
)

type SearchHandler struct {
	profileService services.ProfileService
	groupService   services.GroupService
}

func NewSearchHandler(ps services.ProfileService, gs services.GroupService) *SearchHandler {
	return &SearchHandler{profileService: ps, groupService: gs}
}

func (h *SearchHandler) RegisterRoutes(mux *http.ServeMux) {
	auth := middleware.WithAuth
	mux.Handle("GET /api/search", middleware.Chain(h.Search, auth))
}

func (h *SearchHandler) Search(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.UserIDFromContext(r.Context())
	if err != nil {
		responses.SendError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	query := strings.TrimSpace(r.URL.Query().Get("q"))
	if query == "" {
		query = strings.TrimSpace(r.URL.Query().Get("query"))
	}
	if query == "" {
		responses.SendError(w, http.StatusBadRequest, "q query parameter is required")
		return
	}

	limit := 10
	if limitStr := strings.TrimSpace(r.URL.Query().Get("limit")); limitStr != "" {
		parsed, err := strconv.Atoi(limitStr)
		if err != nil || parsed <= 0 {
			responses.SendError(w, http.StatusBadRequest, "limit must be a positive integer")
			return
		}
		if parsed > 25 {
			parsed = 25
		}
		limit = parsed
	}

	users, err := h.profileService.SearchUsers(r.Context(), userID, query, limit)
	if err != nil {
		responses.SendError(w, http.StatusInternalServerError, "failed to search users: "+err.Error())
		return
	}

	groups, err := h.groupService.SearchGroups(r.Context(), userID, query, limit)
	if err != nil {
		responses.SendError(w, http.StatusInternalServerError, "failed to search groups: "+err.Error())
		return
	}

	responses.SendSuccess(w, "search results", map[string]interface{}{
		"query":  query,
		"limit":  limit,
		"users":  users,
		"groups": groups,
	})
}
