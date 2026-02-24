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

func attachImageFromField(r *http.Request, fieldName, logPrefix string) (string, error) {
	log.Printf("[INFO] %s: Processing %s upload", logPrefix, fieldName)

	file, header, err := r.FormFile(fieldName)
	if err != nil {
		if err == http.ErrMissingFile {
			log.Printf("[INFO] %s: No %s file found in request", logPrefix, fieldName)
			return "", nil
		}
		log.Printf("[ERROR] %s: Error reading form file: %v", logPrefix, err)
		return "", errors.New("Failed to read image from form")
	}
	defer file.Close()

	log.Printf("[INFO] %s: Received file %s (Size: %d bytes)", logPrefix, header.Filename, header.Size)

	if header.Size > MaxFileSize {
		log.Printf("[WARN] %s: File size %d exceeds limit", logPrefix, header.Size)
		return "", errors.New("uploaded file too large")
	}

	buf := make([]byte, 512)
	n, _ := file.Read(buf)
	contentType := http.DetectContentType(buf[:n])
	log.Printf("[INFO] %s: Detected Content-Type: %s", logPrefix, contentType)

	allowed := map[string]bool{
		"image/jpeg": true,
		"image/png":  true,
		"image/gif":  true,
	}

	if !allowed[contentType] {
		log.Printf("[WARN] %s: Rejected invalid Content-Type: %s", logPrefix, contentType)
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
	log.Printf("[INFO] %s: Using extension: %s", logPrefix, ext)

	newUUID, _ := uuid.NewV4()
	filename := newUUID.String() + ext

	uploadDir := "uploads"

	log.Printf("[INFO] %s: Saving file as %s in %s", logPrefix, filename, uploadDir)
	if err := SaveFile(file, filename, uploadDir); err != nil {
		log.Printf("[ERROR] %s: SaveFile failed: %v", logPrefix, err)
		return "", errors.New("Failed to save image file")
	}

	filename = "/uploads/" + filename
	log.Printf("[SUCCESS] %s: Image processed successfully: %s", logPrefix, filename)
	return filename, nil
}
