package logic

import (
	"database/sql"
	"fmt"
	"net/http"
)

// DeleteTag deletes a post along with its relations in the post_tag table based on the ID
func DeleteTag(db *sql.DB, postID int) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// query to delete post-tag relations from the post_tag table
		deletePostTagQuery := `DELETE FROM post_tag WHERE tag_id = $1`
		_, err := db.Exec(deletePostTagQuery, postID)
		if err != nil {
			http.Error(w, "Failed to delete post-tag relations: "+err.Error(), http.StatusInternalServerError)
			return
		}

		// query to delete post from the post table
		deletePostQuery := `DELETE FROM tag WHERE id = $1`
		_, err = db.Exec(deletePostQuery, postID)
		if err != nil {
			http.Error(w, "Failed to delete post: "+err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "Post with ID %d and its relations deleted successfully", postID)
	}
}
