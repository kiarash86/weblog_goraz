package db

import (
	"context"
	"log"
	"weblog/internal/config"

	"github.com/jackc/pgx/v5/pgxpool"
)

func Connect(config *config.Config) *pgxpool.Pool {

	pool, err := pgxpool.New(context.Background(), config.DatabaseURL)
	if err != nil {
		log.Fatal("ok no database")

	}

	err = pool.Ping(context.Background())
	if err != nil {
		log.Fatal("ok no ping")
	}

	log.Println("ok connected succesfully")
	return pool
}
