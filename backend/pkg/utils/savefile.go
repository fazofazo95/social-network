package utils

import (
	"io"
	"log"
	"mime/multipart"
	"os"
	"path/filepath"
)

func SaveFile(src multipart.File, filename string, uploadDir string) error {
	if err := os.MkdirAll(uploadDir, 0o755); err != nil {
		log.Printf("[ERROR] SaveFile: Failed to create directory %s: %v", uploadDir, err)
		return err
	}

	safeName := filepath.Base(filename)
	dstPath := filepath.Join(uploadDir, safeName)

	dst, err := os.Create(dstPath)
	if err != nil {
		log.Printf("[ERROR] SaveFile: Failed to create destination file %s: %v", dstPath, err)
		return err
	}
	defer dst.Close()

	if _, err := io.Copy(dst, src); err != nil {
		log.Printf("[ERROR] SaveFile: Failed to copy file content to %s: %v", dstPath, err)
		return err
	}

	return nil
}
