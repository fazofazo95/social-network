package middleware

import (
	"backend/pkg/db/queries"
	"backend/pkg/responses"
	"context"
	"errors"
	"log"
	"net/http"
)

type contextKey string

const userKey contextKey = "userID"

func ContextWithUserID(ctx context.Context, userID int) context.Context {
	return context.WithValue(ctx, userKey, userID)
}

func UserIDFromContext(ctx context.Context) (int, error) {
	userID, ok := ctx.Value(userKey).(int)
	if !ok {
		return 0, errors.New("user ID not found in context")
	}
	return userID, nil
}

func WithAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c, err := r.Cookie("session_id")
		if err != nil {
			responses.SendError(w, http.StatusUnauthorized, "unauthorized")
			return
		}

		userID, err := queries.AuthenticateSession(r.Context(), c.Value)
		if err != nil {
			log.Printf("WithAuth: AuthenticateSession error: %v", err)
			responses.SendError(w, http.StatusUnauthorized, "unauthorized")
			return
		}

		log.Printf("WithAuth: session valid for user %d", userID)
		ctx := ContextWithUserID(r.Context(), userID)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
