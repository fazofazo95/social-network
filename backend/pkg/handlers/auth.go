package handlers

import (
	"backend/pkg/db/queries"
	"backend/pkg/middleware"
	"backend/pkg/models"
	"backend/pkg/responses"
	"backend/pkg/services"
	"backend/pkg/utils"
	"encoding/json"
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
	if err := r.ParseMultipartForm(20 << 20); err != nil {
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
		responses.SendError(w, http.StatusInternalServerError, err.Error())
		return
	}

	if imageURL != "" {
		signUpInput.Avatar = imageURL
	}

	if err := h.Service.SignUp(r.Context(), signUpInput); err != nil {
		switch err {
		case services.ErrEmailTaken:
			responses.SendError(w, http.StatusConflict, "email already in use")
			return
		case services.ErrUsernameTaken:
			responses.SendError(w, http.StatusConflict, "username already in use")
			return
		default:
			responses.SendError(w, http.StatusInternalServerError, err.Error())
			return
		}
	}

	responses.SendCreated(w, "user created successfully", nil)
}

func (h *AuthHandler) LogInHandler(w http.ResponseWriter, r *http.Request) {
	var req models.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		responses.SendError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	loginResp, err := h.Service.Login(r.Context(), req)
	if err != nil {
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

func (h *AuthHandler) LogOutHandler(w http.ResponseWriter, r *http.Request) {
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

	if err := h.Service.Logout(r.Context(), c.Value, userID); err != nil {
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

func (h *AuthHandler) VerifySession(w http.ResponseWriter, r *http.Request) {
	c, err := r.Cookie("session_id")
	if err != nil {
		responses.SendError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	userID, err := queries.AuthenticateSession(r.Context(), c.Value)
	if err != nil {
		responses.SendError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	responses.SendSuccess(w, "session exists", userID)
}
