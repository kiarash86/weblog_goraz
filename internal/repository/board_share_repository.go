package repository

import (
	"context"
	"weblog/internal/models"

	"github.com/jackc/pgx/v5/pgxpool"
)

type BoardShareRepository struct {
	db *pgxpool.Pool
}

func NewBoardShareRepository(pp *pgxpool.Pool) *BoardShareRepository {
	return &BoardShareRepository{db: pp}
}

func (ur *BoardShareRepository) Delete(ctx context.Context, userID int ,  boardID int) error {
	query := `DELETE FROM board_shares WHERE user_id = $1 AND  board_id = $2`
	_, err := ur.db.Exec(ctx, query, userID , boardID)
	if err != nil {
		return err
	}
	return nil
}

func (ur *BoardShareRepository) Add( ctx context.Context, userID int ,  boardID int) (*models.Board_share, error) {
	query :=
		`
		INSERT INTO board_shares (user_id, board_id)
		VALUES ($1, $2)
		ON CONFLICT (user_id, board_id) DO NOTHING
		RETURNING user_id, board_id
		`

	var boardShare models.Board_share
	err := ur.db.QueryRow(ctx, query,userID, boardID).Scan(
		&boardShare.UserID, &boardShare.BoardID,
	)
	if err != nil {
		return nil, err
	}
	return &boardShare, nil
}


func (ur *BoardShareRepository) HasAccess(ctx context.Context, userID int ,  boardID int) (hasAccess bool ,err error) {
	query := `SELECT EXISTS(SELECT 1 FROM board_shares WHERE user_id = $1 AND  board_id = $2)`
	 err = ur.db.QueryRow(ctx, query, userID , boardID).Scan(&hasAccess)
	if err != nil {
		return
	}
	return
}