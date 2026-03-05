package handlers

import (
	"net/http"
	"path/filepath"
)

func UploadsFileServer() http.Handler {
	uploadDir := filepath.Clean("uploads")
	fs := http.FileServer(http.Dir(uploadDir))
	return http.StripPrefix("/uploads/", fs)
}
