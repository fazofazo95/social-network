package middleware

import (
	"database/sql"
	"fmt"
	"net/http"
	"strconv"
)

// OwnershipMiddleware ελέγχει αν ο συνδεδεμένος χρήστης είναι ο ιδιοκτήτης του πόρου (post/comment)
func OwnershipMiddleware(db *sql.DB, tableName string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// 1. Παίρνουμε το ID από το URL (π.χ. /api/posts/{id})
			idParam := r.PathValue("id")
			resourceID, err := strconv.Atoi(idParam)
			if err != nil {
				http.Error(w, "Invalid ID parameter", http.StatusBadRequest)
				return
			}

			// 2. Παίρνουμε το UserID από το Context (που έβαλε το WithAuth middleware)
			userID, ok := r.Context().Value("userID").(int)
			if !ok {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			// 3. Ελέγχουμε στη βάση αν το resource ανήκει στον χρήστη
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

			// 4. Σύγκριση IDs
			if ownerID != userID {
				http.Error(w, "Forbidden: You are not the owner of this "+tableName, http.StatusForbidden)
				return
			}

			// Αν όλα είναι OK, προχωράμε στον επόμενο handler
			next.ServeHTTP(w, r)
		})
	}
}
