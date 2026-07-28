package models

type Board struct {
	ID        int    `json:"id"`
	AuthorID  int    `json:"author_id"`
	Title      string `json:"title"`
	Content   string `json:"content"`
	ISPrivate bool   `json:"is_private"`
	ImagePath string `json:"img_path"`
}
