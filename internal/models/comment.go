package models

type comment struct {
	ID       int    `json:"id"`
	AuthorID int    `json:"author_id"`
	BoardID  int    `json:"board_id"`
	Content  string `json:"content"`
}
