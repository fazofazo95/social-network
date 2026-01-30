package middleware

import (
	"backend/pkg/db/queries"
	database "backend/pkg/db/sqlite"
	"backend/pkg/responses"
	"context"
	"errors"
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

func WithAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		c, err := r.Cookie("session_id")
		if err != nil {
			responses.SendError(w, http.StatusUnauthorized, "unauthorized")
			return
		}

		userID, err := queries.AuthenticateSession(r.Context(), database.DB, c.Value)
		if err != nil {
			responses.SendError(w, http.StatusUnauthorized, "unauthorized")
			return
		}

		ctx := ContextWithUserID(r.Context(), userID)

		next.ServeHTTP(w, r.WithContext(ctx))
	}
}
