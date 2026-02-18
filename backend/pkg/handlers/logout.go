package handlers

import (
	database "backend/pkg/db/sqlite"
	"backend/pkg/middleware"
	"backend/pkg/responses"
	"backend/pkg/services"
	"log"
	"net/http"
)

func LogOutHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("[INFO] LogOutHandler: Received request")

	userID, err := middleware.UserIDFromContext(r.Context())
	if err != nil {
		log.Printf("[ERROR] LogOutHandler: UserID context error: %v", err)
		responses.SendError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	c, err := r.Cookie("session_id")
	if err != nil {
		log.Printf("[WARN] LogOutHandler: Cookie not found for UserID: %d", userID)
		responses.SendError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	// Initialize auth service
	authService := services.NewAuthService(database.DB)

	// Call service layer to handle logout business logic
	log.Printf("[INFO] LogOutHandler: Attempting logout for UserID: %d", userID)
	if err := authService.Logout(r.Context(), c.Value, userID); err != nil {
		log.Printf("[ERROR] LogOutHandler: Service logout failed for UserID %d: %v", userID, err)
		responses.SendError(w, http.StatusInternalServerError, "failed to logout")
		return
	}

	// Clear session cookie
	log.Printf("[INFO] LogOutHandler: Clearing session cookie for UserID: %d", userID)
	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
	})

	log.Printf("[SUCCESS] LogOutHandler: Logout successful for UserID: %d", userID)
	responses.SendSuccess(w, "logout successful", nil)
}
