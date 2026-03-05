package handlers

import (
	"encoding/json"
	"net/http"

	"backend/pkg/middleware"
	"backend/pkg/models"
	"backend/pkg/responses"
	"backend/pkg/services"
)

type SettingsHandler struct {
	profileService services.ProfileService
}

func NewSettingsHandler(ps services.ProfileService) *SettingsHandler{
	return &SettingsHandler{profileService: ps}
}

func (h *SettingsHandler) RegisterRoutes(mux *http.ServeMux){
	auth := middleware.WithAuth
	
	mux.Handle("GET /api/users/settings", middleware.Chain(h.GetVisibilitySettingsHandler, auth))
	mux.Handle("PATCH /api/users/settings", middleware.Chain(h.UpdateVisibilitySettingsHandler, auth))
	mux.Handle("PUT /api/users/settings", middleware.Chain(h.UpdateVisibilitySettingsHandler, auth))
	mux.Handle("GET /api/users/settings/content", middleware.Chain(h.GetUserContentSettingsHandler, auth))
	mux.Handle("PATCH /api/users/settings/content", middleware.Chain(h.UpdateUserContentSettingsHandler, auth))
	mux.Handle("PUT /api/users/settings/content", middleware.Chain(h.UpdateUserContentSettingsHandler, auth))
}

func (h *SettingsHandler) GetVisibilitySettingsHandler(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.UserIDFromContext(r.Context())
	if err != nil {
		responses.SendError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	settings, err := h.profileService.GetUserVisibilitySettings(r.Context(),userID)
	if err != nil {
		responses.SendError(w, http.StatusInternalServerError, "failed to fetch visibility settings: "+err.Error())
		return
	}

	responses.SendSuccess(w, "visibility settings", settings)
}

func (h *SettingsHandler) UpdateVisibilitySettingsHandler(w http.ResponseWriter, r *http.Request) {
    userID, err := middleware.UserIDFromContext(r.Context())
	if err != nil {
		responses.SendError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

    var req models.UpdateVisibilityRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        responses.SendError(w, http.StatusBadRequest, "invalid JSON")
        return
    }

    updatedSettings, err := h.profileService.UpdateVisibility(r.Context(), userID, req)
    if err != nil {
        responses.SendError(w, http.StatusInternalServerError, err.Error())
        return
    }

    responses.SendSuccess(w, "settings updated", updatedSettings)
}

func (h *SettingsHandler) GetUserContentSettingsHandler(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.UserIDFromContext(r.Context())
	if err != nil {
		responses.SendError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	data, err := h.profileService.GetUserContentSettings(r.Context(),userID)
	if err != nil {
		responses.SendError(w, http.StatusInternalServerError, "failed to fetch content settings: "+err.Error())
		return
	}

	responses.SendSuccess(w, "content settings", data)
}

func (h *SettingsHandler) UpdateUserContentSettingsHandler(w http.ResponseWriter, r *http.Request) {
    userID, err := middleware.UserIDFromContext(r.Context())

    var req models.UserProfileRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        responses.SendError(w, http.StatusBadRequest, "invalid JSON")
        return
    }

    updated, err := h.profileService.UpdateUserContent(r.Context(), userID, req)
    if err != nil {
        responses.SendError(w, http.StatusInternalServerError, "Update failed: "+err.Error())
        return
    }

    responses.SendSuccess(w, "Profile updated successfully", updated)
}