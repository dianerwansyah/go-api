package logic

import (
	"api-go/model"
	"database/sql"
	"encoding/json"
	"net/http"
)

// UpdateTag to update data tag
func UpdateTag(db *sql.DB, tagID int) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var updatedtag model.Tag
		err := json.NewDecoder(r.Body).Decode(&updatedtag)
		if err != nil {
			http.Error(w, "Invalid request payload", http.StatusBadRequest)
			return
		}

		updateQuery := "UPDATE tag SET label = $1 WHERE id = $2"
		_, err = db.Exec(updateQuery, updatedtag.Label, tagID)
		if err != nil {
			http.Error(w, "Failed to update tag: "+err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(updatedtag)
	}
}
