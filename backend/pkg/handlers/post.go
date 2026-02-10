package handlers

import (
	"errors"
	"fmt"
	"io"
	"mime"
	"net/http"
	"strconv"
	"strings"

	database "backend/pkg/db/sqlite"
	"backend/pkg/models"
	"backend/pkg/responses"
	"backend/pkg/services"
	"backend/pkg/utils"

	"github.com/gofrs/uuid/v5"
)

func CreatePost(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
	w.Header().Set("Access-Control-Allow-Credentials", "true")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")

	if r.Method != http.MethodPost {
		responses.SendError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	if err := r.ParseMultipartForm(20 << 20); err != nil {
		responses.SendError(w, http.StatusBadRequest, "Invalid Form")
		return
	}

	userID, err := strconv.Atoi(r.FormValue("user_id"))
	if err != nil {
		responses.SendError(w, http.StatusBadRequest, "Invalid Form")
		return
	}

	privacy := r.FormValue("privacy")

	var wlIDs []int

	if privacy == "custom" {
		wl := r.MultipartForm.Value["whitelisted_users"]

		for _, idStr := range wl {
			id, err := strconv.Atoi(idStr)
			if err != nil {
				responses.SendError(w, http.StatusInternalServerError, err.Error())
				return
			}
			wlIDs = append(wlIDs, id)
		}
	}

	post := models.Post{
		UserID:           userID,
		Content:          r.FormValue("content"),
		Privacy:          privacy,
		WhitelistedUsers: wlIDs,
	}

	if err := AttachImage(r, &post); err != nil {
		responses.SendError(w, http.StatusInternalServerError, err.Error())
		return
	}

	postService := services.NewPostService(database.DB)

	if err := postService.CreatePost(r.Context(), post); err != nil {
		responses.SendError(w, http.StatusInternalServerError, err.Error())
		return
	}

	responses.SendCreated(w, "user created successfully", nil)
}

func AttachImage(r *http.Request, form *models.Post) error {
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

	form.Image = "/uploads/" + filename
	return nil
}
