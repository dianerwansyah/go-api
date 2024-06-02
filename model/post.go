package model

import "time"

type Post struct {
	ID          int       `json:"id"`
	Title       string    `json:"title"`
	Content     string    `json:"content"`
	Tags        []Tag     `json:"tags" key: many"`
	Status      string    `json:"status"`
	PublishDate time.Time `json:"publish_dte"`
}
