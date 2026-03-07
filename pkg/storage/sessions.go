package storage

import (
	"context"
	"time"
)

type Session struct {
	ID          int
	UserID      int64
	ChatID      int64
	MessageID   int
	StartedAt   time.Time
	LastVideoAt time.Time
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
	query := `INSERT INTO sessions (user_id, chat_id, message_id, started_at, last_video_at, session_date)
	VALUES ($1, $2, $3, NOW(), NOW(), CURRENT_DATE)`
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
		&session.ID,
		&session.UserID,
		&session.StartedAt,
		&session.LastVideoAt,
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

func (s *Storage) DeleteSessionTodayAffected(ctx context.Context, chatID int64, messageID int) (bool, error) {
	query := `DELETE FROM sessions WHERE chat_id = $1 AND message_id = $2 AND session_date = CURRENT_DATE`
	result, err := s.db.ExecContext(ctx, query, chatID, messageID)
	if err != nil {
		return false, err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return false, err
	}

	return rows > 0, nil
}
