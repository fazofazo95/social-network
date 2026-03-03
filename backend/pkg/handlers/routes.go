package handlers

import (
	"net/http"

	"backend/pkg/middleware"
)

var auth = middleware.WithAuth

func GroupRoutes(mux *http.ServeMux, handler *GroupHandler) {
	handler.RegisterRoutes(mux)
}
