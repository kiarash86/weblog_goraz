package repository

import (
	"context"
	"weblog/internal/models"

	"github.com/jackc/pgx/v5/pgxpool"
)

type BoardRepository struct {
	db *pgxpool.Pool
}

func NewBoardRepository(pp *pgxpool.Pool) *BoardRepository {
	return &BoardRepository{db: pp}
}

func (ur *BoardRepository) DeleteByID(ctx context.Context, id int, authorID int) error {
	query := `DELETE FROM boards WHERE id = $1 AND author_id = $2`
	_, err := ur.db.Exec(ctx, query, id, authorID)
	if err != nil {
		return err
	}
	return nil
}

func (ur *BoardRepository) Add(ctx context.Context, AuthorID int, title string, content string, isPrivate bool, imgPath string) (*models.Board, error) {
	query :=
		`
		INSERT INTO boards (author_id, title , content , is_private , img_path)
		VALUES ($1, $2 , $3 ,$4 , $5)
		RETURNING id, author_id, title , content , is_private , img_path
		`

	var board models.Board
	err := ur.db.QueryRow(ctx, query, AuthorID, title, content, isPrivate, imgPath).Scan(
		&board.ID, &board.AuthorID, &board.Title, &board.Content, &board.ISPrivate, &board.ImagePath,
	)
	if err != nil {
		return nil, err
	}
	return &board, nil
}

func (ur *BoardRepository) FindByID(ctx context.Context, id int) (*models.Board, error) {
	query := `SELECT id, author_id, title , content , is_private , img_path FROM boards WHERE id = $1`

	var board models.Board
	err := ur.db.QueryRow(ctx, query, id).Scan(
		&board.ID, &board.AuthorID, &board.Title, &board.Content, &board.ISPrivate, &board.ImagePath,
	)
	if err != nil {
		return nil, err
	}
	return &board, nil
}

func (ur *BoardRepository) IsOwner(ctx context.Context, boardID int, userID int) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM boards WHERE id = $1 AND author_id = $2)`
	var isOwner bool
	err := ur.db.QueryRow(ctx, query, boardID, userID).Scan(&isOwner)

	if err != nil {
		return false, err
	}
	return isOwner, nil
}

func (ur *BoardRepository) ListFeed(ctx context.Context, userID int, page int, search string) (boards []*models.Board, err error) {
	limit := 10
	offset := (page - 1) * limit

	query := `
	SELECT id, author_id, title, content, is_private, img_path
	FROM boards
	WHERE (is_private = false
	   OR author_id = $1
	   OR id IN (SELECT board_id FROM board_shares WHERE user_id = $1))
	  AND title ILIKE $2
	ORDER BY id DESC
	LIMIT $3 OFFSET $4
`
	rows, err := ur.db.Query(ctx, query, userID, search, limit, offset)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var board models.Board
		err = rows.Scan(&board.ID, &board.AuthorID, &board.Title, &board.Content, &board.ISPrivate, &board.ImagePath)
		if err != nil {
			return nil, err
		}
		boards = append(boards, &board)

	}
	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return
}
