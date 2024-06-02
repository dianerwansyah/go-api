package logic

import (
	"api-go/model"
	"database/sql"
	"encoding/json"
	"net/http"
	"time"
)

// GetPostByID get post by its ID
func GetPostByID(db *sql.DB, postID int) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Query to post by ID
		query := `
			SELECT id, title, content, status, publishdate
			FROM post
			WHERE id = $1
		`
		row := db.QueryRow(query, postID)
		var post model.Post

		var publishDate sql.NullTime

		values := []interface{}{
			&post.ID, &post.Title, &post.Content, &post.Status, &publishDate,
		}

		err := row.Scan(values...)
		if err != nil {
			if err == sql.ErrNoRows {
				http.Error(w, "Post not found", http.StatusNotFound)
			} else {
				http.Error(w, "Failed to get post: "+err.Error(), http.StatusInternalServerError)
			}
			return
		}

		// Check if publishDate is valid
		if publishDate.Valid {
			post.PublishDate = publishDate.Time
		}

		// Query to get tag related post
		tagsQuery := `
			SELECT tag.id, tag.label
			FROM tag
			INNER JOIN post_tag ON tag.id = post_tag.tag_id
			WHERE post_tag.post_id = $1
		`
		rows, err := db.Query(tagsQuery, postID)
		if err != nil {
			http.Error(w, "Failed to get tags: "+err.Error(), http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		for rows.Next() {
			var label string
			var id int
			if err := rows.Scan(&id, &label); err != nil {
				http.Error(w, "Failed to scan tags: "+err.Error(), http.StatusInternalServerError)
				return
			}
			post.Tags = append(post.Tags, model.Tag{ID: id, Label: label})
		}

		if err := rows.Err(); err != nil {
			http.Error(w, "Error processing tags: "+err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(post)
	}
}

// GetAllPosts all data posts
func GetAllPosts(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Query for all posts
		query := `
			SELECT p.id, p.title, p.content, p.status, p.publishdate, t.label
			FROM post p
			INNER JOIN post_tag pt ON p.id = pt.post_id
			INNER JOIN tag t ON pt.tag_id = t.id
		`
		rows, err := db.Query(query)
		if err != nil {
			http.Error(w, "Failed to get posts: "+err.Error(), http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var postsMap = make(map[int]model.Post)

		for rows.Next() {
			var postID int
			var title, content, status string
			var publishDate sql.NullTime
			var tagLabel string

			err := rows.Scan(&postID, &title, &content, &status, &publishDate, &tagLabel)
			if err != nil {
				http.Error(w, "Failed to scan row: "+err.Error(), http.StatusInternalServerError)
				return
			}

			// Check if publishDate is valid
			var publishTime time.Time
			if publishDate.Valid {
				publishTime = publishDate.Time
			}

			// Check if post already exists in map, if not create new post
			post, ok := postsMap[postID]
			if !ok {
				post = model.Post{
					ID:          postID,
					Title:       title,
					Content:     content,
					Status:      status,
					PublishDate: publishTime,
					Tags:        []model.Tag{},
				}
			}

			post.Tags = append(post.Tags, model.Tag{Label: tagLabel})
			postsMap[postID] = post
		}

		// Convert map to slice of posts
		var posts []model.Post
		for _, post := range postsMap {
			posts = append(posts, post)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(posts)
	}
}
