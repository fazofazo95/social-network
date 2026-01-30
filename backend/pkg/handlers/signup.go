package handlers

import (
	"errors"
	"fmt"
	"io"
	"mime"
	"net/http"
	"strings"

	database "backend/pkg/db/sqlite"
	"backend/pkg/models"
	"backend/pkg/responses"
	"backend/pkg/services"
	"backend/pkg/utils"

	"github.com/gofrs/uuid/v5"
)

// SignupHandler handles POST /signup requests for creating a new user.
func SignUpHandler(w http.ResponseWriter, r *http.Request) {
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

	if err := r.ParseMultipartForm(20 << 20); err != nil {
		responses.SendError(w, http.StatusBadRequest, "Invalid Form")
		return
	}

	signUpInput := models.Signup_fields{
		Email: r.FormValue("email"),
		Password: r.FormValue("password"),
		FirstName: r.FormValue("firstname"),
		LastName: r.FormValue("lastname"),
		Username: r.FormValue("username"),
		Birthday: r.FormValue("date_of_birth"),
		Nickname: r.FormValue("nickname"),
		AboutMe: r.FormValue("about_me"),
	}

	if err := AttachAvatar(r,&signUpInput); err != nil {
		responses.SendError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Initialize auth service
	authService := services.NewAuthService(database.DB)

	// Call service layer to handle signup business logic
	if err := authService.SignUp(r.Context(), signUpInput); err != nil {
		// Map service errors to HTTP responses
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

const MaxFileSize = 20 << 20 // 20 MiB

func AttachAvatar(r *http.Request, form *models.Signup_fields) error{
	file, header, err := r.FormFile("avatar")
	if err != nil {
		if err == http.ErrMissingFile {
			return nil
		}
		return errors.New("Failed to read image from form")
	}
	defer file.Close()

	if header.Size > MaxFileSize {
		return errors.New("uploaded file too large")
	}

	buf := make([]byte, 512)
	n, _ := file.Read(buf)
	contentType := http.DetectContentType(buf[:n])

	allowed := map[string]bool{
		"image/jpeg": true,
		"image/png":  true,
		"image/gif":  true,
	}
	
	if !allowed[contentType] {
		return fmt.Errorf("invalid content-type %s", contentType)
	}

	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return errors.New("Failed to process uploaded image")
	}

	extList, _ := mime.ExtensionsByType(contentType)
	ext := ""
	if len(extList) > 0 {
		ext = extList[0]
	} else {
		// fallback to header-based ext
		parts := strings.Split(header.Filename, ".")
		if len(parts) > 1 {
			ext = "." + parts[len(parts)-1]
		}
	}

	newUUID, _ := uuid.NewV4()
	filename := newUUID.String() + ext

	uploadDir := "uploads"
	// 4) Save to disk
	if err := utils.SaveFile(file, filename, uploadDir); err != nil {
		return errors.New("Failed to save image file")
	}


	form.Avatar = "/uploads/" + filename
	return nil
}



