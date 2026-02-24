package handlers

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"backend/pkg/db/queries"
	database "backend/pkg/db/sqlite"
	"backend/pkg/middleware"
	"backend/pkg/responses"
)

// GetUserContentSettingsHandler returns editable user fields used by settings content manager.
func GetUserContentSettingsHandler(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.UserIDFromContext(r.Context())
	if err != nil {
		responses.SendError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	data, err := queries.GetUserContentSettings(r.Context(), database.DB, userID)
	if err != nil {
		responses.SendError(w, http.StatusInternalServerError, "failed to fetch content settings: "+err.Error())
		return
	}

	responses.SendSuccess(w, "content settings", data)
}

// UpdateUserContentSettingsHandler updates editable user fields used by settings content manager.
// Accepts any subset of:
// first_name, last_name, birthday_date, relationship_status, employed_at, location, phone_number, nickname, about_me.
func UpdateUserContentSettingsHandler(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.UserIDFromContext(r.Context())
	if err != nil {
		responses.SendError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var in map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		responses.SendError(w, http.StatusBadRequest, "invalid JSON")
		return
	}

	if len(in) == 0 {
		responses.SendError(w, http.StatusBadRequest, "no fields to update")
		return
	}

	parseStringField := func(key string) (*string, error) {
		v, ok := in[key]
		if !ok {
			return nil, nil
		}
		s, ok := v.(string)
		if !ok {
			return nil, http.ErrBodyNotAllowed
		}
		trimmed := strings.TrimSpace(s)
		return &trimmed, nil
	}

	firstName, err := parseStringField("first_name")
	if err != nil {
		responses.SendError(w, http.StatusBadRequest, "invalid first_name value")
		return
	}
	lastName, err := parseStringField("last_name")
	if err != nil {
		responses.SendError(w, http.StatusBadRequest, "invalid last_name value")
		return
	}
	birthdayDate, err := parseStringField("birthday_date")
	if err != nil {
		responses.SendError(w, http.StatusBadRequest, "invalid birthday_date value")
		return
	}
	if birthdayDate != nil && *birthdayDate != "" {
		if _, perr := time.Parse("2006-01-02", *birthdayDate); perr != nil {
			responses.SendError(w, http.StatusBadRequest, "birthday_date must be YYYY-MM-DD")
			return
		}
	}
	relationshipStatus, err := parseStringField("relationship_status")
	if err != nil {
		responses.SendError(w, http.StatusBadRequest, "invalid relationship_status value")
		return
	}
	employedAt, err := parseStringField("employed_at")
	if err != nil {
		responses.SendError(w, http.StatusBadRequest, "invalid employed_at value")
		return
	}
	location, err := parseStringField("location")
	if err != nil {
		responses.SendError(w, http.StatusBadRequest, "invalid location value")
		return
	}
	phoneNumber, err := parseStringField("phone_number")
	if err != nil {
		responses.SendError(w, http.StatusBadRequest, "invalid phone_number value")
		return
	}
	nickname, err := parseStringField("nickname")
	if err != nil {
		responses.SendError(w, http.StatusBadRequest, "invalid nickname value")
		return
	}
	aboutMe, err := parseStringField("about_me")
	if err != nil {
		responses.SendError(w, http.StatusBadRequest, "invalid about_me value")
		return
	}

	if err := queries.UpdateUserContentSettings(r.Context(), database.DB, userID,
		firstName, lastName, birthdayDate, relationshipStatus, employedAt, location, phoneNumber, nickname, aboutMe,
	); err != nil {
		if err == sql.ErrNoRows {
			responses.SendError(w, http.StatusNotFound, "user not found")
			return
		}
		responses.SendError(w, http.StatusInternalServerError, "failed to update content settings: "+err.Error())
		return
	}

	updated, err := queries.GetUserContentSettings(r.Context(), database.DB, userID)
	if err != nil {
		log.Printf("[WARN] UpdateUserContentSettingsHandler: updated user %d but failed to fetch content settings: %v", userID, err)
		responses.SendSuccess(w, "content settings updated", in)
		return
	}

	responses.SendSuccess(w, "content settings updated", updated)
}
