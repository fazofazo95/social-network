package handlers

import (
	"backend/pkg/middleware"
	"backend/pkg/responses"
	"backend/pkg/services"
	"backend/pkg/utils"
	"net/http"
	"strconv"
)

type ProfileHandler struct {
	profileService services.ProfileService
	postService    services.PostService
}

func NewProfileHandler(profileServ services.ProfileService,
	postServ services.PostService) *ProfileHandler {
	return &ProfileHandler{profileService: profileServ, postService: postServ}
}

func (h *ProfileHandler) RegisterRoutes(mux *http.ServeMux) {
	auth := middleware.WithAuth

	mux.Handle("GET /api/users/{id}", middleware.Chain(h.GetUserHandler, auth))
	mux.Handle("PUT /api/users/{id}", middleware.Chain(h.UpdateUserHandler, auth))

	mux.Handle("GET /api/users/{id}/posts", middleware.Chain(h.GetUserPostsHandler, auth))
}

func (h *ProfileHandler) GetUserHandler(w http.ResponseWriter, r *http.Request) {
	viewerID, err := middleware.UserIDFromContext(r.Context())
	if err != nil {
		responses.SendError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	targetID := r.PathValue("id")

	data, err := h.profileService.GetUserProfileView(r.Context(), viewerID, targetID)
	if err != nil {
		responses.SendError(w, http.StatusInternalServerError, "failed to fetch profile")
		return
	}

	responses.SendSuccess(w, "profile", data)
}

func (h *ProfileHandler) UpdateUserHandler(w http.ResponseWriter, r *http.Request) {
	viewerID, err := middleware.UserIDFromContext(r.Context())
	if err != nil {
		responses.SendError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	targetIDstr := r.PathValue("id")

	if targetIDstr != "me" {
		responses.SendError(w, http.StatusForbidden, "forbidden")
		return
	}

	targetID := viewerID

	if err := r.ParseMultipartForm(20 << 20); err != nil {
		responses.SendError(w, http.StatusBadRequest, "invalid form")
		return
	}

	imageURL, err := utils.AttachAvatar(r)
	if err != nil {
		responses.SendError(w, http.StatusBadRequest, err.Error())
		return
	}

	coverImageURL, err := utils.AttachCover(r)
	if err != nil {
		responses.SendError(w, http.StatusBadRequest, err.Error())
		return
	}

	err = h.profileService.UpdateUserMedia(r.Context(), targetID, imageURL, coverImageURL)
	if err != nil {
		if err.Error() == "no avatar or cover file provided" {
			responses.SendError(w, http.StatusBadRequest, err.Error())
			return
		}
		responses.SendError(w, http.StatusInternalServerError, "failed to update profile media")
		return
	}

	responses.SendSuccess(w, "profile updated", map[string]interface{}{
		"id":              targetID,
		"profile_picture": imageURL,
		"cover_image":     coverImageURL,
	})
}

func (h *ProfileHandler) GetUserPostsHandler(w http.ResponseWriter, r *http.Request) {
	targetUserID := r.PathValue("id")

	targetUserIDint, err := strconv.Atoi(targetUserID)
	if err != nil {
		responses.SendError(w, http.StatusBadRequest, "Invalid post ID")
		return
	}

	viewerID, _ := middleware.UserIDFromContext(r.Context())

	pageStr := r.URL.Query().Get("page")
	page, _ := strconv.Atoi(pageStr)

	posts, err := h.postService.GetProfilePosts(r.Context(), targetUserIDint, viewerID, page)
	if err != nil {
		responses.SendError(w, http.StatusInternalServerError, err.Error())
		return
	}

	responses.SendSuccess(w, "user posts retrieved successfully", posts)
}
