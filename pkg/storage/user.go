package storage

import (
	"context"
	"time"
)

type User struct {
	UserID      int64
	Username    string
	DaysTrained int
	IsActive    bool
	FailedAt    time.Time
}

func (s *Storage) CreateUser(ctx context.Context, user User) error {
	query := `INSERT INTO users (telegram_id, username, is_active)
	VALUES ($1, $2, $3)
	ON CONFLICT (telegram_id) DO NOTHING
	`

	_, err := s.db.ExecContext(ctx, query, user.UserID, user.Username, user.IsActive)

	return err
}

/*func (s *Storage) ReadUser(ctx context.Context, userID int) (*User, error) {
	var myuser User
}*/
