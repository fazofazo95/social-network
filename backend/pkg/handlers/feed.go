package handlers

import (
	database "backend/pkg/db/sqlite"
	"backend/pkg/middleware"
	"backend/pkg/models"
	"backend/pkg/responses"
	"backend/pkg/services"
	"log"
	"net/http"
	"strconv"
)

func GetFeedHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("[INFO] GetFeedHandler: Received request")

	userID, _ := middleware.UserIDFromContext(r.Context())
	log.Printf("[INFO] GetFeedHandler: Fetching feed for UserID: %d", userID)

	pageStr := r.URL.Query().Get("page")
	page, _ := strconv.Atoi(pageStr)
	if page < 1 {
		page = 1
	}

	limit := 10
	offset := (page - 1) * limit
	log.Printf("[INFO] GetFeedHandler: Pagination - Page: %d, Limit: %d, Offset: %d", page, limit, offset)

	postService := services.NewPostService(database.DB)
	userService := services.NewUserService(database.DB)

	posts, err := postService.GetFeedPosts(r.Context(), userID, limit, offset)
	if err != nil {
		log.Printf("[ERROR] GetFeedHandler: Failed to get feed posts for UserID %d: %v", userID, err)
		responses.SendError(w, http.StatusInternalServerError, "Failed to load feed")
		return
	}

	var suggestions []models.DiscoveredUser
	if page == 1 {
		log.Printf("[INFO] GetFeedHandler: Page 1 detected, fetching user suggestions for UserID: %d", userID)
		suggestions, err = userService.DiscoveredUser(r.Context(), userID, 5)
		if err != nil {
			log.Printf("[WARN] GetFeedHandler: Failed to fetch suggestions: %v", err)
			suggestions = []models.DiscoveredUser{}
		}
	} else {
		suggestions = nil
	}

	log.Printf("[SUCCESS] GetFeedHandler: Returning %d posts and %d suggestions", len(posts), len(suggestions))
	responses.SendSuccess(w, "Feed loaded", map[string]interface{}{
		"posts":       posts,
		"suggestions": suggestions,
		"page":        page,
	})
}
