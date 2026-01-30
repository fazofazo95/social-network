package handlers

import (
	"net/http"
	"path/filepath"
)

func UploadsFileServer() http.Handler {
	uploadDir := "uploads"

	uploadDir = filepath.Clean(uploadDir)
	fs := http.FileServer(http.Dir(uploadDir))
	// strip the URL prefix so requests to /uploads/filename map to uploadDir/filename
	return http.StripPrefix("/uploads/", fs)
}