package handlers

import (
	queries "backend/pkg/db/queries"
	database "backend/pkg/db/sqlite"
	"backend/pkg/middleware"
	"backend/pkg/responses"
	"context"
	"net/http"
)

func LogOutHandler(w http.ResponseWriter, r *http.Request){

	userID, err := middleware.UserIDFromContext(r.Context())
	if err != nil {
		responses.SendError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	c, err := r.Cookie("session_id") 
	if err != nil {
		responses.SendError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	if err := queries.LogOut(context.Background(), database.DB, c.Value, userID); err != nil {
		responses.SendError(w, http.StatusInternalServerError, "failed to logout")
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
	})

	responses.SendSuccess(w, "logout successful", nil)
}