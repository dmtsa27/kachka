package storage

import (
	"context"
	"time"
)

type Session struct {
	Id            int
	User_id       int64
	ChatID        int64
	MessageID     int
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

func (s *Storage) StartSession(ctx context.Context, userID int64, chatID int64, messageID int) error {
	query := `INSERT INTO sessions (user_id, chat_id, message_id, started_at, session_date)
	VALUES ($1, $2, $3, NOW(), CURRENT_DATE)`
	_, err := s.db.ExecContext(ctx, query, userID, chatID, messageID)
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

func (s *Storage) GetSession(ctx context.Context, userID int64) (*Session, error) {
	var session Session
	query := `SELECT id, user_id, started_at, last_video_at
	          FROM sessions
	          WHERE user_id = $1 AND session_date = CURRENT_DATE`

	err := s.db.QueryRowContext(ctx, query, userID).Scan(
		&session.Id,
		&session.User_id,
		&session.Started_at,
		&session.Last_video_at,
	)
	if err != nil {
		return nil, err
	}

	return &session, nil
}

func (s *Storage) DeleteSessionToday(ctx context.Context, chatID int64, messageID int) error {
	query := `DELETE FROM sessions WHERE chat_id = $1 AND message_id = $2 AND session_date = CURRENT_DATE`
	_, err := s.db.ExecContext(ctx, query, chatID, messageID)
	return err
}
