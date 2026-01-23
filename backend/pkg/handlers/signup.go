package handlers

import (
	"context"
	"encoding/json"
	"net/http"

	queries "backend/pkg/db/queries"
	database "backend/pkg/db/sqlite"
	"backend/pkg/responses"
)

// SignupRequest represents the expected JSON payload
type SignupRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

// SignupHandler handles POST /signup requests for creating a new user.
func SignupHandler(w http.ResponseWriter, r *http.Request) {
	// Allow CORS for local frontend testing. For production, tighten this.
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	if r.Method != http.MethodPost {
		responses.SendError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req SignupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		responses.SendError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	input := queries.SignUpInput{
		Email:    req.Email,
		Username: req.Username,
		Password: req.Password,
	}

	if err := queries.SignUp(context.Background(), database.DB, input); err != nil {
		switch err {
		case queries.ErrEmailTaken:
			responses.SendError(w, http.StatusConflict, "email already in use")
			return
		case queries.ErrUsernameTaken:
			responses.SendError(w, http.StatusConflict, "username already in use")
			return
		default:
			responses.SendError(w, http.StatusInternalServerError, "internal server error")
			return
		}
	}

	responses.SendCreated(w, "user created successfully", nil)
}
