package model

type Tag struct {
	ID    int    `json:"id"`
	Label string `json:"label" key:"uniq"`
}
