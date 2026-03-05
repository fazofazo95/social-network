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

type PostHandler struct {
	Service services.PostService
}

func NewPostHandler(s services.PostService) *PostHandler {
	return &PostHandler{Service: s}
}

func (h *PostHandler) RegisterRoutes(mux *http.ServeMux) {
	auth := middleware.WithAuth

	mux.Handle("POST /api/posts", middleware.Chain(h.CreatePost, auth))

	mux.Handle("PUT /api/posts/{id}", middleware.Chain(h.UpdatePost, auth))
	mux.Handle("DELETE /api/posts/{id}", middleware.Chain(h.DeletePost, auth))
	mux.Handle("PUT /api/posts/{id}/restore", middleware.Chain(h.RestorePost, auth))

	mux.Handle("GET /api/posts/{id}", middleware.Chain(h.GetPostHandler, auth))
}

func (h *PostHandler) CreatePost(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(20 << 20); err != nil {
		log.Printf("[ERROR] CreatePostHandler: ParseMultipartForm failed: %v", err)
		responses.SendError(w, http.StatusBadRequest, "Invalid Form")
		return
	}

	userID, _ := middleware.UserIDFromContext(r.Context())

	privacy := r.FormValue("privacy")

	var wlIDs []int
	if len(r.MultipartForm.Value["whitelisted_users"]) > 0 {
		wl := r.MultipartForm.Value["whitelisted_users"]
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
		post.Image = imageURL
	}

	if _, err := h.Service.CreatePost(r.Context(), userID, &post); err != nil {
		log.Printf("[ERROR] CreatePostHandler: Service call failed: %v", err)
		responses.SendError(w, http.StatusInternalServerError, err.Error())
		return
	}

	responses.SendCreated(w, "user created successfully", nil)
}

func (h *PostHandler) UpdatePost(w http.ResponseWriter, r *http.Request) {
	postID, _ := strconv.Atoi(r.PathValue("id"))
	userID, _ := middleware.UserIDFromContext(r.Context())

	var data struct {
		Content string `json:"content"`
	}
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		log.Printf("[ERROR] UpdatePostHandler: JSON decode failed: %v", err)
		responses.SendError(w, http.StatusBadRequest, "Invalid JSON")
		return
	}

	if err := h.Service.UpdatePost(r.Context(), postID, userID, data.Content); err != nil {
		log.Printf("[ERROR] UpdatePostHandler: Service call failed: %v", err)
		responses.SendError(w, http.StatusInternalServerError, err.Error())
		return
	}

	responses.SendSuccess(w, "post updated successfully", nil)
}

func (h *PostHandler) DeletePost(w http.ResponseWriter, r *http.Request) {
	postIDStr := r.PathValue("id")
	postIDInt, err := strconv.Atoi(postIDStr)

	userID, _ := middleware.UserIDFromContext(r.Context())

	if err != nil {
		log.Printf("[ERROR] DeletePostHandler: Invalid PostID format: %v", err)
		responses.SendError(w, http.StatusBadRequest, "Invalid post ID")
		return
	}

	err = h.Service.DeletePost(r.Context(), postIDInt, userID)
	if err != nil {
		log.Printf("[ERROR] DeletePostHandler: Service call failed: %v", err)
		responses.SendError(w, http.StatusInternalServerError, err.Error())
		return
	}

	responses.SendSuccess(w, "post deleted successfully", nil)
}

func (h *PostHandler) RestorePost(w http.ResponseWriter, r *http.Request) {
	postIDStr := r.PathValue("id")
	postIDInt, err := strconv.Atoi(postIDStr)

	userID, _ := middleware.UserIDFromContext(r.Context())

	if err != nil {
		log.Printf("[ERROR] RestorePostHandler: Invalid PostID format: %v", err)
		responses.SendError(w, http.StatusBadRequest, "Invalid post ID")
		return
	}

	err = h.Service.RestorePost(r.Context(), postIDInt, userID)
	if err != nil {
		log.Printf("[ERROR] RestorePostHandler: Service call failed: %v", err)
		responses.SendError(w, http.StatusInternalServerError, err.Error())
		return
	}

	responses.SendSuccess(w, "post restored successfully", nil)
}

func (h *PostHandler) GetPostHandler(w http.ResponseWriter, r *http.Request) {
	postIDStr := r.PathValue("id")
	postIDInt, err := strconv.Atoi(postIDStr)

	if err != nil {
		log.Printf("[ERROR] GetPostHandler: Invalid PostID format: %v", err)
		responses.SendError(w, http.StatusBadRequest, "Invalid post ID")
		return
	}

	userID, _ := middleware.UserIDFromContext(r.Context())

	post, err := h.Service.GetPost(r.Context(), postIDInt, userID)
	if err != nil {
		log.Printf("[ERROR] GetPostHandler: Service failed to retrieve post %d: %v", postIDInt, err)
		responses.SendError(w, http.StatusInternalServerError, err.Error())
		return
	}

	responses.SendSuccess(w, "post retrieved successfully", post)
}
