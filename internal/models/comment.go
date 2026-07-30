package models

type Comment struct {
	ID             int    `json:"id"`
	AuthorID       int    `json:"author_id"`
	BoardID        int    `json:"board_id"`
	Content        string `json:"content"`
	AuthorUsername string `json:"author_username"`
}
