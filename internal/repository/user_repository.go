package repository

import (
	"context"
	"weblog/internal/models"

	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepository struct {
	db *pgxpool.Pool
}

func NewUserRepository(pp *pgxpool.Pool) *UserRepository {
	return &UserRepository{db: pp}
}

func (ur *UserRepository) DeleteByID(ctx context.Context, id int) error {
	query := `DELETE FROM users WHERE id = $1`
	_, err := ur.db.Exec(ctx, query, id)
	if err != nil {
		return err
	}
	return nil
}

func (ur *UserRepository) Add(ctx context.Context, username string, password string) (*models.User, error) {
	query :=
		`
		INSERT INTO users (username, password)
		VALUES ($1, $2)
		RETURNING id, username, password
		`

	var user models.User
	err := ur.db.QueryRow(ctx, query, username, password).Scan(
		&user.ID, &user.UserName, &user.Password,
	)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (ur *UserRepository) FindByID(ctx context.Context, id int) (*models.User, error) {
	query := `SELECT id, username, password FROM users WHERE id = $1`

	var user models.User
	err := ur.db.QueryRow(ctx, query, id).Scan(
		&user.ID, &user.UserName, &user.Password,
	)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (ur *UserRepository) FindByUserName(ctx context.Context, username string) (*models.User, error) {
	query := `SELECT id, username, password FROM users WHERE username = $1`

	var user models.User
	err := ur.db.QueryRow(ctx, query, username).Scan(
		&user.ID, &user.UserName, &user.Password,
	)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (ur *UserRepository) IsTakenThisUsername(ctx context.Context, username string) (isTaken bool) {
	query := `SELECT EXISTS (SELECT 1  FROM users WHERE username = $1)`

	err := ur.db.QueryRow(ctx, query, username).Scan(&isTaken)
	if err != nil {

		return
	}

	return
}
