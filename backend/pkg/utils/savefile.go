package utils

import (
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
)

func SaveFile(src multipart.File, filename string, uploadDir string) error {

	if err := os.MkdirAll(uploadDir, 0o755); err != nil {
		return err
	}

	safeName := filepath.Base(filename)
	dstPath := filepath.Join(uploadDir, safeName)

	dst, err := os.Create(dstPath)
	if err != nil {
		return err
	}
	defer dst.Close()

	if _, err := io.Copy(dst, src); err != nil {
		return err
	}

	return nil
}