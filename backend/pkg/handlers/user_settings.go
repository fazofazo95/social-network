package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"backend/pkg/db/queries"
	database "backend/pkg/db/sqlite"
	"backend/pkg/middleware"
	"backend/pkg/responses"
)

// GetVisibilitySettingsHandler returns the current user's visibility settings
// mapping numeric flags to human-friendly strings (visible/hidden) and
// profile_type to public/private.
func GetVisibilitySettingsHandler(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.UserIDFromContext(r.Context())
	if err != nil {
		responses.SendError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	settings, err := queries.GetUserVisibilitySettings(r.Context(), database.DB, userID)
	if err != nil {
		responses.SendError(w, http.StatusInternalServerError, "failed to fetch visibility settings: "+err.Error())
		return
	}

	responses.SendSuccess(w, "visibility settings", settings)
}

// UpdateVisibilitySettingsHandler updates the authenticated user's visibility settings.
// Accepts a JSON body with any of: email_vis, birthday_date_vis, relationship_status_vis,
// employed_at_vis, phone_number_vis (values: "visible"/"hidden" or boolean), and profile_type ("public"/"private").
func UpdateVisibilitySettingsHandler(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.UserIDFromContext(r.Context())
	if err != nil {
		responses.SendError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	// incoming payload
	var in map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		responses.SendError(w, http.StatusBadRequest, "invalid JSON")
		return
	}

	if len(in) == 0 {
		responses.SendError(w, http.StatusBadRequest, "no fields to update")
		return
	}

	// helper: parse visibility value (string "visible"/"hidden" or boolean) -> *int
	parseVis := func(v interface{}) (*int, error) {
		if v == nil {
			return nil, nil
		}
		// boolean
		if b, ok := v.(bool); ok {
			val := 0
			if b {
				val = 1
			}
			return &val, nil
		}
		// string
		if s, ok := v.(string); ok {
			s = strings.ToLower(strings.TrimSpace(s))
			if s == "visible" || s == "true" {
				val := 1
				return &val, nil
			}
			if s == "hidden" || s == "false" {
				val := 0
				return &val, nil
			}
			return nil, http.ErrBodyNotAllowed
		}
		return nil, http.ErrBodyNotAllowed
	}

	// parse profile_type
	var profileTypePtr *int
	if v, ok := in["profile_type"]; ok {
		if v == nil {
			profileTypePtr = nil
		} else if s, ok := v.(string); ok {
			s = strings.ToLower(strings.TrimSpace(s))
			if s == "public" {
				val := 0
				profileTypePtr = &val
			} else if s == "private" {
				val := 1
				profileTypePtr = &val
			} else {
				responses.SendError(w, http.StatusBadRequest, "invalid profile_type value")
				return
			}
		} else {
			responses.SendError(w, http.StatusBadRequest, "invalid profile_type value")
			return
		}
	}

	// parse each visibility field
	emailVis, err := parseVis(in["email_vis"])
	if err != nil {
		responses.SendError(w, http.StatusBadRequest, "invalid email_vis value")
		return
	}
	birthdayVis, err := parseVis(in["birthday_date_vis"])
	if err != nil {
		responses.SendError(w, http.StatusBadRequest, "invalid birthday_date_vis value")
		return
	}
	relVis, err := parseVis(in["relationship_status_vis"])
	if err != nil {
		responses.SendError(w, http.StatusBadRequest, "invalid relationship_status_vis value")
		return
	}
	employedVis, err := parseVis(in["employed_at_vis"])
	if err != nil {
		responses.SendError(w, http.StatusBadRequest, "invalid employed_at_vis value")
		return
	}
	phoneVis, err := parseVis(in["phone_number_vis"])
	if err != nil {
		responses.SendError(w, http.StatusBadRequest, "invalid phone_number_vis value")
		return
	}
	aboutVis, err := parseVis(in["about_me_vis"])
	if err != nil {
		responses.SendError(w, http.StatusBadRequest, "invalid about_me_vis value")
		return
	}
	nickVis, err := parseVis(in["nickname_vis"])
	if err != nil {
		responses.SendError(w, http.StatusBadRequest, "invalid nickname_vis value")
		return
	}
	followVis, err := parseVis(in["follow_vis"])
	if err != nil {
		responses.SendError(w, http.StatusBadRequest, "invalid follow_vis value")
		return
	}

	// Call DB updater
	if err := queries.UpdateUserVisibilitySettings(r.Context(), database.DB, userID, emailVis, birthdayVis, relVis, employedVis, phoneVis, aboutVis, nickVis, followVis, profileTypePtr); err != nil {
		responses.SendError(w, http.StatusInternalServerError, "failed to update settings: "+err.Error())
		return
	}

	// Return the updated settings (read again)
	settings, err := queries.GetUserVisibilitySettings(r.Context(), database.DB, userID)
	if err != nil {
		responses.SendError(w, http.StatusInternalServerError, "failed to fetch settings after update: "+err.Error())
		return
	}

	responses.SendSuccess(w, "settings updated", settings)
}
