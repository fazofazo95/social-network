package handlers

import (
	database "backend/pkg/db/sqlite"
	"backend/pkg/middleware"
	"backend/pkg/models"
	"backend/pkg/responses"
	"backend/pkg/services"
	"backend/pkg/utils"
	"encoding/json"
	"net/http"
	"strconv"
)

func CreatePostHandler(w http.ResponseWriter, r *http.Request) {

	if err := r.ParseMultipartForm(20 << 20); err != nil {
		responses.SendError(w, http.StatusBadRequest, "Invalid Form")
		return
	}

	userID, err := strconv.Atoi(r.FormValue("user_id"))
	if err != nil {
		responses.SendError(w, http.StatusBadRequest, "Invalid Form")
		return
	}

	privacy := r.FormValue("privacy")

	var wlIDs []int

	if privacy == "custom" {
		wl := r.MultipartForm.Value["whitelisted_users"]

		for _, idStr := range wl {
			id, err := strconv.Atoi(idStr)
			if err != nil {
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
		responses.SendError(w, http.StatusInternalServerError, err.Error())
		return
	}

	if imageURL != "" {
		post.Image = imageURL
	}

	postService := services.NewPostService(database.DB)

	if err := postService.CreatePost(r.Context(), post); err != nil {
		responses.SendError(w, http.StatusInternalServerError, err.Error())
		return
	}

	responses.SendCreated(w, "user created successfully", nil)
}

func UpdatePostHandler(w http.ResponseWriter, r *http.Request) {
	postID := r.PathValue("id")
	postIDInt, err := strconv.Atoi(postID)
	if err != nil {
		responses.SendError(w, http.StatusBadRequest, "Invalid post ID")
		return
	}

	userID, _ := middleware.UserIDFromContext(r.Context())

	var updateData models.UpdateData
	if err := json.NewDecoder(r.Body).Decode(&updateData); err != nil {
		responses.SendError(w, http.StatusBadRequest, "Invalid JSON body")
		return
	}

	updateData.ParentID = postIDInt

	postService := services.NewPostService(database.DB)
	owner, err := postService.IsOwner(r.Context(), userID, postIDInt)
	if err != nil {
		responses.SendError(w, http.StatusInternalServerError, err.Error())
		return
	}

	if !owner {
		responses.SendError(w, http.StatusForbidden, "You do not have permission to update this post")
		return
	}

	err = postService.UpdatePost(r.Context(), userID, updateData)
	if err != nil {
		responses.SendError(w, http.StatusInternalServerError, err.Error())
		return
	}

	responses.SendSuccess(w, "post updated successfully", nil)
}

func DeletePostHandler(w http.ResponseWriter, r *http.Request) {

}

func RestorePostHandler(w http.ResponseWriter, r *http.Request) {
}

func GetPostHandler(w http.ResponseWriter, r *http.Request) {

}
