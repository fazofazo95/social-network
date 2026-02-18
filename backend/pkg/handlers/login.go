package handlers

import (
	database "backend/pkg/db/sqlite"
	"backend/pkg/models"
	"backend/pkg/responses"
	"backend/pkg/services"
	"encoding/json"
	"log"
	"net/http"
)

func LogInHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("[INFO] LogInHandler: Received request")

	var req models.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("[ERROR] LogInHandler: Decode failed: %v", err)
		responses.SendError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Initialize auth service
	authService := services.NewAuthService(database.DB)

	// Call service layer to handle login business logic
	log.Println("[INFO] LogInHandler: Attempting login")
	loginResp, err := authService.Login(r.Context(), req)
	if err != nil {
		// Map service errors to HTTP responses
		switch err {
		case services.ErrInvalidCredentials:
			log.Println("[WARN] LogInHandler: Invalid credentials")
			responses.SendError(w, http.StatusUnauthorized, "invalid username, email, or password")
			return
		case services.ErrSessionFailed:
			log.Println("[ERROR] LogInHandler: Session creation failed")
			responses.SendError(w, http.StatusInternalServerError, "failed to create session")
			return
		default:
			log.Printf("[ERROR] LogInHandler: Internal error: %v", err)
			responses.SendError(w, http.StatusInternalServerError, "internal server error")
			return
		}
	}

	// Set session cookie
	log.Printf("[INFO] LogInHandler: Setting cookie for UserID: %d", loginResp.UserID)
	cookie := &http.Cookie{
		Name:     "session_id",
		Value:    loginResp.SessionID,
		Path:     "/",
		HttpOnly: true,
		Secure:   false, // Set to true in production with HTTPS
	}
	http.SetCookie(w, cookie)

	log.Printf("[SUCCESS] LogInHandler: Login successful for UserID: %d", loginResp.UserID)
	responses.SendSuccess(w, "login successful", nil)
}
