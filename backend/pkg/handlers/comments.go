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

func CreateCommentHandler(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(20 << 20); err != nil {
		responses.SendError(w, http.StatusBadRequest, "Invalid Form")
		return
	}

	userID, _ := middleware.UserIDFromContext(r.Context())

	comment := models.Comment{
		UserID:     userID,
		Content:    r.FormValue("content"),
		ParentType: r.FormValue("parent_type"),
	}

	parentIDint, err := strconv.Atoi(r.FormValue("parent_id"))
	if err != nil {
		responses.SendError(w, http.StatusBadRequest, "Invalid parent_id")
		return
	}

	comment.ParentID = parentIDint

	imageURL, err := utils.AttachAvatar(r)
	if err != nil {
		responses.SendError(w, http.StatusInternalServerError, err.Error())
		return
	}

	if imageURL != "" {
		comment.ExtraContent = imageURL
	}

	commentService := services.NewCommentService(database.DB)

	if err := commentService.CreateComment(r.Context(), comment); err != nil {
		responses.SendError(w, http.StatusInternalServerError, err.Error())
		return
	}

	responses.SendCreated(w, "comment created successfully", nil)
}

func UpdateCommentHandler(w http.ResponseWriter, r *http.Request) {
	commentID := r.PathValue("id")
	commentIDInt, err := strconv.Atoi(commentID)
	if err != nil {
		responses.SendError(w, http.StatusBadRequest, "Invalid comment ID")
		return
	}

	var updateData struct {
		Content string `json:"content"`
	}
	if err := json.NewDecoder(r.Body).Decode(&updateData); err != nil {
		responses.SendError(w, http.StatusBadRequest, "Invalid JSON body")
		return
	}

	commentService := services.NewCommentService(database.DB)

	err = commentService.UpdateComment(r.Context(), commentIDInt, updateData.Content)
	if err != nil {
		responses.SendError(w, http.StatusInternalServerError, err.Error())
		return
	}

	responses.SendSuccess(w, "comment updated successfully", nil)
}

func DeleteCommentHandler(w http.ResponseWriter, r *http.Request) {
	commentID, _ := strconv.Atoi(r.PathValue("id"))

	commentService := services.NewCommentService(database.DB)

	if err := commentService.DeleteComment(r.Context(), commentID); err != nil {
		responses.SendError(w, http.StatusInternalServerError, err.Error())
		return
	}

	responses.SendSuccess(w, "comment deleted successfully", nil)
}

func RestoreCommentHandler(w http.ResponseWriter, r *http.Request) {
	commentID, _ := strconv.Atoi(r.PathValue("id"))

	commentService := services.NewCommentService(database.DB)

	err := commentService.RestoreComment(r.Context(), commentID)
	if err != nil {
		responses.SendError(w, http.StatusInternalServerError, err.Error())
		return
	}

	responses.SendSuccess(w, "comment restored successfully", nil)
}

func GetPostCommentsHandler(w http.ResponseWriter, r *http.Request) {
	postID := r.PathValue("id")
	postIDInt, err := strconv.Atoi(postID)
	if err != nil {
		responses.SendError(w, http.StatusBadRequest, "Invalid post ID")
		return
	}

	userID, _ := middleware.UserIDFromContext(r.Context())

	commentService := services.NewCommentService(database.DB)

	comments, err := commentService.GetPostComments(r.Context(), postIDInt, userID)
	if err != nil {
		responses.SendError(w, http.StatusInternalServerError, "Failed to retrieve comments")
		return
	}

	responses.SendSuccess(w, "comments retrieved successfully", comments)
}
