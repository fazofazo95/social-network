package handlers

import (
	"backend/pkg/db/queries"
	database "backend/pkg/db/sqlite"
	"backend/pkg/responses"
	"net/http"
)

func VerifySession(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
    w.Header().Set("Access-Control-Allow-Credentials", "true")
    w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
    w.Header().Set("Access-Control-Allow-Methods", "GET")

	c, err := r.Cookie("session_id")
	if err != nil {
		responses.SendError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	userID, err := queries.SessionExists(r.Context(), database.DB, c.Value)
	if err != nil {
		responses.SendError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	responses.SendSuccess(w, "session exists", userID)
}