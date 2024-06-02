package logic

import (
	"api-go/model"
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/lib/pq"
)

// CreatePost to Insert table post and post_tag
func CreatePost(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var post model.Post
		err := json.NewDecoder(r.Body).Decode(&post)
		if err != nil {
			http.Error(w, "Invalid request payload", http.StatusBadRequest)
			return
		}

		postID, err := InsertPost(db, post)
		if err != nil {
			http.Error(w, "Failed to create post: "+err.Error(), http.StatusInternalServerError)
			return
		}

		if postID > 0 {
			err = InsertPostTag(db, postID, post.Tags)
			if err != nil {
				http.Error(w, "Failed to insert post-tag relationships: "+err.Error(), http.StatusInternalServerError)
				return
			}

			response := map[string]interface{}{
				"id": postID,
			}

			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(response)
		}
	}
}

// Insert tabel post
func InsertPost(db *sql.DB, post model.Post) (int, error) {
	// cek all label tag
	labels := make([]string, len(post.Tags))
	for i, tag := range post.Tags {
		labels[i] = tag.Label
	}

	// take id from all tag in database
	idMap := make(map[string]int)
	rows, err := db.Query("SELECT id, label FROM tag WHERE label = ANY($1)", pq.Array(labels))
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	for rows.Next() {
		var id int
		var label string
		if err := rows.Scan(&id, &label); err != nil {
			return 0, err
		}
		idMap[label] = id
	}
	if err := rows.Err(); err != nil {
		return 0, err
	}

	if post.Status == "" {
		post.Status = "Draft"
	}

	// query Insert post after get all id
	postQuery := `
		INSERT INTO post (title, content, status, publishdate) 
		VALUES ($1, $2, $3, $4) 
		RETURNING id
	`
	row := db.QueryRow(postQuery, post.Title, post.Content, post.Status, post.PublishDate)
	var postID int
	err = row.Scan(&postID)
	if err != nil {
		return 0, err
	}

	return postID, nil
}

// insert post_tag
func InsertPostTag(db *sql.DB, postID int, tags []model.Tag) error {
	for _, tag := range tags {
		_, err := db.Exec("INSERT INTO post_tag (post_id, tag_id) VALUES ($1, (SELECT id FROM tag WHERE label = $2))", postID, tag.Label)
		if err != nil {
			return err
		}
	}
	return nil
}
