package handlers

import (
	"backend/pkg/middleware"
	"backend/pkg/models"
	"backend/pkg/responses"
	"backend/pkg/services"
	"backend/pkg/utils"
	"encoding/json"
	"log"
	"net/http"
)

type AuthHandler struct {
	Service services.AuthService
}

func NewAuthHandler(s services.AuthService) *AuthHandler {
	return &AuthHandler{Service: s}
}

func (h *AuthHandler) RegisterRoutes(mux *http.ServeMux) {
	auth := middleware.WithAuth
	
	mux.HandleFunc("POST /api/users", h.CreateUserHandler)
	mux.HandleFunc("POST /api/login", h.LogInHandler)
	mux.Handle("DELETE /api/logout", middleware.Chain(h.LogOutHandler, auth))
	mux.HandleFunc("GET /api/verify-session", h.VerifySession)
}

func (h *AuthHandler) CreateUserHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("[INFO] CreateUserHandler: Received request")

	if err := r.ParseMultipartForm(20 << 20); err != nil {
		log.Printf("[ERROR] CreateUserHandler: ParseMultipartForm failed: %v", err)
		responses.SendError(w, http.StatusBadRequest, "Invalid Form")
		return
	}

	signUpInput := models.SignupFields{
		Email:     r.FormValue("email"),
		Password:  r.FormValue("password"),
		FirstName: r.FormValue("firstname"),
		LastName:  r.FormValue("lastname"),
		Username:  r.FormValue("username"),
		Birthday:  r.FormValue("date_of_birth"),
		Nickname:  r.FormValue("nickname"),
		AboutMe:   r.FormValue("about_me"),
	}

	imageURL, err := utils.AttachAvatar(r)
	if err != nil {
		log.Printf("[ERROR] CreateUserHandler: Avatar attachment failed: %v", err)
		responses.SendError(w, http.StatusInternalServerError, err.Error())
		return
	}

	if imageURL != "" {
		log.Printf("[INFO] CreateUserHandler: Avatar uploaded: %s", imageURL)
		signUpInput.Avatar = imageURL
	}

	log.Printf("[INFO] CreateUserHandler: Attempting SignUp for Email: %s, Username: %s", signUpInput.Email, signUpInput.Username)
	if err := h.Service.SignUp(r.Context(), signUpInput); err != nil {
		switch err {
		case services.ErrEmailTaken:
			log.Printf("[WARN] CreateUserHandler: Email taken: %s", signUpInput.Email)
			responses.SendError(w, http.StatusConflict, "email already in use")
			return
		case services.ErrUsernameTaken:
			log.Printf("[WARN] CreateUserHandler: Username taken: %s", signUpInput.Username)
			responses.SendError(w, http.StatusConflict, "username already in use")
			return
		default:
			log.Printf("[ERROR] CreateUserHandler: SignUp service error: %v", err)
			responses.SendError(w, http.StatusInternalServerError, err.Error())
			return
		}
	}

	log.Printf("[SUCCESS] CreateUserHandler: User %s created successfully", signUpInput.Username)
	responses.SendCreated(w, "user created successfully", nil)
}

func (h *AuthHandler) LogInHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("[INFO] LogInHandler: Received request")

	var req models.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("[ERROR] LogInHandler: Decode failed: %v", err)
		responses.SendError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Call service layer to handle login business logic
	log.Println("[INFO] LogInHandler: Attempting login")
	loginResp, err := h.Service.Login(r.Context(), req)
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

func (h *AuthHandler) LogOutHandler(w http.ResponseWriter, r *http.Request) {
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

	// Call service layer to handle logout business logic
	log.Printf("[INFO] LogOutHandler: Attempting logout for UserID: %d", userID)
	if err := h.Service.Logout(r.Context(), c.Value, userID); err != nil {
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

func  (h *AuthHandler) VerifySession(w http.ResponseWriter, r *http.Request) {
	c, err := r.Cookie("session_id")
	if err != nil {
		responses.SendError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	userID, err := h.Service.AuthenticateSession(r.Context(), c.Value)
	if err != nil {
		responses.SendError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	responses.SendSuccess(w, "session exists", userID)
}