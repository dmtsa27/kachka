package storage

import (
	"context"
	"database/sql"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type Storage struct {
	db *sql.DB
}

func NewPostgresDB(ctx context.Context, connStr string) (*Storage, error) {
	mydb, err := sql.Open("pgx", connStr)
	if err != nil {
		return nil, err
	}
	if err := mydb.Ping(); err != nil {
		return nil, err
	}
	return &Storage{db: mydb}, nil
}

func (s *Storage) Close() error {
	return s.db.Close()
}
