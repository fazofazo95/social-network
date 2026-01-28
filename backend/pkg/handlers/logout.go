package handlers

import (
	database "backend/pkg/db/sqlite"
	"backend/pkg/middleware"
	"backend/pkg/responses"
	"backend/pkg/services"
	"net/http"
)

func LogOutHandler(w http.ResponseWriter, r *http.Request) {

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

	// Initialize auth service
	authService := services.NewAuthService(database.DB)

	// Call service layer to handle logout business logic
	if err := authService.Logout(r.Context(), c.Value, userID); err != nil {
		responses.SendError(w, http.StatusInternalServerError, "failed to logout")
		return
	}

	// Clear session cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
	})

	responses.SendSuccess(w, "logout successful", nil)
}