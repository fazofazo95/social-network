package handlers

import (
	database "backend/pkg/db/sqlite"
	"backend/pkg/models"
	"backend/pkg/responses"
	"backend/pkg/services"
	"encoding/json"
	"net/http"
)

func LogInHandler(w http.ResponseWriter, r *http.Request) {
	var req models.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		responses.SendError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Initialize auth service
	authService := services.NewAuthService(database.DB)

	// Call service layer to handle login business logic
	loginResp, err := authService.Login(r.Context(), req)
	if err != nil {
		// Map service errors to HTTP responses
		switch err {
		case services.ErrInvalidCredentials:
			responses.SendError(w, http.StatusUnauthorized, "invalid username, email, or password")
			return
		case services.ErrSessionFailed:
			responses.SendError(w, http.StatusInternalServerError, "failed to create session")
			return
		default:
			responses.SendError(w, http.StatusInternalServerError, "internal server error")
			return
		}
	}

	// Set session cookie
	cookie := &http.Cookie{
		Name:     "session_id",
		Value:    loginResp.SessionID,
		Path:     "/",
		HttpOnly: true,
		Secure:   false, // Set to true in production with HTTPS
	}
	http.SetCookie(w, cookie)

	responses.SendSuccess(w, "login successful", nil)
}
