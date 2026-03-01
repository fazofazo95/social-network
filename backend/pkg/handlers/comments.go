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

func (h *CommentHandler)  CreateCommentHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("[INFO] CreateCommentHandler: Received request")
	if err := r.ParseMultipartForm(20 << 20); err != nil {
		log.Printf("[ERROR] CreateCommentHandler: ParseMultipartForm failed: %v", err)
		responses.SendError(w, http.StatusBadRequest, "Invalid Form")
		return
	}

	userID, _ := middleware.UserIDFromContext(r.Context())
	log.Printf("[INFO] CreateCommentHandler: UserID from context: %d", userID)

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
		log.Printf("[INFO] CreateCommentHandler: Image attached: %s", imageURL)
		comment.ExtraContent = imageURL
	}

	if _, err := h.Service.CreateComment(r.Context(), userID, comment); err != nil {
		log.Printf("[ERROR] CreateCommentHandler: Service call failed: %v", err)
		responses.SendError(w, http.StatusInternalServerError, err.Error())
		return
	}

	log.Println("[SUCCESS] CreateCommentHandler: Comment created")
	responses.SendCreated(w, "comment created successfully", nil)
}

func (h *CommentHandler)  UpdateCommentHandler(w http.ResponseWriter, r *http.Request) {
	commentID := r.PathValue("id")
	log.Printf("[INFO] UpdateCommentHandler: Received request for ID: %s", commentID)

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

	log.Printf("[SUCCESS] UpdateCommentHandler: Comment %d updated", commentIDInt)
	responses.SendSuccess(w, "comment updated successfully", nil)
}

func (h *CommentHandler)  DeleteCommentHandler(w http.ResponseWriter, r *http.Request) {
	commentIDStr := r.PathValue("id")
	commentID, _ := strconv.Atoi(commentIDStr)
	log.Printf("[INFO] DeleteCommentHandler: Received request for ID: %d", commentID)

	userID, _ := middleware.UserIDFromContext(r.Context())

	if err := h.Service.DeleteComment(r.Context(), commentID, userID); err != nil {
		log.Printf("[ERROR] DeleteCommentHandler: Service call failed: %v", err)
		responses.SendError(w, http.StatusInternalServerError, err.Error())
		return
	}

	log.Printf("[SUCCESS] DeleteCommentHandler: Comment %d deleted", commentID)
	responses.SendSuccess(w, "comment deleted successfully", nil)
}

func (h *CommentHandler)  RestoreCommentHandler(w http.ResponseWriter, r *http.Request) {
	commentIDStr := r.PathValue("id")
	commentID, _ := strconv.Atoi(commentIDStr)
	log.Printf("[INFO] RestoreCommentHandler: Received request for ID: %d", commentID)

	userID, _ := middleware.UserIDFromContext(r.Context())

	err := h.Service.RestoreComment(r.Context(), commentID, userID)
	if err != nil {
		log.Printf("[ERROR] RestoreCommentHandler: Service call failed: %v", err)
		responses.SendError(w, http.StatusInternalServerError, err.Error())
		return
	}

	log.Printf("[SUCCESS] RestoreCommentHandler: Comment %d restored", commentID)
	responses.SendSuccess(w, "comment restored successfully", nil)
}

func (h *CommentHandler)  GetPostCommentsHandler(w http.ResponseWriter, r *http.Request) {
	postID := r.PathValue("id")
	log.Printf("[INFO] GetPostCommentsHandler: Fetching for PostID: %s", postID)

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

	log.Printf("[SUCCESS] GetPostCommentsHandler: Retrieved %d comments", len(comments))
	responses.SendSuccess(w, "comments retrieved successfully", comments)
}
