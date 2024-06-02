package logic

import (
	"api-go/model"
	"database/sql"
	"encoding/json"
	"net/http"
)

// GetTagByID get tag by its ID
func GetTagByID(db *sql.DB, tagID int) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Query to get tag by ID
		query := `SELECT id, label FROM tag WHERE id = $1`
		row := db.QueryRow(query, tagID)
		var tag model.Tag

		err := row.Scan(&tag.ID, &tag.Label)
		if err != nil {
			http.Error(w, "Failed to get tag: "+err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(tag)
	}
}
