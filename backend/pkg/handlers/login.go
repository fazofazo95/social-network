package handlers

import (
	queries "backend/pkg/db/queries"
	database "backend/pkg/db/sqlite"
	"backend/pkg/responses"
	"context"
	"encoding/json"
	"net/http"
)

func LogInHandler(w http.ResponseWriter, r *http.Request){

	type LogInRequest struct {
		Username string
		Email    string
		Password string
	}

	var req LogInRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		responses.SendError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Email == "" && req.Username == "" {
		responses.SendError(w, http.StatusBadRequest, "email or username is required")
		return
	}

	input := queries.LogInInput{
		Email:    req.Email,
		Username: req.Username,
		Password: req.Password,
	}


	userID, err := queries.LogIn(context.Background(), database.DB, input)
	if err != nil {
		switch err {
		case queries.ErrInvalidUsernameOrEmail:
			responses.SendError(w, http.StatusUnauthorized, "wrong username or email")
			return
		case queries.ErrInvalidPassword:
			responses.SendError(w, http.StatusUnauthorized, "invalid password")
			return
		default:
			responses.SendError(w, http.StatusInternalServerError, "internal server error")
			return
		}
	}

	// Generate session
	sessionID, err := queries.GenerateSession(context.Background(), database.DB, userID)
	if err != nil {
		responses.SendError(w, http.StatusInternalServerError, "failed to create session")
		return
	}

	// Set session cookie
	cookie := &http.Cookie{
		Name:     "session_id",
		Value:    sessionID,
		Path:     "/",
		HttpOnly: true,
		Secure:   false, // Set to true in production with HTTPS
	}
	http.SetCookie(w, cookie)

	responses.SendSuccess(w, "login successful", nil)
}