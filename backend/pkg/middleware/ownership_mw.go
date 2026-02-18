package middleware

import (
	"backend/pkg/responses"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"strconv"
)

func OwnershipMiddleware(db *sql.DB, tableName string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			log.Printf("[INFO] OwnershipMiddleware: Checking ownership for table: %s", tableName)

			idParam := r.PathValue("id")
			resourceID, err := strconv.Atoi(idParam)
			if err != nil {
				log.Printf("[ERROR] OwnershipMiddleware: Invalid ID parameter '%s': %v", idParam, err)
				responses.SendError(w, http.StatusBadRequest, "Invalid ID parameter")
				return
			}

			userID, ok := r.Context().Value("userID").(int)
			if !ok {
				log.Println("[WARN] OwnershipMiddleware: UserID missing from context")
				responses.SendError(w, http.StatusUnauthorized, "Unauthorized")
				return
			}

			var ownerID int
			query := fmt.Sprintf("SELECT user_id FROM %s WHERE id = ?", tableName)
			err = db.QueryRowContext(r.Context(), query, resourceID).Scan(&ownerID)

			if err != nil {
				if err == sql.ErrNoRows {
					log.Printf("[WARN] OwnershipMiddleware: Resource %d not found in %s", resourceID, tableName)
					responses.SendError(w, http.StatusNotFound, "Resource not found")
				} else {
					log.Printf("[ERROR] OwnershipMiddleware: Database error for %s ID %d: %v", tableName, resourceID, err)
					responses.SendError(w, http.StatusInternalServerError, "Database error")
				}
				return
			}

			if ownerID != userID {
				log.Printf("[WARN] OwnershipMiddleware: Forbidden - User %d does not own %s %d (Owner: %d)", userID, tableName, resourceID, ownerID)
				responses.SendError(w, http.StatusForbidden, "Forbidden: You are not the owner of this "+tableName)
				return
			}

			log.Printf("[SUCCESS] OwnershipMiddleware: Ownership verified for User %d on %s %d", userID, tableName, resourceID)
			next.ServeHTTP(w, r)
		})
	}
}
