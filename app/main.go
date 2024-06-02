package main

import (
	"log"
	"net/http"
	"strconv"
	"strings"

	"api-go/helper"
	"api-go/model"

	"api-go/logic"
)

func main() {
	config, err := helper.LoadConfig()
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}

	db, err := helper.SetupDatabase(config)
	if err != nil {
		log.Fatalf("Error setting up database: %v", err)
	}

	modelsToCreate := []interface{}{
		model.Post{},
		model.Tag{},
	}

	for _, model := range modelsToCreate {
		err := helper.CreateTableFromModel(db, model)
		if err != nil {
			log.Fatalf("Error setting up table for model %T: %v", model, err)
		}

		err = helper.UpdateTableFromModel(db, model)
		if err != nil {
			log.Fatalf("Error updating table for model %T: %v", model, err)
		}
	}

	// Array of pairs for join tables
	joinTablePairs := [][]string{
		{"post", "tag"},
	}

	err = helper.CreateJoinTables(db, joinTablePairs)
	if err != nil {
		log.Fatalf("Error creating join tables: %v", err)
	}

	// Define API routes
	http.HandleFunc("/api/posts", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			logic.CreatePost(db)(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	http.HandleFunc("/api/posts/", func(w http.ResponseWriter, r *http.Request) {
		postID, err := getIDFromURL(r)
		if err != nil {
			http.Error(w, "Invalid post ID", http.StatusBadRequest)
			return
		}
		switch r.Method {
		case http.MethodPut:
			logic.UpdatePost(db, postID)(w, r)
		case http.MethodDelete:
			logic.DeletePost(db, postID)(w, r)
		case http.MethodGet:
			logic.GetPostByID(db, postID)(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	http.HandleFunc("/api/tag", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			logic.CreatePost(db)(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	http.HandleFunc("/api/tag/", func(w http.ResponseWriter, r *http.Request) {
		postID, err := getIDFromURL(r)
		if err != nil {
			http.Error(w, "Invalid tag ID", http.StatusBadRequest)
			return
		}
		switch r.Method {
		case http.MethodPut:
			logic.UpdateTag(db, postID)(w, r)
		case http.MethodDelete:
			logic.DeleteTag(db, postID)(w, r)
		case http.MethodGet:
			logic.GetTagByID(db, postID)(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	log.Println("Starting server on :8081")
	if err := http.ListenAndServe(":8081", nil); err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
}

func getIDFromURL(r *http.Request) (int, error) {
	urlParts := strings.Split(r.URL.Path, "/")
	idStr := urlParts[len(urlParts)-1]
	return strconv.Atoi(idStr)
}
