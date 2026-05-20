package storage

import (
	"context"
	"database/sql"
	"os"

	"github.com/dmtsa27/kachka.git/pkg/service"
	_ "github.com/jackc/pgx/v5/stdlib"
	"go.uber.org/fx"
)

type Storage struct {
	db *sql.DB
}

var Module = fx.Options(
	fx.Provide(
		func(lc fx.Lifecycle) (*Storage, error) {
			dsn := os.Getenv("DATABASE_URL")
			s, err := NewPostgresDB(context.Background(), dsn)
			if err != nil {
				return nil, err
			}
			lc.Append(fx.Hook{
				OnStop: func(ctx context.Context) error {
					return s.Close()
				},
			})
			return s, nil
		},
		func(s *Storage) service.UserRepository { return s },
		func(s *Storage) service.SessionRepository { return s },
		func(s *Storage) service.WorkoutRepository { return s },
		func(s *Storage) service.ChallengeRepository { return s },
		func(s *Storage) service.BootstrapRepository { return s },
		func(s *Storage) service.ModerationRepository { return s },
		func(s *Storage) service.VoteRepository { return s },
	),
)

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
