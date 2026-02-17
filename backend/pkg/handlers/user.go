package handlers

import (
	"net/http"
	"strconv"

	database "backend/pkg/db/sqlite"
	"backend/pkg/middleware"
	"backend/pkg/models"
	"backend/pkg/responses"
	"backend/pkg/services"
	"backend/pkg/utils"
)

func CreateUserHandler(w http.ResponseWriter, r *http.Request) {

	if err := r.ParseMultipartForm(20 << 20); err != nil {
		responses.SendError(w, http.StatusBadRequest, "Invalid Form")
		return
	}

	signUpInput := models.Signup_fields{
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

	authService := services.NewAuthService(database.DB)

	if err := authService.SignUp(r.Context(), signUpInput); err != nil {
		switch err {
		case services.ErrEmailTaken:
			responses.SendError(w, http.StatusConflict, "email already in use")
			return
		case services.ErrUsernameTaken:
			responses.SendError(w, http.StatusConflict, "username already in use")
			return
		default:
			responses.SendError(w, http.StatusInternalServerError, "internal server error")
			return
		}
	}

	responses.SendCreated(w, "user created successfully", nil)
}

func GetUserHandler(w http.ResponseWriter, r *http.Request) {}

func UpdateUserHandler(w http.ResponseWriter, r *http.Request) {}

func GetUserPostsHandler(w http.ResponseWriter, r *http.Request) {
	targetUserID := r.PathValue("id")
	targetUserIDint, err := strconv.Atoi(targetUserID)
	if err != nil {
		responses.SendError(w, http.StatusBadRequest, "Invalid post ID")
		return
	}

	viewerID, _ := middleware.UserIDFromContext(r.Context())

	postService := services.NewPostService(database.DB)
	posts, err := postService.GetUserPosts(r.Context(), targetUserIDint, viewerID)
	if err != nil {
		responses.SendError(w, http.StatusInternalServerError, err.Error())
		return
	}

	responses.SendSuccess(w, "user posts retrieved successfully", posts)
}
