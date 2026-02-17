package middleware

import "net/http"

// Middleware type για ευκολία
type Middleware func(http.Handler) http.Handler

// Chain δέχεται έναν handler και μια λίστα από middlewares
// και τα εφαρμόζει με τη σειρά που δίνονται.
func Chain(h http.HandlerFunc, middlewares ...Middleware) http.Handler {
	var current http.Handler = h
	// Τα εφαρμόζουμε από το τέλος προς την αρχή για να διατηρηθεί η σειρά εκτέλεσης
	for i := len(middlewares) - 1; i >= 0; i-- {
		current = middlewares[i](current)
	}
	return current
}
