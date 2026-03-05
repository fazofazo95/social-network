package utils

import (
	"errors"
	"fmt"
	"io"
	"log"
	"mime"
	"net/http"
	"strings"

	"github.com/gofrs/uuid/v5"
)

const MaxFileSize = 20 << 20 // 20 MiB

func AttachAvatar(r *http.Request) (string, error) {
	return attachImageFromField(r, "avatar", "AttachAvatar")
}

func AttachCover(r *http.Request) (string, error) {
	return attachImageFromField(r, "cover", "AttachCover")
}

func AttachGroupImage(r *http.Request) (string, error) {
	return attachImageFromField(r, "group_picture", "AttachGroupImage")
}

func attachImageFromField(r *http.Request, fieldName, logPrefix string) (string, error) {
	file, header, err := r.FormFile(fieldName)
	if err != nil {
		if err == http.ErrMissingFile {
			return "", nil
		}
		log.Printf("[ERROR] %s: Error reading form file: %v", logPrefix, err)
		return "", errors.New("Failed to read image from form")
	}
	defer file.Close()

	if header.Size > MaxFileSize {
		return "", errors.New("uploaded file too large")
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
		return "", fmt.Errorf("invalid content-type %s", contentType)
	}

	if _, err := file.Seek(0, io.SeekStart); err != nil {
		log.Printf("[ERROR] %s: Failed to seek file: %v", logPrefix, err)
		return "", errors.New("Failed to process uploaded image")
	}

	extList, _ := mime.ExtensionsByType(contentType)
	ext := ""
	if len(extList) > 0 {
		ext = extList[0]
	} else {
		parts := strings.Split(header.Filename, ".")
		if len(parts) > 1 {
			ext = "." + parts[len(parts)-1]
		}
	}

	newUUID, _ := uuid.NewV4()
	filename := newUUID.String() + ext

	uploadDir := "uploads"

	if err := SaveFile(file, filename, uploadDir); err != nil {
		log.Printf("[ERROR] %s: SaveFile failed: %v", logPrefix, err)
		return "", errors.New("Failed to save image file")
	}

	filename = "/uploads/" + filename
	return filename, nil
}
