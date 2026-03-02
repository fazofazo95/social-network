package handlers

import (
	"backend/pkg/middleware"
	"backend/pkg/models"
	"backend/pkg/responses"
	"backend/pkg/services"
	"log"
	"net/http"
	"strconv"
)

type FeedHandler struct {
	postService    services.PostService
	profileService services.ProfileService
}

func NewFeedHandler(postServ services.PostService, profileServ services.ProfileService) *FeedHandler {
	return &FeedHandler{postService: postServ, profileService: profileServ}
}

func (h *FeedHandler) RegisterRoutes(mux *http.ServeMux) {
	auth := middleware.WithAuth

	mux.Handle("GET /api/feed", middleware.Chain(h.GetFeedHandler, auth))
	mux.Handle("GET /api/discover", middleware.Chain(h.DiscoverHandler, auth))
}

func (h *FeedHandler) GetFeedHandler(w http.ResponseWriter, r *http.Request) {

	userID, _ := middleware.UserIDFromContext(r.Context())

	pageStr := r.URL.Query().Get("page")
	page, _ := strconv.Atoi(pageStr)

	posts, err := h.postService.GetUserFeed(r.Context(), userID, page)
	if err != nil {
		responses.SendError(w, http.StatusInternalServerError, "Failed to load feed")
		return
	}

	var suggestions []*models.DiscoveredUser
	if page == 1 {
		suggestions, err = h.profileService.DiscoveredUser(r.Context(), userID, 5)
		if err != nil {
			suggestions = nil
		}
	} else {
		suggestions = nil
	}

	responses.SendSuccess(w, "Feed loaded", map[string]interface{}{
		"posts":       posts,
		"suggestions": suggestions,
		"page":        page,
	})
}

func (h *FeedHandler) DiscoverHandler(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.UserIDFromContext(r.Context())
	if err != nil {
		log.Printf("[ERROR] DiscoverHandler: Authentication failed: %v", err)
		http.Error(w, "User not authenticated", http.StatusUnauthorized)
		return
	}

	const defaultLimit = 5

	users, err := h.profileService.DiscoveredUser(r.Context(), userID, defaultLimit)
	if err != nil {
		responses.SendError(w, http.StatusInternalServerError, "Failed to discover users: "+err.Error())
		return
	}

	responses.SendSuccess(w, "Discovered users", users)
}
