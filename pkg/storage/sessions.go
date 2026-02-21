package storage

import (
	"context"
	"time"
)

type Session struct {
	Id            int
	User_id       int64
	Started_at    time.Time
	Last_video_at time.Time
}

func (s *Storage) HasTrainedToday(ctx context.Context, userID int64) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1 FROM sessions 
			WHERE user_id = $1 AND CAST(session_date AS DATE) = CURRENT_DATE
		)`

	var exists bool
	err := s.db.QueryRowContext(ctx, query, userID).Scan(&exists)

	if err != nil {
		return false, err
	}
	return exists, nil

}

func (s *Storage) StartSession(ctx context.Context, userID int64) error {
	query := `INSERT INTO sessions (user_id, started_at, session_date)
	VALUES ($1, NOW(), CURRENT_DATE)

	`
	_, err := s.db.ExecContext(ctx, query, userID)

	return err
}

func (s *Storage) AddLatestSession(ctx context.Context, userID int64) error {
	query := `UPDATE sessions
					SET last_video_at = NOW()
					WHERE user_id = $1 AND session_date = CURRENT_DATE

	`
	_, err := s.db.ExecContext(ctx, query, userID)
	return err
}

func (s *Storage) GetLatestSession(ctx context.Context, userID int64) (time.Time, error) {
	var sessionTime time.Time
	query := `SELECT last_video_at
              FROM sessions
              WHERE user_id = $1 AND session_date = CURRENT_DATE`

	err := s.db.QueryRowContext(ctx, query, userID).Scan(&sessionTime)

	if err != nil {
		return time.Time{}, err
	}

	return sessionTime, nil
}
