package logic

import (
	"api-go/model"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
)

// UpdatePost to update data post
func UpdatePost(db *sql.DB, postID int) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var updatedPost model.Post
		err := json.NewDecoder(r.Body).Decode(&updatedPost)
		if err != nil {
			http.Error(w, "Invalid request payload", http.StatusBadRequest)
			return
		}

		// get all tag for table
		tagsMap, err := getAllTagsMap(db)
		if err != nil {
			http.Error(w, "Failed to retrieve tags: "+err.Error(), http.StatusInternalServerError)
			return
		}

		// Cek tag in post
		for _, tag := range updatedPost.Tags {
			if _, exists := tagsMap[tag.Label]; !exists {
				http.Error(w, "Tag '"+tag.Label+"' does not exist", http.StatusBadRequest)
				return
			}
		}

		err = UpdateQueryPost(db, updatedPost, postID)
		if err != nil {
			http.Error(w, "Failed to UpdateQueryPost: "+err.Error(), http.StatusInternalServerError)
		}

		// Cek if any change to tag
		if len(updatedPost.Tags) > 0 {
			// Update post_tag
			err = UpdatePostTags(db, postID, updatedPost.Tags, tagsMap)
			if err != nil {
				http.Error(w, "Failed to update post tags: "+err.Error(), http.StatusInternalServerError)
				return
			}
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(updatedPost)
	}
}

func UpdateQueryPost(db *sql.DB, updatedPost model.Post, postID int) error {
	args := []interface{}{}

	if updatedPost.Title != "" {
		args = append(args, updatedPost.Title)
	}

	if updatedPost.Content != "" {
		args = append(args, updatedPost.Content)
	}

	if len(args) == 0 {
		return fmt.Errorf("no valid fields to update")
	}

	args = append(args, postID)
	updateQuery := fmt.Sprintf("UPDATE post SET title = $1, content = $2 WHERE id = $%d", len(args))

	_, err := db.Exec(updateQuery, args...)
	if err != nil {
		return fmt.Errorf("failed to update post: %w", err)
	}

	return nil
}

// getAllTagsMap get all tags from table and return map
func getAllTagsMap(db *sql.DB) (map[string]int, error) {
	rows, err := db.Query("SELECT id, label FROM tag")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tagsMap := make(map[string]int)
	for rows.Next() {
		var id int
		var label string
		if err := rows.Scan(&id, &label); err != nil {
			return nil, err
		}
		tagsMap[label] = id
	}
	return tagsMap, nil
}

// UpdatePostTags updates the tags with a post
func UpdatePostTags(db *sql.DB, postID int, tags []model.Tag, tagsMap map[string]int) error {
	// Delete existing tags for the post
	_, err := db.Exec("DELETE FROM post_tag WHERE post_id = $1", postID)
	if err != nil {
		return err
	}

	// Insert new tags for the post
	for _, tag := range tags {
		tagID, exists := tagsMap[tag.Label]
		if !exists {
			return fmt.Errorf("Tag '%s' does not exist", tag.Label)
		}
		_, err = db.Exec("INSERT INTO post_tag (post_id, tag_id) VALUES ($1, $2)", postID, tagID)
		if err != nil {
			return err
		}
	}

	return nil
}
