package middleware

import (
	"database/sql"
	"fmt"
	"net/http"
	"strconv"
)

func OwnershipMiddleware(db *sql.DB, tableName string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			idParam := r.PathValue("id")
			resourceID, err := strconv.Atoi(idParam)
			if err != nil {
				http.Error(w, "Invalid ID parameter", http.StatusBadRequest)
				return
			}

			userID, ok := r.Context().Value("userID").(int)
			if !ok {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			var ownerID int
			query := fmt.Sprintf("SELECT user_id FROM %s WHERE id = ?", tableName)
			err = db.QueryRowContext(r.Context(), query, resourceID).Scan(&ownerID)

			if err != nil {
				if err == sql.ErrNoRows {
					http.Error(w, "Resource not found", http.StatusNotFound)
				} else {
					http.Error(w, "Database error", http.StatusInternalServerError)
				}
				return
			}

			if ownerID != userID {
				http.Error(w, "Forbidden: You are not the owner of this "+tableName, http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
