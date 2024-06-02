package logic

import (
	"api-go/model"
	"database/sql"
	"encoding/json"
	"net/http"
)

// CreateTag to Insert table tag and post_tag
func CreateTag(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var tag model.Tag
		err := json.NewDecoder(r.Body).Decode(&tag)
		if err != nil {
			http.Error(w, "Invalid request payload", http.StatusBadRequest)
			return
		}

		tagID, err := InsertTag(db, tag)
		if err != nil {
			http.Error(w, "Failed to create tag: "+err.Error(), http.StatusInternalServerError)
			return
		}

		if tagID > 0 {
			response := map[string]interface{}{
				"id": tagID,
			}

			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(response)
		} else {
			response := map[string]interface{}{
				"message": "Cannot create post because data already exists",
			}
			w.WriteHeader(http.StatusConflict)
			json.NewEncoder(w).Encode(response)
		}
	}
}

// Insert tabel tag
func InsertTag(db *sql.DB, tag model.Tag) (int, error) {
	var existingID int
	// take id from all tag in database
	err := db.QueryRow("SELECT id FROM tag WHERE label = $1", tag.Label).Scan(&existingID)
	if err != nil {
		return 0, err
	}
	if err == nil {
		return 0, nil
	}

	var tagID int
	err = db.QueryRow("INSERT INTO tag (label) VALUES ($1) RETURNING id", tag.Label).Scan(&tagID)
	if err != nil {
		return 0, err
	}

	return tagID, nil
}
