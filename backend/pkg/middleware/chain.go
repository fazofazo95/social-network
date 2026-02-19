package middleware

import "net/http"

type Middleware func(http.Handler) http.Handler

func Chain(h http.HandlerFunc, middlewares ...Middleware) http.Handler {
	var current http.Handler = h
	for i := len(middlewares) - 1; i >= 0; i-- {
		current = middlewares[i](current)
	}
	return current
}
