package handlers

import (
	"database/sql"
	"log"
	"net/http"
	"strconv"

	"backend/pkg/db/queries"
	database "backend/pkg/db/sqlite"
	"backend/pkg/middleware"
	"backend/pkg/models"
	"backend/pkg/responses"
	"backend/pkg/services"
	"backend/pkg/utils"
)

func CreateUserHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("[INFO] CreateUserHandler: Received request")

	if err := r.ParseMultipartForm(20 << 20); err != nil {
		log.Printf("[ERROR] CreateUserHandler: ParseMultipartForm failed: %v", err)
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
		log.Printf("[ERROR] CreateUserHandler: Avatar attachment failed: %v", err)
		responses.SendError(w, http.StatusInternalServerError, err.Error())
		return
	}

	if imageURL != "" {
		log.Printf("[INFO] CreateUserHandler: Avatar uploaded: %s", imageURL)
		signUpInput.Avatar = imageURL
	}

	authService := services.NewAuthService(database.DB)

	log.Printf("[INFO] CreateUserHandler: Attempting SignUp for Email: %s, Username: %s", signUpInput.Email, signUpInput.Username)
	if err := authService.SignUp(r.Context(), signUpInput); err != nil {
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

func GetUserHandler(w http.ResponseWriter, r *http.Request) {
	viewerID, err := middleware.UserIDFromContext(r.Context())
	if err != nil {
		responses.SendError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	targetID, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || targetID <= 0 {
		responses.SendError(w, http.StatusBadRequest, "invalid user id")
		return
	}

	data, err := queries.GetUserProfileView(r.Context(), database.DB, viewerID, targetID)
	if err != nil {
		if err == sql.ErrNoRows {
			responses.SendError(w, http.StatusNotFound, "user not found")
			return
		}
		responses.SendError(w, http.StatusInternalServerError, "failed to fetch profile: "+err.Error())
		return
	}

	responses.SendSuccess(w, "profile", data)
}

func UpdateUserHandler(w http.ResponseWriter, r *http.Request) {}

func GetUserPostsHandler(w http.ResponseWriter, r *http.Request) {
	targetUserID := r.PathValue("id")
	log.Printf("[INFO] GetUserPostsHandler: Fetching posts for TargetUserID: %s", targetUserID)

	targetUserIDint, err := strconv.Atoi(targetUserID)
	if err != nil {
		log.Printf("[ERROR] GetUserPostsHandler: Invalid ID format: %v", err)
		responses.SendError(w, http.StatusBadRequest, "Invalid post ID")
		return
	}

	viewerID, _ := middleware.UserIDFromContext(r.Context())
	log.Printf("[INFO] GetUserPostsHandler: ViewerID: %d fetching posts of TargetID: %d", viewerID, targetUserIDint)

	postService := services.NewPostService(database.DB)
	posts, err := postService.GetUserPosts(r.Context(), targetUserIDint, viewerID)
	if err != nil {
		log.Printf("[ERROR] GetUserPostsHandler: Service call failed: %v", err)
		responses.SendError(w, http.StatusInternalServerError, err.Error())
		return
	}

	log.Printf("[SUCCESS] GetUserPostsHandler: Retrieved %d posts", len(posts))
	responses.SendSuccess(w, "user posts retrieved successfully", posts)
}
