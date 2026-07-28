package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"weblog/internal/models"
)

type CommentRepository struct {
	db *pgxpool.Pool
}

func NewCommentRepository(pp *pgxpool.Pool) *CommentRepository {
	return &CommentRepository{db: pp}
}

func (ur *CommentRepository) DeleteByID(ctx context.Context, id int) error {
	query := `DELETE FROM comments WHERE id = $1`
	_, err := ur.db.Exec(ctx, query, id)
	if err != nil {
		return err
	}
	return nil
}

func (ur *CommentRepository) Add(ctx context.Context, authorID int, boardID int, content string) (*models.Comment, error) {
	query := `
		INSERT INTO comments (author_id, board_id, content)
		VALUES ($1, $2, $3)
		RETURNING id, author_id, board_id, content
	`

	var comment models.Comment
	err := ur.db.QueryRow(ctx, query, authorID, boardID, content).Scan(
		&comment.ID, &comment.AuthorID, &comment.BoardID, &comment.Content,
	)
	if err != nil {
		return nil, err
	}
	return &comment, nil
}

func (ur *CommentRepository) FindByID(ctx context.Context, id int) (*models.Comment, error) {
	query := `SELECT id, author_id, board_id, content FROM comments WHERE id = $1`

	var comment models.Comment
	err := ur.db.QueryRow(ctx, query, id).Scan(
		&comment.ID, &comment.AuthorID, &comment.BoardID, &comment.Content,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &comment, nil
}

func (ur *CommentRepository) ListCommentsOfBoard(ctx context.Context, boardID int) ([]*models.Comment, error) {
	query := `SELECT id, author_id, board_id, content FROM comments WHERE board_id = $1`

	rows, err := ur.db.Query(ctx, query, boardID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var comments []*models.Comment
	for rows.Next() {
		var comment models.Comment
		if err := rows.Scan(&comment.ID, &comment.AuthorID, &comment.BoardID, &comment.Content); err != nil {
			return nil, err
		}
		comments = append(comments, &comment)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return comments, nil
}
