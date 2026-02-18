package handlers

import (
	database "backend/pkg/db/sqlite"
	"backend/pkg/middleware"
	"backend/pkg/models"
	"backend/pkg/responses"
	"backend/pkg/services"
	"backend/pkg/utils"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
)

func CreatePostHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("[INFO] CreatePostHandler: Received request")

	if err := r.ParseMultipartForm(20 << 20); err != nil {
		log.Printf("[ERROR] CreatePostHandler: ParseMultipartForm failed: %v", err)
		responses.SendError(w, http.StatusBadRequest, "Invalid Form")
		return
	}

	userID, _ := middleware.UserIDFromContext(r.Context())
	log.Printf("[INFO] CreatePostHandler: Creating post for UserID: %d", userID)

	privacy := r.FormValue("privacy")
	var wlIDs []int

	if privacy == "custom" {
		wl := r.MultipartForm.Value["whitelisted_users"]
		log.Printf("[INFO] CreatePostHandler: Custom privacy detected, processing %d whitelisted users", len(wl))

		for _, idStr := range wl {
			id, err := strconv.Atoi(idStr)
			if err != nil {
				log.Printf("[ERROR] CreatePostHandler: Invalid whitelisted ID '%s': %v", idStr, err)
				responses.SendError(w, http.StatusInternalServerError, err.Error())
				return
			}
			wlIDs = append(wlIDs, id)
		}
	}

	post := models.Post{
		UserID:           userID,
		Content:          r.FormValue("content"),
		Privacy:          privacy,
		WhitelistedUsers: wlIDs,
	}

	imageURL, err := utils.AttachAvatar(r)
	if err != nil {
		log.Printf("[ERROR] CreatePostHandler: File attachment failed: %v", err)
		responses.SendError(w, http.StatusInternalServerError, err.Error())
		return
	}

	if imageURL != "" {
		log.Printf("[INFO] CreatePostHandler: Image attached: %s", imageURL)
		post.Image = imageURL
	}

	postService := services.NewPostService(database.DB)

	if err := postService.CreatePost(r.Context(), post); err != nil {
		log.Printf("[ERROR] CreatePostHandler: Service call failed: %v", err)
		responses.SendError(w, http.StatusInternalServerError, err.Error())
		return
	}

	log.Printf("[SUCCESS] CreatePostHandler: Post created for UserID: %d", userID)
	responses.SendCreated(w, "user created successfully", nil)
}

func UpdatePostHandler(w http.ResponseWriter, r *http.Request) {
	postID, _ := strconv.Atoi(r.PathValue("id"))
	log.Printf("[INFO] UpdatePostHandler: Updating PostID: %d", postID)

	var data struct {
		Content string `json:"content"`
	}
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		log.Printf("[ERROR] UpdatePostHandler: JSON decode failed: %v", err)
		responses.SendError(w, http.StatusBadRequest, "Invalid JSON")
		return
	}

	postService := services.NewPostService(database.DB)
	if err := postService.UpdatePost(r.Context(), postID, data.Content); err != nil {
		log.Printf("[ERROR] UpdatePostHandler: Service call failed: %v", err)
		responses.SendError(w, http.StatusInternalServerError, err.Error())
		return
	}

	log.Printf("[SUCCESS] UpdatePostHandler: Post %d updated", postID)
	responses.SendSuccess(w, "post updated successfully", nil)
}

func DeletePostHandler(w http.ResponseWriter, r *http.Request) {
	postIDStr := r.PathValue("id")
	postIDInt, err := strconv.Atoi(postIDStr)
	log.Printf("[INFO] DeletePostHandler: Deleting PostID: %s", postIDStr)

	if err != nil {
		log.Printf("[ERROR] DeletePostHandler: Invalid PostID format: %v", err)
		responses.SendError(w, http.StatusBadRequest, "Invalid post ID")
		return
	}

	postService := services.NewPostService(database.DB)

	err = postService.DeletePost(r.Context(), postIDInt)
	if err != nil {
		log.Printf("[ERROR] DeletePostHandler: Service call failed: %v", err)
		responses.SendError(w, http.StatusInternalServerError, err.Error())
		return
	}

	log.Printf("[SUCCESS] DeletePostHandler: Post %d marked as deleted", postIDInt)
	responses.SendSuccess(w, "post deleted successfully", nil)
}

func RestorePostHandler(w http.ResponseWriter, r *http.Request) {
	postIDStr := r.PathValue("id")
	postIDInt, err := strconv.Atoi(postIDStr)
	log.Printf("[INFO] RestorePostHandler: Restoring PostID: %s", postIDStr)

	if err != nil {
		log.Printf("[ERROR] RestorePostHandler: Invalid PostID format: %v", err)
		responses.SendError(w, http.StatusBadRequest, "Invalid post ID")
		return
	}

	postService := services.NewPostService(database.DB)

	err = postService.RestorePost(r.Context(), postIDInt)
	if err != nil {
		log.Printf("[ERROR] RestorePostHandler: Service call failed: %v", err)
		responses.SendError(w, http.StatusInternalServerError, err.Error())
		return
	}

	log.Printf("[SUCCESS] RestorePostHandler: Post %d restored", postIDInt)
	responses.SendSuccess(w, "post restored successfully", nil)
}

func GetPostHandler(w http.ResponseWriter, r *http.Request) {
	postIDStr := r.PathValue("id")
	postIDInt, err := strconv.Atoi(postIDStr)
	log.Printf("[INFO] GetPostHandler: Fetching PostID: %s", postIDStr)

	if err != nil {
		log.Printf("[ERROR] GetPostHandler: Invalid PostID format: %v", err)
		responses.SendError(w, http.StatusBadRequest, "Invalid post ID")
		return
	}

	userID, _ := middleware.UserIDFromContext(r.Context())

	postService := services.NewPostService(database.DB)

	post, err := postService.GetPostByID(r.Context(), userID, postIDInt)
	if err != nil {
		log.Printf("[ERROR] GetPostHandler: Service failed to retrieve post %d: %v", postIDInt, err)
		responses.SendError(w, http.StatusInternalServerError, err.Error())
		return
	}

	log.Printf("[SUCCESS] GetPostHandler: Post %d retrieved for UserID %d", postIDInt, userID)
	responses.SendSuccess(w, "post retrieved successfully", post)
}
