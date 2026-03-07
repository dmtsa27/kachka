package storage

import (
	"context"
	"database/sql"
)

type User struct {
	TelegramID  int64
	Username    string
	DaysTrained int
	IsActive    bool
	FailedAt    sql.NullTime
}

func (s *Storage) CreateUser(ctx context.Context, user User) error {
	query := `INSERT INTO users (telegram_id, username, is_active)
	VALUES ($1, $2, $3)
	ON CONFLICT (telegram_id) DO UPDATE
	SET username = EXCLUDED.username,
		is_active = true,
		failed_at = NULL
	`

	_, err := s.db.ExecContext(ctx, query, user.TelegramID, user.Username, user.IsActive)

	return err
}

func (s *Storage) ReadUser(ctx context.Context, userID int64) (*User, error) {
	var myuser User
	query := `SELECT telegram_id, username, days_trained, is_active, failed_at FROM users WHERE telegram_id = $1`

	err := s.db.QueryRowContext(ctx, query, userID).Scan(&myuser.TelegramID, &myuser.Username, &myuser.DaysTrained, &myuser.IsActive, &myuser.FailedAt)
	if err != nil {
		return nil, err
	}

	return &myuser, nil
}

func (s *Storage) DeactivateUser(ctx context.Context, userID int64) error {
	query := `UPDATE users SET is_active = false, failed_at = NOW() WHERE telegram_id = $1`
	_, err := s.db.ExecContext(ctx, query, userID)
	return err
}

// BatchDeactivateUsers деактивує переданих юзерів атомарно в одній транзакції.
func (s *Storage) BatchDeactivateUsers(ctx context.Context, userIDs []int64) error {
	if len(userIDs) == 0 {
		return nil
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for _, id := range userIDs {
		if _, err := tx.ExecContext(ctx,
			`UPDATE users SET is_active = false, failed_at = NOW() WHERE telegram_id = $1`, id,
		); err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (s *Storage) GetAllActiveUsers(ctx context.Context) ([]User, error) {
	query := `SELECT telegram_id, username, days_trained, is_active, failed_at FROM users WHERE is_active = true`

	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var u User
		if err := rows.Scan(&u.TelegramID, &u.Username, &u.DaysTrained, &u.IsActive, &u.FailedAt); err != nil {
			return nil, err
		}
		users = append(users, u)
	}

	return users, rows.Err()
}
