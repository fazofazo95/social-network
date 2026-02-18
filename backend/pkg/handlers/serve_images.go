package handlers

import (
	"log"
	"net/http"
	"path/filepath"
)

func UploadsFileServer() http.Handler {
	uploadDir := "uploads"

	log.Printf("[INFO] UploadsFileServer: Initializing file server for directory: %s", uploadDir)

	uploadDir = filepath.Clean(uploadDir)
	fs := http.FileServer(http.Dir(uploadDir))

	// strip the URL prefix so requests to /uploads/filename map to uploadDir/filename
	handler := http.StripPrefix("/uploads/", fs)

	log.Println("[SUCCESS] UploadsFileServer: Handler ready at /uploads/")
	return handler
}
