package handlers

import (
	database "backend/pkg/db/sqlite"
	"backend/pkg/middleware"
	"backend/pkg/models"
	"backend/pkg/responses"
	"backend/pkg/services"
	"net/http"
	"strconv"
)

func GetFeedHandler(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.UserIDFromContext(r.Context())

	pageStr := r.URL.Query().Get("page")
	page, _ := strconv.Atoi(pageStr)
	if page < 1 {
		page = 1
	}

	limit := 10
	offset := (page - 1) * limit

	postService := services.NewPostService(database.DB)
	userService := services.NewUserService(database.DB)

	posts, err := postService.GetFeedPosts(r.Context(), userID, limit, offset)
	if err != nil {
		responses.SendError(w, http.StatusInternalServerError, "Failed to load feed")
		return
	}

	var suggestions []models.DiscoveredUser
	if page == 1 {
		suggestions, err = userService.DiscoveredUser(r.Context(), userID, 5)
		if err != nil {
			suggestions = []models.DiscoveredUser{}
		}
	} else {
		suggestions = nil
	}

	// 4. Response
	responses.SendSuccess(w, "Feed loaded", map[string]interface{}{
		"posts":       posts,
		"suggestions": suggestions,
		"page":        page,
	})
}
