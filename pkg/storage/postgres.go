package storage

import (
	"database/sql"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type Storage struct {
	db *sql.DB
}

func NewPostgresDB(connStr string) (*Storage, error) {
	mydb, err := sql.Open("pgx", connStr)
	if err != nil {
		return nil, err
	}
	if err := mydb.Ping(); err != nil {
		return nil, err
	}
	return &Storage{db: mydb}, nil
}
