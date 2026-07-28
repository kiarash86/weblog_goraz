package models

type Board struct {
	ID        int    `json:"id"`
	AuthorID  int    `json:"author_id"`
	Name      string `json:"name"`
	Content   string `json:"content"`
	ISPrivate bool   `json:"is_private"`
	ImagePath string `json:"img_path"`
}
