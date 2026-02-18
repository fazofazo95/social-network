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
	log.Println("[INFO] AttachAvatar: Processing avatar upload")

	file, header, err := r.FormFile("avatar")
	if err != nil {
		if err == http.ErrMissingFile {
			log.Println("[INFO] AttachAvatar: No avatar file found in request")
			return "", nil
		}
		log.Printf("[ERROR] AttachAvatar: Error reading form file: %v", err)
		return "", errors.New("Failed to read image from form")
	}
	defer file.Close()

	log.Printf("[INFO] AttachAvatar: Received file %s (Size: %d bytes)", header.Filename, header.Size)

	if header.Size > MaxFileSize {
		log.Printf("[WARN] AttachAvatar: File size %d exceeds limit", header.Size)
		return "", errors.New("uploaded file too large")
	}

	buf := make([]byte, 512)
	n, _ := file.Read(buf)
	contentType := http.DetectContentType(buf[:n])
	log.Printf("[INFO] AttachAvatar: Detected Content-Type: %s", contentType)

	allowed := map[string]bool{
		"image/jpeg": true,
		"image/png":  true,
		"image/gif":  true,
	}

	if !allowed[contentType] {
		log.Printf("[WARN] AttachAvatar: Rejected invalid Content-Type: %s", contentType)
		return "", fmt.Errorf("invalid content-type %s", contentType)
	}

	if _, err := file.Seek(0, io.SeekStart); err != nil {
		log.Printf("[ERROR] AttachAvatar: Failed to seek file: %v", err)
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
	log.Printf("[INFO] AttachAvatar: Using extension: %s", ext)

	newUUID, _ := uuid.NewV4()
	filename := newUUID.String() + ext

	uploadDir := "uploads"

	log.Printf("[INFO] AttachAvatar: Saving file as %s in %s", filename, uploadDir)
	if err := SaveFile(file, filename, uploadDir); err != nil {
		log.Printf("[ERROR] AttachAvatar: SaveFile failed: %v", err)
		return "", errors.New("Failed to save image file")
	}

	filename = "/uploads/" + filename
	log.Printf("[SUCCESS] AttachAvatar: Avatar processed successfully: %s", filename)
	return filename, nil
}
