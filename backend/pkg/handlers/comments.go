package handlers

import (
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

type CommentHandler struct {
	Service services.CommentService
}

func NewCommentHandler(s services.CommentService) *CommentHandler {
	return &CommentHandler{Service: s}
}

func (h *CommentHandler) RegisterRoutes(mux *http.ServeMux) {
	auth := middleware.WithAuth

	mux.Handle("GET /api/posts/{id}/comments", middleware.Chain(h.GetPostCommentsHandler, auth))

	mux.Handle("POST /api/comments", middleware.Chain(h.CreateCommentHandler, auth))
	mux.Handle("PUT /api/comments/{id}", middleware.Chain(h.UpdateCommentHandler, auth))
	mux.Handle("PUT /api/comments/{id}/delete", middleware.Chain(h.DeleteCommentHandler, auth))
	mux.Handle("PUT /api/comments/{id}/restore", middleware.Chain(h.RestoreCommentHandler, auth))
}

func (h *CommentHandler) CreateCommentHandler(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(20 << 20); err != nil {
		log.Printf("[ERROR] CreateCommentHandler: ParseMultipartForm failed: %v", err)
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
		log.Printf("[ERROR] CreateCommentHandler: Invalid parent_id '%s': %v", r.FormValue("parent_id"), err)
		responses.SendError(w, http.StatusBadRequest, "Invalid parent_id")
		return
	}

	comment.ParentID = parentIDint

	imageURL, err := utils.AttachAvatar(r)
	if err != nil {
		log.Printf("[ERROR] CreateCommentHandler: AttachAvatar failed: %v", err)
		responses.SendError(w, http.StatusInternalServerError, err.Error())
		return
	}

	if imageURL != "" {
		comment.ExtraContent = imageURL
	}

	if _, err := h.Service.CreateComment(r.Context(), userID, comment); err != nil {
		log.Printf("[ERROR] CreateCommentHandler: Service call failed: %v", err)
		responses.SendError(w, http.StatusInternalServerError, err.Error())
		return
	}

	responses.SendCreated(w, "comment created successfully", nil)
}

func (h *CommentHandler) UpdateCommentHandler(w http.ResponseWriter, r *http.Request) {
	commentID := r.PathValue("id")

	commentIDInt, err := strconv.Atoi(commentID)
	if err != nil {
		log.Printf("[ERROR] UpdateCommentHandler: Invalid ID format: %v", err)
		responses.SendError(w, http.StatusBadRequest, "Invalid comment ID")
		return
	}

	userID, _ := middleware.UserIDFromContext(r.Context())

	var updateData struct {
		Content string `json:"content"`
	}
	if err := json.NewDecoder(r.Body).Decode(&updateData); err != nil {
		log.Printf("[ERROR] UpdateCommentHandler: JSON decode failed: %v", err)
		responses.SendError(w, http.StatusBadRequest, "Invalid JSON body")
		return
	}

	err = h.Service.UpdateComment(r.Context(), commentIDInt, userID, updateData.Content)
	if err != nil {
		log.Printf("[ERROR] UpdateCommentHandler: Service call failed: %v", err)
		responses.SendError(w, http.StatusInternalServerError, err.Error())
		return
	}

	responses.SendSuccess(w, "comment updated successfully", nil)
}

func (h *CommentHandler) DeleteCommentHandler(w http.ResponseWriter, r *http.Request) {
	commentIDStr := r.PathValue("id")
	commentID, _ := strconv.Atoi(commentIDStr)

	userID, _ := middleware.UserIDFromContext(r.Context())

	if err := h.Service.DeleteComment(r.Context(), commentID, userID); err != nil {
		log.Printf("[ERROR] DeleteCommentHandler: Service call failed: %v", err)
		responses.SendError(w, http.StatusInternalServerError, err.Error())
		return
	}

	responses.SendSuccess(w, "comment deleted successfully", nil)
}

func (h *CommentHandler) RestoreCommentHandler(w http.ResponseWriter, r *http.Request) {
	commentIDStr := r.PathValue("id")
	commentID, _ := strconv.Atoi(commentIDStr)

	userID, _ := middleware.UserIDFromContext(r.Context())

	err := h.Service.RestoreComment(r.Context(), commentID, userID)
	if err != nil {
		log.Printf("[ERROR] RestoreCommentHandler: Service call failed: %v", err)
		responses.SendError(w, http.StatusInternalServerError, err.Error())
		return
	}

	responses.SendSuccess(w, "comment restored successfully", nil)
}

func (h *CommentHandler) GetPostCommentsHandler(w http.ResponseWriter, r *http.Request) {
	postID := r.PathValue("id")

	postIDInt, err := strconv.Atoi(postID)
	if err != nil {
		log.Printf("[ERROR] GetPostCommentsHandler: Invalid PostID: %v", err)
		responses.SendError(w, http.StatusBadRequest, "Invalid post ID")
		return
	}

	userID, _ := middleware.UserIDFromContext(r.Context())

	comments, err := h.Service.GetPostComments(r.Context(), postIDInt, userID)
	if err != nil {
		log.Printf("[ERROR] GetPostCommentsHandler: Service call failed: %v", err)
		responses.SendError(w, http.StatusInternalServerError, "Failed to retrieve comments")
		return
	}

	responses.SendSuccess(w, "comments retrieved successfully", comments)
}
