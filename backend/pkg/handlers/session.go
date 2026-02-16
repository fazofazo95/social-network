package handlers

import (
	"backend/pkg/db/queries"
	database "backend/pkg/db/sqlite"
	"backend/pkg/responses"
	"net/http"
)

func VerifySession(w http.ResponseWriter, r *http.Request) {
	c, err := r.Cookie("session_id")
	if err != nil {
		responses.SendError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	userID, err := queries.AuthenticateSession(r.Context(), database.DB, c.Value)
	if err != nil {
		responses.SendError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	responses.SendSuccess(w, "session exists", userID)
}